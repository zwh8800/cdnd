// Package llm 提供 LLM 提供者抽象和实现。
package llm

import (
	"context"
)

// MessageRole 表示消息发送者的角色。
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

// Message 表示对话中的单条消息。
type Message struct {
	Role       MessageRole `json:"role"`
	Content    string      `json:"content"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`   // assistant 消息包含的工具调用
	ToolCallID string      `json:"tool_call_id,omitempty"` // tool 角色消息关联的调用ID
	Name       string      `json:"name,omitempty"`         // tool 角色消息的工具名称
}

// Request 表示生成文本的请求。
type Request struct {
	Messages    []Message        `json:"messages"`
	Model       string           `json:"model,omitempty"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Temperature float64          `json:"temperature,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
	Tools       []ToolDefinition `json:"tools,omitempty"`
	ToolChoice  interface{}      `json:"tool_choice,omitempty"` // "auto"|"none"|"required"|具体工具名
}

// Response 表示来自 LLM 的响应。
type Response struct {
	ID           string     `json:"id"`
	Content      string     `json:"content"`
	Model        string     `json:"model"`
	Usage        Usage      `json:"usage"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	FinishReason string     `json:"finish_reason,omitempty"` // "stop"|"tool_calls"|"length"
}

// Usage 表示令牌使用量信息。
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk 表示流式响应的数据块。
type StreamChunk struct {
	Content      string     `json:"content"`
	Done         bool       `json:"done"`
	Error        error      `json:"error,omitempty"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	FinishReason string     `json:"finish_reason,omitempty"`
}

// Provider 定义 LLM 提供者的接口。
type Provider interface {
	// Name 返回提供者名称。
	Name() string

	// Generate 为给定请求生成补全。
	Generate(ctx context.Context, req *Request) (*Response, error)

	// GenerateStream 生成流式补全。
	GenerateStream(ctx context.Context, req *Request) (<-chan StreamChunk, error)

	// SetModel 设置要使用的模型。
	SetModel(model string)

	// SetMaxTokens 设置生成的最大令牌数。
	SetMaxTokens(maxTokens int)

	// SetTemperature 设置生成的温度。
	SetTemperature(temp float64)
}

// ProviderConfig 包含提供者的配置。
type ProviderConfig struct {
	APIKey      string
	Model       string
	BaseURL     string
	MaxTokens   int
	Temperature float64
}

// ToolDefinition 工具定义（用于发送给 LLM）
type ToolDefinition struct {
	Type     string                 `json:"type"`
	Function ToolFunctionDefinition `json:"function"`
}

// ToolFunctionDefinition 函数定义
type ToolFunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCall 工具调用请求（从 LLM 返回）
type ToolCall struct {
	ID        string `json:"id"`
	Type      string `json:"type,omitempty"` // 通常为 "function"
	Name      string `json:"name"`           // 工具名称
	Arguments string `json:"arguments"`      // JSON 字符串格式的参数
}
