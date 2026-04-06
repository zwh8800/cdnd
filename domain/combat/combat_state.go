package combat

import (
	"time"

	"github.com/zwh8800/cdnd/domain/character"
	"github.com/zwh8800/cdnd/domain/llm"
)

// Position 位置坐标
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

// InitiativeEntry 先攻顺序条目
type InitiativeEntry struct {
	EntityID   string `json:"entity_id"`
	Initiative int    `json:"initiative"`
	HasActed   bool   `json:"has_acted"`
	IsPlayer   bool   `json:"is_player"`
}

// Combatant 战斗参与者
type Combatant struct {
	ID           string               `json:"id"`
	Name         string               `json:"name"`
	IsPlayer     bool                 `json:"is_player"`
	IsNPC        bool                 `json:"is_npc"`
	HP           int                  `json:"hp"`
	MaxHP        int                  `json:"max_hp"`
	AC           int                  `json:"ac"`
	Initiative   int                  `json:"initiative"`
	Conditions   []string             `json:"conditions"`
	Position     *Position            `json:"position,omitempty"`
	Abilities    character.Attributes `json:"abilities,omitempty"`
	SavingThrows map[string]int       `json:"saving_throws,omitempty"`
}

// CombatState 战斗状态
type CombatState struct {
	Active                bool              `json:"active"`
	Round                 int               `json:"round"`
	CurrentTurn           int               `json:"current_turn"`
	Initiative            []InitiativeEntry `json:"initiative"`
	Participants          []*Combatant      `json:"participants"`
	StartedAt             time.Time         `json:"started_at"`
	PlayerActionUsed      bool              `json:"player_action_used"`
	PlayerBonusActionUsed bool              `json:"player_bonus_action_used"`
	History               []llm.Message     `json:"history,omitempty"` // 战斗期间的历史记录
}
