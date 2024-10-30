package ddgchat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"
)

func AuthMiddleware(config *Config) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		if len(config.Tokens) == 0 {
			return c.Next()
		}

		// check bearer token
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.Error("no authorization header")
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			logger.Error("no token")
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		if !slices.Contains(config.Tokens, token) {
			logger.Error("invalid token", zap.String("token", token))
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		return c.Next()
	}
}

func ListModels(config *Config) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var models []ModelInfo
		for model := range config.ModelMapping {
			models = append(models, ModelInfo{
				ID:      model,
				Object:  "model",
				Created: time.Now().Unix(),
				OwnedBy: "custom",
			})
		}

		return c.JSON(fiber.Map{
			"data":   models,
			"object": "list",
		})
	}
}

func ChatCompletions(config *Config) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		logger.Debug("received chat completions request")
		var req ChatCompletionRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		conversationId := generateUUID()
		logger.Debug("generated conversation id", zap.String("conversation_id", conversationId))

		if req.Stream {
			c.Set("Content-Type", "text/event-stream")
			c.Set("Cache-Control", "no-cache")
			c.Set("Connection", "keep-alive")
			c.Set("Transfer-Encoding", "chunked")

			channel := make(chan string)
			go streamResponse(req, conversationId, channel, config)

			// 定义一个符合 fasthttp.StreamWriter 类型的函数
			writer := func(w *bufio.Writer) {
				for msg := range channel {
					if _, err := w.WriteString(msg); err != nil {
						logger.Error("Error writing to stream", zap.Error(err))
						return
					}
					if err := w.Flush(); err != nil {
						logger.Error("Error flushing stream", zap.Error(err))
						return
					}
				}
			}

			c.Context().SetBodyStreamWriter(writer)
			return nil
		}

		response, err := generateResponse(req, conversationId, config)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(response)
	}
}

func EndConversation(config *Config) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotImplemented)
	}
}

func streamResponse(req ChatCompletionRequest, conversationId string, channel chan string, config *Config) {
	defer close(channel)

	// 将当前对话历史加入到conversations中
	conversationMutex.Lock()
	conversations[conversationId] = req.Messages
	conversationHistory := conversations[conversationId]
	conversationMutex.Unlock()

	// 创建用于存储完整响应的buffer
	var fullResponse string

	// 创建响应通道
	responseChan := make(chan string)
	errorChan := make(chan error)

	// 在goroutine中处理DuckDuckGo的响应
	go func() {
		query := strings.Join(func() []string {
			var contents []string
			for _, msg := range req.Messages {
				contents = append(contents, msg.Content)
			}
			return contents
		}(), " ")

		if err := chatWithDuckDuckGo(query, req.Model, conversationHistory, responseChan, config); err != nil {
			errorChan <- err
			return
		}
	}()

	// 处理响应
	for {
		select {
		case chunk, ok := <-responseChan:
			if !ok {
				// 发送完成标记
				response := ChatCompletionStreamResponse{
					ID:      conversationId,
					Created: time.Now().Unix(),
					Model:   req.Model,
					Object:  "chat.completion.chunk",
					Choices: []ChatCompletionStreamResponseChoice{
						{
							Index: 0,
							Delta: DeltaMessage{},
							FinishReason: func() *string {
								s := "stop"
								return &s
							}(),
						},
					},
				}

				responseJSON, _ := json.Marshal(response)
				channel <- fmt.Sprintf("data: %s\n\n", string(responseJSON))
				channel <- "data: [DONE]\n\n"

				// 更新conversations
				conversationMutex.Lock()
				conversations[conversationId] = append(
					conversations[conversationId],
					ChatMessage{
						Role:    "assistant",
						Content: fullResponse,
					},
				)
				conversationMutex.Unlock()

				return
			}

			fullResponse += chunk

			// 创建流式响应
			response := ChatCompletionStreamResponse{
				ID:      conversationId,
				Created: time.Now().Unix(),
				Model:   req.Model,
				Object:  "chat.completion.chunk",
				Choices: []ChatCompletionStreamResponseChoice{
					{
						Index: 0,
						Delta: DeltaMessage{
							Content: &chunk,
						},
					},
				},
			}

			// 序列化并发送响应
			responseJSON, err := json.Marshal(response)
			if err != nil {
				logger.Error("Error marshaling response", zap.Error(err))
				return
			}

			channel <- fmt.Sprintf("data: %s\n\n", string(responseJSON))

			// 添加一个小延迟模拟真实的流式响应
			time.Sleep(time.Duration(50+rand.Intn(50)) * time.Millisecond)

		case err := <-errorChan:
			// 处理错误
			errorResponse := struct {
				Error string `json:"error"`
			}{
				Error: err.Error(),
			}

			errorJSON, _ := json.Marshal(errorResponse)
			channel <- fmt.Sprintf("data: %s\n\n", string(errorJSON))
			return

		case <-time.After(30 * time.Second):
			// 超时处理
			logger.Error("Stream response timeout")
			return
		}
	}
}

func generateResponse(req ChatCompletionRequest, conversationId string, config *Config) (*ChatCompletionResponse, error) {
	// 将当前对话历史加入到conversations中
	conversationMutex.Lock()
	conversations[conversationId] = req.Messages
	conversationHistory := conversations[conversationId]
	conversationMutex.Unlock()

	// 创建响应channel
	responseChan := make(chan string)
	done := make(chan bool)
	errorChan := make(chan error)
	var fullResponse string

	// 在goroutine中处理DuckDuckGo的响应
	go func() {
		logger.Debug("deal with duckduckgo response")
		query := strings.Join(func() []string {
			var contents []string
			for _, msg := range req.Messages {
				contents = append(contents, msg.Content)
			}
			return contents
		}(), " ")

		if err := chatWithDuckDuckGo(query, req.Model, conversationHistory, responseChan, config); err != nil {
			logger.Error("failed to chat with duckduckgo", zap.Error(err))
			errorChan <- err
			return
		}
		done <- true
	}()

	// 收集完整响应
	for {
		select {
		case chunk, ok := <-responseChan:
			if !ok {
				continue
			}
			fullResponse += chunk

		case err := <-errorChan:
			return nil, err

		case <-done:
			// 计算token数量（这里用简单的分词方式估算，实际项目中可能需要更准确的token计算方法）
			promptTokens := 0
			for _, msg := range conversationHistory {
				promptTokens += len(strings.Split(msg.Content, " "))
			}
			completionTokens := len(strings.Split(fullResponse, " "))
			totalTokens := promptTokens + completionTokens

			// 构建响应
			response := &ChatCompletionResponse{
				ID:      conversationId,
				Object:  "chat.completion",
				Created: time.Now().Unix(),
				Model:   req.Model,
				Choices: []ChatCompletionResponseChoice{
					{
						Index: 0,
						Message: ChatMessage{
							Role:    "assistant",
							Content: fullResponse,
						},
						FinishReason: func() *string {
							s := "stop"
							return &s
						}(),
					},
				},
				Usage: ChatCompletionResponseUsage{
					PromptTokens:     promptTokens,
					CompletionTokens: completionTokens,
					TotalTokens:      totalTokens,
				},
			}

			// 更新对话历史
			conversationMutex.Lock()
			conversations[conversationId] = append(
				conversations[conversationId],
				ChatMessage{
					Role:    "assistant",
					Content: fullResponse,
				},
			)
			conversationMutex.Unlock()

			return response, nil

		case <-time.After(30 * time.Second):
			return nil, fmt.Errorf("response generation timeout")
		}
	}
}

func RegisterRoutes(app *fiber.App, config *Config) {
	api := app.Group("/v1", AuthMiddleware(config))

	api.Get("/models", ListModels(config))
	api.Post("/chat/completions", ChatCompletions(config))
	api.Delete("/conversations/:id", EndConversation(config))
}
