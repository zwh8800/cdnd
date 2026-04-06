package character

// SkillType 表示 D&D 技能。
type SkillType string

const (
	// 力量技能
	Athletics SkillType = "athletics"

	// 敏捷技能
	Acrobatics    SkillType = "acrobatics"
	SleightOfHand SkillType = "sleight_of_hand"
	Stealth       SkillType = "stealth"

	// 智力技能
	Arcana        SkillType = "arcana"
	History       SkillType = "history"
	Investigation SkillType = "investigation"
	Nature        SkillType = "nature"
	Religion      SkillType = "religion"

	// 感知技能
	AnimalHandling SkillType = "animal_handling"
	Insight        SkillType = "insight"
	Medicine       SkillType = "medicine"
	Perception     SkillType = "perception"
	Survival       SkillType = "survival"

	// 魅力技能
	Deception    SkillType = "deception"
	Intimidation SkillType = "intimidation"
	Performance  SkillType = "performance"
	Persuasion   SkillType = "persuasion"
)

// AllSkillTypes 返回所有技能类型。
func AllSkillTypes() []SkillType {
	return []SkillType{
		Athletics,
		Acrobatics, SleightOfHand, Stealth,
		Arcana, History, Investigation, Nature, Religion,
		AnimalHandling, Insight, Medicine, Perception, Survival,
		Deception, Intimidation, Performance, Persuasion,
	}
}

// SkillAbility 返回与技能关联的属性。
func SkillAbility(skill SkillType) Ability {
	switch skill {
	case Athletics:
		return Strength
	case Acrobatics, SleightOfHand, Stealth:
		return Dexterity
	case Arcana, History, Investigation, Nature, Religion:
		return Intelligence
	case AnimalHandling, Insight, Medicine, Perception, Survival:
		return Wisdom
	case Deception, Intimidation, Performance, Persuasion:
		return Charisma
	default:
		return Strength
	}
}

// Skill 表示角色在某个技能上的熟练度。
type Skill struct {
	Type       SkillType `json:"type"`
	Ability    Ability   `json:"ability"`
	Proficient bool      `json:"proficient"`
	Expertise  bool      `json:"expertise"` // 双倍熟练加值
	Bonus      int       `json:"bonus"`     // 杂项加值
}

// Modifier 计算技能调整值。
func (s *Skill) Modifier(abilityMod, profBonus int) int {
	mod := abilityMod
	if s.Proficient {
		mod += profBonus
	}
	if s.Expertise {
		mod += profBonus // 双倍熟练
	}
	mod += s.Bonus
	return mod
}

// SavingThrow 表示豁免熟练度。
type SavingThrow struct {
	Ability    Ability `json:"ability"`
	Proficient bool    `json:"proficient"`
}

// Modifier 计算豁免调整值。
func (st *SavingThrow) Modifier(abilityMod, profBonus int) int {
	mod := abilityMod
	if st.Proficient {
		mod += profBonus
	}
	return mod
}

// SkillInfo 技能信息
type SkillInfo struct {
	Type        SkillType `json:"type"`
	Name        string    `json:"name"`
	Ability     Ability   `json:"ability"`
	AbilityName string    `json:"ability_name"`
}

// SkillNames 技能中文名称映射
var SkillNames = map[SkillType]string{
	Athletics:      "运动",
	Acrobatics:     "体操",
	SleightOfHand:  "手法",
	Stealth:        "隐匿",
	Arcana:         "奥秘",
	History:        "历史",
	Investigation:  "调查",
	Nature:         "自然",
	Religion:       "宗教",
	AnimalHandling: "驯兽",
	Insight:        "洞察",
	Medicine:       "医药",
	Perception:     "察觉",
	Survival:       "求生",
	Deception:      "欺瞒",
	Intimidation:   "威吓",
	Performance:    "表演",
	Persuasion:     "说服",
}

// AbilityNames 属性中文名称映射
var AbilityNames = map[Ability]string{
	Strength:     "力量",
	Dexterity:    "敏捷",
	Constitution: "体质",
	Intelligence: "智力",
	Wisdom:       "感知",
	Charisma:     "魅力",
}

// GetSkillInfo 获取技能信息
func GetSkillInfo(skill SkillType) (*SkillInfo, bool) {
	name, ok := SkillNames[skill]
	if !ok {
		return nil, false
	}
	ability := SkillAbility(skill)
	return &SkillInfo{
		Type:        skill,
		Name:        name,
		Ability:     ability,
		AbilityName: AbilityNames[ability],
	}, true
}

// GetSkillName 获取技能中文名称
func GetSkillName(skill SkillType) string {
	if name, ok := SkillNames[skill]; ok {
		return name
	}
	return string(skill)
}

// GetAbilityName 获取属性中文名称
func GetAbilityName(ability Ability) string {
	if name, ok := AbilityNames[ability]; ok {
		return name
	}
	return string(ability)
}
