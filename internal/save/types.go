package save

import (
	"time"

	"github.com/zwh8800/cdnd/internal/character"
	"github.com/zwh8800/cdnd/internal/llm"
	"github.com/zwh8800/cdnd/internal/world"
)

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

// Position 位置坐标
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

// QuestState 任务状态
type QuestState struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Stage       int         `json:"stage"`
	Completed   bool        `json:"completed"`
	Objectives  []Objective `json:"objectives"`
	StartedAt   time.Time   `json:"started_at"`
	CompletedAt time.Time   `json:"completed_at,omitempty"`
}

// Objective 任务目标
type Objective struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	Target      int    `json:"target,omitempty"`
	Current     int    `json:"current,omitempty"`
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
	Active       bool              `json:"active"`
	Round        int               `json:"round"`
	CurrentTurn  int               `json:"current_turn"`
	Initiative   []InitiativeEntry `json:"initiative"`
	Participants []*Combatant      `json:"participants"`
	StartedAt    time.Time         `json:"started_at"`
}

// SaveData 存档数据
type SaveData struct {
	// 元数据
	Slot      int       `json:"slot"`
	SaveName  string    `json:"save_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	PlayTime  int       `json:"play_time"` // 游戏时间（秒）

	// 游戏状态
	SessionID string    `json:"session_id"`
	Phase     GamePhase `json:"phase"`
	TurnCount int       `json:"turn_count"`

	// 角色数据
	Character *character.Character `json:"character"`

	// 世界数据
	CurrentScene  *world.Scene    `json:"current_scene"`
	VisitedScenes map[string]bool `json:"visited_scenes"`
	WorldFlags    map[string]bool `json:"world_flags"`
	WorldCounters map[string]int  `json:"world_counters"`
	Quests        []*QuestState   `json:"quests"`

	// 场景和NPC数据
	Scenes []*world.Scene `json:"scenes"`
	NPCs   []*world.NPC   `json:"npcs"`

	// 对话历史
	History   []llm.Message `json:"history"`
	DMContext string        `json:"dm_context"`

	// 战斗状态
	Combat *CombatState `json:"combat,omitempty"`

	// 版本信息
	Version string `json:"version"`
}

// SaveSlot 存档槽位信息（用于显示存档列表）
type SaveSlot struct {
	Slot           int       `json:"slot"`
	SaveName       string    `json:"save_name"`
	CharacterName  string    `json:"character_name"`
	CharacterLevel int       `json:"character_level"`
	CharacterClass string    `json:"character_class"`
	Phase          GamePhase `json:"phase"`
	PlayTime       int       `json:"play_time"`
	UpdatedAt      time.Time `json:"updated_at"`
	Preview        string    `json:"preview"` // 场景名称或简要描述
}

// SaveMetadata 存档元数据
type SaveMetadata struct {
	TotalSaves    int       `json:"total_saves"`
	TotalPlayTime int       `json:"total_play_time"`
	LastPlayed    time.Time `json:"last_played"`
	Version       string    `json:"version"`
}

// NewSaveData 创建新的存档数据
func NewSaveData(slot int) *SaveData {
	return &SaveData{
		Slot:          slot,
		SaveName:      "新存档",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		VisitedScenes: make(map[string]bool),
		WorldFlags:    make(map[string]bool),
		WorldCounters: make(map[string]int),
		Quests:        make([]*QuestState, 0),
		Scenes:        make([]*world.Scene, 0),
		NPCs:          make([]*world.NPC, 0),
		History:       make([]llm.Message, 0),
		Version:       "1.0.0",
	}
}

// GetWorldData 获取世界数据
func (d *SaveData) GetWorldData() ([]*world.Scene, []*world.NPC) {
	return d.Scenes, d.NPCs
}

// ToSlot 转换为存档槽位信息
func (d *SaveData) ToSlot() *SaveSlot {
	slot := &SaveSlot{
		Slot:      d.Slot,
		SaveName:  d.SaveName,
		Phase:     d.Phase,
		PlayTime:  d.PlayTime,
		UpdatedAt: d.UpdatedAt,
	}

	if d.Character != nil {
		slot.CharacterName = d.Character.Name
		slot.CharacterLevel = d.Character.Level
		if d.Character.HasClass() {
			slot.CharacterClass = d.Character.Class.Name
		}
	}

	if d.CurrentScene != nil {
		slot.Preview = d.CurrentScene.Name
	}

	return slot
}
