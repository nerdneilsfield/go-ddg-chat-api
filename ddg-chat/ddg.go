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

func chatWithDuckDuckGo(query string, model string, history []ChatMessage, channel chan string, config *Config) error {
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

	// 处理系统消息
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

	if systemMsg != nil && len(userMsgs) > 0 {
		userMsgs[0]["content"] = fmt.Sprintf("%s\n\n%s", systemMsg.Content, userMsgs[0]["content"])
	}

	payload := map[string]interface{}{
		"messages": userMsgs,
		"model":    originalModel,
	}

	// 发送请求到DuckDuckGo
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

func streamDuckDuckGoResponse(client *fasthttp.Client, req *fasthttp.Request, channel chan string, config *Config) error {
	// 创建响应对象
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// 设置请求为流式
	req.SetConnectionClose()

	// 发送请求并获取响应
	if err := client.Do(req, resp); err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	// 检查响应状态码
	statusCode := resp.StatusCode()
	if statusCode != fasthttp.StatusOK {
		return fmt.Errorf("unexpected status code: %d", statusCode)
	}

	// 读取响应体
	body := resp.Body()
	reader := bufio.NewReader(bytes.NewReader(body))

	maxRetries := 5
	retryCount := 0

	for {
		// 读取每一行
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading response: %v", err)
		}

		// 处理 Server-Sent Events 格式
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			data = strings.TrimSpace(data)

			// 检查是否是结束标记
			if data == "[DONE]" {
				break
			}

			// 解析JSON响应
			var jsonResponse struct {
				Message string `json:"message"`
			}

			if err := json.Unmarshal([]byte(data), &jsonResponse); err != nil {
				logger.Error("Error parsing JSON", zap.Error(err))
				continue
			}

			// 发送消息到channel
			channel <- jsonResponse.Message
		}

		// 处理响应状态码429 (Rate Limit)
		if statusCode == 429 && retryCount < maxRetries {
			retryCount++
			logger.Warn("Rate limit exceeded, retry attempt %d of %d", zap.Int("retryCount", retryCount), zap.Int("maxRetries", maxRetries))

			// 获取新的User-Agent和VQD token
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

			// 更新请求头
			req.Header.Set("User-Agent", userAgent)
			req.Header.Set("x-vqd-4", vqdToken)

			// 重新发送请求
			if err := client.Do(req, resp); err != nil {
				logger.Error("Retry request failed: %v", zap.Error(err))
				continue
			}

			// 重置reader
			body = resp.Body()
			reader = bufio.NewReader(bytes.NewReader(body))
			statusCode = resp.StatusCode()
			continue
		}
	}

	return nil
}

func createProxyClient() *fasthttp.Client {
	proxyURL := os.Getenv("https_proxy")
	if proxyURL == "" {
		proxyURL = os.Getenv("HTTPS_PROXY")
	}
	if proxyURL == "" {
		return &fasthttp.Client{}
	}

	// 解析代理URL
	parsedProxyURL, err := url.Parse(proxyURL)
	if err != nil {
		log.Printf("Error parsing proxy URL: %v", err)
		return &fasthttp.Client{}
	}
	logger.Debug("parsed proxy URL", zap.String("proxy_url", parsedProxyURL.String()))

	return &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			logger.Debug("dialing to proxy", zap.String("addr", addr))
			// 连接到代理服务器
			proxyConn, err := fasthttp.Dial(parsedProxyURL.Host)
			if err != nil {
				logger.Error("error connecting to proxy", zap.Error(err))
				return nil, fmt.Errorf("error connecting to proxy: %w", err)
			}

			// 发送 CONNECT 请求
			_, err = fmt.Fprintf(proxyConn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", addr, addr)
			if err != nil {
				proxyConn.Close()
				logger.Error("error sending CONNECT", zap.Error(err))
				return nil, fmt.Errorf("error sending CONNECT: %w", err)
			}

			// 读取响应
			res := make([]byte, 1024)
			n, err := proxyConn.Read(res)
			if err != nil {
				proxyConn.Close()
				logger.Error("error reading CONNECT response", zap.Error(err))
				return nil, fmt.Errorf("error reading CONNECT response: %w", err)
			}

			// 检查是否 200 OK
			if !bytes.Contains(res[:n], []byte("200 Connection established")) {
				proxyConn.Close()
				logger.Error("proxy connection failed", zap.String("response", string(res[:n])))
				return nil, fmt.Errorf("proxy connection failed: %s", res[:n])
			}

			return proxyConn, nil
		},
	}
}

// 辅助函数：随机生成User-Agent
func getRandomUserAgent() string {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edge/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"}
	return userAgents[rand.Intn(len(userAgents))]
}
