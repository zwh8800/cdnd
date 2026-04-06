package dnd5e_engine

import (
	"github.com/zwh8800/cdnd/domain"
	"github.com/zwh8800/cdnd/domain/character"
)

type Engine struct {
	state State
}

type State struct {
	Phase     domain.GamePhase `json:"phase"`
	TurnCount int              `json:"turn_count"`

	// 角色信息
	Character *character.Character `json:"character"`
}
