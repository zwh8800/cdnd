package tools

import (
	"context"
)

// SetOptionsTool 设置选项工具
type SetOptionsTool struct {
	BaseTool
	state StateAccessor
}

// NewSetOptionsTool 创建设置选项工具
func NewSetOptionsTool(state StateAccessor) *SetOptionsTool {
	return &SetOptionsTool{
		BaseTool: NewBaseTool("set_options", "设置玩家当前可用的操作选项。每次响应时必须调用此工具提供可选操作列表。参数: options (字符串数组，每个元素是一个可选操作)"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *SetOptionsTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"options": map[string]interface{}{
				"type":        "array",
				"description": "玩家可用的操作选项列表",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"options"},
	}
}

// Execute 执行设置选项
func (t *SetOptionsTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if t.state == nil {
		return nil, ErrStateNotAvailable
	}

	optionsInterface, ok := args["options"].([]interface{})
	if !ok {
		return nil, ErrInvalidArguments
	}

	options := make([]string, 0, len(optionsInterface))
	for _, opt := range optionsInterface {
		if str, ok := opt.(string); ok {
			options = append(options, str)
		}
	}

	// 将选项存储到状态中
	t.state.SetCurrentOptions(options)

	return &ToolResult{
		Success:   true,
		Narrative: "",
		Data: map[string]interface{}{
			"options": options,
		},
	}, nil
}
