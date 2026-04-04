package world

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Manager 世界管理器
type Manager struct {
	mu     sync.RWMutex
	scenes map[string]*Scene
	npcs   map[string]*NPC
}

// NewManager 创建新的世界管理器
func NewManager() *Manager {
	return &Manager{
		scenes: make(map[string]*Scene),
		npcs:   make(map[string]*NPC),
	}
}

// AddScene 添加场景
func (m *Manager) AddScene(scene *Scene) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if scene.ID == "" {
		scene.ID = uuid.New().String()
	}
	scene.CreatedAt = time.Now()
	scene.UpdatedAt = time.Now()
	m.scenes[scene.ID] = scene
}

// GetScene 获取场景
func (m *Manager) GetScene(id string) *Scene {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.scenes[id]
}

// RemoveScene 移除场景
func (m *Manager) RemoveScene(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.scenes, id)
}

// ListScenes 列出所有场景
func (m *Manager) ListScenes() []*Scene {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Scene, 0, len(m.scenes))
	for _, s := range m.scenes {
		result = append(result, s)
	}
	return result
}

// AddNPC 添加NPC
func (m *Manager) AddNPC(npc *NPC) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if npc.ID == "" {
		npc.ID = uuid.New().String()
	}
	npc.CreatedAt = time.Now()
	npc.UpdatedAt = time.Now()
	m.npcs[npc.ID] = npc
}

// GetNPC 获取NPC
func (m *Manager) GetNPC(id string) *NPC {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.npcs[id]
}

// RemoveNPC 移除NPC
func (m *Manager) RemoveNPC(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.npcs, id)
}

// ListNPCs 列出所有NPC
func (m *Manager) ListNPCs() []*NPC {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*NPC, 0, len(m.npcs))
	for _, n := range m.npcs {
		result = append(result, n)
	}
	return result
}

// SpawnNPC 在场景中生成NPC
func (m *Manager) SpawnNPC(sceneID string, npcID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	scene, ok := m.scenes[sceneID]
	if !ok {
		return false
	}

	_, ok = m.npcs[npcID]
	if !ok {
		return false
	}

	scene.AddNPC(npcID)
	scene.UpdatedAt = time.Now()
	return true
}

// DespawnNPC 从场景中移除NPC
func (m *Manager) DespawnNPC(sceneID string, npcID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	scene, ok := m.scenes[sceneID]
	if !ok {
		return false
	}

	scene.RemoveNPC(npcID)
	scene.UpdatedAt = time.Now()
	return true
}

// GetSceneNPCs 获取场景中的所有NPC
func (m *Manager) GetSceneNPCs(sceneID string) []*NPC {
	m.mu.RLock()
	defer m.mu.RUnlock()

	scene, ok := m.scenes[sceneID]
	if !ok {
		return nil
	}

	npcs := make([]*NPC, 0, len(scene.NPCs))
	for _, npcID := range scene.NPCs {
		if npc, ok := m.npcs[npcID]; ok {
			npcs = append(npcs, npc)
		}
	}
	return npcs
}

// MoveNPC 移动NPC到另一个场景
func (m *Manager) MoveNPC(npcID string, fromSceneID string, toSceneID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	fromScene, ok := m.scenes[fromSceneID]
	if !ok {
		return false
	}

	toScene, ok := m.scenes[toSceneID]
	if !ok {
		return false
	}

	_, ok = m.npcs[npcID]
	if !ok {
		return false
	}

	fromScene.RemoveNPC(npcID)
	toScene.AddNPC(npcID)
	fromScene.UpdatedAt = time.Now()
	toScene.UpdatedAt = time.Now()
	return true
}

// LinkScenes 连接两个场景（双向）
func (m *Manager) LinkScenes(scene1ID string, scene2ID string, name1 string, name2 string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	scene1, ok1 := m.scenes[scene1ID]
	scene2, ok2 := m.scenes[scene2ID]
	if !ok1 || !ok2 {
		return false
	}

	// 场景1到场景2的出口
	scene1.Exits = append(scene1.Exits, Exit{
		ID:          uuid.New().String(),
		Name:        name1,
		TargetScene: scene2ID,
		Visible:     true,
		OneWay:      false,
	})

	// 场景2到场景1的出口
	scene2.Exits = append(scene2.Exits, Exit{
		ID:          uuid.New().String(),
		Name:        name2,
		TargetScene: scene1ID,
		Visible:     true,
		OneWay:      false,
	})

	scene1.UpdatedAt = time.Now()
	scene2.UpdatedAt = time.Now()
	return true
}

// LinkScenesOneWay 单向连接两个场景
func (m *Manager) LinkScenesOneWay(fromSceneID string, toSceneID string, name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	fromScene, ok1 := m.scenes[fromSceneID]
	_, ok2 := m.scenes[toSceneID]
	if !ok1 || !ok2 {
		return false
	}

	fromScene.Exits = append(fromScene.Exits, Exit{
		ID:          uuid.New().String(),
		Name:        name,
		TargetScene: toSceneID,
		Visible:     true,
		OneWay:      true,
	})

	fromScene.UpdatedAt = time.Now()
	return true
}

// GetConnectedScenes 获取与指定场景相连的所有场景
func (m *Manager) GetConnectedScenes(sceneID string) []*Scene {
	m.mu.RLock()
	defer m.mu.RUnlock()

	scene, ok := m.scenes[sceneID]
	if !ok {
		return nil
	}

	connected := make([]*Scene, 0, len(scene.Exits))
	for _, exit := range scene.Exits {
		if targetScene, ok := m.scenes[exit.TargetScene]; ok {
			connected = append(connected, targetScene)
		}
	}
	return connected
}

// Clear 清空所有数据
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.scenes = make(map[string]*Scene)
	m.npcs = make(map[string]*NPC)
}

// Import 导入世界数据
func (m *Manager) Import(scenes []*Scene, npcs []*NPC) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, s := range scenes {
		m.scenes[s.ID] = s
	}
	for _, n := range npcs {
		m.npcs[n.ID] = n
	}
}

// Export 导出世界数据
func (m *Manager) Export() ([]*Scene, []*NPC) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	scenes := make([]*Scene, 0, len(m.scenes))
	for _, s := range m.scenes {
		scenes = append(scenes, s)
	}

	npcs := make([]*NPC, 0, len(m.npcs))
	for _, n := range m.npcs {
		npcs = append(npcs, n)
	}

	return scenes, npcs
}
