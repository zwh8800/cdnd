package llm

import (
	llmd "github.com/zwh8800/cdnd/domain/llm"
)

// Type aliases from domain/llm
type (
	Provider               = llmd.Provider
	ProviderConfig         = llmd.ProviderConfig
	Request                = llmd.Request
	Response               = llmd.Response
	StreamChunk            = llmd.StreamChunk
	Message                = llmd.Message
	MessageRole            = llmd.MessageRole
	ToolCall               = llmd.ToolCall
	ToolDefinition         = llmd.ToolDefinition
	ToolFunctionDefinition = llmd.ToolFunctionDefinition
	Usage                  = llmd.Usage
)
