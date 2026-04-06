package tools

import (
	"context"
	"errors"

	"github.com/zwh8800/cdnd/domain/character"
	"github.com/zwh8800/cdnd/domain/combat"
	"github.com/zwh8800/cdnd/domain/llm"
)

// StateAccessor 状态访问接口（解耦 tools 和 game 包）
type StateAccessor interface {
	// 角色相关
	GetCharacter() *character.Character

	// 世界标志
	GetFlag(key string) bool
	SetFlag(key string, value bool)

	// 计数器
	GetCounter(key string) int
	SetCounter(key string, value int)

	// 当前选项
	SetCurrentOptions(options []string)
	GetCurrentOptions() []string

	// 战斗相关
	GetCombat() *combat.CombatState
	StartCombat(participants []*combat.Combatant)
	EndCombat()
	NextTurn() *combat.Combatant
	GetCurrentCombatant() *combat.Combatant
	GetCombatant(id string) *combat.Combatant
	RemoveCombatant(id string) bool
	GetEnemies() []*combat.Combatant
	GetAliveEnemies() []*combat.Combatant
	IsCombatOver() (over bool, victory bool)
	AddCombatHistory(msg llm.Message)
	GetCombatHistory() []llm.Message
	ClearCombatHistory()
}

// Tool 工具接口
type Tool interface {
	// Name 工具名称
	Name() string
	// Description 工具描述 (LLM 可见)
	Description() string
	// Parameters JSON Schema 参数定义
	Parameters() map[string]interface{}
	// Execute 执行工具
	Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
}

// ToolResult 工具执行结果
type ToolResult struct {
	Success   bool        `json:"success"`         // 是否成功
	Data      interface{} `json:"data"`            // 返回数据
	Narrative string      `json:"narrative"`       // 用于叙述的文本
	Error     string      `json:"error,omitempty"` // 错误信息
}

// ToolDefinition 工具定义（用于 LLM API）
type ToolDefinition struct {
	Type     string       `json:"type"`     // 工具类型（function）
	Function ToolFunction `json:"function"` // 函数定义
}

// ToolFunction 函数定义
type ToolFunction struct {
	Name        string                 `json:"name"`        // 函数名称
	Description string                 `json:"description"` // 函数描述
	Parameters  map[string]interface{} `json:"parameters"`  // 参数定义（JSON Schema）
}

// ToolCall 工具调用请求
type ToolCall struct {
	ID        string                 `json:"id"`        // 唯一调用ID
	Name      string                 `json:"name"`      // 工具名称
	Arguments map[string]interface{} `json:"arguments"` // JSON格式的参数
}

// ToDefinition 将工具转换为 LLM API 定义
func ToDefinition(tool Tool) *ToolDefinition {
	return &ToolDefinition{
		Type: "function",
		Function: ToolFunction{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  tool.Parameters(),
		},
	}
}

// BaseTool 基础工具结构（可选嵌入）
type BaseTool struct {
	name        string // 工具名称
	description string // 工具描述
}

// NewBaseTool 创建基础工具
func NewBaseTool(name, description string) BaseTool {
	return BaseTool{name: name, description: description}
}

// Name 返回工具名称
func (t BaseTool) Name() string {
	return t.name
}

// Description 返回工具描述
func (t BaseTool) Description() string {
	return t.description
}

// 错误定义
var (
	ErrNotImplemented    = errors.New("工具未实现")
	ErrInvalidArguments  = errors.New("无效的参数")
	ErrPermissionDenied  = errors.New("权限不足")
	ErrToolNotFound      = errors.New("工具不存在")
	ErrStateNotAvailable = errors.New("游戏状态不可用")
)
