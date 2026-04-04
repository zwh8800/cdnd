package tools

import (
	"context"
	"fmt"

	"github.com/zwh8800/cdnd/internal/character"
	"github.com/zwh8800/cdnd/internal/rules"
	"github.com/zwh8800/cdnd/pkg/dice"
)

// RollDiceTool 投骰子工具
type RollDiceTool struct {
	BaseTool
}

// NewRollDiceTool 创建投骰子工具
func NewRollDiceTool() *RollDiceTool {
	return &RollDiceTool{
		BaseTool: NewBaseTool("roll_dice", "投骰子。参数: notation (骰子表达式，如 1d20+5)"),
	}
}

// Parameters 返回参数定义
func (t *RollDiceTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"notation": map[string]interface{}{
				"type":        "string",
				"description": "骰子表达式，如 1d20+5, 2d6, 1d20adv (优势), 1d20dis (劣势)",
			},
		},
		"required": []string{"notation"},
	}
}

// Execute 执行投骰子
func (t *RollDiceTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	notation, ok := args["notation"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	result, err := dice.ParseAndRoll(notation)
	if err != nil {
		return nil, fmt.Errorf("投骰失败: %w", err)
	}

	criticalText := ""
	if result.Critical == dice.CritSuccess {
		criticalText = " (大成功!)"
	} else if result.Critical == dice.CritFail {
		criticalText = " (大失败!)"
	}

	narrative := fmt.Sprintf("投掷 %s: %d%s", notation, result.Total, criticalText)

	return &ToolResult{
		Success:   true,
		Narrative: narrative,
		Data: map[string]interface{}{
			"notation":  notation,
			"total":     result.Total,
			"dice":      result.Dice,
			"modifier":  result.Modifier,
			"critical":  int(result.Critical),
			"roll_type": int(result.RollType),
		},
	}, nil
}

// SkillCheckTool 技能检定工具
type SkillCheckTool struct {
	BaseTool
	state StateAccessor
	rules *rules.Engine
}

// NewSkillCheckTool 创建技能检定工具
func NewSkillCheckTool(state StateAccessor, rulesEngine *rules.Engine) *SkillCheckTool {
	return &SkillCheckTool{
		BaseTool: NewBaseTool("skill_check", "进行技能检定。参数: skill (技能名称), dc (难度等级), advantage (是否优势，可选)"),
		state:    state,
		rules:    rulesEngine,
	}
}

// Parameters 返回参数定义
func (t *SkillCheckTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"skill": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"运动", "体操", "手法", "隐匿", "奥秘", "历史", "调查", "自然", "宗教", "驯兽", "洞察", "医药", "察觉", "求生", "欺瞒", "威吓", "表演", "说服"},
				"description": "要检定的技能名称",
			},
			"dc": map[string]interface{}{
				"type":        "integer",
				"minimum":     1,
				"maximum":     30,
				"description": "难度等级 (DC)",
			},
			"advantage": map[string]interface{}{
				"type":        "boolean",
				"description": "是否具有优势",
				"default":     false,
			},
		},
		"required": []string{"skill", "dc"},
	}
}

// skillNameToType 技能名称映射
var skillNameToType = map[string]character.SkillType{
	"运动": character.Athletics,
	"体操": character.Acrobatics,
	"手法": character.SleightOfHand,
	"隐匿": character.Stealth,
	"奥秘": character.Arcana,
	"历史": character.History,
	"调查": character.Investigation,
	"自然": character.Nature,
	"宗教": character.Religion,
	"驯兽": character.AnimalHandling,
	"洞察": character.Insight,
	"医药": character.Medicine,
	"察觉": character.Perception,
	"求生": character.Survival,
	"欺瞒": character.Deception,
	"威吓": character.Intimidation,
	"表演": character.Performance,
	"说服": character.Persuasion,
}

// Execute 执行技能检定
func (t *SkillCheckTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if t.state == nil || t.state.GetCharacter() == nil {
		return nil, ErrStateNotAvailable
	}

	skillName, ok := args["skill"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	dcFloat, ok := args["dc"].(float64)
	if !ok {
		return nil, ErrInvalidArguments
	}
	dc := int(dcFloat)

	advantage := false
	if v, ok := args["advantage"].(bool); ok {
		advantage = v
	}

	skillType, ok := skillNameToType[skillName]
	if !ok {
		return nil, fmt.Errorf("未知技能: %s", skillName)
	}

	rollType := dice.NormalRoll
	if advantage {
		rollType = dice.AdvantageRoll
	}

	result := t.rules.SkillCheck(t.state.GetCharacter(), skillType, dc, rollType)

	resultText := "失败"
	if result.Success {
		resultText = "成功"
	}

	criticalText := ""
	if result.Critical == rules.CriticalSuccess {
		criticalText = " (大成功!)"
	} else if result.Critical == rules.CriticalFailure {
		criticalText = " (大失败!)"
	}

	narrative := fmt.Sprintf("%s检定: 投出 %d (DC %d) - %s%s", skillName, result.Total, dc, resultText, criticalText)

	return &ToolResult{
		Success:   result.Success,
		Narrative: narrative,
		Data: map[string]interface{}{
			"skill":    skillName,
			"roll":     result.Roll.Total,
			"modifier": result.Total - result.Roll.Total,
			"total":    result.Total,
			"dc":       dc,
			"success":  result.Success,
			"critical": result.Critical.String(),
		},
	}, nil
}

// SavingThrowTool 豁免检定工具
type SavingThrowTool struct {
	BaseTool
	state StateAccessor
	rules *rules.Engine
}

// NewSavingThrowTool 创建豁免检定工具
func NewSavingThrowTool(state StateAccessor, rulesEngine *rules.Engine) *SavingThrowTool {
	return &SavingThrowTool{
		BaseTool: NewBaseTool("saving_throw", "进行豁免检定。参数: ability (属性), dc (难度等级), advantage (是否优势，可选)"),
		state:    state,
		rules:    rulesEngine,
	}
}

// Parameters 返回参数定义
func (t *SavingThrowTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"ability": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"力量", "敏捷", "体质", "智力", "感知", "魅力"},
				"description": "豁免属性",
			},
			"dc": map[string]interface{}{
				"type":        "integer",
				"minimum":     1,
				"maximum":     30,
				"description": "难度等级 (DC)",
			},
			"advantage": map[string]interface{}{
				"type":        "boolean",
				"description": "是否具有优势",
				"default":     false,
			},
		},
		"required": []string{"ability", "dc"},
	}
}

// abilityNameToType 属性名称映射
var abilityNameToType = map[string]character.Ability{
	"力量": character.Strength,
	"敏捷": character.Dexterity,
	"体质": character.Constitution,
	"智力": character.Intelligence,
	"感知": character.Wisdom,
	"魅力": character.Charisma,
}

// Execute 执行豁免检定
func (t *SavingThrowTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if t.state == nil || t.state.GetCharacter() == nil {
		return nil, ErrStateNotAvailable
	}

	abilityName, ok := args["ability"].(string)
	if !ok {
		return nil, ErrInvalidArguments
	}

	dcFloat, ok := args["dc"].(float64)
	if !ok {
		return nil, ErrInvalidArguments
	}
	dc := int(dcFloat)

	advantage := false
	if v, ok := args["advantage"].(bool); ok {
		advantage = v
	}

	ability, ok := abilityNameToType[abilityName]
	if !ok {
		return nil, fmt.Errorf("未知属性: %s", abilityName)
	}

	rollType := dice.NormalRoll
	if advantage {
		rollType = dice.AdvantageRoll
	}

	result := t.rules.SavingThrow(t.state.GetCharacter(), ability, dc, rollType)

	resultText := "失败"
	if result.Success {
		resultText = "成功"
	}

	criticalText := ""
	if result.Critical == rules.CriticalSuccess {
		criticalText = " (大成功!)"
	} else if result.Critical == rules.CriticalFailure {
		criticalText = " (大失败!)"
	}

	narrative := fmt.Sprintf("%s豁免: 投出 %d (DC %d) - %s%s", abilityName, result.Total, dc, resultText, criticalText)

	return &ToolResult{
		Success:   result.Success,
		Narrative: narrative,
		Data: map[string]interface{}{
			"ability":  abilityName,
			"roll":     result.Roll.Total,
			"modifier": result.Total - result.Roll.Total,
			"total":    result.Total,
			"dc":       dc,
			"success":  result.Success,
			"critical": result.Critical.String(),
		},
	}, nil
}
