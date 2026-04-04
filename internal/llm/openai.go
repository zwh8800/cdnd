package llm

import (
	"context"
	"errors"
	"io"

	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider 为 OpenAI API 实现 Provider 接口。
type OpenAIProvider struct {
	client      *openai.Client
	model       string
	maxTokens   int
	temperature float64
	baseURL     string
}

// NewOpenAIProvider 创建一个新的 OpenAI 提供者。
func NewOpenAIProvider(cfg ProviderConfig) *OpenAIProvider {
	config := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		config.BaseURL = cfg.BaseURL
	}

	return &OpenAIProvider{
		client:      openai.NewClientWithConfig(config),
		model:       cfg.Model,
		maxTokens:   cfg.MaxTokens,
		temperature: cfg.Temperature,
		baseURL:     cfg.BaseURL,
	}
}

// Name 返回提供者名称。
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Generate 为给定请求生成补全。
func (p *OpenAIProvider) Generate(ctx context.Context, req *Request) (*Response, error) {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	model := req.Model
	if model == "" {
		model = p.model
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = p.maxTokens
	}

	temp := req.Temperature
	if temp == 0 {
		temp = p.temperature
	}

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: float32(temp),
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no response choices returned")
	}

	return &Response{
		ID:      resp.ID,
		Content: resp.Choices[0].Message.Content,
		Model:   resp.Model,
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

// GenerateStream 生成流式补全。
func (p *OpenAIProvider) GenerateStream(ctx context.Context, req *Request) (<-chan StreamChunk, error) {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	model := req.Model
	if model == "" {
		model = p.model
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = p.maxTokens
	}

	temp := req.Temperature
	if temp == 0 {
		temp = p.temperature
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: float32(temp),
		Stream:      true,
	})
	if err != nil {
		return nil, err
	}

	chunkChan := make(chan StreamChunk, 100)

	go func() {
		defer close(chunkChan)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				chunkChan <- StreamChunk{Done: true}
				return
			}
			if err != nil {
				chunkChan <- StreamChunk{Error: err}
				return
			}

			if len(response.Choices) > 0 {
				chunkChan <- StreamChunk{
					Content: response.Choices[0].Delta.Content,
				}
			}
		}
	}()

	return chunkChan, nil
}

// SetModel 设置要使用的模型。
func (p *OpenAIProvider) SetModel(model string) {
	p.model = model
}

// SetMaxTokens 设置生成的最大令牌数。
func (p *OpenAIProvider) SetMaxTokens(maxTokens int) {
	p.maxTokens = maxTokens
}

// SetTemperature 设置生成的温度。
func (p *OpenAIProvider) SetTemperature(temp float64) {
	p.temperature = temp
}
