package tools

import (
	"context"
	"fmt"
)

// AddItemTool 获得物品工具
type AddItemTool struct {
	BaseTool
	state StateAccessor
}

// NewAddItemTool 创建获得物品工具
func NewAddItemTool(state StateAccessor) *AddItemTool {
	return &AddItemTool{
		BaseTool: NewBaseTool("add_item", "角色获得物品。参数: item_id (物品ID), name (物品名称), quantity (数量)"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *AddItemTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"item_id": map[string]interface{}{
				"type":        "string",
				"description": "物品唯一标识",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "物品名称",
			},
			"quantity": map[string]interface{}{
				"type":        "integer",
				"minimum":     1,
				"default":     1,
				"description": "数量",
			},
		},
		"required": []string{"item_id", "name"},
	}
}

// Execute 执行获得物品
func (t *AddItemTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if t.state == nil || t.state.GetCharacter() == nil {
		return nil, ErrStateNotAvailable
	}

	itemID, ok := args["item_id"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	name, ok := args["name"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	quantity := 1
	if v, ok := args["quantity"].(float64); ok {
		quantity = int(v)
	}

	// 简化实现：记录到世界标志
	key := fmt.Sprintf("item_%s", itemID)
	t.state.SetCounter(key, t.state.GetCounter(key)+quantity)

	quantityText := ""
	if quantity > 1 {
		quantityText = fmt.Sprintf(" x%d", quantity)
	}

	c := t.state.GetCharacter()
	narrative := fmt.Sprintf("%s 获得了 [%s]%s", c.Name, name, quantityText)

	return &ToolResult{
		Success:   true,
		Narrative: narrative,
		Data: map[string]interface{}{
			"item_id":  itemID,
			"name":     name,
			"quantity": quantity,
		},
	}, nil
}

// RemoveItemTool 失去物品工具
type RemoveItemTool struct {
	BaseTool
	state StateAccessor
}

// NewRemoveItemTool 创建失去物品工具
func NewRemoveItemTool(state StateAccessor) *RemoveItemTool {
	return &RemoveItemTool{
		BaseTool: NewBaseTool("remove_item", "角色失去物品。参数: item_id (物品ID), quantity (数量)"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *RemoveItemTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"item_id": map[string]interface{}{
				"type":        "string",
				"description": "物品唯一标识",
			},
			"quantity": map[string]interface{}{
				"type":        "integer",
				"minimum":     1,
				"default":     1,
				"description": "数量",
			},
		},
		"required": []string{"item_id"},
	}
}

// Execute 执行失去物品
func (t *RemoveItemTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if t.state == nil || t.state.GetCharacter() == nil {
		return nil, ErrStateNotAvailable
	}

	itemID, ok := args["item_id"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	quantity := 1
	if v, ok := args["quantity"].(float64); ok {
		quantity = int(v)
	}

	key := fmt.Sprintf("item_%s", itemID)
	current := t.state.GetCounter(key)
	if current < quantity {
		return &ToolResult{
			Success: false,
			Error:   "物品数量不足",
		}, nil
	}

	t.state.SetCounter(key, current-quantity)

	c := t.state.GetCharacter()
	narrative := fmt.Sprintf("%s 失去了 [%s] x%d", c.Name, itemID, quantity)

	return &ToolResult{
		Success:   true,
		Narrative: narrative,
		Data: map[string]interface{}{
			"item_id":  itemID,
			"quantity": quantity,
		},
	}, nil
}

// SpendGoldTool 花费金币工具
type SpendGoldTool struct {
	BaseTool
	state StateAccessor
}

// NewSpendGoldTool 创建花费金币工具
func NewSpendGoldTool(state StateAccessor) *SpendGoldTool {
	return &SpendGoldTool{
		BaseTool: NewBaseTool("spend_gold", "花费金币。参数: amount (金额)"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *SpendGoldTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"amount": map[string]interface{}{
				"type":        "integer",
				"minimum":     1,
				"description": "花费金额（金币）",
			},
		},
		"required": []string{"amount"},
	}
}

// Execute 执行花费金币
func (t *SpendGoldTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if t.state == nil || t.state.GetCharacter() == nil {
		return nil, ErrStateNotAvailable
	}

	amountFloat, ok := args["amount"].(float64)
	if !ok {
		return nil, ErrInvalidArguments
	}
	amount := int(amountFloat)

	c := t.state.GetCharacter()
	if c.Gold < amount {
		return &ToolResult{
			Success:   false,
			Error:     "金币不足",
			Narrative: fmt.Sprintf("%s 没有足够的金币（需要 %d，仅有 %d）", c.Name, amount, c.Gold),
		}, nil
	}

	oldGold := c.Gold
	c.Gold -= amount

	narrative := fmt.Sprintf("%s 花费了 %d 金币（%d -> %d）", c.Name, amount, oldGold, c.Gold)

	return &ToolResult{
		Success:   true,
		Narrative: narrative,
		Data: map[string]interface{}{
			"amount":   amount,
			"old_gold": oldGold,
			"new_gold": c.Gold,
		},
	}, nil
}

// GainGoldTool 获得金币工具
type GainGoldTool struct {
	BaseTool
	state StateAccessor
}

// NewGainGoldTool 创建获得金币工具
func NewGainGoldTool(state StateAccessor) *GainGoldTool {
	return &GainGoldTool{
		BaseTool: NewBaseTool("gain_gold", "获得金币。参数: amount (金额)"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *GainGoldTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"amount": map[string]interface{}{
				"type":        "integer",
				"minimum":     1,
				"description": "获得金额（金币）",
			},
		},
		"required": []string{"amount"},
	}
}

// Execute 执行获得金币
func (t *GainGoldTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if t.state == nil || t.state.GetCharacter() == nil {
		return nil, ErrStateNotAvailable
	}

	amountFloat, ok := args["amount"].(float64)
	if !ok {
		return nil, ErrInvalidArguments
	}
	amount := int(amountFloat)

	c := t.state.GetCharacter()
	oldGold := c.Gold
	c.Gold += amount

	narrative := fmt.Sprintf("%s 获得了 %d 金币（%d -> %d）", c.Name, amount, oldGold, c.Gold)

	return &ToolResult{
		Success:   true,
		Narrative: narrative,
		Data: map[string]interface{}{
			"amount":   amount,
			"old_gold": oldGold,
			"new_gold": c.Gold,
		},
	}, nil
}
