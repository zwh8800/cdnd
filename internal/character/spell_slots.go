package character

// SpellSlots 表示某个等级可用的法术槽数量
type SpellSlots struct {
	// 1-9环法术槽
	Level1 int `json:"level_1"` // 1环法术槽
	Level2 int `json:"level_2"` // 2环法术槽
	Level3 int `json:"level_3"` // 3环法术槽
	Level4 int `json:"level_4"` // 4环法术槽
	Level5 int `json:"level_5"` // 5环法术槽
	Level6 int `json:"level_6"` // 6环法术槽
	Level7 int `json:"level_7"` // 7环法术槽
	Level8 int `json:"level_8"` // 8环法术槽
	Level9 int `json:"level_9"` // 9环法术槽
}

// SpellcastingType 施法类型
type SpellcastingType int

const (
	SpellcastingNone  SpellcastingType = iota // 非施法者
	SpellcastingFull                          // 全施法者（法师、牧师、德鲁伊、吟游诗人、术士）
	SpellcastingHalf                          // 半施法者（圣武士、游侠）
	SpellcastingPact                          // 契约魔法（邪术师）
	SpellcastingThird                         // 三分之一的施法者（奥法骑士、诡术师）
)

// GetSlotsByLevel 获取指定环阶的法术槽数量
func (s SpellSlots) GetSlotsByLevel(spellLevel int) int {
	switch spellLevel {
	case 1:
		return s.Level1
	case 2:
		return s.Level2
	case 3:
		return s.Level3
	case 4:
		return s.Level4
	case 5:
		return s.Level5
	case 6:
		return s.Level6
	case 7:
		return s.Level7
	case 8:
		return s.Level8
	case 9:
		return s.Level9
	default:
		return 0
	}
}

// SetSlotsByLevel 设置指定环阶的法术槽数量
func (s *SpellSlots) SetSlotsByLevel(spellLevel int, count int) {
	switch spellLevel {
	case 1:
		s.Level1 = count
	case 2:
		s.Level2 = count
	case 3:
		s.Level3 = count
	case 4:
		s.Level4 = count
	case 5:
		s.Level5 = count
	case 6:
		s.Level6 = count
	case 7:
		s.Level7 = count
	case 8:
		s.Level8 = count
	case 9:
		s.Level9 = count
	}
}

// Total 返回总法术槽数量
func (s SpellSlots) Total() int {
	return s.Level1 + s.Level2 + s.Level3 + s.Level4 + s.Level5 +
		s.Level6 + s.Level7 + s.Level8 + s.Level9
}

// IsEmpty 检查是否有任何法术槽
func (s SpellSlots) IsEmpty() bool {
	return s.Total() == 0
}

// 全施法者法术槽成长表（法师、牧师、德鲁伊、吟游诗人、术士）
// 按照 D&D 5e 玩家手册
var fullCasterSlots = []SpellSlots{
	{Level1: 2, Level2: 0, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0},
	{Level1: 3, Level2: 0, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0},
	{Level1: 4, Level2: 2, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0},
	{Level1: 4, Level2: 3, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0},
	{Level1: 4, Level2: 3, Level3: 2, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0},
	{Level1: 4, Level2: 3, Level3: 3, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0},
	{Level1: 4, Level2: 3, Level3: 3, Level4: 1, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0},
	{Level1: 4, Level2: 3, Level3: 3, Level4: 2, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0},
	{Level1: 4, Level2: 3, Level3: 3, Level4: 3, Level5: 1, Level6: 0, Level7: 0, Level8: 0, Level9: 0},
	{Level1: 4, Level2: 3, Level3: 3, Level4: 3, Level5: 2, Level6: 0, Level7: 0, Level8: 0, Level9: 0},
	{Level1: 4, Level2: 3, Level3: 3, Level4: 3, Level5: 2, Level6: 1, Level7: 0, Level8: 0, Level9: 0},
	{Level1: 4, Level2: 3, Level3: 3, Level4: 3, Level5: 2, Level6: 1, Level7: 0, Level8: 0, Level9: 0},
	{Level1: 4, Level2: 3, Level3: 3, Level4: 3, Level5: 2, Level6: 1, Level7: 1, Level8: 0, Level9: 0},
	{Level1: 4, Level2: 3, Level3: 3, Level4: 3, Level5: 2, Level6: 1, Level7: 1, Level8: 0, Level9: 0},
	{Level1: 4, Level2: 3, Level3: 3, Level4: 3, Level5: 2, Level6: 1, Level7: 1, Level8: 1, Level9: 0},
	{Level1: 4, Level2: 3, Level3: 3, Level4: 3, Level5: 2, Level6: 1, Level7: 1, Level8: 1, Level9: 0},
	{Level1: 4, Level2: 3, Level3: 3, Level4: 3, Level5: 2, Level6: 1, Level7: 1, Level8: 1, Level9: 1},
	{Level1: 4, Level2: 3, Level3: 3, Level4: 3, Level5: 3, Level6: 1, Level7: 1, Level8: 1, Level9: 1},
	{Level1: 4, Level2: 3, Level3: 3, Level4: 3, Level5: 3, Level6: 2, Level7: 1, Level8: 1, Level9: 1},
	{Level1: 4, Level2: 3, Level3: 3, Level4: 3, Level5: 3, Level6: 2, Level7: 2, Level8: 1, Level9: 1},
}

// 半施法者法术槽成长表（圣武士、游侠）
// 从2级开始获得法术槽，使用等级/2向下取整作为有效施法者等级
var halfCasterSlots = []SpellSlots{
	{Level1: 0, Level2: 0, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 1级：无
	{Level1: 2, Level2: 0, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 2级：有效施法者等级1
	{Level1: 3, Level2: 0, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 3级：有效施法者等级1
	{Level1: 3, Level2: 0, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 4级：有效施法者等级2
	{Level1: 4, Level2: 2, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 5级：有效施法者等级2
	{Level1: 4, Level2: 2, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 6级：有效施法者等级3
	{Level1: 4, Level2: 3, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 7级：有效施法者等级3
	{Level1: 4, Level2: 3, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 8级：有效施法者等级4
	{Level1: 4, Level2: 3, Level3: 2, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 9级：有效施法者等级4
	{Level1: 4, Level2: 3, Level3: 2, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 10级：有效施法者等级5
	{Level1: 4, Level2: 3, Level3: 3, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 11级：有效施法者等级5
	{Level1: 4, Level2: 3, Level3: 3, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 12级：有效施法者等级6
	{Level1: 4, Level2: 3, Level3: 3, Level4: 1, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 13级：有效施法者等级6
	{Level1: 4, Level2: 3, Level3: 3, Level4: 1, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 14级：有效施法者等级7
	{Level1: 4, Level2: 3, Level3: 3, Level4: 2, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 15级：有效施法者等级7
	{Level1: 4, Level2: 3, Level3: 3, Level4: 2, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 16级：有效施法者等级8
	{Level1: 4, Level2: 3, Level3: 3, Level4: 3, Level5: 1, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 17级：有效施法者等级8
	{Level1: 4, Level2: 3, Level3: 3, Level4: 3, Level5: 1, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 18级：有效施法者等级9
	{Level1: 4, Level2: 3, Level3: 3, Level4: 3, Level5: 2, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 19级：有效施法者等级9
	{Level1: 4, Level2: 3, Level3: 3, Level4: 3, Level5: 2, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 20级：有效施法者等级10
}

// 三分之一施法者法术槽成长表（奥法骑士、诡术师）
// 使用等级/3向下取整作为有效施法者等级
var thirdCasterSlots = []SpellSlots{
	{Level1: 0, Level2: 0, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0},
	{Level1: 0, Level2: 0, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0},
	{Level1: 2, Level2: 0, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 3级：有效施法者等级1
	{Level1: 2, Level2: 0, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 4级：有效施法者等级1
	{Level1: 2, Level2: 0, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 5级：有效施法者等级1
	{Level1: 3, Level2: 0, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 6级：有效施法者等级2
	{Level1: 3, Level2: 0, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 7级：有效施法者等级2
	{Level1: 3, Level2: 0, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 8级：有效施法者等级2
	{Level1: 4, Level2: 2, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 9级：有效施法者等级3
	{Level1: 4, Level2: 2, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 10级：有效施法者等级3
	{Level1: 4, Level2: 2, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 11级：有效施法者等级3
	{Level1: 4, Level2: 3, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 12级：有效施法者等级4
	{Level1: 4, Level2: 3, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 13级：有效施法者等级4
	{Level1: 4, Level2: 3, Level3: 0, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 14级：有效施法者等级4
	{Level1: 4, Level2: 3, Level3: 2, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 15级：有效施法者等级5
	{Level1: 4, Level2: 3, Level3: 2, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 16级：有效施法者等级5
	{Level1: 4, Level2: 3, Level3: 2, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 17级：有效施法者等级5
	{Level1: 4, Level2: 3, Level3: 3, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 18级：有效施法者等级6
	{Level1: 4, Level2: 3, Level3: 3, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 19级：有效施法者等级6
	{Level1: 4, Level2: 3, Level3: 3, Level4: 0, Level5: 0, Level6: 0, Level7: 0, Level8: 0, Level9: 0}, // 20级：有效施法者等级6
}

// PactSlots 邪术师契约魔法法术槽
// 邪术师使用独特的 Pact Magic 系统，所有法术槽为同一环阶
type PactSlots struct {
	Level       int `json:"level"`       // 角色等级
	SlotLevel   int `json:"slot_level"`  // 法术槽环阶
	SlotCount   int `json:"slot_count"`  // 法术槽数量
	Invocations int `json:"invocations"` // 魔能祈唤数量
}

// 邪术师契约魔法法术槽成长表
var warlockPactSlots = []PactSlots{
	{Level: 1, SlotLevel: 1, SlotCount: 1, Invocations: 0},
	{Level: 2, SlotLevel: 1, SlotCount: 2, Invocations: 2},
	{Level: 3, SlotLevel: 2, SlotCount: 2, Invocations: 2},
	{Level: 4, SlotLevel: 2, SlotCount: 2, Invocations: 2},
	{Level: 5, SlotLevel: 3, SlotCount: 2, Invocations: 2},
	{Level: 6, SlotLevel: 3, SlotCount: 2, Invocations: 3},
	{Level: 7, SlotLevel: 4, SlotCount: 2, Invocations: 3},
	{Level: 8, SlotLevel: 4, SlotCount: 2, Invocations: 4},
	{Level: 9, SlotLevel: 5, SlotCount: 2, Invocations: 4},
	{Level: 10, SlotLevel: 5, SlotCount: 2, Invocations: 5},
	{Level: 11, SlotLevel: 5, SlotCount: 3, Invocations: 5},
	{Level: 12, SlotLevel: 5, SlotCount: 3, Invocations: 6},
	{Level: 13, SlotLevel: 5, SlotCount: 3, Invocations: 6},
	{Level: 14, SlotLevel: 5, SlotCount: 3, Invocations: 6},
	{Level: 15, SlotLevel: 5, SlotCount: 3, Invocations: 7},
	{Level: 16, SlotLevel: 5, SlotCount: 3, Invocations: 7},
	{Level: 17, SlotLevel: 5, SlotCount: 4, Invocations: 7},
	{Level: 18, SlotLevel: 5, SlotCount: 4, Invocations: 8},
	{Level: 19, SlotLevel: 5, SlotCount: 4, Invocations: 8},
	{Level: 20, SlotLevel: 5, SlotCount: 4, Invocations: 8},
}

// GetFullCasterSlots 获取全施法者指定等级的法术槽
func GetFullCasterSlots(level int) SpellSlots {
	if level < 1 || level > 20 {
		return SpellSlots{}
	}
	return fullCasterSlots[level-1]
}

// GetHalfCasterSlots 获取半施法者指定等级的法术槽
func GetHalfCasterSlots(level int) SpellSlots {
	if level < 1 || level > 20 {
		return SpellSlots{}
	}
	return halfCasterSlots[level-1]
}

// GetThirdCasterSlots 获取三分之一施法者指定等级的法术槽
func GetThirdCasterSlots(level int) SpellSlots {
	if level < 1 || level > 20 {
		return SpellSlots{}
	}
	return thirdCasterSlots[level-1]
}

// GetWarlockPactSlots 获取邪术师指定等级的契约魔法法术槽
func GetWarlockPactSlots(level int) PactSlots {
	if level < 1 || level > 20 {
		return PactSlots{}
	}
	return warlockPactSlots[level-1]
}

// GetSpellSlotsByType 根据施法类型获取法术槽
func GetSpellSlotsByType(casterType SpellcastingType, level int) SpellSlots {
	switch casterType {
	case SpellcastingFull:
		return GetFullCasterSlots(level)
	case SpellcastingHalf:
		return GetHalfCasterSlots(level)
	case SpellcastingThird:
		return GetThirdCasterSlots(level)
	default:
		return SpellSlots{}
	}
}

// MaxSpellLevel 返回可使用的最高环阶法术
func (s SpellSlots) MaxSpellLevel() int {
	if s.Level9 > 0 {
		return 9
	}
	if s.Level8 > 0 {
		return 8
	}
	if s.Level7 > 0 {
		return 7
	}
	if s.Level6 > 0 {
		return 6
	}
	if s.Level5 > 0 {
		return 5
	}
	if s.Level4 > 0 {
		return 4
	}
	if s.Level3 > 0 {
		return 3
	}
	if s.Level2 > 0 {
		return 2
	}
	if s.Level1 > 0 {
		return 1
	}
	return 0
}

// SpellLevelName 中文法术环阶名称
var SpellLevelName = map[int]string{
	0: "戏法",
	1: "1环",
	2: "2环",
	3: "3环",
	4: "4环",
	5: "5环",
	6: "6环",
	7: "7环",
	8: "8环",
	9: "9环",
}

// GetSpellLevelName 获取法术环阶中文名称
func GetSpellLevelName(level int) string {
	if name, ok := SpellLevelName[level]; ok {
		return name
	}
	return "未知"
}

// SchoolName 中文魔法学派名称
var SchoolName = map[SpellSchool]string{
	SchoolAbjuration:    "防护学派",
	SchoolConjuration:   "咒法学派",
	SchoolDivination:    "预言学派",
	SchoolEnchantment:   "惑控学派",
	SchoolEvocation:     "塑能学派",
	SchoolIllusion:      "幻术学派",
	SchoolNecromancy:    "死灵学派",
	SchoolTransmutation: "变化学派",
}

// GetSchoolName 获取魔法学派中文名称
func GetSchoolName(school SpellSchool) string {
	if name, ok := SchoolName[school]; ok {
		return name
	}
	return string(school)
}
