package state

import (
	"time"

	"github.com/google/uuid"
	"github.com/zwh8800/cdnd/internal/character"
	"github.com/zwh8800/cdnd/internal/combat"
	"github.com/zwh8800/cdnd/internal/llm"
	"github.com/zwh8800/cdnd/internal/quest"
	"github.com/zwh8800/cdnd/internal/world"
)

// State 游戏状态
type State struct {
	// 基本信息
	SessionID string    `json:"session_id"`
	Phase     GamePhase `json:"phase"`
	TurnCount int       `json:"turn_count"`
	SubTurn   int       `json:"sub_turn"` // 子回合（用于战斗等）

	// 角色信息
	Character *character.Character `json:"character"`

	// 世界信息
	CurrentScene  *world.Scene        `json:"current_scene"`
	VisitedScenes map[string]bool     `json:"visited_scenes"`
	WorldFlags    map[string]bool     `json:"world_flags"`    // 世界标志（用于任务状态等）
	WorldCounters map[string]int      `json:"world_counters"` // 世界计数器
	Quests        []*quest.QuestState `json:"quests"`

	// 对话历史
	History   []llm.Message `json:"history"`
	DMContext string        `json:"dm_context"` // DM上下文（场景描述等）

	// 战斗状态
	Combat *combat.CombatState `json:"combat,omitempty"`

	// 当前可用的操作选项（由DM通过set_options工具设置）
	CurrentOptions []string `json:"current_options,omitempty"`

	// 时间戳
	CreatedAt   time.Time `json:"created_at"`
	LastSavedAt time.Time `json:"last_saved_at"`
	PlayedTime  int       `json:"played_time"` // 游戏时间（秒）
}

// NewState 创建新的游戏状态
func NewState() *State {
	return &State{
		SessionID:     uuid.New().String(),
		Phase:         PhaseCharacterCreation,
		TurnCount:     0,
		VisitedScenes: make(map[string]bool),
		WorldFlags:    make(map[string]bool),
		WorldCounters: make(map[string]int),
		History:       make([]llm.Message, 0),
		Quests:        make([]*quest.QuestState, 0),
		CreatedAt:     time.Now(),
		LastSavedAt:   time.Now(),
	}
}

// SetPhase 设置游戏阶段
func (s *State) SetPhase(phase GamePhase) {
	s.Phase = phase
}

// GetPhase 获取当前游戏阶段
func (s *State) GetPhase() GamePhase {
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
func (s *State) AddQuest(quest *quest.QuestState) {
	s.Quests = append(s.Quests, quest)
}

// GetQuest 获取任务
func (s *State) GetQuest(id string) *quest.QuestState {
	for _, q := range s.Quests {
		if q.ID == id {
			return q
		}
	}
	return nil
}

// StartCombat 开始战斗
func (s *State) StartCombat(participants []*combat.Combatant) {
	// 按先攻排序
	initiative := make([]combat.InitiativeEntry, 0, len(participants))
	for _, p := range participants {
		initiative = append(initiative, combat.InitiativeEntry{
			EntityID:   p.ID,
			Initiative: p.Initiative,
			IsPlayer:   p.IsPlayer,
		})
	}

	// 排序（高先攻先行动）
	sortInitiative(initiative)

	s.Combat = &combat.CombatState{
		Active:       true,
		Round:        1,
		CurrentTurn:  0,
		Initiative:   initiative,
		Participants: participants,
		StartedAt:    time.Now(),
	}
	s.Phase = PhaseCombat
}

// EndCombat 结束战斗
func (s *State) EndCombat() {
	s.Combat = nil
	s.Phase = PhaseExploration
}

// NextTurn 下一回合
func (s *State) NextTurn() *combat.Combatant {
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

// GetCombat 获取战斗状态
func (s *State) GetCombat() *combat.CombatState {
	return s.Combat
}

// GetCurrentCombatant 获取当前行动的战斗参与者
func (s *State) GetCurrentCombatant() *combat.Combatant {
	if s.Combat == nil || !s.Combat.Active {
		return nil
	}

	if s.Combat.CurrentTurn < len(s.Combat.Participants) {
		return s.Combat.Participants[s.Combat.CurrentTurn]
	}
	return nil
}

// GetCombatant 按ID查找战斗参与者
func (s *State) GetCombatant(id string) *combat.Combatant {
	if s.Combat == nil {
		return nil
	}

	for _, p := range s.Combat.Participants {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// RemoveCombatant 移除战斗参与者（用于死亡敌人）
func (s *State) RemoveCombatant(id string) bool {
	if s.Combat == nil {
		return false
	}

	for i, p := range s.Combat.Participants {
		if p.ID == id {
			// 从参与者列表中移除
			s.Combat.Participants = append(s.Combat.Participants[:i], s.Combat.Participants[i+1:]...)

			// 同时从先攻列表中移除
			for j, entry := range s.Combat.Initiative {
				if entry.EntityID == id {
					s.Combat.Initiative = append(s.Combat.Initiative[:j], s.Combat.Initiative[j+1:]...)
					break
				}
			}

			// 调整当前回合索引
			if s.Combat.CurrentTurn >= len(s.Combat.Participants) {
				s.Combat.CurrentTurn = 0
			}

			return true
		}
	}
	return false
}

// IsPlayerTurn 判断是否玩家回合
func (s *State) IsPlayerTurn() bool {
	if s.Combat == nil || !s.Combat.Active {
		return false
	}

	current := s.GetCurrentCombatant()
	return current != nil && current.IsPlayer
}

// GetEnemies 获取所有敌人列表
func (s *State) GetEnemies() []*combat.Combatant {
	if s.Combat == nil {
		return nil
	}

	var enemies []*combat.Combatant
	for _, p := range s.Combat.Participants {
		if !p.IsPlayer {
			enemies = append(enemies, p)
		}
	}
	return enemies
}

// GetAliveEnemies 获取存活的敌人列表
func (s *State) GetAliveEnemies() []*combat.Combatant {
	if s.Combat == nil {
		return nil
	}

	var enemies []*combat.Combatant
	for _, p := range s.Combat.Participants {
		if !p.IsPlayer && p.HP > 0 {
			enemies = append(enemies, p)
		}
	}
	return enemies
}

// IsCombatOver 检查战斗是否结束
func (s *State) IsCombatOver() (over bool, victory bool) {
	if s.Combat == nil || !s.Combat.Active {
		return true, false
	}

	enemies := s.GetAliveEnemies()
	if len(enemies) == 0 {
		return true, true // 所有敌人死亡，胜利
	}

	if s.Character != nil && s.Character.HitPoints.Current <= 0 {
		return true, false // 玩家死亡，失败
	}

	return false, false
}

// AddCombatHistory 添加战斗历史消息
func (s *State) AddCombatHistory(msg llm.Message) {
	if s.Combat == nil {
		return
	}
	s.Combat.History = append(s.Combat.History, msg)
}

// GetCombatHistory 获取战斗历史
func (s *State) GetCombatHistory() []llm.Message {
	if s.Combat == nil {
		return nil
	}
	result := make([]llm.Message, len(s.Combat.History))
	copy(result, s.Combat.History)
	return result
}

// ClearCombatHistory 清空战斗历史
func (s *State) ClearCombatHistory() {
	if s.Combat == nil {
		return
	}
	s.Combat.History = nil
}

// GetCombatStats 获取战斗统计信息
func (s *State) GetCombatStats() map[string]interface{} {
	if s.Combat == nil {
		return nil
	}

	enemies := s.GetEnemies()
	aliveEnemies := s.GetAliveEnemies()
	deadEnemies := len(enemies) - len(aliveEnemies)

	return map[string]interface{}{
		"round":         s.Combat.Round,
		"turn":          s.Combat.CurrentTurn,
		"total_enemies": len(enemies),
		"alive_enemies": len(aliveEnemies),
		"dead_enemies":  deadEnemies,
		"player_hp":     s.Character.HitPoints.Current,
		"player_max_hp": s.Character.HitPoints.Max,
	}
}

// SetCurrentOptions 设置当前选项
func (s *State) SetCurrentOptions(options []string) {
	s.CurrentOptions = options
}

// GetCurrentOptions 获取当前选项
func (s *State) GetCurrentOptions() []string {
	return s.CurrentOptions
}

// ClearCurrentOptions 清除当前选项
func (s *State) ClearCurrentOptions() {
	s.CurrentOptions = nil
}

// sortInitiative 按先攻值排序（高到低）
func sortInitiative(entries []combat.InitiativeEntry) {
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Initiative > entries[i].Initiative {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
}

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
