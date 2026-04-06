package llm

import (
	"context"
	"errors"
	"io"

	"github.com/sashabaranov/go-openai"
	"github.com/zwh8800/cdnd/domain/llm"
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
		messages[i] = p.convertMessage(msg)
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

	// 构建请求
	chatRequest := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: float32(temp),
	}

	// 添加工具定义
	if len(req.Tools) > 0 {
		chatRequest.Tools = make([]openai.Tool, len(req.Tools))
		for i, tool := range req.Tools {
			chatRequest.Tools[i] = openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
		}
		if req.ToolChoice != nil {
			chatRequest.ToolChoice = req.ToolChoice
		}
	}

	resp, err := p.client.CreateChatCompletion(ctx, chatRequest)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no response choices returned")
	}

	// 解析响应
	response := &Response{
		ID:           resp.ID,
		Content:      resp.Choices[0].Message.Content,
		Model:        resp.Model,
		FinishReason: string(resp.Choices[0].FinishReason),
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}

	// 解析工具调用
	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		response.ToolCalls = make([]ToolCall, len(resp.Choices[0].Message.ToolCalls))
		for i, tc := range resp.Choices[0].Message.ToolCalls {
			response.ToolCalls[i] = ToolCall{
				ID:        tc.ID,
				Type:      string(tc.Type),
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			}
		}
	}

	return response, nil
}

// GenerateStream 生成流式补全。
func (p *OpenAIProvider) GenerateStream(ctx context.Context, req *Request) (<-chan StreamChunk, error) {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = p.convertMessage(msg)
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

	// 构建请求
	chatRequest := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: float32(temp),
		Stream:      true,
	}

	// 添加工具定义
	if len(req.Tools) > 0 {
		chatRequest.Tools = make([]openai.Tool, len(req.Tools))
		for i, tool := range req.Tools {
			chatRequest.Tools[i] = openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
		}
		if req.ToolChoice != nil {
			chatRequest.ToolChoice = req.ToolChoice
		}
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, chatRequest)
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
				chunk := StreamChunk{
					Content: response.Choices[0].Delta.Content,
				}
				if response.Choices[0].FinishReason != "" {
					chunk.FinishReason = string(response.Choices[0].FinishReason)
				}
				chunkChan <- chunk
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

// convertMessage 将 llm.Message 转换为 openai.ChatCompletionMessage
func (p *OpenAIProvider) convertMessage(msg Message) openai.ChatCompletionMessage {
	om := openai.ChatCompletionMessage{
		Role:    string(msg.Role),
		Content: msg.Content,
	}

	// 处理 tool 角色消息
	if msg.Role == llm.RoleTool {
		om.ToolCallID = msg.ToolCallID
	}

	// 处理 assistant 消息中的工具调用
	if len(msg.ToolCalls) > 0 {
		om.ToolCalls = make([]openai.ToolCall, len(msg.ToolCalls))
		for i, tc := range msg.ToolCalls {
			om.ToolCalls[i] = openai.ToolCall{
				ID:   tc.ID,
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      tc.Name,
					Arguments: tc.Arguments,
				},
			}
		}
	}

	return om
}
