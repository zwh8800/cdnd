package events

import (
	"sync"
)

// EventType 事件类型
type EventType int

const (
	EventNone EventType = iota
	// 角色事件
	EventCharacterCreated
	EventCharacterDamaged
	EventCharacterHealed
	EventCharacterDied
	EventConditionAdded
	EventConditionRemoved
	EventLevelUp
	// 物品事件
	EventItemAdded
	EventItemRemoved
	EventItemEquipped
	EventItemUsed
	// 场景事件
	EventSceneChanged
	EventSceneEntered
	EventSceneExited
	EventNPCSpawned
	EventNPCRemoved
	EventNPCInteract
	// 战斗事件
	EventCombatStarted
	EventCombatEnded
	EventTurnStarted
	EventTurnEnded
	EventAttackRolled
	EventDamageDealt
	// 任务事件
	EventQuestAdded
	EventQuestUpdated
	EventQuestCompleted
	// 工具事件
	EventToolExecuted
	// 系统事件
	EventGameSaved
	EventGameLoaded
	EventPhaseChanged
	EventError
)

// String 返回事件类型的中文名称
func (e EventType) String() string {
	switch e {
	case EventCharacterCreated:
		return "角色创建"
	case EventCharacterDamaged:
		return "角色受伤"
	case EventCharacterHealed:
		return "角色治疗"
	case EventCharacterDied:
		return "角色死亡"
	case EventConditionAdded:
		return "添加状态"
	case EventConditionRemoved:
		return "移除状态"
	case EventLevelUp:
		return "升级"
	case EventItemAdded:
		return "获得物品"
	case EventItemRemoved:
		return "失去物品"
	case EventItemEquipped:
		return "装备物品"
	case EventItemUsed:
		return "使用物品"
	case EventSceneChanged:
		return "场景变更"
	case EventSceneEntered:
		return "进入场景"
	case EventSceneExited:
		return "离开场景"
	case EventNPCSpawned:
		return "NPC出现"
	case EventNPCRemoved:
		return "NPC消失"
	case EventNPCInteract:
		return "NPC互动"
	case EventCombatStarted:
		return "战斗开始"
	case EventCombatEnded:
		return "战斗结束"
	case EventTurnStarted:
		return "回合开始"
	case EventTurnEnded:
		return "回合结束"
	case EventAttackRolled:
		return "攻击检定"
	case EventDamageDealt:
		return "造成伤害"
	case EventQuestAdded:
		return "任务添加"
	case EventQuestUpdated:
		return "任务更新"
	case EventQuestCompleted:
		return "任务完成"
	case EventToolExecuted:
		return "工具执行"
	case EventGameSaved:
		return "游戏保存"
	case EventGameLoaded:
		return "游戏加载"
	case EventPhaseChanged:
		return "阶段变更"
	case EventError:
		return "错误"
	default:
		return "未知事件"
	}
}

// Event 游戏事件
type Event struct {
	Type      EventType `json:"type"`
	Timestamp int64     `json:"timestamp"`
	Source    string    `json:"source,omitempty"`  // 事件来源
	Target    string    `json:"target,omitempty"`  // 事件目标
	Data      any       `json:"data,omitempty"`    // 事件数据
	Message   string    `json:"message,omitempty"` // 事件消息
}

// EventHandler 事件处理函数
type EventHandler func(Event)

// EventDispatcher 事件分发器
type EventDispatcher struct {
	mu       sync.RWMutex
	handlers map[EventType][]EventHandler
	queued   []Event
}

// NewEventDispatcher 创建新的事件分发器
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[EventType][]EventHandler),
		queued:   make([]Event, 0),
	}
}

// Subscribe 订阅事件
func (d *EventDispatcher) Subscribe(eventType EventType, handler EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

// Unsubscribe 取消订阅事件
func (d *EventDispatcher) Unsubscribe(eventType EventType, handler EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	handlers := d.handlers[eventType]
	for i, h := range handlers {
		// 使用指针比较
		if &h == &handler {
			d.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Dispatch 分发事件
func (d *EventDispatcher) Dispatch(event Event) {
	d.mu.RLock()
	handlers := d.handlers[event.Type]
	d.mu.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}
}

// DispatchSync 同步分发事件（阻塞直到所有处理器完成）
func (d *EventDispatcher) DispatchSync(event Event) {
	d.Dispatch(event)
}

// Queue 将事件加入队列
func (d *EventDispatcher) Queue(event Event) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.queued = append(d.queued, event)
}

// ProcessQueue 处理队列中的所有事件
func (d *EventDispatcher) ProcessQueue() {
	d.mu.Lock()
	queued := d.queued
	d.queued = make([]Event, 0)
	d.mu.Unlock()

	for _, event := range queued {
		d.Dispatch(event)
	}
}

// Clear 清除所有订阅
func (d *EventDispatcher) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers = make(map[EventType][]EventHandler)
	d.queued = make([]Event, 0)
}

// HasHandlers 检查是否有指定事件的处理器
func (d *EventDispatcher) HasHandlers(eventType EventType) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.handlers[eventType]) > 0
}

// NewEvent 创建新事件
func NewEvent(eventType EventType, data any) Event {
	return Event{
		Type:      eventType,
		Timestamp: currentTimeMillis(),
		Data:      data,
	}
}

// NewEventWithMessage 创建带消息的事件
func NewEventWithMessage(eventType EventType, message string, data any) Event {
	return Event{
		Type:      eventType,
		Timestamp: currentTimeMillis(),
		Data:      data,
		Message:   message,
	}
}

// currentTimeMillis 获取当前时间戳（毫秒）
func currentTimeMillis() int64 {
	return 0 // 简化实现，实际使用 time.Now().UnixMilli()
}
