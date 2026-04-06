package storage

import (
	"time"

	"github.com/zwh8800/cdnd/domain"
	"github.com/zwh8800/cdnd/domain/character"
	"github.com/zwh8800/cdnd/domain/combat"
	"github.com/zwh8800/cdnd/domain/llm"
	"github.com/zwh8800/cdnd/domain/quest"
	"github.com/zwh8800/cdnd/domain/world"
)

// SaveData 存档数据
type SaveData struct {
	// 元数据
	Slot      int       `json:"slot"`
	SaveName  string    `json:"save_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	PlayTime  int       `json:"play_time"` // 游戏时间（秒）

	// 游戏状态
	SessionID string           `json:"session_id"`
	Phase     domain.GamePhase `json:"phase"`
	TurnCount int              `json:"turn_count"`

	// 角色数据
	Character *character.Character `json:"character"`

	// 世界数据
	CurrentScene  *world.Scene        `json:"current_scene"`
	VisitedScenes map[string]bool     `json:"visited_scenes"`
	WorldFlags    map[string]bool     `json:"world_flags"`
	WorldCounters map[string]int      `json:"world_counters"`
	Quests        []*quest.QuestState `json:"quests"`

	// 场景和NPC数据
	Scenes []*world.Scene `json:"scenes"`
	NPCs   []*world.NPC   `json:"npcs"`

	// 对话历史
	History   []llm.Message `json:"history"`
	DMContext string        `json:"dm_context"`

	// 战斗状态
	Combat *combat.CombatState `json:"combat,omitempty"`

	// 版本信息
	Version string `json:"version"`
}

// SaveSlot 存档槽位信息（用于显示存档列表）
type SaveSlot struct {
	Slot           int              `json:"slot"`
	SaveName       string           `json:"save_name"`
	CharacterName  string           `json:"character_name"`
	CharacterLevel int              `json:"character_level"`
	CharacterClass string           `json:"character_class"`
	Phase          domain.GamePhase `json:"phase"`
	PlayTime       int              `json:"play_time"`
	UpdatedAt      time.Time        `json:"updated_at"`
	Preview        string           `json:"preview"` // 场景名称或简要描述
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
		Quests:        make([]*quest.QuestState, 0),
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
