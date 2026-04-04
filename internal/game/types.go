package game

import (
	"time"

	"github.com/zwh8800/cdnd/internal/save"
)

// ActionType 行动类型
type ActionType int

const (
	ActionNone ActionType = iota
	// 探索行动
	ActionMove     // 移动
	ActionExamine  // 检查
	ActionInteract // 互动
	ActionRest     // 休息
	// 对话行动
	ActionSay       // 说话
	ActionAsk       // 询问
	ActionNegotiate // 谈判
	// 战斗行动
	ActionAttack  // 攻击
	ActionDefend  // 防御
	ActionCast    // 施法
	ActionUseItem // 使用物品
	ActionFlee    // 逃跑
	// 系统行动
	ActionSave // 保存
	ActionLoad // 加载
	ActionQuit // 退出
)

// String 返回行动类型的中文名称
func (a ActionType) String() string {
	switch a {
	case ActionMove:
		return "移动"
	case ActionExamine:
		return "检查"
	case ActionInteract:
		return "互动"
	case ActionRest:
		return "休息"
	case ActionSay:
		return "说话"
	case ActionAsk:
		return "询问"
	case ActionNegotiate:
		return "谈判"
	case ActionAttack:
		return "攻击"
	case ActionDefend:
		return "防御"
	case ActionCast:
		return "施法"
	case ActionUseItem:
		return "使用物品"
	case ActionFlee:
		return "逃跑"
	case ActionSave:
		return "保存"
	case ActionLoad:
		return "加载"
	case ActionQuit:
		return "退出"
	default:
		return "未知行动"
	}
}

// Difficulty 难度等级常量
type Difficulty int

const (
	DifficultyVeryEasy Difficulty = 5
	DifficultyEasy     Difficulty = 10
	DifficultyMedium   Difficulty = 15
	DifficultyHard     Difficulty = 20
	DifficultyVeryHard Difficulty = 25
	DifficultyHeroic   Difficulty = 30
)

// DC 返回难度等级对应的DC值
func (d Difficulty) DC() int {
	return int(d)
}

// String 返回难度等级的中文名称
func (d Difficulty) String() string {
	switch d {
	case DifficultyVeryEasy:
		return "非常简单"
	case DifficultyEasy:
		return "简单"
	case DifficultyMedium:
		return "中等"
	case DifficultyHard:
		return "困难"
	case DifficultyVeryHard:
		return "非常困难"
	case DifficultyHeroic:
		return "英雄级"
	default:
		return "未知"
	}
}

// ActionResult 行动结果
type ActionResult struct {
	Action    ActionType `json:"action"`
	Success   bool       `json:"success"`
	Message   string     `json:"message"`
	Data      any        `json:"data,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

// TurnInfo 回合信息（用于战斗）
type TurnInfo struct {
	Round       int       `json:"round"`
	Turn        int       `json:"turn"`
	CurrentTurn int       `json:"current_turn"` // 当前行动者索引
	Order       []string  `json:"order"`        // 先攻顺序 (实体ID列表)
	StartedAt   time.Time `json:"started_at"`
}

// GameConfig 游戏配置
type GameConfig struct {
	MaxHistoryTurns   int    `json:"max_history_turns"`   // 最大历史记录回合数
	AutoSave          bool   `json:"auto_save"`           // 自动保存
	AutoSaveInterval  int    `json:"auto_save_interval"`  // 自动保存间隔（回合数）
	ShowDiceRolls     bool   `json:"show_dice_rolls"`     // 显示骰子结果
	ShowDamageNumbers bool   `json:"show_damage_numbers"` // 显示伤害数值
	StreamResponses   bool   `json:"stream_responses"`    // 流式响应
	Language          string `json:"language"`            // 游戏语言
}

// DefaultGameConfig 返回默认游戏配置
func DefaultGameConfig() *GameConfig {
	return &GameConfig{
		MaxHistoryTurns:   50,
		AutoSave:          true,
		AutoSaveInterval:  5,
		ShowDiceRolls:     true,
		ShowDamageNumbers: true,
		StreamResponses:   true,
		Language:          "zh-CN",
	}
}

// Phase 常量 - save 包的别名
const (
	PhaseCharacterCreation = save.PhaseCharacterCreation
	PhaseIntroduction      = save.PhaseIntroduction
	PhaseExploration       = save.PhaseExploration
	PhaseDialogue          = save.PhaseDialogue
	PhaseCombat            = save.PhaseCombat
	PhaseRest              = save.PhaseRest
	PhaseGameOver          = save.PhaseGameOver
)
