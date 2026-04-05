// Package character 提供 D&D 5e 角色管理功能。
package character

import (
	"github.com/google/uuid"
)

// Character 表示 D&D 5e 玩家角色。
type Character struct {
	// 基本信息
	ID         string `json:"id"`
	Name       string `json:"name"`
	PlayerName string `json:"player_name,omitempty"`
	Race       Race   `json:"race"`
	Class      Class  `json:"class"`
	Level      int    `json:"level"`
	Background string `json:"background"`
	Alignment  string `json:"alignment"`
	Experience int    `json:"experience"`

	// 属性
	Attributes Attributes `json:"attributes"`

	// 生命值
	HitPoints HitPoints `json:"hit_points"`

	// 速度
	Speed int `json:"speed"`

	// 防御等级
	ArmorClass int `json:"armor_class"`

	// 先攻
	Initiative int `json:"initiative"`

	// 熟练加值
	ProficiencyBonus int `json:"proficiency_bonus"`

	// 技能
	Skills map[SkillType]Skill `json:"skills"`

	// 豁免
	SavingThrows map[Ability]SavingThrow `json:"saving_throws"`

	// 装备与物品
	Equipment []Item `json:"equipment"`
	Inventory []Item `json:"inventory"`
	Gold      int    `json:"gold"`

	// 特性与熟练
	Features      []Feature     `json:"features"`
	Proficiencies []Proficiency `json:"proficiencies"`

	// 法术（如果是施法者）
	Spells              []Spell    `json:"spells,omitempty"`
	SpellSlots          SpellSlots `json:"spell_slots,omitempty"`
	SpellcastingAbility Ability    `json:"spellcasting_ability,omitempty"`

	// 状态效果
	Conditions []string `json:"conditions,omitempty"`
}

// NewCharacter 创建带有默认值的新角色。
func NewCharacter(name string, race Race, class Class) *Character {
	c := &Character{
		ID:               uuid.New().String(),
		Name:             name,
		Race:             race,
		Class:            class,
		Level:            1,
		Attributes:       DefaultAttributes(),
		ProficiencyBonus: 2,
		Skills:           make(map[SkillType]Skill),
		SavingThrows:     make(map[Ability]SavingThrow),
		Equipment:        []Item{},
		Inventory:        []Item{},
		Features:         []Feature{},
		Proficiencies:    []Proficiency{},
	}

	// 初始化技能
	for _, skillType := range AllSkillTypes() {
		c.Skills[skillType] = Skill{
			Type:       skillType,
			Ability:    SkillAbility(skillType),
			Proficient: false,
			Bonus:      0,
		}
	}

	// 初始化豁免
	for _, ability := range AllAbilities() {
		c.SavingThrows[ability] = SavingThrow{
			Ability:    ability,
			Proficient: false,
		}
	}

	return c
}

// HitPoints 表示角色的生命值。
type HitPoints struct {
	Current int `json:"current"`
	Max     int `json:"max"`
	Temp    int `json:"temp"`
}

// TakeDamage 减少指定量的生命值。
func (hp *HitPoints) TakeDamage(damage int) {
	// 先扣除临时生命值
	if hp.Temp > 0 {
		if damage <= hp.Temp {
			hp.Temp -= damage
			return
		}
		damage -= hp.Temp
		hp.Temp = 0
	}

	hp.Current -= damage
	if hp.Current < 0 {
		hp.Current = 0
	}
}

// Heal 恢复生命值。
func (hp *HitPoints) Heal(amount int) {
	hp.Current += amount
	if hp.Current > hp.Max {
		hp.Current = hp.Max
	}
}

// HasSkillProficiency 检查是否有技能熟练
func (c *Character) HasSkillProficiency(skill SkillType) bool {
	if s, ok := c.Skills[skill]; ok {
		return s.Proficient
	}
	return false
}

// HasSavingThrowProficiency 检查是否有豁免熟练
func (c *Character) HasSavingThrowProficiency(ability Ability) bool {
	if st, ok := c.SavingThrows[ability]; ok {
		return st.Proficient
	}
	return false
}

// GetSkillModifier 获取技能调整值
func (c *Character) GetSkillModifier(skill SkillType) int {
	abilityMod := c.Attributes.Modifier(SkillAbility(skill))
	if s, ok := c.Skills[skill]; ok {
		return s.Modifier(abilityMod, c.ProficiencyBonus)
	}
	return abilityMod
}

// GetSavingThrowModifier 获取豁免调整值
func (c *Character) GetSavingThrowModifier(ability Ability) int {
	abilityMod := c.Attributes.Modifier(ability)
	if st, ok := c.SavingThrows[ability]; ok {
		return st.Modifier(abilityMod, c.ProficiencyBonus)
	}
	return abilityMod
}

// SetSkillProficiency 设置技能熟练
func (c *Character) SetSkillProficiency(skill SkillType, proficient bool) {
	if s, ok := c.Skills[skill]; ok {
		s.Proficient = proficient
		c.Skills[skill] = s
	}
}

// SetSavingThrowProficiency 设置豁免熟练
func (c *Character) SetSavingThrowProficiency(ability Ability, proficient bool) {
	if st, ok := c.SavingThrows[ability]; ok {
		st.Proficient = proficient
		c.SavingThrows[ability] = st
	}
}

// HasClass 检查是否有职业
func (c *Character) HasClass() bool {
	return c.Class.ID != ""
}

// HasCondition 检查是否有指定状态效果
func (c *Character) HasCondition(condition string) bool {
	for _, cond := range c.Conditions {
		if cond == condition {
			return true
		}
	}
	return false
}

// AddCondition 添加状态效果
func (c *Character) AddCondition(condition string) {
	if !c.HasCondition(condition) {
		c.Conditions = append(c.Conditions, condition)
	}
}

// RemoveCondition 移除状态效果
func (c *Character) RemoveCondition(condition string) {
	for i, cond := range c.Conditions {
		if cond == condition {
			c.Conditions = append(c.Conditions[:i], c.Conditions[i+1:]...)
			return
		}
	}
}

// GetConditions 获取状态效果副本
func (c *Character) GetConditions() []string {
	result := make([]string, len(c.Conditions))
	copy(result, c.Conditions)
	return result
}
