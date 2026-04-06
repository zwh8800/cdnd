package domain

// GamePhase 游戏阶段
type GamePhase int

const (
	PhaseCharacterCreation GamePhase = iota // 角色创建
	PhaseIntroduction                       // 引言/开场
	PhaseExploration                        // 探索
	PhaseDialogue                           // 对话
	PhaseCombat                             // 战斗
	PhaseRest                               // 休息
	PhaseGameOver                           // 游戏结束
)

// String 返回游戏阶段的中文名称
func (p GamePhase) String() string {
	switch p {
	case PhaseCharacterCreation:
		return "角色创建"
	case PhaseIntroduction:
		return "开场"
	case PhaseExploration:
		return "探索"
	case PhaseDialogue:
		return "对话"
	case PhaseCombat:
		return "战斗"
	case PhaseRest:
		return "休息"
	case PhaseGameOver:
		return "游戏结束"
	default:
		return "未知"
	}
}
