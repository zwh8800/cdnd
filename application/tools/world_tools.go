package tools

import (
	"context"
	"fmt"
)

// MoveToSceneTool 移动到场景工具
type MoveToSceneTool struct {
	BaseTool
	state StateAccessor
}

// NewMoveToSceneTool 创建移动到场景工具
func NewMoveToSceneTool(state StateAccessor) *MoveToSceneTool {
	return &MoveToSceneTool{
		BaseTool: NewBaseTool("move_to_scene", "移动角色到新场景。参数: scene_id (场景ID), scene_name (场景名称), description (场景描述)"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *MoveToSceneTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"scene_id": map[string]interface{}{
				"type":        "string",
				"description": "场景唯一标识",
			},
			"scene_name": map[string]interface{}{
				"type":        "string",
				"description": "场景名称",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "场景描述",
			},
		},
		"required": []string{"scene_id", "scene_name"},
	}
}

// Execute 执行移动到场景
func (t *MoveToSceneTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if t.state == nil {
		return nil, ErrStateNotAvailable
	}

	sceneID, ok := args["scene_id"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	sceneName, ok := args["scene_name"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	description, _ := args["description"].(string)

	t.state.SetFlag(fmt.Sprintf("visited_%s", sceneID), true)
	t.state.SetFlag("current_scene_id", true)
	t.state.SetCounter("scene_transition", t.state.GetCounter("scene_transition")+1)

	narrative := fmt.Sprintf("进入: **%s**", sceneName)
	if description != "" {
		narrative += fmt.Sprintf("\n\n%s", description)
	}

	return &ToolResult{
		Success:   true,
		Narrative: narrative,
		Data: map[string]interface{}{
			"scene_id":    sceneID,
			"scene_name":  sceneName,
			"description": description,
		},
	}, nil
}

// SpawnNPCTool 生成NPC工具
type SpawnNPCTool struct {
	BaseTool
	state StateAccessor
}

// NewSpawnNPCTool 创建生成NPC工具
func NewSpawnNPCTool(state StateAccessor) *SpawnNPCTool {
	return &SpawnNPCTool{
		BaseTool: NewBaseTool("spawn_npc", "在当前场景生成NPC。参数: npc_id, name, role, description"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *SpawnNPCTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"npc_id": map[string]interface{}{
				"type":        "string",
				"description": "NPC唯一标识",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "NPC名称",
			},
			"role": map[string]interface{}{
				"type":        "string",
				"description": "NPC角色/职业",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "NPC外观描述",
			},
		},
		"required": []string{"npc_id", "name"},
	}
}

// Execute 执行生成NPC
func (t *SpawnNPCTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	npcID, ok := args["npc_id"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	name, ok := args["name"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	role, _ := args["role"].(string)
	description, _ := args["description"].(string)

	key := fmt.Sprintf("npc_%s", npcID)
	t.state.SetFlag(key, true)

	roleText := ""
	if role != "" {
		roleText = fmt.Sprintf(" (%s)", role)
	}

	narrative := fmt.Sprintf("**%s**%s 出现了", name, roleText)
	if description != "" {
		narrative += fmt.Sprintf(" - %s", description)
	}

	return &ToolResult{
		Success:   true,
		Narrative: narrative,
		Data: map[string]interface{}{
			"npc_id":      npcID,
			"name":        name,
			"role":        role,
			"description": description,
		},
	}, nil
}

// RemoveNPCTool 移除NPC工具
type RemoveNPCTool struct {
	BaseTool
	state StateAccessor
}

// NewRemoveNPCTool 创建移除NPC工具
func NewRemoveNPCTool(state StateAccessor) *RemoveNPCTool {
	return &RemoveNPCTool{
		BaseTool: NewBaseTool("remove_npc", "从当前场景移除NPC。参数: npc_id, reason"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *RemoveNPCTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"npc_id": map[string]interface{}{
				"type":        "string",
				"description": "NPC唯一标识",
			},
			"reason": map[string]interface{}{
				"type":        "string",
				"description": "移除原因",
			},
		},
		"required": []string{"npc_id"},
	}
}

// Execute 执行移除NPC
func (t *RemoveNPCTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	npcID, ok := args["npc_id"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	reason, _ := args["reason"].(string)

	key := fmt.Sprintf("npc_%s", npcID)
	t.state.SetFlag(key, false)

	narrative := fmt.Sprintf("NPC [%s] 离开了", npcID)
	if reason != "" {
		narrative += fmt.Sprintf(" (%s)", reason)
	}

	return &ToolResult{
		Success:   true,
		Narrative: narrative,
		Data: map[string]interface{}{
			"npc_id": npcID,
			"reason": reason,
		},
	}, nil
}

// SetFlagTool 设置标志工具
type SetFlagTool struct {
	BaseTool
	state StateAccessor
}

// NewSetFlagTool 创建设置标志工具
func NewSetFlagTool(state StateAccessor) *SetFlagTool {
	return &SetFlagTool{
		BaseTool: NewBaseTool("set_flag", "设置游戏世界标志。参数: key, value"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *SetFlagTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"key": map[string]interface{}{
				"type":        "string",
				"description": "标志键名",
			},
			"value": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "标志值",
			},
		},
		"required": []string{"key"},
	}
}

// Execute 执行设置标志
func (t *SetFlagTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if t.state == nil {
		return nil, ErrStateNotAvailable
	}

	key, ok := args["key"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	value := true
	if v, ok := args["value"].(bool); ok {
		value = v
	}

	t.state.SetFlag(key, value)

	return &ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"key":   key,
			"value": value,
		},
	}, nil
}

// GetFlagTool 获取标志工具
type GetFlagTool struct {
	BaseTool
	state StateAccessor
}

// NewGetFlagTool 创建获取标志工具
func NewGetFlagTool(state StateAccessor) *GetFlagTool {
	return &GetFlagTool{
		BaseTool: NewBaseTool("get_flag", "获取游戏世界标志的值。参数: key"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *GetFlagTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"key": map[string]interface{}{
				"type":        "string",
				"description": "标志键名",
			},
		},
		"required": []string{"key"},
	}
}

// Execute 执行获取标志
func (t *GetFlagTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if t.state == nil {
		return nil, ErrStateNotAvailable
	}

	key, ok := args["key"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	value := t.state.GetFlag(key)

	return &ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"key":   key,
			"value": value,
		},
	}, nil
}
