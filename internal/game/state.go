package game

import (
	"time"

	"github.com/google/uuid"
	"github.com/zwh8800/cdnd/internal/character"
	"github.com/zwh8800/cdnd/internal/llm"
	"github.com/zwh8800/cdnd/internal/save"
	"github.com/zwh8800/cdnd/internal/world"
)

// State 游戏状态
type State struct {
	// 基本信息
	SessionID string         `json:"session_id"`
	Phase     save.GamePhase `json:"phase"`
	TurnCount int            `json:"turn_count"`
	SubTurn   int            `json:"sub_turn"` // 子回合（用于战斗等）

	// 角色信息
	Character *character.Character `json:"character"`

	// 世界信息
	CurrentScene  *world.Scene       `json:"current_scene"`
	VisitedScenes map[string]bool    `json:"visited_scenes"`
	WorldFlags    map[string]bool    `json:"world_flags"`    // 世界标志（用于任务状态等）
	WorldCounters map[string]int     `json:"world_counters"` // 世界计数器
	Quests        []*save.QuestState `json:"quests"`

	// 对话历史
	History   []llm.Message `json:"history"`
	DMContext string        `json:"dm_context"` // DM上下文（场景描述等）

	// 战斗状态
	Combat *save.CombatState `json:"combat,omitempty"`

	// 时间戳
	CreatedAt   time.Time `json:"created_at"`
	LastSavedAt time.Time `json:"last_saved_at"`
	PlayedTime  int       `json:"played_time"` // 游戏时间（秒）
}

// NewState 创建新的游戏状态
func NewState() *State {
	return &State{
		SessionID:     uuid.New().String(),
		Phase:         save.PhaseCharacterCreation,
		TurnCount:     0,
		VisitedScenes: make(map[string]bool),
		WorldFlags:    make(map[string]bool),
		WorldCounters: make(map[string]int),
		History:       make([]llm.Message, 0),
		Quests:        make([]*save.QuestState, 0),
		CreatedAt:     time.Now(),
		LastSavedAt:   time.Now(),
	}
}

// SetPhase 设置游戏阶段
func (s *State) SetPhase(phase save.GamePhase) {
	s.Phase = phase
}

// GetPhase 获取当前游戏阶段
func (s *State) GetPhase() save.GamePhase {
	return s.Phase
}

// SetCharacter 设置角色
func (s *State) SetCharacter(c *character.Character) {
	s.Character = c
}

// GetCharacter 获取角色
func (s *State) GetCharacter() *character.Character {
	return s.Character
}

// SetCurrentScene 设置当前场景
func (s *State) SetCurrentScene(scene *world.Scene) {
	s.CurrentScene = scene
	if scene != nil {
		s.VisitedScenes[scene.ID] = true
	}
}

// GetCurrentScene 获取当前场景
func (s *State) GetCurrentScene() *world.Scene {
	return s.CurrentScene
}

// AddHistory 添加对话历史
func (s *State) AddHistory(msg llm.Message) {
	s.History = append(s.History, msg)
}

// GetHistory 获取对话历史
func (s *State) GetHistory() []llm.Message {
	result := make([]llm.Message, len(s.History))
	copy(result, s.History)
	return result
}

// IncrementTurn 增加回合数
func (s *State) IncrementTurn() {
	s.TurnCount++
}

// SetFlag 设置世界标志
func (s *State) SetFlag(key string, value bool) {
	s.WorldFlags[key] = value
}

// GetFlag 获取世界标志
func (s *State) GetFlag(key string) bool {
	return s.WorldFlags[key]
}

// SetCounter 设置世界计数器
func (s *State) SetCounter(key string, value int) {
	s.WorldCounters[key] = value
}

// GetCounter 获取世界计数器
func (s *State) GetCounter(key string) int {
	return s.WorldCounters[key]
}

// IncrementCounter 增加计数器值
func (s *State) IncrementCounter(key string, delta int) int {
	s.WorldCounters[key] += delta
	return s.WorldCounters[key]
}

// AddQuest 添加任务
func (s *State) AddQuest(quest *save.QuestState) {
	s.Quests = append(s.Quests, quest)
}

// GetQuest 获取任务
func (s *State) GetQuest(id string) *save.QuestState {
	for _, q := range s.Quests {
		if q.ID == id {
			return q
		}
	}
	return nil
}

// StartCombat 开始战斗
func (s *State) StartCombat(participants []*save.Combatant) {
	// 按先攻排序
	initiative := make([]save.InitiativeEntry, 0, len(participants))
	for _, p := range participants {
		initiative = append(initiative, save.InitiativeEntry{
			EntityID:   p.ID,
			Initiative: p.Initiative,
			IsPlayer:   p.IsPlayer,
		})
	}

	// 排序（高先攻先行动）
	sortInitiative(initiative)

	s.Combat = &save.CombatState{
		Active:       true,
		Round:        1,
		CurrentTurn:  0,
		Initiative:   initiative,
		Participants: participants,
		StartedAt:    time.Now(),
	}
	s.Phase = save.PhaseCombat
}

// EndCombat 结束战斗
func (s *State) EndCombat() {
	s.Combat = nil
	s.Phase = save.PhaseExploration
}

// NextTurn 下一回合
func (s *State) NextTurn() *save.Combatant {
	if s.Combat == nil || !s.Combat.Active {
		return nil
	}

	// 标记当前行动者已行动
	if s.Combat.CurrentTurn < len(s.Combat.Initiative) {
		s.Combat.Initiative[s.Combat.CurrentTurn].HasActed = true
	}

	// 移动到下一个
	s.Combat.CurrentTurn++

	// 如果一轮结束，开始新回合
	if s.Combat.CurrentTurn >= len(s.Combat.Initiative) {
		s.Combat.Round++
		s.Combat.CurrentTurn = 0
		// 重置所有行动标记
		for i := range s.Combat.Initiative {
			s.Combat.Initiative[i].HasActed = false
		}
	}

	// 返回当前行动者
	if s.Combat.CurrentTurn < len(s.Combat.Participants) {
		return s.Combat.Participants[s.Combat.CurrentTurn]
	}
	return nil
}

// GetCurrentCombatant 获取当前行动的战斗参与者
func (s *State) GetCurrentCombatant() *save.Combatant {
	if s.Combat == nil || !s.Combat.Active {
		return nil
	}

	if s.Combat.CurrentTurn < len(s.Combat.Participants) {
		return s.Combat.Participants[s.Combat.CurrentTurn]
	}
	return nil
}

// sortInitiative 按先攻值排序（高到低）
func sortInitiative(entries []save.InitiativeEntry) {
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Initiative > entries[i].Initiative {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
}
