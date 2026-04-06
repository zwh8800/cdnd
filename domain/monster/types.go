package monster

import (
	"github.com/zwh8800/cdnd/domain/character"
)

// Size 体型
type Size string

const (
	SizeTiny       Size = "微型"
	SizeSmall      Size = "小型"
	SizeMedium     Size = "中型"
	SizeLarge      Size = "大型"
	SizeHuge       Size = "巨型"
	SizeGargantuan Size = "超巨型"
)

// Type 怪物类型
type Type string

const (
	TypeAberration  Type = "异怪"
	TypeBeast       Type = "野兽"
	TypeCelestial   Type = "天界生物"
	TypeConstruct   Type = "构装生物"
	TypeDragon      Type = "龙"
	TypeElemental   Type = "元素生物"
	TypeFey         Type = "精类"
	TypeFiend       Type = "邪魔"
	TypeGiant       Type = "巨人"
	TypeHumanoid    Type = "人形生物"
	TypeMonstrosity Type = "怪兽"
	TypeOoze        Type = "泥怪"
	TypePlant       Type = "植物"
	TypeUndead      Type = "亡灵"
)

// Alignment 阵营
type Alignment string

const (
	AlignmentLawfulGood     Alignment = "守序善良"
	AlignmentNeutralGood    Alignment = "中立善良"
	AlignmentChaoticGood    Alignment = "混乱善良"
	AlignmentLawfulNeutral  Alignment = "守序中立"
	AlignmentTrueNeutral    Alignment = "绝对中立"
	AlignmentChaoticNeutral Alignment = "混乱中立"
	AlignmentLawfulEvil     Alignment = "守序邪恶"
	AlignmentNeutralEvil    Alignment = "中立邪恶"
	AlignmentChaoticEvil    Alignment = "混乱邪恶"
	AlignmentUnaligned      Alignment = "无阵营"
)

// ActionType 动作类型
type ActionType string

const (
	ActionTypeMelee   ActionType = "近战"
	ActionTypeRanged  ActionType = "远程"
	ActionTypeSpell   ActionType = "法术"
	ActionTypeSpecial ActionType = "特殊"
)

// MonsterAction 怪物动作
type MonsterAction struct {
	Name        string     `json:"name"`         // 动作名称
	Type        ActionType `json:"type"`         // 动作类型
	AttackBonus int        `json:"attack_bonus"` // 攻击加值（-1表示无需攻击检定）
	Damage      string     `json:"damage"`       // 伤害表达式（如 "1d6+2"）
	DamageType  string     `json:"damage_type"`  // 伤害类型
	Range       string     `json:"range"`        // 射程（如 "5尺" 或 "30/120尺"）
	Description string     `json:"description"`  // 动作描述
	SaveDC      int        `json:"save_dc"`      // 豁免DC（如需）
	SaveAbility string     `json:"save_ability"` // 豁免属性（如需）
}

// MonsterTemplate 怪物模板
type MonsterTemplate struct {
	ID        string    `json:"id"`        // 唯一标识
	Name      string    `json:"name"`      // 显示名称
	Size      Size      `json:"size"`      // 体型
	Type      Type      `json:"type"`      // 类型
	Alignment Alignment `json:"alignment"` // 阵营
	CR        float64   `json:"cr"`        // 挑战等级
	XP        int       `json:"xp"`        // 经验值
	HP        string    `json:"hp"`        // 生命值骰表达式（如 "2d6+6"）
	AC        int       `json:"ac"`        // 护甲等级
	Speed     int       `json:"speed"`     // 移动速度（尺）

	// 属性
	Abilities character.Attributes `json:"abilities"`

	// 豁免加值（可选，默认使用属性调整值）
	SavingThrows map[string]int `json:"saving_throws,omitempty"`

	// 技能加值（可选）
	Skills map[string]int `json:"skills,omitempty"`

	// 伤害抗性
	DamageResistances []string `json:"damage_resistances,omitempty"`

	// 伤害免疫
	DamageImmunities []string `json:"damage_immunities,omitempty"`

	// 状态免疫
	ConditionImmunities []string `json:"condition_immunities,omitempty"`

	// 感官
	Senses []string `json:"senses,omitempty"`

	// 语言
	Languages []string `json:"languages,omitempty"`

	// 动作
	Actions []MonsterAction `json:"actions"`

	// 反应动作（可选）
	Reactions []MonsterAction `json:"reactions,omitempty"`

	// 传奇动作（可选，用于Boss）
	LegendaryActions []MonsterAction `json:"legendary_actions,omitempty"`

	// 描述
	Description string `json:"description"`

	// DM隐藏信息
	DMNotes string `json:"dm_notes,omitempty"`
}

// GetHPExpression 获取生命值骰表达式
func (m *MonsterTemplate) GetHPExpression() string {
	return m.HP
}

// GetXPByCR 根据CR获取经验值（标准D&D 5e表格）
func GetXPByCR(cr float64) int {
	switch cr {
	case 0:
		return 10
	case 0.125:
		return 25
	case 0.25:
		return 50
	case 0.5:
		return 100
	case 1:
		return 200
	case 2:
		return 450
	case 3:
		return 700
	case 4:
		return 1100
	case 5:
		return 1800
	case 6:
		return 2300
	case 7:
		return 2900
	case 8:
		return 3900
	case 9:
		return 5000
	case 10:
		return 5900
	default:
		if cr < 0.125 {
			return 10
		}
		return int(cr) * 1000
	}
}
