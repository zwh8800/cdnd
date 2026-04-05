package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// Registry 工具注册表
type Registry struct {
	tools       map[string]Tool
	permissions map[string][]string // 工具 -> 允许的游戏阶段
}

// NewRegistry 创建新的工具注册表
func NewRegistry() *Registry {
	return &Registry{
		tools:       make(map[string]Tool),
		permissions: make(map[string][]string),
	}
}

// Register 注册工具
func (r *Registry) Register(tool Tool, allowedPhases ...string) {
	r.tools[tool.Name()] = tool
	if len(allowedPhases) > 0 {
		r.permissions[tool.Name()] = allowedPhases
	}
}

// Get 获取工具
func (r *Registry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// Execute 执行工具
func (r *Registry) Execute(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	tool, ok := r.tools[name]

	if !ok {
		return nil, ErrToolNotFound
	}

	return tool.Execute(ctx, args)
}

// ExecuteFromJSON 从 JSON 执行工具
func (r *Registry) ExecuteFromJSON(ctx context.Context, name string, argsJSON string) (*ToolResult, error) {
	var args map[string]interface{}
	if argsJSON != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return nil, fmt.Errorf("解析参数失败: %w", err)
		}
	}
	return r.Execute(ctx, name, args)
}

// GetToolDefinitions 获取所有工具定义（用于 LLM API）
func (r *Registry) GetToolDefinitions() []*ToolDefinition {
	definitions := make([]*ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		definitions = append(definitions, ToDefinition(tool))
	}
	return definitions
}

// ListTools 列出所有工具名称
func (r *Registry) ListTools() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// HasTool 检查工具是否存在
func (r *Registry) HasTool(name string) bool {
	_, ok := r.tools[name]
	return ok
}

// IsAllowedInPhase 检查工具在指定阶段是否允许
func (r *Registry) IsAllowedInPhase(name string, phase string) bool {
	phases, ok := r.permissions[name]
	if !ok {
		// 没有权限限制，默认允许
		return true
	}

	for _, p := range phases {
		if p == phase || p == "*" {
			return true
		}
	}
	return false
}

// Clear 清除所有工具
func (r *Registry) Clear() {
	r.tools = make(map[string]Tool)
	r.permissions = make(map[string][]string)
}

// ToolCount 返回工具数量
func (r *Registry) ToolCount() int {
	return len(r.tools)
}
