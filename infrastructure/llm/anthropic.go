package llm

import (
	"context"
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/zwh8800/cdnd/domain/llm"
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
func (p *AnthropicProvider) Generate(ctx context.Context, req *Request) (*Response, error) {
	messages := make([]anthropic.MessageParam, 0, len(req.Messages))
	var systemPrompt string

	for _, msg := range req.Messages {
		switch msg.Role {
		case llm.RoleUser:
			messages = append(messages, p.convertUserMessage(msg))
		case llm.RoleAssistant:
			messages = append(messages, p.convertAssistantMessage(msg))
		case llm.RoleSystem:
			systemPrompt = msg.Content
		case llm.RoleTool:
			messages = append(messages, p.convertToolResultMessage(msg))
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

	params := anthropic.MessageNewParams{
		MaxTokens: maxTokens,
		Messages:  messages,
		Model:     model,
	}

	if systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{{Text: systemPrompt}}
	}

	if len(req.Tools) > 0 {
		tools := make([]anthropic.ToolUnionParam, len(req.Tools))
		for i, tool := range req.Tools {
			inputSchema := anthropic.ToolInputSchemaParam{
				Properties: tool.Function.Parameters,
			}
			if req, ok := tool.Function.Parameters["required"].([]string); ok {
				inputSchema.Required = req
			}

			toolParam := anthropic.ToolParam{
				Name:        tool.Function.Name,
				InputSchema: inputSchema,
			}
			if tool.Function.Description != "" {
				toolParam.Description = anthropic.Opt(tool.Function.Description)
			}
			tools[i] = anthropic.ToolUnionParam{OfTool: &toolParam}
		}
		params.Tools = tools
	}

	message, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return nil, err
	}

	response := &Response{
		ID:    message.ID,
		Model: message.Model,
		Usage: Usage{
			PromptTokens:     int(message.Usage.InputTokens),
			CompletionTokens: int(message.Usage.OutputTokens),
			TotalTokens:      int(message.Usage.InputTokens + message.Usage.OutputTokens),
		},
	}

	for _, block := range message.Content {
		switch block.Type {
		case "text":
			response.Content += block.Text
		case "tool_use":
			argsJSON, _ := json.Marshal(block.Input)
			response.ToolCalls = append(response.ToolCalls, ToolCall{
				ID:        block.ID,
				Type:      "function",
				Name:      block.Name,
				Arguments: string(argsJSON),
			})
		}
	}

	switch message.StopReason {
	case anthropic.StopReasonEndTurn:
		response.FinishReason = "stop"
	case anthropic.StopReasonToolUse:
		response.FinishReason = "tool_calls"
	}

	return response, nil
}

// GenerateStream 生成流式补全。
func (p *AnthropicProvider) GenerateStream(ctx context.Context, req *Request) (<-chan StreamChunk, error) {
	messages := make([]anthropic.MessageParam, 0, len(req.Messages))
	var systemPrompt string

	for _, msg := range req.Messages {
		switch msg.Role {
		case llm.RoleUser:
			messages = append(messages, p.convertUserMessage(msg))
		case llm.RoleAssistant:
			messages = append(messages, p.convertAssistantMessage(msg))
		case llm.RoleSystem:
			systemPrompt = msg.Content
		case llm.RoleTool:
			messages = append(messages, p.convertToolResultMessage(msg))
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

	msgParams := anthropic.MessageNewParams{
		MaxTokens: maxTokens,
		Messages:  messages,
		Model:     model,
	}

	if systemPrompt != "" {
		msgParams.System = []anthropic.TextBlockParam{{Text: systemPrompt}}
	}

	if len(req.Tools) > 0 {
		tools := make([]anthropic.ToolUnionParam, len(req.Tools))
		for i, tool := range req.Tools {
			inputSchema := anthropic.ToolInputSchemaParam{
				Properties: tool.Function.Parameters,
			}
			if req, ok := tool.Function.Parameters["required"].([]string); ok {
				inputSchema.Required = req
			}

			toolParam := anthropic.ToolParam{
				Name:        tool.Function.Name,
				InputSchema: inputSchema,
			}
			if tool.Function.Description != "" {
				toolParam.Description = anthropic.Opt(tool.Function.Description)
			}
			tools[i] = anthropic.ToolUnionParam{OfTool: &toolParam}
		}
		msgParams.Tools = tools
	}

	stream := p.client.Messages.NewStreaming(ctx, msgParams)
	chunkChan := make(chan StreamChunk, 100)

	go func() {
		defer close(chunkChan)

		for stream.Next() {
			event := stream.Current()
			switch event.Type {
			case "content_block_delta":
				delta := event.AsContentBlockDelta()
				if delta.Delta.Type == "text_delta" {
					chunkChan <- StreamChunk{Content: delta.Delta.Text}
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

// convertUserMessage 转换用户消息
func (p *AnthropicProvider) convertUserMessage(msg Message) anthropic.MessageParam {
	return anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content))
}

// convertAssistantMessage 转换助手消息
func (p *AnthropicProvider) convertAssistantMessage(msg Message) anthropic.MessageParam {
	blocks := make([]anthropic.ContentBlockParamUnion, 0)
	if msg.Content != "" {
		blocks = append(blocks, anthropic.NewTextBlock(msg.Content))
	}
	for _, tc := range msg.ToolCalls {
		var input map[string]interface{}
		json.Unmarshal([]byte(tc.Arguments), &input)
		blocks = append(blocks, anthropic.NewToolUseBlock(tc.ID, input, tc.Name))
	}
	return anthropic.NewAssistantMessage(blocks...)
}

// convertToolResultMessage 转换工具结果消息
func (p *AnthropicProvider) convertToolResultMessage(msg Message) anthropic.MessageParam {
	return anthropic.NewUserMessage(
		anthropic.NewToolResultBlock(msg.ToolCallID, msg.Content, true),
	)
}
