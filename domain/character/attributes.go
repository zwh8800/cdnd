package character

import "math"

// Ability 表示 D&D 属性类型。
type Ability string

const (
	Strength     Ability = "strength"
	Dexterity    Ability = "dexterity"
	Constitution Ability = "constitution"
	Intelligence Ability = "intelligence"
	Wisdom       Ability = "wisdom"
	Charisma     Ability = "charisma"
)

// AllAbilities 返回所有属性类型。
func AllAbilities() []Ability {
	return []Ability{Strength, Dexterity, Constitution, Intelligence, Wisdom, Charisma}
}

// Attributes 表示六项属性值。
type Attributes struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

// DefaultAttributes 返回默认属性值（全为 10）。
func DefaultAttributes() Attributes {
	return Attributes{
		Strength:     10,
		Dexterity:    10,
		Constitution: 10,
		Intelligence: 10,
		Wisdom:       10,
		Charisma:     10,
	}
}

// Get 按名称获取属性值。
func (a *Attributes) Get(ability Ability) int {
	switch ability {
	case Strength:
		return a.Strength
	case Dexterity:
		return a.Dexterity
	case Constitution:
		return a.Constitution
	case Intelligence:
		return a.Intelligence
	case Wisdom:
		return a.Wisdom
	case Charisma:
		return a.Charisma
	default:
		return 0
	}
}

// Set 按名称设置属性值。
func (a *Attributes) Set(ability Ability, value int) {
	switch ability {
	case Strength:
		a.Strength = value
	case Dexterity:
		a.Dexterity = value
	case Constitution:
		a.Constitution = value
	case Intelligence:
		a.Intelligence = value
	case Wisdom:
		a.Wisdom = value
	case Charisma:
		a.Charisma = value
	}
}

// Modifier 返回给定属性的调整值。
// 公式：floor((属性值 - 10) / 2)
func (a *Attributes) Modifier(ability Ability) int {
	score := a.Get(ability)
	return int(math.Floor(float64(score-10) / 2))
}

// ModifierString 返回带符号的调整值字符串。
func (a *Attributes) ModifierString(ability Ability) string {
	mod := a.Modifier(ability)
	if mod >= 0 {
		return "+" + string(rune('0'+mod))
	}
	return string(rune('0' + mod))
}
