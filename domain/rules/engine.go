package rules

import (
	character2 "github.com/zwh8800/cdnd/domain/character"
	dice2 "github.com/zwh8800/cdnd/domain/dice"
)

// Engine 规则引擎
type Engine struct{}

// NewEngine 创建新的规则引擎
func NewEngine() *Engine {
	return &Engine{}
}

// CheckResult 检定结果
type CheckResult struct {
	Success  bool         `json:"success"`
	Roll     dice2.Result `json:"roll"`
	Total    int          `json:"total"`
	DC       int          `json:"dc"`
	Margin   int          `json:"margin"` // 差值（正数为成功，负数为失败）
	Critical CriticalType `json:"critical"`
}

// CriticalType 豁免/攻击类型
type CriticalType int

const (
	CriticalNone    CriticalType = iota
	CriticalSuccess              // 大成功 (自然20)
	CriticalFailure              // 大失败 (自然1)
)

// String 返回暴击类型的中文名称
func (c CriticalType) String() string {
	switch c {
	case CriticalSuccess:
		return "大成功"
	case CriticalFailure:
		return "大失败"
	default:
		return ""
	}
}

// convertCritical 转换暴击类型
func convertCritical(c dice2.CriticalType) CriticalType {
	switch c {
	case dice2.CritSuccess:
		return CriticalSuccess
	case dice2.CritFail:
		return CriticalFailure
	default:
		return CriticalNone
	}
}

// AbilityCheck 属性检定
func (e *Engine) AbilityCheck(c *character2.Character, ability character2.Ability, dc int, rollType dice2.RollType) *CheckResult {
	result := &CheckResult{
		DC:       dc,
		Critical: CriticalNone,
	}

	// 投骰
	roll := dice2.RollDice(1, 20, 0, rollType)
	result.Roll = roll
	result.Critical = convertCritical(roll.Critical)

	// 检查大成功/大失败
	if roll.Critical == dice2.CritSuccess {
		modifier := c.Attributes.Modifier(ability)
		result.Total = roll.Total + modifier
		result.Success = true
	} else if roll.Critical == dice2.CritFail {
		modifier := c.Attributes.Modifier(ability)
		result.Total = roll.Total + modifier
		result.Success = false
	} else {
		// 计算总值
		modifier := c.Attributes.Modifier(ability)
		result.Total = roll.Total + modifier
		result.Margin = result.Total - dc
		result.Success = result.Total >= dc
	}

	return result
}

// SkillCheck 技能检定
func (e *Engine) SkillCheck(c *character2.Character, skill character2.SkillType, dc int, rollType dice2.RollType) *CheckResult {
	result := &CheckResult{
		DC:       dc,
		Critical: CriticalNone,
	}

	// 获取技能信息
	skillInfo, ok := character2.GetSkillInfo(skill)
	if !ok {
		result.Success = false
		return result
	}

	// 投骰
	roll := dice2.RollDice(1, 20, 0, rollType)
	result.Roll = roll
	result.Critical = convertCritical(roll.Critical)

	// 检查大成功/大失败
	if roll.Critical == dice2.CritSuccess {
		modifier := c.Attributes.Modifier(skillInfo.Ability)
		if c.HasSkillProficiency(skill) {
			modifier += c.ProficiencyBonus
		}
		result.Total = roll.Total + modifier
		result.Success = true
	} else if roll.Critical == dice2.CritFail {
		modifier := c.Attributes.Modifier(skillInfo.Ability)
		if c.HasSkillProficiency(skill) {
			modifier += c.ProficiencyBonus
		}
		result.Total = roll.Total + modifier
		result.Success = false
	} else {
		// 计算总值
		modifier := c.Attributes.Modifier(skillInfo.Ability)

		// 检查熟练加成
		if c.HasSkillProficiency(skill) {
			modifier += c.ProficiencyBonus
		}

		result.Total = roll.Total + modifier
		result.Margin = result.Total - dc
		result.Success = result.Total >= dc
	}

	return result
}

// SavingThrow 豁免检定
func (e *Engine) SavingThrow(c *character2.Character, ability character2.Ability, dc int, rollType dice2.RollType) *CheckResult {
	result := &CheckResult{
		DC:       dc,
		Critical: CriticalNone,
	}

	// 投骰
	roll := dice2.RollDice(1, 20, 0, rollType)
	result.Roll = roll
	result.Critical = convertCritical(roll.Critical)

	// 检查大成功/大失败
	if roll.Critical == dice2.CritSuccess {
		modifier := c.Attributes.Modifier(ability)
		if c.HasSavingThrowProficiency(ability) {
			modifier += c.ProficiencyBonus
		}
		result.Total = roll.Total + modifier
		result.Success = true
	} else if roll.Critical == dice2.CritFail {
		modifier := c.Attributes.Modifier(ability)
		if c.HasSavingThrowProficiency(ability) {
			modifier += c.ProficiencyBonus
		}
		result.Total = roll.Total + modifier
		result.Success = false
	} else {
		// 计算总值
		modifier := c.Attributes.Modifier(ability)

		// 检查豁免熟练
		if c.HasSavingThrowProficiency(ability) {
			modifier += c.ProficiencyBonus
		}

		result.Total = roll.Total + modifier
		result.Margin = result.Total - dc
		result.Success = result.Total >= dc
	}

	return result
}

// AttackRoll 攻击检定
func (e *Engine) AttackRoll(c *character2.Character, ability character2.Ability, ac int, rollType dice2.RollType) *CheckResult {
	result := &CheckResult{
		DC:       ac,
		Critical: CriticalNone,
	}

	// 投骰
	roll := dice2.RollDice(1, 20, 0, rollType)
	result.Roll = roll
	result.Critical = convertCritical(roll.Critical)

	// 检查暴击
	if roll.Critical == dice2.CritSuccess {
		modifier := c.Attributes.Modifier(ability)
		modifier += c.ProficiencyBonus
		result.Total = roll.Total + modifier
		result.Success = true
	} else if roll.Critical == dice2.CritFail {
		modifier := c.Attributes.Modifier(ability)
		modifier += c.ProficiencyBonus
		result.Total = roll.Total + modifier
		result.Success = false
	} else {
		// 计算总值
		modifier := c.Attributes.Modifier(ability)

		// 武器熟练加成（简化处理，假设熟练）
		modifier += c.ProficiencyBonus

		result.Total = roll.Total + modifier
		result.Margin = result.Total - ac
		result.Success = result.Total >= ac
	}

	return result
}

// RollDamage 投伤害骰
func (e *Engine) RollDamage(notation string, modifier int, critical bool) *DamageResult {
	result := &DamageResult{
		Modifier: modifier,
		Critical: critical,
	}

	// 投基础伤害
	baseRoll, err := dice2.ParseAndRoll(notation)
	if err != nil {
		result.Total = modifier
		return result
	}
	result.Base = baseRoll
	result.Total = baseRoll.Total + modifier

	// 暴击伤害（额外投一次伤害骰）
	if critical {
		extra, err := dice2.ParseAndRoll(notation)
		if err == nil {
			result.CriticalDamage = extra.Total
			result.Total += extra.Total
		}
	}

	return result
}

// DamageResult 伤害结果
type DamageResult struct {
	Base           dice2.Result `json:"base"`
	Modifier       int          `json:"modifier"`
	Critical       bool         `json:"critical"`
	CriticalDamage int          `json:"critical_damage,omitempty"`
	Total          int          `json:"total"`
}

// CalculateAC 计算AC
func (e *Engine) CalculateAC(c *character2.Character) int {
	// 基础AC = 10 + 敏捷调整值
	baseAC := 10 + c.Attributes.Modifier(character2.Dexterity)

	// 装备护甲会影响AC（简化处理）
	// TODO: 根据装备计算AC

	return baseAC
}
