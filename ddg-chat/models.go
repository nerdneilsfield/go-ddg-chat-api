package ddgchat

// 常量定义
var DEFAULT_MODEL_MAPPING = map[string]string{
	"ddg/gpt-4o-mini":                       "gpt-4o-mini",
	"ddg/claude-3-haiku":                    "claude-3-haiku-20240307",
	"ddg/mixtral-8x7b":                      "mistralai/Mixtral-8x7B-Instruct-v0.1",
	"ddg/meta-Llama-3-1-70B-Instruct-Turbo": "meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo",
}

// 模型结构体定义
type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model           string             `json:"model"`
	Messages        []ChatMessage      `json:"messages"`
	Temperature     *float64           `json:"temperature,omitempty"`
	TopP            *float64           `json:"top_p,omitempty"`
	N               *int               `json:"n,omitempty"`
	Stream          bool               `json:"stream"`
	Stop            interface{}        `json:"stop,omitempty"`
	MaxTokens       *int               `json:"max_tokens,omitempty"`
	PresencePenalty *float64           `json:"presence_penalty,omitempty"`
	FreqPenalty     *float64           `json:"frequency_penalty,omitempty"`
	LogitBias       map[string]float64 `json:"logit_bias,omitempty"`
	User            *string            `json:"user,omitempty"`
}

type ChatCompletionResponseChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason *string     `json:"finish_reason,omitempty"`
}

type ChatCompletionResponseUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatCompletionResponse struct {
	ID      string                         `json:"id"`
	Object  string                         `json:"object"`
	Created int64                          `json:"created"`
	Model   string                         `json:"model"`
	Choices []ChatCompletionResponseChoice `json:"choices"`
	Usage   ChatCompletionResponseUsage    `json:"usage"`
}

type DeltaMessage struct {
	Role    *string `json:"role,omitempty"`
	Content *string `json:"content,omitempty"`
}

type ChatCompletionStreamResponseChoice struct {
	Index        int          `json:"index"`
	Delta        DeltaMessage `json:"delta"`
	FinishReason *string      `json:"finish_reason,omitempty"`
}

type ChatCompletionStreamResponse struct {
	ID      string                               `json:"id"`
	Object  string                               `json:"object"`
	Created int64                                `json:"created"`
	Model   string                               `json:"model"`
	Choices []ChatCompletionStreamResponseChoice `json:"choices"`
}
