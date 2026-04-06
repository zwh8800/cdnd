package world

import (
	"time"
)

// SceneType 场景类型
type SceneType string

const (
	SceneTypeTown       SceneType = "town"       // 城镇
	SceneTypeDungeon    SceneType = "dungeon"    // 地下城
	SceneTypeWilderness SceneType = "wilderness" // 荒野
	SceneTypeBuilding   SceneType = "building"   // 建筑
	SceneTypeRoom       SceneType = "room"       // 房间
	SceneTypeCombat     SceneType = "combat"     // 战斗场景
)

// Scene 场景
type Scene struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`        // 场景名称（中文）
	Type        SceneType `json:"type"`        // 场景类型
	Description string    `json:"description"` // 场景描述（中文）

	// 连接信息
	Exits []Exit `json:"exits"` // 出口

	// 内容
	NPCs     []string  `json:"npcs"`     // NPC ID列表
	Items    []string  `json:"items"`    // 物品ID列表
	Features []Feature `json:"features"` // 场景特性

	// 环境信息
	LightLevel LightLevel `json:"light_level"` // 光照等级
	Terrain    Terrain    `json:"terrain"`     // 地形类型
	Danger     int        `json:"danger"`      // 危险等级 (1-10)

	// 元数据
	Tags       []string       `json:"tags"`       // 标签
	Properties map[string]any `json:"properties"` // 自定义属性
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// Exit 出口/通道
type Exit struct {
	ID          string `json:"id"`
	Name        string `json:"name"`             // 出口名称（中文）
	Description string `json:"description"`      // 出口描述
	TargetScene string `json:"target_scene"`     // 目标场景ID
	Locked      bool   `json:"locked"`           // 是否锁定
	KeyID       string `json:"key_id,omitempty"` // 开锁需要的钥匙ID
	DC          int    `json:"dc,omitempty"`     // 开锁DC
	Visible     bool   `json:"visible"`          // 是否可见
	OneWay      bool   `json:"one_way"`          // 是否单向通道
}

// Feature 场景特性
type Feature struct {
	ID          string `json:"id"`
	Name        string `json:"name"`        // 特性名称（中文）
	Description string `json:"description"` // 特性描述
	Interactive bool   `json:"interactive"` // 是否可互动
	Used        bool   `json:"used"`        // 是否已使用
}

// LightLevel 光照等级
type LightLevel int

const (
	LightBright LightLevel = iota // 明亮
	LightDim                      // 昏暗
	LightDark                     // 黑暗
)

// String 返回光照等级的中文名称
func (l LightLevel) String() string {
	switch l {
	case LightBright:
		return "明亮"
	case LightDim:
		return "昏暗"
	case LightDark:
		return "黑暗"
	default:
		return "未知"
	}
}

// Terrain 地形类型
type Terrain int

const (
	TerrainNormal    Terrain = iota // 普通
	TerrainDifficult                // 困难地形
	TerrainWater                    // 水域
	TerrainLava                     // 岩浆
	TerrainIce                      // 冰面
	TerrainCliff                    // 悬崖
)

// String 返回地形类型的中文名称
func (t Terrain) String() string {
	switch t {
	case TerrainNormal:
		return "普通"
	case TerrainDifficult:
		return "困难地形"
	case TerrainWater:
		return "水域"
	case TerrainLava:
		return "岩浆"
	case TerrainIce:
		return "冰面"
	case TerrainCliff:
		return "悬崖"
	default:
		return "未知"
	}
}

// GetExit 获取指定ID的出口
func (s *Scene) GetExit(id string) *Exit {
	for i := range s.Exits {
		if s.Exits[i].ID == id {
			return &s.Exits[i]
		}
	}
	return nil
}

// GetExitByName 根据名称获取出口
func (s *Scene) GetExitByName(name string) *Exit {
	for i := range s.Exits {
		if s.Exits[i].Name == name {
			return &s.Exits[i]
		}
	}
	return nil
}

// GetFeature 获取指定ID的特性
func (s *Scene) GetFeature(id string) *Feature {
	for i := range s.Features {
		if s.Features[i].ID == id {
			return &s.Features[i]
		}
	}
	return nil
}

// AddNPC 添加NPC
func (s *Scene) AddNPC(npcID string) {
	for _, id := range s.NPCs {
		if id == npcID {
			return
		}
	}
	s.NPCs = append(s.NPCs, npcID)
}

// RemoveNPC 移除NPC
func (s *Scene) RemoveNPC(npcID string) {
	for i, id := range s.NPCs {
		if id == npcID {
			s.NPCs = append(s.NPCs[:i], s.NPCs[i+1:]...)
			return
		}
	}
}

// AddItem 添加物品
func (s *Scene) AddItem(itemID string) {
	for _, id := range s.Items {
		if id == itemID {
			return
		}
	}
	s.Items = append(s.Items, itemID)
}

// RemoveItem 移除物品
func (s *Scene) RemoveItem(itemID string) {
	for i, id := range s.Items {
		if id == itemID {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
			return
		}
	}
}

// HasTag 检查是否有指定标签
func (s *Scene) HasTag(tag string) bool {
	for _, t := range s.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// SetProperty 设置自定义属性
func (s *Scene) SetProperty(key string, value any) {
	if s.Properties == nil {
		s.Properties = make(map[string]any)
	}
	s.Properties[key] = value
}

// GetProperty 获取自定义属性
func (s *Scene) GetProperty(key string) (any, bool) {
	if s.Properties == nil {
		return nil, false
	}
	v, ok := s.Properties[key]
	return v, ok
}
