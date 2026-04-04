package save

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	MaxSlots       = 10
	SaveDirName    = "cdnd"
	SaveSubDirName = "saves"
	SaveFileExt    = ".json"
)

// Manager 存档管理器
type Manager struct {
	mu      sync.RWMutex
	saveDir string
	cache   map[int]*SaveData
}

// NewManager 创建新的存档管理器
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("无法获取用户目录: %w", err)
	}

	saveDir := filepath.Join(homeDir, "."+SaveDirName, SaveSubDirName)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return nil, fmt.Errorf("无法创建存档目录: %w", err)
	}

	return &Manager{
		saveDir: saveDir,
		cache:   make(map[int]*SaveData),
	}, nil
}

// NewManagerWithPath 使用指定路径创建存档管理器
func NewManagerWithPath(saveDir string) (*Manager, error) {
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return nil, fmt.Errorf("无法创建存档目录: %w", err)
	}

	return &Manager{
		saveDir: saveDir,
		cache:   make(map[int]*SaveData),
	}, nil
}

// Save 保存存档
func (m *Manager) Save(slot int, data *SaveData) error {
	if slot < 1 || slot > MaxSlots {
		return fmt.Errorf("无效的存档槽位: %d (有效范围: 1-%d)", slot, MaxSlots)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 更新时间戳
	data.UpdatedAt = time.Now()
	data.Slot = slot

	// 序列化为JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化存档失败: %w", err)
	}

	// 写入文件
	filePath := m.getSaveFilePath(slot)
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("写入存档文件失败: %w", err)
	}

	// 更新缓存
	m.cache[slot] = data

	return nil
}

// Load 加载存档
func (m *Manager) Load(slot int) (*SaveData, error) {
	if slot < 1 || slot > MaxSlots {
		return nil, fmt.Errorf("无效的存档槽位: %d (有效范围: 1-%d)", slot, MaxSlots)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查缓存
	if data, ok := m.cache[slot]; ok {
		return data, nil
	}

	// 读取文件
	filePath := m.getSaveFilePath(slot)
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("存档槽位 %d 为空", slot)
		}
		return nil, fmt.Errorf("读取存档文件失败: %w", err)
	}

	// 反序列化
	var data SaveData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("解析存档数据失败: %w", err)
	}

	// 更新缓存
	m.cache[slot] = &data

	return &data, nil
}

// Delete 删除存档
func (m *Manager) Delete(slot int) error {
	if slot < 1 || slot > MaxSlots {
		return fmt.Errorf("无效的存档槽位: %d (有效范围: 1-%d)", slot, MaxSlots)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	filePath := m.getSaveFilePath(slot)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除存档文件失败: %w", err)
	}

	// 清除缓存
	delete(m.cache, slot)

	return nil
}

// ListSlots 列出所有存档槽位信息
func (m *Manager) ListSlots() ([]*SaveSlot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	slots := make([]*SaveSlot, 0, MaxSlots)

	for slot := 1; slot <= MaxSlots; slot++ {
		// 检查缓存
		if data, ok := m.cache[slot]; ok {
			slots = append(slots, data.ToSlot())
			continue
		}

		// 尝试读取文件
		filePath := m.getSaveFilePath(slot)
		jsonData, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				// 空槽位
				slots = append(slots, &SaveSlot{Slot: slot})
				continue
			}
			return nil, fmt.Errorf("读取存档 %d 失败: %w", slot, err)
		}

		var data SaveData
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return nil, fmt.Errorf("解析存档 %d 失败: %w", slot, err)
		}

		slots = append(slots, data.ToSlot())
		// 更新缓存
		m.cache[slot] = &data
	}

	return slots, nil
}

// GetEmptySlots 获取空存档槽位列表
func (m *Manager) GetEmptySlots() ([]int, error) {
	slots, err := m.ListSlots()
	if err != nil {
		return nil, err
	}

	empty := make([]int, 0)
	for _, s := range slots {
		if s.CharacterName == "" {
			empty = append(empty, s.Slot)
		}
	}

	return empty, nil
}

// GetUsedSlots 获取已使用的存档槽位列表
func (m *Manager) GetUsedSlots() ([]int, error) {
	slots, err := m.ListSlots()
	if err != nil {
		return nil, err
	}

	used := make([]int, 0)
	for _, s := range slots {
		if s.CharacterName != "" {
			used = append(used, s.Slot)
		}
	}

	return used, nil
}

// Exists 检查存档是否存在
func (m *Manager) Exists(slot int) bool {
	if slot < 1 || slot > MaxSlots {
		return false
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// 检查缓存
	if _, ok := m.cache[slot]; ok {
		return true
	}

	// 检查文件
	filePath := m.getSaveFilePath(slot)
	_, err := os.Stat(filePath)
	return err == nil
}

// GetMetadata 获取存档元数据
func (m *Manager) GetMetadata() (*SaveMetadata, error) {
	slots, err := m.ListSlots()
	if err != nil {
		return nil, err
	}

	meta := &SaveMetadata{}
	for _, s := range slots {
		if s.CharacterName != "" {
			meta.TotalSaves++
			meta.TotalPlayTime += s.PlayTime
			if s.UpdatedAt.After(meta.LastPlayed) {
				meta.LastPlayed = s.UpdatedAt
			}
		}
	}
	meta.Version = "1.0.0"

	return meta, nil
}

// QuickSave 快速保存（使用第一个可用槽位或最近的槽位）
func (m *Manager) QuickSave(data *SaveData) (int, error) {
	// 查找空槽位
	empty, err := m.GetEmptySlots()
	if err != nil {
		return 0, err
	}

	var slot int
	if len(empty) > 0 {
		slot = empty[0]
	} else {
		// 所有槽位已满，覆盖最旧的
		slots, err := m.ListSlots()
		if err != nil {
			return 0, err
		}
		// 按更新时间排序
		sort.Slice(slots, func(i, j int) bool {
			return slots[i].UpdatedAt.Before(slots[j].UpdatedAt)
		})
		slot = slots[0].Slot
	}

	if err := m.Save(slot, data); err != nil {
		return 0, err
	}

	return slot, nil
}

// QuickLoad 快速加载（加载最近的存档）
func (m *Manager) QuickLoad() (*SaveData, error) {
	slots, err := m.ListSlots()
	if err != nil {
		return nil, err
	}

	// 找到最近的存档
	var latestSlot *SaveSlot
	for _, s := range slots {
		if s.CharacterName != "" {
			if latestSlot == nil || s.UpdatedAt.After(latestSlot.UpdatedAt) {
				latestSlot = s
			}
		}
	}

	if latestSlot == nil {
		return nil, fmt.Errorf("没有可用的存档")
	}

	return m.Load(latestSlot.Slot)
}

// ClearCache 清除缓存
func (m *Manager) ClearCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache = make(map[int]*SaveData)
}

// GetSaveDir 获取存档目录路径
func (m *Manager) GetSaveDir() string {
	return m.saveDir
}

// getSaveFilePath 获取存档文件路径
func (m *Manager) getSaveFilePath(slot int) string {
	return filepath.Join(m.saveDir, fmt.Sprintf("slot_%d%s", slot, SaveFileExt))
}

// ImportSave 从文件导入存档
func (m *Manager) ImportSave(filePath string, slot int) error {
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取导入文件失败: %w", err)
	}

	var data SaveData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("解析导入数据失败: %w", err)
	}

	return m.Save(slot, &data)
}

// ExportSave 导出存档到文件
func (m *Manager) ExportSave(slot int, filePath string) error {
	data, err := m.Load(slot)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化存档失败: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("写入导出文件失败: %w", err)
	}

	return nil
}
