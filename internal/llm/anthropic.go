package llm

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// AnthropicProvider 为 Anthropic Claude API 实现 Provider 接口。
type AnthropicProvider struct {
	client      *anthropic.Client
	model       string
	maxTokens   int
	temperature float64
}

// NewAnthropicProvider 创建一个新的 Anthropic 提供者。
func NewAnthropicProvider(cfg ProviderConfig) *AnthropicProvider {
	opts := []option.RequestOption{}
	if cfg.APIKey != "" {
		opts = append(opts, option.WithAPIKey(cfg.APIKey))
	}

	client := anthropic.NewClient(opts...)

	return &AnthropicProvider{
		client:      &client,
		model:       cfg.Model,
		maxTokens:   cfg.MaxTokens,
		temperature: cfg.Temperature,
	}
}

// Name 返回提供者名称。
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

// Generate 为给定请求生成补全。
// 注意：此实现需要正确的 SDK 集成。
// Anthropic SDK 有特定的参数类型需要匹配。
func (p *AnthropicProvider) Generate(ctx context.Context, req *Request) (*Response, error) {
	// 构建消息
	messages := make([]anthropic.MessageParam, 0, len(req.Messages))

	for _, msg := range req.Messages {
		switch msg.Role {
		case RoleUser:
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		case RoleAssistant:
			messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
		// 系统消息在新 SDK 中以不同方式处理
		case RoleSystem:
			// 系统提示需要单独添加到参数中
		}
	}

	model := req.Model
	if model == "" {
		model = p.model
	}

	maxTokens := int64(req.MaxTokens)
	if maxTokens == 0 {
		maxTokens = int64(p.maxTokens)
	}

	// 正确使用 SDK 的参数类型
	params := anthropic.MessageNewParams{
		MaxTokens: maxTokens,
		Messages:  messages,
	}

	// 设置模型
	if model != "" {
		params.Model = model
	}

	message, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return nil, err
	}

	// 提取文本内容
	var content string
	for _, block := range message.Content {
		content += block.Text
	}

	return &Response{
		ID:      message.ID,
		Content: content,
		Model:   message.Model,
		Usage: Usage{
			PromptTokens:     int(message.Usage.InputTokens),
			CompletionTokens: int(message.Usage.OutputTokens),
			TotalTokens:      int(message.Usage.InputTokens + message.Usage.OutputTokens),
		},
	}, nil
}

// GenerateStream 生成流式补全。
func (p *AnthropicProvider) GenerateStream(ctx context.Context, req *Request) (<-chan StreamChunk, error) {
	// 构建消息
	messages := make([]anthropic.MessageParam, 0, len(req.Messages))
	var systemPrompt string

	for _, msg := range req.Messages {
		switch msg.Role {
		case RoleUser:
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		case RoleAssistant:
			messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
		case RoleSystem:
			// 系统消息在 Anthropic SDK 中单独处理
			systemPrompt = msg.Content
		}
	}

	model := req.Model
	if model == "" {
		model = p.model
	}

	maxTokens := int64(req.MaxTokens)
	if maxTokens == 0 {
		maxTokens = int64(p.maxTokens)
	}

	// 构建流式请求参数
	params := anthropic.MessageNewParams{
		MaxTokens: maxTokens,
		Messages:  messages,
		Model:     model,
	}

	if systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{{Text: systemPrompt}}
	}

	// 创建流式请求
	stream := p.client.Messages.NewStreaming(ctx, params)

	chunkChan := make(chan StreamChunk, 100)

	go func() {
		defer close(chunkChan)

		for stream.Next() {
			event := stream.Current()

			switch event.Type {
			case "content_block_delta":
				// 使用 AsContentBlockDelta 获取增量事件
				delta := event.AsContentBlockDelta()
				// 检查是否是文本增量
				if delta.Delta.Type == "text_delta" {
					chunkChan <- StreamChunk{
						Content: delta.Delta.Text,
					}
				}
			case "message_stop":
				chunkChan <- StreamChunk{Done: true}
				return
			}
		}

		if err := stream.Err(); err != nil {
			chunkChan <- StreamChunk{Error: err}
		}
	}()

	return chunkChan, nil
}

// SetModel 设置要使用的模型。
func (p *AnthropicProvider) SetModel(model string) {
	p.model = model
}

// SetMaxTokens 设置生成的最大令牌数。
func (p *AnthropicProvider) SetMaxTokens(maxTokens int) {
	p.maxTokens = maxTokens
}

// SetTemperature 设置生成的温度。
func (p *AnthropicProvider) SetTemperature(temp float64) {
	p.temperature = temp
}
