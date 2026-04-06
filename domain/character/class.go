package character

// HitDice 生命骰类型
type HitDice string

const (
	D6  HitDice = "d6"
	D8  HitDice = "d8"
	D10 HitDice = "d10"
	D12 HitDice = "d12"
)

// ClassType 职业类型枚举
type ClassType string

const (
	Barbarian ClassType = "barbarian" // 野蛮人
	Bard      ClassType = "bard"      // 吟游诗人
	Cleric    ClassType = "cleric"    // 牧师
	Druid     ClassType = "druid"     // 德鲁伊
	Fighter   ClassType = "fighter"   // 战士
	Monk      ClassType = "monk"      // 武僧
	Paladin   ClassType = "paladin"   // 圣武士
	Ranger    ClassType = "ranger"    // 游侠
	Rogue     ClassType = "rogue"     // 游荡者
	Sorcerer  ClassType = "sorcerer"  // 术士
	Warlock   ClassType = "warlock"   // 邪术师
	Wizard    ClassType = "wizard"    // 法师
)

// SubClass 子职业（又称原型/原型路径）
type SubClass struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`        // 中文名称
	Description string    `json:"description"` // 中文描述
	Level       int       `json:"level"`       // 可选择的等级
	Features    []Feature `json:"features"`    // 子职业特性
}

// SpellSlotTable 法术槽成长表（按等级）
type SpellSlotTable struct {
	// 每个等级对应的各环阶法术槽数量
	// Level1Spells, Level2Spells, etc.
	Slots map[int]SpellSlots `json:"slots"`
}

// Class 职业（官方中文翻译）
type Class struct {
	ID                  string      `json:"id"`
	Name                string      `json:"name"`                 // 中文名称
	NameEn              string      `json:"name_en"`              // 英文名称
	Type                ClassType   `json:"type"`                 // 职业类型
	Description         string      `json:"description"`          // 中文描述
	HitDice             HitDice     `json:"hit_dice"`             // 生命骰
	PrimaryAbility      []Ability   `json:"primary_ability"`      // 主要属性
	SavingThrows        []Ability   `json:"saving_throws"`        // 豁免熟练
	SkillCount          int         `json:"skill_count"`          // 可选技能数量
	SkillOptions        []SkillType `json:"skill_options"`        // 可选技能列表
	ArmorProficiencies  []string    `json:"armor_proficiencies"`  // 护甲熟练
	WeaponProficiencies []string    `json:"weapon_proficiencies"` // 武器熟练
	Tools               []string    `json:"tools"`                // 工具熟练
	Features            []Feature   `json:"features"`             // 职业特性
	Spellcasting        bool        `json:"spellcasting"`         // 是否施法者
	// 扩展字段
	SubClasses          []SubClass `json:"sub_classes,omitempty"`          // 子职业选项
	SpellcastingAbility Ability    `json:"spellcasting_ability,omitempty"` // 施法属性
	RitualCaster        bool       `json:"ritual_caster,omitempty"`        // 仪式施法
	CantripCount        int        `json:"cantrip_count,omitempty"`        // 戏法数量（1级）
}

// Feature 职业特性
type Feature struct {
	Name        string `json:"name"`        // 特性名称（中文）
	Level       int    `json:"level"`       // 获得等级
	Description string `json:"description"` // 特性描述（中文）
}

// GetHitDiceValue 返回生命骰的数值
func (hd HitDice) GetHitDiceValue() int {
	switch hd {
	case D6:
		return 6
	case D8:
		return 8
	case D10:
		return 10
	case D12:
		return 12
	default:
		return 8
	}
}

// GetSubClass 根据ID获取子职业
func (c *Class) GetSubClass(id string) *SubClass {
	for i := range c.SubClasses {
		if c.SubClasses[i].ID == id {
			return &c.SubClasses[i]
		}
	}
	return nil
}

// HasSubClasses 检查是否有子职业选项
func (c *Class) HasSubClasses() bool {
	return len(c.SubClasses) > 0
}

// GetAllClasses 获取所有职业
func GetAllClasses() []*Class {
	classes := make([]*Class, 0, len(StandardClasses))
	for _, c := range StandardClasses {
		class := c // 创建副本
		classes = append(classes, &class)
	}
	return classes
}
