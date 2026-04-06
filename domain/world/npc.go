package world

import (
	"time"

	"github.com/zwh8800/cdnd/domain/character"
)

// NPCDisposition NPC态度
type NPCDisposition int

const (
	DispositionHostile    NPCDisposition = iota // 敌对
	DispositionUnfriendly                       // 不友好
	DispositionNeutral                          // 中立
	DispositionFriendly                         // 友好
	DispositionAllied                           // 同盟
)

// String 返回态度的中文名称
func (d NPCDisposition) String() string {
	switch d {
	case DispositionHostile:
		return "敌对"
	case DispositionUnfriendly:
		return "不友好"
	case DispositionNeutral:
		return "中立"
	case DispositionFriendly:
		return "友好"
	case DispositionAllied:
		return "同盟"
	default:
		return "未知"
	}
}

// NPCType NPC类型
type NPCType int

const (
	NPCTypeGeneric    NPCType = iota // 普通
	NPCTypeMerchant                  // 商人
	NPCTypeQuestGiver                // 任务发布者
	NPCTypeEnemy                     // 敌人
	NPCTypeAlly                      // 盟友
	NPCTypeTrainer                   // 训练师
)

// String 返回NPC类型的中文名称
func (t NPCType) String() string {
	switch t {
	case NPCTypeGeneric:
		return "普通NPC"
	case NPCTypeMerchant:
		return "商人"
	case NPCTypeQuestGiver:
		return "任务发布者"
	case NPCTypeEnemy:
		return "敌人"
	case NPCTypeAlly:
		return "盟友"
	case NPCTypeTrainer:
		return "训练师"
	default:
		return "未知"
	}
}

// NPC 非玩家角色
type NPC struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`        // 名称（中文）
	Type        NPCType        `json:"type"`        // 类型
	Disposition NPCDisposition `json:"disposition"` // 对玩家态度
	Description string         `json:"description"` // 描述（中文）
	Appearance  string         `json:"appearance"`  // 外貌描述（中文）
	Personality string         `json:"personality"` // 性格特征

	// 对话信息
	Greeting     string           `json:"greeting"`      // 问候语
	DialogueTree []DialogueOption `json:"dialogue_tree"` // 对话选项树

	// 属性信息（用于战斗NPC）
	HP        int                         `json:"hp,omitempty"`
	MaxHP     int                         `json:"max_hp,omitempty"`
	AC        int                         `json:"ac,omitempty"`
	Speed     int                         `json:"speed,omitempty"` // 移动速度（尺）
	Abilities character.Attributes        `json:"abilities,omitempty"`
	Skills    map[character.SkillType]int `json:"skills,omitempty"`

	// 战斗相关
	CR         float64     `json:"cr,omitempty"`         // 挑战等级
	XP         int         `json:"xp,omitempty"`         // 经验值
	Actions    []NPCAction `json:"actions,omitempty"`    // 可用动作
	Conditions []string    `json:"conditions,omitempty"` // 当前状态

	// 商人相关
	ShopInventory  []string `json:"shop_inventory,omitempty"`  // 商品ID列表
	BuyMultiplier  float64  `json:"buy_multiplier,omitempty"`  // 收购价格倍率
	SellMultiplier float64  `json:"sell_multiplier,omitempty"` // 出售价格倍率

	// 任务相关
	QuestsGiven []string `json:"quests_given,omitempty"` // 可发布的任务ID

	// AI提示
	HiddenContext string   `json:"hidden_context,omitempty"` // DM可见的隐藏信息
	Keywords      []string `json:"keywords,omitempty"`       // 关键词标签

	// 元数据
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DialogueOption 对话选项
type DialogueOption struct {
	ID          string           `json:"id"`
	Text        string           `json:"text"`                   // 选项文本（中文）
	Response    string           `json:"response"`               // NPC回应（中文）
	Condition   string           `json:"condition,omitempty"`    // 显示条件
	Effect      string           `json:"effect,omitempty"`       // 选择后的效果
	NextOptions []DialogueOption `json:"next_options,omitempty"` // 后续选项
	Repeatable  bool             `json:"repeatable"`             // 是否可重复
	Used        bool             `json:"used"`                   // 是否已使用
}

// NPCAction NPC动作
type NPCAction struct {
	Name        string `json:"name"`        // 动作名称（中文）
	Description string `json:"description"` // 动作描述
	Type        string `json:"type"`        // 攻击类型: melee, ranged, spell
	AttackBonus int    `json:"attack_bonus"`
	Damage      string `json:"damage"`      // 伤害表达式，如 "1d8+3"
	DamageType  string `json:"damage_type"` // 伤害类型
	Range       string `json:"range"`       // 射程
}

// TakeDamage 受到伤害
func (n *NPC) TakeDamage(amount int) int {
	n.HP -= amount
	if n.HP < 0 {
		n.HP = 0
	}
	return n.HP
}

// Heal 恢复生命值
func (n *NPC) Heal(amount int) int {
	n.HP += amount
	if n.HP > n.MaxHP {
		n.HP = n.MaxHP
	}
	return n.HP
}

// IsDead 检查是否死亡
func (n *NPC) IsDead() bool {
	return n.HP <= 0
}

// HasCondition 检查是否有状态
func (n *NPC) HasCondition(condition string) bool {
	for _, c := range n.Conditions {
		if c == condition {
			return true
		}
	}
	return false
}

// AddCondition 添加状态
func (n *NPC) AddCondition(condition string) {
	if !n.HasCondition(condition) {
		n.Conditions = append(n.Conditions, condition)
	}
}

// RemoveCondition 移除状态
func (n *NPC) RemoveCondition(condition string) {
	for i, c := range n.Conditions {
		if c == condition {
			n.Conditions = append(n.Conditions[:i], n.Conditions[i+1:]...)
			return
		}
	}
}

// GetDialogueOption 获取对话选项
func (n *NPC) GetDialogueOption(id string) *DialogueOption {
	return findDialogueOption(n.DialogueTree, id)
}

// findDialogueOption 递归查找对话选项
func findDialogueOption(options []DialogueOption, id string) *DialogueOption {
	for i := range options {
		if options[i].ID == id {
			return &options[i]
		}
		if len(options[i].NextOptions) > 0 {
			if found := findDialogueOption(options[i].NextOptions, id); found != nil {
				return found
			}
		}
	}
	return nil
}

// SetDisposition 设置态度
func (n *NPC) SetDisposition(d NPCDisposition) {
	n.Disposition = d
}

// ImproveDisposition 改善态度
func (n *NPC) ImproveDisposition() bool {
	if n.Disposition < DispositionAllied {
		n.Disposition++
		return true
	}
	return false
}

// WorsenDisposition 恶化态度
func (n *NPC) WorsenDisposition() bool {
	if n.Disposition > DispositionHostile {
		n.Disposition--
		return true
	}
	return false
}
