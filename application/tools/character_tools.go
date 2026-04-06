package tools

import (
	"context"
	"fmt"
)

// DealDamageTool 造成伤害工具
type DealDamageTool struct {
	BaseTool
	state StateAccessor
}

// NewDealDamageTool 创建造成伤害工具
func NewDealDamageTool(state StateAccessor) *DealDamageTool {
	return &DealDamageTool{
		BaseTool: NewBaseTool("deal_damage", "对目标造成伤害。参数: target (目标), amount (伤害值), type (伤害类型)"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *DealDamageTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"target": map[string]interface{}{
				"type":        "string",
				"description": "目标名称（player 或 NPC名称）",
			},
			"amount": map[string]interface{}{
				"type":        "integer",
				"minimum":     1,
				"description": "伤害值",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"强酸", "钝击", "冷冻", "火焰", "力场", "闪电", "黯蚀", "穿刺", "毒素", "心灵", "光耀", "挥砍", "雷鸣"},
				"description": "伤害类型",
			},
		},
		"required": []string{"target", "amount", "type"},
	}
}

// Execute 执行造成伤害
func (t *DealDamageTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if t.state == nil || t.state.GetCharacter() == nil {
		return nil, ErrStateNotAvailable
	}

	target, ok := args["target"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	amountFloat, ok := args["amount"].(float64)
	if !ok {
		return nil, ErrInvalidArguments
	}
	amount := int(amountFloat)

	damageType, _ := args["type"].(string)

	// 如果目标是玩家
	c := t.state.GetCharacter()
	if target == "player" || target == c.Name {
		oldHP := c.HitPoints.Current
		c.HitPoints.TakeDamage(amount)
		newHP := c.HitPoints.Current

		narrative := fmt.Sprintf("%s 受到 %d 点%s伤害！生命值: %d -> %d", c.Name, amount, damageType, oldHP, newHP)
		if newHP == 0 {
			narrative += " (倒地!)"
		}

		return &ToolResult{
			Success:   true,
			Narrative: narrative,
			Data: map[string]interface{}{
				"target":  target,
				"amount":  amount,
				"type":    damageType,
				"old_hp":  oldHP,
				"new_hp":  newHP,
				"is_down": newHP == 0,
			},
		}, nil
	}

	// 如果目标是NPC（简化处理，实际需要查找NPC）
	return &ToolResult{
		Success:   true,
		Narrative: fmt.Sprintf("%s 受到 %d 点%s伤害", target, amount, damageType),
		Data: map[string]interface{}{
			"target": target,
			"amount": amount,
			"type":   damageType,
		},
	}, nil
}

// HealCharacterTool 治疗工具
type HealCharacterTool struct {
	BaseTool
	state StateAccessor
}

// NewHealCharacterTool 创建治疗工具
func NewHealCharacterTool(state StateAccessor) *HealCharacterTool {
	return &HealCharacterTool{
		BaseTool: NewBaseTool("heal_character", "治疗目标。参数: target (目标), amount (治疗量)"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *HealCharacterTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"target": map[string]interface{}{
				"type":        "string",
				"description": "目标名称（player 或 NPC名称）",
			},
			"amount": map[string]interface{}{
				"type":        "integer",
				"minimum":     1,
				"description": "治疗量",
			},
		},
		"required": []string{"target", "amount"},
	}
}

// Execute 执行治疗
func (t *HealCharacterTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if t.state == nil || t.state.GetCharacter() == nil {
		return nil, ErrStateNotAvailable
	}

	target, ok := args["target"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	amountFloat, ok := args["amount"].(float64)
	if !ok {
		return nil, ErrInvalidArguments
	}
	amount := int(amountFloat)

	// 如果目标是玩家
	c := t.state.GetCharacter()
	if target == "player" || target == c.Name {
		oldHP := c.HitPoints.Current
		c.HitPoints.Heal(amount)
		newHP := c.HitPoints.Current

		narrative := fmt.Sprintf("%s 恢复了 %d 点生命值！生命值: %d -> %d/%d",
			c.Name, amount, oldHP, newHP, c.HitPoints.Max)

		return &ToolResult{
			Success:   true,
			Narrative: narrative,
			Data: map[string]interface{}{
				"target": target,
				"amount": amount,
				"old_hp": oldHP,
				"new_hp": newHP,
				"max_hp": c.HitPoints.Max,
			},
		}, nil
	}

	return &ToolResult{
		Success:   true,
		Narrative: fmt.Sprintf("%s 恢复了 %d 点生命值", target, amount),
		Data: map[string]interface{}{
			"target": target,
			"amount": amount,
		},
	}, nil
}

// AddConditionTool 添加状态工具
type AddConditionTool struct {
	BaseTool
	state StateAccessor
}

// NewAddConditionTool 创建添加状态工具
func NewAddConditionTool(state StateAccessor) *AddConditionTool {
	return &AddConditionTool{
		BaseTool: NewBaseTool("add_condition", "为目标添加状态。参数: target (目标), condition (状态名), duration (持续时间，可选)"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *AddConditionTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"target": map[string]interface{}{
				"type":        "string",
				"description": "目标名称",
			},
			"condition": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"失明", "魅惑", "耳聋", "恐惧", "擒抱", "束缚", "失能", "隐形", "中毒", "倒地", "震慑", "昏迷"},
				"description": "状态名称",
			},
			"duration": map[string]interface{}{
				"type":        "integer",
				"description": "持续时间（回合数），0表示永久或直到解除",
				"default":     0,
			},
		},
		"required": []string{"target", "condition"},
	}
}

// Execute 执行添加状态
func (t *AddConditionTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	target, ok := args["target"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	condition, ok := args["condition"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	duration := 0
	if v, ok := args["duration"].(float64); ok {
		duration = int(v)
	}

	// 存储状态（简化实现，实际需要完整的状态管理系统）
	key := fmt.Sprintf("condition_%s_%s", target, condition)
	t.state.SetFlag(key, true)

	durationText := ""
	if duration > 0 {
		durationText = fmt.Sprintf("，持续 %d 回合", duration)
	}

	narrative := fmt.Sprintf("%s 获得 [%s] 状态%s", target, condition, durationText)

	return &ToolResult{
		Success:   true,
		Narrative: narrative,
		Data: map[string]interface{}{
			"target":    target,
			"condition": condition,
			"duration":  duration,
		},
	}, nil
}

// RemoveConditionTool 移除状态工具
type RemoveConditionTool struct {
	BaseTool
	state StateAccessor
}

// NewRemoveConditionTool 创建移除状态工具
func NewRemoveConditionTool(state StateAccessor) *RemoveConditionTool {
	return &RemoveConditionTool{
		BaseTool: NewBaseTool("remove_condition", "移除目标的状态。参数: target (目标), condition (状态名)"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *RemoveConditionTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"target": map[string]interface{}{
				"type":        "string",
				"description": "目标名称",
			},
			"condition": map[string]interface{}{
				"type":        "string",
				"description": "状态名称",
			},
		},
		"required": []string{"target", "condition"},
	}
}

// Execute 执行移除状态
func (t *RemoveConditionTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	target, ok := args["target"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	condition, ok := args["condition"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	key := fmt.Sprintf("condition_%s_%s", target, condition)
	t.state.SetFlag(key, false)

	narrative := fmt.Sprintf("%s 的 [%s] 状态已解除", target, condition)

	return &ToolResult{
		Success:   true,
		Narrative: narrative,
		Data: map[string]interface{}{
			"target":    target,
			"condition": condition,
		},
	}, nil
}
