package ddgchat

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"
)

func generateUUID() string {
	return uuid.New().String()
}

// Get VQD token from DuckDuckGo API for authentication
func updateVQDToken(userAgent string, config *Config) (string, error) {
	logger.Debug("updating VQD token")
	client := createProxyClient()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(config.DDGChatAPIURL + "/country.json")
	req.Header.Set("User-Agent", userAgent)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := client.Do(req, resp); err != nil {
		logger.Error("failed to get country.json", zap.Error(err))
		return "", err
	}

	req.Reset()
	req.SetRequestURI(config.DDGChatAPIURL + "/duckchat/v1/status")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("x-vqd-accept", "1")

	if err := client.Do(req, resp); err != nil {
		logger.Error("failed to get duckchat/v1/status", zap.Error(err))
		return "", err
	}

	vqdToken := string(resp.Header.Peek("x-vqd-4"))
	if vqdToken == "" {
		logger.Error("failed to get VQD token")
		return "", fmt.Errorf("failed to get VQD token")
	}

	logger.Debug("got VQD token", zap.String("vqd_token", vqdToken))

	return vqdToken, nil
}

// Main chat function to interact with DuckDuckGo API
func chatWithDuckDuckGo(query string, model string, history []ChatMessage, channel chan string, config *Config) error {
	logger.Debug("chat with duckduckgo", zap.String("query", query), zap.String("model", model))
	originalModel := config.ModelMapping[model]
	if originalModel == "" {
		originalModel = model
	}

	var userAgent string
	if config.UserAgent == "" {
		userAgent = getRandomUserAgent()
	} else {
		userAgent = config.UserAgent
	}
	vqdToken, err := updateVQDToken(userAgent, config)
	if err != nil {
		return err
	}

	// Process system message and user messages
	var systemMsg *ChatMessage
	var userMsgs []map[string]string

	for _, msg := range history {
		if msg.Role == "system" {
			systemMsg = &msg
		} else if msg.Role == "user" {
			userMsgs = append(userMsgs, map[string]string{
				"role":    msg.Role,
				"content": msg.Content,
			})
		}
	}

	// Combine system message with first user message if exists
	if systemMsg != nil {
		for i := range userMsgs {
			userMsgs[i]["content"] = fmt.Sprintf("%s\n\n%s", systemMsg.Content, userMsgs[i]["content"])
		}
	}

	payload := map[string]interface{}{
		"messages": userMsgs,
		"model":    originalModel,
	}

	logger.Debug("payload", zap.Any("payload", payload))

	client := createProxyClient()
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(config.DDGChatAPIURL + "/duckchat/v1/chat")
	req.Header.SetMethod("POST")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("x-vqd-4", vqdToken)

	jsonPayload, _ := json.Marshal(payload)
	req.SetBody(jsonPayload)

	return streamDuckDuckGoResponse(client, req, channel, config)
}

// Handle streaming response from DuckDuckGo API with retry mechanism
func streamDuckDuckGoResponse(client *fasthttp.Client, req *fasthttp.Request, channel chan string, config *Config) error {
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetConnectionClose()

	if err := client.Do(req, resp); err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != fasthttp.StatusOK {
		return fmt.Errorf("unexpected status code: %d", statusCode)
	}

	body := resp.Body()
	reader := bufio.NewReader(bytes.NewReader(body))

	maxRetries := 5
	retryCount := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading response: %v", err)
		}

		// Process Server-Sent Events
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			data = strings.TrimSpace(data)

			if data == "[DONE]" {
				break
			}

			var jsonResponse struct {
				Message string `json:"message"`
			}

			if err := json.Unmarshal([]byte(data), &jsonResponse); err != nil {
				logger.Error("Error parsing JSON", zap.Error(err))
				continue
			}

			channel <- jsonResponse.Message
		}

		// Handle rate limiting with retries
		if statusCode == 429 && retryCount < maxRetries {
			retryCount++
			logger.Warn("Rate limit exceeded, retry attempt %d of %d", zap.Int("retryCount", retryCount), zap.Int("maxRetries", maxRetries))

			var userAgent string
			if config.UserAgent == "" {
				userAgent = getRandomUserAgent()
			} else {
				userAgent = config.UserAgent
			}
			vqdToken, err := updateVQDToken(userAgent, config)
			if err != nil {
				logger.Error("Failed to update VQD token: %v", zap.Error(err))
				continue
			}

			req.Header.Set("User-Agent", userAgent)
			req.Header.Set("x-vqd-4", vqdToken)

			if err := client.Do(req, resp); err != nil {
				logger.Error("Retry request failed: %v", zap.Error(err))
				continue
			}

			body = resp.Body()
			reader = bufio.NewReader(bytes.NewReader(body))
			statusCode = resp.StatusCode()
			continue
		}
	}

	return nil
}

// Create HTTP client with proxy support if configured
func createProxyClient() *fasthttp.Client {
	proxyURL := os.Getenv("https_proxy")
	if proxyURL == "" {
		proxyURL = os.Getenv("HTTPS_PROXY")
	}
	if proxyURL == "" {
		return &fasthttp.Client{}
	}

	parsedProxyURL, err := url.Parse(proxyURL)
	if err != nil {
		log.Printf("Error parsing proxy URL: %v", err)
		return &fasthttp.Client{}
	}
	logger.Debug("parsed proxy URL", zap.String("proxy_url", parsedProxyURL.String()))

	return &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			logger.Debug("dialing to proxy", zap.String("addr", addr))
			proxyConn, err := fasthttp.Dial(parsedProxyURL.Host)
			if err != nil {
				logger.Error("error connecting to proxy", zap.Error(err))
				return nil, fmt.Errorf("error connecting to proxy: %w", err)
			}

			_, err = fmt.Fprintf(proxyConn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", addr, addr)
			if err != nil {
				proxyConn.Close()
				logger.Error("error sending CONNECT", zap.Error(err))
				return nil, fmt.Errorf("error sending CONNECT: %w", err)
			}

			res := make([]byte, 1024)
			n, err := proxyConn.Read(res)
			if err != nil {
				proxyConn.Close()
				logger.Error("error reading CONNECT response", zap.Error(err))
				return nil, fmt.Errorf("error reading CONNECT response: %w", err)
			}

			if !bytes.Contains(res[:n], []byte("200 Connection established")) {
				proxyConn.Close()
				logger.Error("proxy connection failed", zap.String("response", string(res[:n])))
				return nil, fmt.Errorf("proxy connection failed: %s", res[:n])
			}

			return proxyConn, nil
		},
	}
}

func getRandomUserAgent() string {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edge/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"}
	return userAgents[rand.Intn(len(userAgents))]
}
