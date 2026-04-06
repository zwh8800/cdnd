package character

// Item 表示游戏中的物品。
type Item struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        ItemType `json:"type"`
	Weight      float64  `json:"weight"`
	Value       int      `json:"value"` // 以金币计
	Quantity    int      `json:"quantity"`
	Rarity      Rarity   `json:"rarity"`
	Attuned     bool     `json:"attuned"`
	Properties  []string `json:"properties"`
}

// ItemType 表示物品的类别。
type ItemType string

const (
	ItemWeapon       ItemType = "weapon"
	ItemArmor        ItemType = "armor"
	ItemShield       ItemType = "shield"
	ItemPotion       ItemType = "potion"
	ItemScroll       ItemType = "scroll"
	ItemWand         ItemType = "wand"
	ItemRing         ItemType = "ring"
	ItemRod          ItemType = "rod"
	ItemStaff        ItemType = "staff"
	ItemWondrousItem ItemType = "wondrous_item"
	ItemAmmunition   ItemType = "ammunition"
	ItemTool         ItemType = "tool"
	ItemGear         ItemType = "gear"
	ItemTreasure     ItemType = "treasure"
)

// Rarity 表示物品的稀有度。
type Rarity string

const (
	RarityCommon    Rarity = "common"
	RarityUncommon  Rarity = "uncommon"
	RarityRare      Rarity = "rare"
	RarityVeryRare  Rarity = "very_rare"
	RarityLegendary Rarity = "legendary"
	RarityArtifact  Rarity = "artifact"
)

// Proficiency 表示角色的熟练项。
type Proficiency struct {
	Type        ProficiencyType `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
}

// ProficiencyType 表示熟练项的类型。
type ProficiencyType string

const (
	ProfArmor       ProficiencyType = "armor"
	ProfWeapon      ProficiencyType = "weapon"
	ProfTool        ProficiencyType = "tool"
	ProfLanguage    ProficiencyType = "language"
	ProfSkill       ProficiencyType = "skill"
	ProfSavingThrow ProficiencyType = "saving_throw"
)

// Spell 表示一个法术。
type Spell struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Level       int         `json:"level"`
	School      SpellSchool `json:"school"`
	CastingTime string      `json:"casting_time"`
	Range       string      `json:"range"`
	Components  []string    `json:"components"`
	Duration    string      `json:"duration"`
	Description string      `json:"description"`
	Classes     []string    `json:"classes"`
}

// SpellSchool 表示魔法学派。
type SpellSchool string

const (
	SchoolAbjuration    SpellSchool = "abjuration"
	SchoolConjuration   SpellSchool = "conjuration"
	SchoolDivination    SpellSchool = "divination"
	SchoolEnchantment   SpellSchool = "enchantment"
	SchoolEvocation     SpellSchool = "evocation"
	SchoolIllusion      SpellSchool = "illusion"
	SchoolNecromancy    SpellSchool = "necromancy"
	SchoolTransmutation SpellSchool = "transmutation"
)

// Inventory 管理角色的物品。
type Inventory struct {
	Items    []Item `json:"items"`
	Capacity int    `json:"capacity"` // 最大重量或物品数量
}

// AddItem 向背包添加物品。
func (inv *Inventory) AddItem(item Item) {
	// 检查物品是否已存在（可堆叠）
	for i, existing := range inv.Items {
		if existing.ID == item.ID && existing.Type != ItemWeapon && existing.Type != ItemArmor {
			inv.Items[i].Quantity += item.Quantity
			return
		}
	}
	inv.Items = append(inv.Items, item)
}

// RemoveItem 从背包移除物品。
func (inv *Inventory) RemoveItem(itemID string, quantity int) bool {
	for i, item := range inv.Items {
		if item.ID == itemID {
			if item.Quantity <= quantity {
				// 移除整个堆叠
				inv.Items = append(inv.Items[:i], inv.Items[i+1:]...)
			} else {
				inv.Items[i].Quantity -= quantity
			}
			return true
		}
	}
	return false
}

// GetTotalWeight 计算所有物品的总重量。
func (inv *Inventory) GetTotalWeight() float64 {
	total := 0.0
	for _, item := range inv.Items {
		total += item.Weight * float64(item.Quantity)
	}
	return total
}
