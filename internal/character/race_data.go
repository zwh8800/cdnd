package character

// StandardRaces 标准 D&D 5e 种族数据（参考《玩家手册》）
// 所有术语使用官方中文翻译
var StandardRaces = map[string]Race{
	// ============== 人类 ==============
	"human": {
		ID:          "human",
		Name:        "人类",
		NameEn:      "Human",
		Description: "人类是所有种族中最多才多艺、适应性最强的。他们的足迹遍布世界各地，拥有多样的文化和道德观念。",
		Size:        SizeMedium,
		Speed:       30,
		AbilityBonuses: map[Ability]int{
			Strength:     1,
			Dexterity:    1,
			Constitution: 1,
			Intelligence: 1,
			Wisdom:       1,
			Charisma:     1,
		},
		Languages: []string{"通用语"},
		Traits: []Trait{
			{Name: "多才多艺", Description: "所有属性值+1"},
		},
		AgeRange:    AgeRange{Adulthood: 18, MaxAge: 100},
		HeightRange: HeightRange{BaseHeight: 56, ModDice: 10, ModCount: 2},
		WeightRange: WeightRange{BaseWeight: 110, ModDice: 4, ModCount: 2},
	},

	// ============== 精灵 ==============
	"elf": {
		ID:          "elf",
		Name:        "精灵",
		NameEn:      "Elf",
		Description: "精灵是拥有超凡脱俗优雅气质的魔法种族。他们长寿且美丽，以诗歌、舞蹈和歌谣闻名。",
		Size:        SizeMedium,
		Speed:       30,
		AbilityBonuses: map[Ability]int{
			Dexterity: 2,
		},
		Languages: []string{"通用语", "精灵语"},
		Traits: []Trait{
			{Name: "黑暗视觉", Description: "在60尺范围内，昏暗中视物如白昼"},
			{Name: "敏锐感官", Description: "察觉技能熟练"},
			{Name: "精灵血脉", Description: "魅惑豁免优势，免疫魔法睡眠"},
			{Name: "恍惚之眠", Description: "不需要睡眠，4小时冥想替代8小时休息"},
		},
		AgeRange:    AgeRange{Adulthood: 100, MaxAge: 750},
		HeightRange: HeightRange{BaseHeight: 54, ModDice: 10, ModCount: 2},
		WeightRange: WeightRange{BaseWeight: 90, ModDice: 4, ModCount: 1},
		SubRaces: []SubRace{
			{
				ID:          "high_elf",
				Name:        "高等精灵",
				Description: "高等精灵是最常见的精灵，以学识和魔法天赋著称。",
				AbilityBonuses: map[Ability]int{
					Intelligence: 1,
				},
				Traits: []Trait{
					{Name: "精灵武器训练", Description: "精通长剑、短剑、短弓、长弓"},
					{Name: "戏法", Description: "选择一个法师戏法"},
					{Name: "额外语言", Description: "学会一门自选语言"},
				},
			},
			{
				ID:          "wood_elf",
				Name:        "木精灵",
				Description: "木精灵远离文明，生活在森林中，以潜行和狩猎技能著称。",
				AbilityBonuses: map[Ability]int{
					Wisdom: 1,
				},
				Traits: []Trait{
					{Name: "精灵武器训练", Description: "精通长剑、短剑、短弓、长弓"},
					{Name: "面具荒野", Description: "可尝试隐藏在自然环境中"},
					{Name: "迅捷", Description: "基础速度35尺"},
				},
			},
			{
				ID:          "drow",
				Name:        "卓尔精灵",
				Description: "卓尔精灵是居住在地下的黑暗精灵，以邪恶著称。",
				AbilityBonuses: map[Ability]int{
					Charisma: 1,
				},
				Traits: []Trait{
					{Name: "卓越黑暗视觉", Description: "在120尺范围内黑暗中视物"},
					{Name: "阳光敏感", Description: "阳光下攻击检定和察觉检定劣势"},
					{Name: "卓尔魔法", Description: "获得舞光术戏法，3级可施展妖火术，5级可施展黑暗术"},
					{Name: "卓尔武器训练", Description: "精通细剑、短剑、手弩"},
				},
			},
		},
	},

	// ============== 矮人 ==============
	"dwarf": {
		ID:          "dwarf",
		Name:        "矮人",
		NameEn:      "Dwarf",
		Description: "矮人身材矮小但体格健壮，以勇气和耐力著称。他们是技艺精湛的工匠。",
		Size:        SizeMedium,
		Speed:       25,
		AbilityBonuses: map[Ability]int{
			Constitution: 2,
		},
		Languages: []string{"通用语", "矮人语"},
		Traits: []Trait{
			{Name: "黑暗视觉", Description: "在60尺范围内，昏暗中视物如白昼"},
			{Name: "矮人韧性", Description: "毒素豁免优势，毒素伤害抗性"},
			{Name: "石匠技艺", Description: "与石造物相关的历史检定双倍熟练加值"},
			{Name: "矮人战斗训练", Description: "精通战斧、手斧、轻锤、战锤"},
		},
		AgeRange:    AgeRange{Adulthood: 50, MaxAge: 400},
		HeightRange: HeightRange{BaseHeight: 44, ModDice: 4, ModCount: 2},
		WeightRange: WeightRange{BaseWeight: 130, ModDice: 6, ModCount: 2},
		SubRaces: []SubRace{
			{
				ID:          "hill_dwarf",
				Name:        "山丘矮人",
				Description: "山丘矮人拥有敏锐的感官和深厚的智慧。",
				AbilityBonuses: map[Ability]int{
					Wisdom: 1,
				},
				Traits: []Trait{
					{Name: "矮人结实", Description: "生命值上限+1，升级时再+1"},
				},
			},
			{
				ID:          "mountain_dwarf",
				Name:        "山地矮人",
				Description: "山地矮人强壮且高大，以铠甲技艺著称。",
				AbilityBonuses: map[Ability]int{
					Strength: 2,
				},
				Traits: []Trait{
					{Name: "矮人护甲训练", Description: "精通轻甲、中甲、重甲"},
				},
			},
		},
	},

	// ============== 半身人 ==============
	"halfling": {
		ID:          "halfling",
		Name:        "半身人",
		NameEn:      "Halfling",
		Description: "半身人身材矮小，性格温和，以幸运和灵活著称。他们喜欢平静的生活。",
		Size:        SizeSmall,
		Speed:       25,
		AbilityBonuses: map[Ability]int{
			Dexterity: 2,
		},
		Languages: []string{"通用语", "半身人语"},
		Traits: []Trait{
			{Name: "幸运", Description: "攻击检定、属性检定、豁免检定骰出1时可以重掷"},
			{Name: "勇敢", Description: "恐惧豁免优势"},
			{Name: "半身人灵活", Description: "可以穿越大型生物占据的空间"},
		},
		AgeRange:    AgeRange{Adulthood: 20, MaxAge: 150},
		HeightRange: HeightRange{BaseHeight: 31, ModDice: 4, ModCount: 2},
		WeightRange: WeightRange{BaseWeight: 35, ModDice: 4, ModCount: 1},
		SubRaces: []SubRace{
			{
				ID:          "lightfoot_halfling",
				Name:        "轻足半身人",
				Description: "轻足半身人善于隐藏，经常与其他种族混居。",
				AbilityBonuses: map[Ability]int{
					Charisma: 1,
				},
				Traits: []Trait{
					{Name: "天生匿踪", Description: "中型或更大体型生物后可以尝试隐藏"},
				},
			},
			{
				ID:          "stout_halfling",
				Name:        "壮实半身人",
				Description: "壮实半身人像矮人一样健壮，适应各种环境。",
				AbilityBonuses: map[Ability]int{
					Constitution: 1,
				},
				Traits: []Trait{
					{Name: "壮实", Description: "毒素豁免优势，毒素伤害抗性"},
				},
			},
		},
	},

	// ============== 龙裔 ==============
	"dragonborn": {
		ID:          "dragonborn",
		Name:        "龙裔",
		NameEn:      "Dragonborn",
		Description: "龙裔是龙族的后裔，拥有龙类血统。他们骄傲而威严，以龙息和龙族血统著称。",
		Size:        SizeMedium,
		Speed:       30,
		AbilityBonuses: map[Ability]int{
			Strength: 2,
			Charisma: 1,
		},
		Languages: []string{"通用语", "龙语"},
		Traits: []Trait{
			{Name: "龙息", Description: "可以喷出5x30尺锥状龙息，造成2d6伤害（体质豁免DC减半），伤害类型取决于龙族血统"},
			{Name: "伤害抗性", Description: "对龙族血统对应的伤害类型具有抗性"},
		},
		AgeRange:    AgeRange{Adulthood: 15, MaxAge: 80},
		HeightRange: HeightRange{BaseHeight: 66, ModDice: 8, ModCount: 2},
		WeightRange: WeightRange{BaseWeight: 175, ModDice: 6, ModCount: 2},
	},

	// ============== 侏儒 ==============
	"gnome": {
		ID:          "gnome",
		Name:        "侏儒",
		NameEn:      "Gnome",
		Description: "侏儒身材矮小，以聪明才智和工程技艺著称。他们充满好奇心和幽默感。",
		Size:        SizeSmall,
		Speed:       25,
		AbilityBonuses: map[Ability]int{
			Intelligence: 2,
		},
		Languages: []string{"通用语", "侏儒语"},
		Traits: []Trait{
			{Name: "黑暗视觉", Description: "在60尺范围内，昏暗中视物如白昼"},
			{Name: "侏儒智慧", Description: "对魔法进行智力、感知、魅力豁免时具有优势"},
		},
		AgeRange:    AgeRange{Adulthood: 40, MaxAge: 500},
		HeightRange: HeightRange{BaseHeight: 35, ModDice: 4, ModCount: 2},
		WeightRange: WeightRange{BaseWeight: 35, ModDice: 4, ModCount: 1},
		SubRaces: []SubRace{
			{
				ID:          "forest_gnome",
				Name:        "森林侏儒",
				Description: "森林侏儒天生具有幻术天赋，与自然生物友好相处。",
				AbilityBonuses: map[Ability]int{
					Dexterity: 1,
				},
				Traits: []Trait{
					{Name: "自然幻术", Description: "可以施展次级幻影戏法"},
					{Name: "与兽语", Description: "可以与小型的野兽简单交流"},
				},
			},
			{
				ID:          "rock_gnome",
				Name:        "岩石侏儒",
				Description: "岩石侏儒以工匠技艺和发明创造著称。",
				AbilityBonuses: map[Ability]int{
					Constitution: 1,
				},
				Traits: []Trait{
					{Name: "工匠知识", Description: "使用工匠工具时可以快速制作简单物品"},
					{Name: "修补匠", Description: "可以使用工匠工具制作微小的机械装置"},
				},
			},
		},
	},

	// ============== 半精灵 ==============
	"half_elf": {
		ID:          "half_elf",
		Name:        "半精灵",
		NameEn:      "Half-Elf",
		Description: "半精灵继承了人类和精灵双方的优点。他们多才多艺，魅力出众，但常在两种文化间感到格格不入。",
		Size:        SizeMedium,
		Speed:       30,
		AbilityBonuses: map[Ability]int{
			Charisma: 2,
		},
		Languages: []string{"通用语", "精灵语"},
		Traits: []Trait{
			{Name: "两项属性提升", Description: "选择两项属性各+1"},
			{Name: "精灵血脉", Description: "魅惑豁免优势，免疫魔法睡眠"},
			{Name: "多才多艺", Description: "选择两项自选技能熟练"},
		},
		AgeRange:    AgeRange{Adulthood: 20, MaxAge: 180},
		HeightRange: HeightRange{BaseHeight: 57, ModDice: 8, ModCount: 2},
		WeightRange: WeightRange{BaseWeight: 110, ModDice: 4, ModCount: 2},
	},

	// ============== 半兽人 ==============
	"half_orc": {
		ID:          "half_orc",
		Name:        "半兽人",
		NameEn:      "Half-Orc",
		Description: "半兽人拥有兽人的力量和人类的智慧。他们通常身体强壮，勇敢无畏。",
		Size:        SizeMedium,
		Speed:       30,
		AbilityBonuses: map[Ability]int{
			Strength:     2,
			Constitution: 1,
		},
		Languages: []string{"通用语", "兽人语"},
		Traits: []Trait{
			{Name: "黑暗视觉", Description: "在60尺范围内，昏暗中视物如白昼"},
			{Name: "野蛮攻击", Description: "重击时额外增加一个武器伤害骰"},
			{Name: "顽强", Description: "生命值降为0时可以恢复1点生命值，长休后恢复此能力"},
			{Name: "威吓熟练", Description: "威吓技能熟练"},
		},
		AgeRange:    AgeRange{Adulthood: 14, MaxAge: 75},
		HeightRange: HeightRange{BaseHeight: 58, ModDice: 10, ModCount: 2},
		WeightRange: WeightRange{BaseWeight: 150, ModDice: 8, ModCount: 2},
	},

	// ============== 提夫林 ==============
	"tiefling": {
		ID:          "tiefling",
		Name:        "提夫林",
		NameEn:      "Tiefling",
		Description: "提夫林拥有恶魔血统，外表带有角和尾巴。他们常被误解，但内心并非邪恶。",
		Size:        SizeMedium,
		Speed:       30,
		AbilityBonuses: map[Ability]int{
			Intelligence: 1,
			Charisma:     2,
		},
		Languages: []string{"通用语", "深渊语"},
		Traits: []Trait{
			{Name: "黑暗视觉", Description: "在60尺范围内，昏暗中视物如白昼"},
			{Name: "地狱抗性", Description: "火焰伤害抗性"},
			{Name: "地狱传承", Description: "1级可施展奇术戏法，3级可施展地狱斥责术，5级可施展黑暗术"},
		},
		AgeRange:    AgeRange{Adulthood: 18, MaxAge: 100},
		HeightRange: HeightRange{BaseHeight: 57, ModDice: 8, ModCount: 2},
		WeightRange: WeightRange{BaseWeight: 110, ModDice: 4, ModCount: 2},
	},
}

// GetRace 根据 ID 获取种族
func GetRace(id string) *Race {
	if race, ok := StandardRaces[id]; ok {
		return &race
	}
	return nil
}

// AllRaces 返回所有种族列表
func AllRaces() []Race {
	races := make([]Race, 0, len(StandardRaces))
	for _, race := range StandardRaces {
		races = append(races, race)
	}
	return races
}

// DragonAncestry 龙族血统类型（用于龙裔）
type DragonAncestry struct {
	Name        string      `json:"name"`         // 名称
	DamageType  string      `json:"damage_type"`  // 吐息伤害类型
	BreathShape BreathShape `json:"breath_shape"` // 吐息形状
}

// BreathShape 吐息形状
type BreathShape string

const (
	BreathCone BreathShape = "锥状" // 5x30尺锥状
	BreathLine BreathShape = "线状" // 5x30尺线状
)

// DragonAncestries 龙族血统选项
var DragonAncestries = map[string]DragonAncestry{
	"black":  {Name: "黑龙", DamageType: "强酸", BreathShape: BreathCone},
	"blue":   {Name: "蓝龙", DamageType: "闪电", BreathShape: BreathLine},
	"brass":  {Name: "黄铜龙", DamageType: "火焰", BreathShape: BreathLine},
	"bronze": {Name: "青铜龙", DamageType: "闪电", BreathShape: BreathLine},
	"copper": {Name: "赤铜龙", DamageType: "强酸", BreathShape: BreathLine},
	"gold":   {Name: "金龙", DamageType: "火焰", BreathShape: BreathCone},
	"green":  {Name: "绿龙", DamageType: "毒素", BreathShape: BreathCone},
	"red":    {Name: "红龙", DamageType: "火焰", BreathShape: BreathCone},
	"silver": {Name: "银龙", DamageType: "冷冻", BreathShape: BreathCone},
	"white":  {Name: "白龙", DamageType: "冷冻", BreathShape: BreathCone},
}
