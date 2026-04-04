package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zwh8800/cdnd/internal/character"
	"github.com/zwh8800/cdnd/internal/config"
	"github.com/zwh8800/cdnd/internal/llm"
	"github.com/zwh8800/cdnd/internal/llm/prompt"
	"github.com/zwh8800/cdnd/internal/rules"
	"github.com/zwh8800/cdnd/internal/save"
	"github.com/zwh8800/cdnd/internal/tools"
	"github.com/zwh8800/cdnd/internal/world"
	"github.com/zwh8800/cdnd/pkg/dice"
)

// Engine 游戏引擎
type Engine struct {
	mu           sync.RWMutex
	state        *State
	llmProvider  llm.Provider
	prompt       *prompt.Builder
	rules        *rules.Engine
	world        *world.Manager
	save         *save.Manager
	toolRegistry *tools.Registry
	events       *EventDispatcher
	config       *config.Config
}

// NewEngine 创建新的游戏引擎
func NewEngine(cfg *config.Config, provider llm.Provider) (*Engine, error) {
	saveManager, err := save.NewManager()
	if err != nil {
		return nil, fmt.Errorf("创建存档管理器失败: %w", err)
	}

	engine := &Engine{
		state:        NewState(),
		llmProvider:  provider,
		prompt:       prompt.NewBuilder(),
		rules:        rules.NewEngine(),
		world:        world.NewManager(),
		save:         saveManager,
		toolRegistry: tools.NewRegistry(),
		events:       NewEventDispatcher(),
		config:       cfg,
	}

	engine.registerTools()
	return engine, nil
}

// registerTools 注册所有DM工具
func (e *Engine) registerTools() {
	e.toolRegistry.Register(tools.NewRollDiceTool())
	e.toolRegistry.Register(tools.NewSkillCheckTool(e.state, e.rules))
	e.toolRegistry.Register(tools.NewSavingThrowTool(e.state, e.rules))
	e.toolRegistry.Register(tools.NewDealDamageTool(e.state))
	e.toolRegistry.Register(tools.NewHealCharacterTool(e.state))
	e.toolRegistry.Register(tools.NewAddConditionTool(e.state))
	e.toolRegistry.Register(tools.NewRemoveConditionTool(e.state))
	e.toolRegistry.Register(tools.NewAddItemTool(e.state))
	e.toolRegistry.Register(tools.NewRemoveItemTool(e.state))
	e.toolRegistry.Register(tools.NewSpendGoldTool(e.state))
	e.toolRegistry.Register(tools.NewGainGoldTool(e.state))
	e.toolRegistry.Register(tools.NewMoveToSceneTool(e.state))
	e.toolRegistry.Register(tools.NewSpawnNPCTool(e.state))
	e.toolRegistry.Register(tools.NewRemoveNPCTool(e.state))
	e.toolRegistry.Register(tools.NewSetFlagTool(e.state))
	e.toolRegistry.Register(tools.NewGetFlagTool(e.state))
}

// Start 开始新游戏
func (e *Engine) Start(c *character.Character) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.state = NewState()
	e.state.SetCharacter(c)
	e.state.SetPhase(save.PhaseIntroduction)
	return nil
}

// LoadGame 加载游戏
func (e *Engine) LoadGame(slot int) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	data, err := e.save.Load(slot)
	if err != nil {
		return fmt.Errorf("加载存档失败: %w", err)
	}
	e.state = &State{
		SessionID:     data.SessionID,
		Phase:         data.Phase,
		TurnCount:     data.TurnCount,
		Character:     data.Character,
		CurrentScene:  data.CurrentScene,
		VisitedScenes: data.VisitedScenes,
		WorldFlags:    data.WorldFlags,
		WorldCounters: data.WorldCounters,
		Quests:        data.Quests,
		History:       data.History,
		DMContext:     data.DMContext,
		Combat:        data.Combat,
		CreatedAt:     data.CreatedAt,
		LastSavedAt:   data.UpdatedAt,
		PlayedTime:    data.PlayTime,
	}
	scenes, npcs := data.GetWorldData()
	e.world.Import(scenes, npcs)
	return nil
}

// SaveGame 保存游戏
func (e *Engine) SaveGame(slot int) error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	data := &save.SaveData{
		Slot:          slot,
		SaveName:      e.state.Character.Name,
		CreatedAt:     e.state.CreatedAt,
		UpdatedAt:     time.Now(),
		PlayTime:      e.state.PlayedTime,
		SessionID:     e.state.SessionID,
		Phase:         e.state.Phase,
		TurnCount:     e.state.TurnCount,
		Character:     e.state.Character,
		CurrentScene:  e.state.CurrentScene,
		VisitedScenes: e.state.VisitedScenes,
		WorldFlags:    e.state.WorldFlags,
		WorldCounters: e.state.WorldCounters,
		Quests:        e.state.Quests,
		History:       e.state.History,
		DMContext:     e.state.DMContext,
		Combat:        e.state.Combat,
		Version:       "1.0.0",
	}
	if e.world != nil {
		data.Scenes, data.NPCs = e.world.Export()
	}
	return e.save.Save(slot, data)
}

// GetState 获取游戏状态
func (e *Engine) GetState() *State {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.state
}

// GetCharacter 获取角色
func (e *Engine) GetCharacter() *character.Character {
	return e.state.GetCharacter()
}

// GetCurrentScene 获取当前场景
func (e *Engine) GetCurrentScene() *world.Scene {
	return e.state.GetCurrentScene()
}

// ProcessPlayerInput 处理玩家输入
func (e *Engine) ProcessPlayerInput(ctx context.Context, input string) (*DMResponse, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.state.IncrementTurn()
	gameCtx := &prompt.GameContext{
		Phase:         e.state.GetPhase().String(),
		Character:     e.state.GetCharacter(),
		CurrentScene:  e.state.GetCurrentScene(),
		DMContext:     e.state.DMContext,
		History:       e.state.GetHistory(),
		TurnCount:     e.state.TurnCount,
		WorldFlags:    e.state.WorldFlags,
		WorldCounters: e.state.WorldCounters,
	}
	messages := e.buildMessages(gameCtx, input)
	resp, err := e.llmProvider.Generate(ctx, &llm.Request{Messages: messages})
	if err != nil {
		return nil, fmt.Errorf("LLM调用失败: %w", err)
	}
	e.state.AddHistory(llm.Message{Role: llm.RoleUser, Content: input})
	e.state.AddHistory(llm.Message{Role: llm.RoleAssistant, Content: resp.Content})
	return &DMResponse{Content: resp.Content, Phase: e.state.GetPhase()}, nil
}

// ProcessWithTools 使用Tool Call处理玩家输入
func (e *Engine) ProcessWithTools(ctx context.Context, input string) (*DMResponse, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.state.IncrementTurn()
	gameCtx := &prompt.GameContext{
		Phase:        e.state.GetPhase().String(),
		Character:    e.state.GetCharacter(),
		CurrentScene: e.state.GetCurrentScene(),
		DMContext:    e.state.DMContext,
		History:      e.state.GetHistory(),
		TurnCount:    e.state.TurnCount,
	}
	systemPrompt := e.prompt.BuildSystemPrompt(gameCtx)
	messages := []llm.Message{{Role: llm.RoleSystem, Content: systemPrompt}}
	messages = append(messages, e.prompt.BuildHistoryContext(e.state.GetHistory(), 20)...)
	messages = append(messages, llm.Message{Role: llm.RoleUser, Content: input})
	resp, err := e.llmProvider.Generate(ctx, &llm.Request{Messages: messages})
	if err != nil {
		return nil, err
	}
	e.state.AddHistory(llm.Message{Role: llm.RoleUser, Content: input})
	e.state.AddHistory(llm.Message{Role: llm.RoleAssistant, Content: resp.Content})
	return &DMResponse{Content: resp.Content, Phase: e.state.GetPhase()}, nil
}

// buildMessages 构建LLM消息
func (e *Engine) buildMessages(ctx *prompt.GameContext, input string) []llm.Message {
	messages := []llm.Message{{Role: llm.RoleSystem, Content: e.prompt.BuildSystemPrompt(ctx)}}
	messages = append(messages, e.prompt.BuildHistoryContext(ctx.History, 20)...)
	messages = append(messages, llm.Message{Role: llm.RoleUser, Content: input})
	return messages
}

// RollDice 投骰子
func (e *Engine) RollDice(notation string) (*dice.Result, error) {
	result, err := dice.ParseAndRoll(notation)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// SkillCheck 技能检定
func (e *Engine) SkillCheck(skill character.SkillType, dc int, advantage bool) *rules.CheckResult {
	rollType := dice.NormalRoll
	if advantage {
		rollType = dice.AdvantageRoll
	}
	return e.rules.SkillCheck(e.state.GetCharacter(), skill, dc, rollType)
}

// SavingThrow 豁免检定
func (e *Engine) SavingThrow(ability character.Ability, dc int, advantage bool) *rules.CheckResult {
	rollType := dice.NormalRoll
	if advantage {
		rollType = dice.AdvantageRoll
	}
	return e.rules.SavingThrow(e.state.GetCharacter(), ability, dc, rollType)
}

// SetPhase 设置游戏阶段
func (e *Engine) SetPhase(phase save.GamePhase) {
	e.state.SetPhase(phase)
	e.events.Dispatch(Event{Type: EventPhaseChanged, Data: phase, Message: fmt.Sprintf("进入%s阶段", phase.String())})
}

// SetScene 设置当前场景
func (e *Engine) SetScene(scene *world.Scene) {
	e.state.SetCurrentScene(scene)
	e.events.Dispatch(Event{Type: EventSceneChanged, Target: scene.ID, Message: fmt.Sprintf("进入: %s", scene.Name)})
}

// TakeDamage 角色受到伤害
func (e *Engine) TakeDamage(amount int) int {
	c := e.state.GetCharacter()
	if c == nil {
		return 0
	}
	oldHP := c.HitPoints.Current
	c.HitPoints.TakeDamage(amount)
	e.events.Dispatch(Event{Type: EventCharacterDamaged, Data: map[string]int{"old_hp": oldHP, "new_hp": c.HitPoints.Current, "damage": amount}, Message: fmt.Sprintf("%s 受到 %d 点伤害", c.Name, amount)})
	return c.HitPoints.Current
}

// Heal 角色治疗
func (e *Engine) Heal(amount int) int {
	c := e.state.GetCharacter()
	if c == nil {
		return 0
	}
	oldHP := c.HitPoints.Current
	c.HitPoints.Heal(amount)
	e.events.Dispatch(Event{Type: EventCharacterHealed, Data: map[string]int{"old_hp": oldHP, "new_hp": c.HitPoints.Current, "heal": amount}, Message: fmt.Sprintf("%s 恢复 %d 点生命值", c.Name, amount)})
	return c.HitPoints.Current
}

// SubscribeEvent 订阅事件
func (e *Engine) SubscribeEvent(eventType EventType, handler EventHandler) {
	e.events.Subscribe(eventType, handler)
}

// GetToolDefinitions 获取工具定义
func (e *Engine) GetToolDefinitions() []*tools.ToolDefinition {
	return e.toolRegistry.GetToolDefinitions()
}

// ExecuteTool 执行工具
func (e *Engine) ExecuteTool(ctx context.Context, name string, args map[string]interface{}) (*tools.ToolResult, error) {
	return e.toolRegistry.Execute(ctx, name, args)
}

// GetSaveSlots 获取存档槽位列表
func (e *Engine) GetSaveSlots() ([]*save.SaveSlot, error) {
	return e.save.ListSlots()
}

// DMResponse DM响应
type DMResponse struct {
	Content   string           `json:"content"`
	Phase     save.GamePhase   `json:"phase"`
	ToolCalls []tools.ToolCall `json:"tool_calls,omitempty"`
}
