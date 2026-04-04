package character

// Size 体型类别（官方中文翻译）
type Size string

const (
	SizeTiny       Size = "微型"  // Tiny
	SizeSmall      Size = "小型"  // Small
	SizeMedium     Size = "中型"  // Medium
	SizeLarge      Size = "大型"  // Large
	SizeHuge       Size = "巨型"  // Huge
	SizeGargantuan Size = "超巨型" // Gargantuan
)

// AgeRange 年龄范围
type AgeRange struct {
	Adulthood int `json:"adulthood"` // 成年年龄
	MaxAge    int `json:"max_age"`   // 最大寿命
}

// HeightRange 身高范围（单位：尺）
type HeightRange struct {
	BaseHeight int `json:"base_height"` // 基础身高
	ModDice    int `json:"mod_dice"`    // 身高变异骰子面数
	ModCount   int `json:"mod_count"`   // 骰子数量
}

// WeightRange 体重范围（单位：磅）
type WeightRange struct {
	BaseWeight int `json:"base_weight"` // 基础体重
	ModDice    int `json:"mod_dice"`    // 体重变异骰子面数
	ModCount   int `json:"mod_count"`   // 骰子数量
}

// SubRace 子种族
type SubRace struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`            // 中文名称
	Description    string          `json:"description"`     // 中文描述
	AbilityBonuses map[Ability]int `json:"ability_bonuses"` // 属性加值
	Traits         []Trait         `json:"traits"`          // 子种族特性
}

// Race 种族（官方中文翻译）
type Race struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`            // 中文名称
	NameEn         string          `json:"name_en"`         // 英文名称
	Description    string          `json:"description"`     // 中文描述
	Size           Size            `json:"size"`            // 体型
	Speed          int             `json:"speed"`           // 速度（尺）
	AbilityBonuses map[Ability]int `json:"ability_bonuses"` // 属性加值
	Traits         []Trait         `json:"traits"`          // 种族特性
	Languages      []string        `json:"languages"`       // 语言
	// 扩展字段
	SubRaces       []SubRace   `json:"sub_races,omitempty"`       // 子种族选项
	AgeRange       AgeRange    `json:"age_range"`                 // 年龄范围
	HeightRange    HeightRange `json:"height_range"`              // 身高范围
	WeightRange    WeightRange `json:"weight_range"`              // 体重范围
	WeaponTraining []string    `json:"weapon_training,omitempty"` // 武器熟练
	Cantrips       []string    `json:"cantrips,omitempty"`        // 天生戏法
}

// Trait 种族特性
type Trait struct {
	Name        string `json:"name"`        // 特性名称（中文）
	Description string `json:"description"` // 特性描述（中文）
}

// GetSubRace 根据ID获取子种族
func (r *Race) GetSubRace(id string) *SubRace {
	for i := range r.SubRaces {
		if r.SubRaces[i].ID == id {
			return &r.SubRaces[i]
		}
	}
	return nil
}

// HasSubRaces 检查是否有子种族选项
func (r *Race) HasSubRaces() bool {
	return len(r.SubRaces) > 0
}

// GetAllRaces 获取所有种族
func GetAllRaces() []*Race {
	races := make([]*Race, 0, len(StandardRaces))
	for _, r := range StandardRaces {
		race := r // 创建副本
		races = append(races, &race)
	}
	return races
}
