package game

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
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
	// 不要创建新的 State 对象，而是重置现有对象
	// 这样工具中保存的 state 引用仍然有效
	e.state.SessionID = uuid.New().String()
	e.state.Phase = save.PhaseIntroduction
	e.state.TurnCount = 0
	e.state.SubTurn = 0
	e.state.Character = c
	e.state.CurrentScene = nil
	e.state.VisitedScenes = make(map[string]bool)
	e.state.WorldFlags = make(map[string]bool)
	e.state.WorldCounters = make(map[string]int)
	e.state.Quests = make([]*save.QuestState, 0)
	e.state.History = make([]llm.Message, 0)
	e.state.DMContext = ""
	e.state.Combat = nil
	e.state.CreatedAt = time.Now()
	e.state.LastSavedAt = time.Now()
	e.state.PlayedTime = 0
	return nil
}

// LoadGame 加载游戏
func (e *Engine) LoadGame(slot int) error {
	data, err := e.save.Load(slot)
	if err != nil {
		return fmt.Errorf("加载存档失败: %w", err)
	}

	// 检查角色数据是否有效
	if data.Character == nil {
		return fmt.Errorf("存档数据不完整：角色数据缺失")
	}

	// 初始化空的 map（如果存档中为空）
	if data.VisitedScenes == nil {
		data.VisitedScenes = make(map[string]bool)
	}
	if data.WorldFlags == nil {
		data.WorldFlags = make(map[string]bool)
	}
	if data.WorldCounters == nil {
		data.WorldCounters = make(map[string]int)
	}

	// 更新现有 state 对象的字段，而不是创建新对象
	// 这样工具中保存的 state 引用仍然有效
	e.state.SessionID = data.SessionID
	e.state.Phase = data.Phase
	e.state.TurnCount = data.TurnCount
	e.state.Character = data.Character
	e.state.CurrentScene = data.CurrentScene
	e.state.VisitedScenes = data.VisitedScenes
	e.state.WorldFlags = data.WorldFlags
	e.state.WorldCounters = data.WorldCounters
	e.state.Quests = data.Quests
	e.state.History = data.History
	e.state.DMContext = data.DMContext
	e.state.Combat = data.Combat
	e.state.CreatedAt = data.CreatedAt
	e.state.LastSavedAt = data.UpdatedAt
	e.state.PlayedTime = data.PlayTime

	// 验证角色是否正确加载
	if e.state.GetCharacter() == nil {
		return fmt.Errorf("角色加载失败：角色数据为空")
	}

	scenes, npcs := data.GetWorldData()
	e.world.Import(scenes, npcs)
	return nil
}

// SaveGame 保存游戏
func (e *Engine) SaveGame(slot int) error {
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

// Process 处理玩家输入，支持Tool Call
// 实现完整的 Agentic Loop：调用LLM -> 执行工具 -> 反馈结果 -> 循环
func (e *Engine) Process(ctx context.Context, input string) (*DMResponse, error) {
	e.state.IncrementTurn()
	e.state.ClearCurrentOptions()

	// 1. 获取工具定义并转换为 LLM 格式
	toolDefs := e.toolRegistry.GetToolDefinitions()
	llmToolDefs := make([]llm.ToolDefinition, len(toolDefs))
	for i, td := range toolDefs {
		llmToolDefs[i] = llm.ToolDefinition{
			Type: td.Type,
			Function: llm.ToolFunctionDefinition{
				Name:        td.Function.Name,
				Description: td.Function.Description,
				Parameters:  td.Function.Parameters,
			},
		}
	}

	// 2. 构建初始消息
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
	systemPrompt := e.prompt.BuildSystemPrompt(gameCtx)
	messages := []llm.Message{{Role: llm.RoleSystem, Content: systemPrompt}}
	messages = append(messages, e.prompt.BuildHistoryContext(e.state.GetHistory(), 20)...)
	messages = append(messages, llm.Message{Role: llm.RoleUser, Content: input})

	// 3. Agentic Loop (最多10次迭代)
	const maxIterations = 10
	var allToolCalls []tools.ToolCall
	var allNarratives []string

	for i := 0; i < maxIterations; i++ {
		// 3.1 调用 LLM
		resp, err := e.llmProvider.Generate(ctx, &llm.Request{
			Messages:   messages,
			Tools:      llmToolDefs,
			ToolChoice: "auto",
		})
		if err != nil {
			return nil, fmt.Errorf("LLM调用失败: %w", err)
		}

		// 3.2 检查是否有工具调用
		if len(resp.ToolCalls) == 0 {
			// 没有工具调用，返回最终响应
			e.state.AddHistory(llm.Message{Role: llm.RoleUser, Content: input})

			// 解析选项（仅用于提取选项列表，不改变原始内容）
			options, _ := prompt.ParseOptions(resp.Content)

			// 更新状态中的选项
			e.state.SetCurrentOptions(options)

			e.state.AddHistory(llm.Message{Role: llm.RoleAssistant, Content: resp.Content})
			return &DMResponse{
				Content:        resp.Content,
				Phase:          e.state.GetPhase(),
				ToolCalls:      allToolCalls,
				ToolNarratives: allNarratives,
				Options:        options,
			}, nil
		}

		// 3.3 添加 assistant 消息（包含工具调用）
		assistantMsg := llm.Message{
			Role:      llm.RoleAssistant,
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		}
		messages = append(messages, assistantMsg)

		// 3.4 执行所有工具调用并添加结果消息
		for _, tc := range resp.ToolCalls {
			// 解析参数
			var args map[string]interface{}
			if tc.Arguments != "" {
				if err := json.Unmarshal([]byte(tc.Arguments), &args); err != nil {
					args = make(map[string]interface{})
				}
			} else {
				args = make(map[string]interface{})
			}

			// 执行工具
			result, err := e.toolRegistry.Execute(ctx, tc.Name, args)

			// 生成 D&D 风格叙述
			narrative := e.generateToolNarrative(tc.Name, args, result, err)
			allNarratives = append(allNarratives, narrative)

			// 分发工具执行事件
			e.events.Dispatch(Event{
				Type:    EventToolExecuted,
				Target:  tc.ID,
				Data:    map[string]interface{}{"tool": tc.Name, "args": args, "result": result, "error": err},
				Message: narrative,
			})

			// 记录工具调用
			allToolCalls = append(allToolCalls, tools.ToolCall{
				ID:        tc.ID,
				Name:      tc.Name,
				Arguments: args,
			})

			// 添加工具结果消息
			toolMsg := llm.Message{
				Role:       llm.RoleTool,
				Content:    formatToolResult(result, err),
				ToolCallID: tc.ID,
				Name:       tc.Name,
			}
			messages = append(messages, toolMsg)
		}
	}

	// 超过最大迭代次数，返回错误
	return nil, fmt.Errorf("工具调用超过最大迭代次数 (%d)", maxIterations)
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
	Content        string           `json:"content"`
	Phase          save.GamePhase   `json:"phase"`
	ToolCalls      []tools.ToolCall `json:"tool_calls,omitempty"`
	ToolNarratives []string         `json:"tool_narratives,omitempty"` // D&D风格工具执行叙述
	Options        []string         `json:"options,omitempty"`         // 当前可用选项
}

// StreamChunk 流式响应数据块
type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}

// formatToolResult 格式化工具执行结果为消息内容
func formatToolResult(result *tools.ToolResult, err error) string {
	if err != nil {
		return fmt.Sprintf("工具执行错误: %s", err.Error())
	}
	if result == nil {
		return "工具执行完成（无结果）"
	}

	var sb strings.Builder
	if result.Success {
		sb.WriteString("成功: ")
	} else {
		sb.WriteString("失败: ")
	}

	if result.Narrative != "" {
		sb.WriteString(result.Narrative)
	}

	if result.Error != "" {
		sb.WriteString(" [")
		sb.WriteString(result.Error)
		sb.WriteString("]")
	}

	return sb.String()
}

// indentLines 给每一行文本添加指定的前缀缩进
func indentLines(text string, prefix string) string {
	lines := strings.Split(text, "\n")
	var result strings.Builder
	for _, line := range lines {
		if line != "" {
			result.WriteString(prefix + line + "\n")
		}
	}
	return result.String()
}

// generateToolNarrative 生成D&D风格的工具执行叙述
func (e *Engine) generateToolNarrative(toolName string, args map[string]interface{}, result *tools.ToolResult, execErr error) string {
	var sb strings.Builder

	// 获取工具类型分类
	toolCategory := getToolCategory(toolName)

	// 根据结果状态添加视觉标记
	var statusMarker string
	if execErr != nil {
		statusMarker = "❌ "
	} else if result != nil && result.Success {
		statusMarker = "✅ "
	} else {
		statusMarker = "⚠️ "
	}

	// 添加明显的工具调用标记
	sb.WriteString("\n")
	sb.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString(indentLines(fmt.Sprintf("⚙️  工具调用: %s", toolName), "  "))
	sb.WriteString("╠════════════════════════════════════════════════════════════════╣\n")

	// 生成叙述标题
	headerText := statusMarker + getToolNarrativeHeader(toolName, toolCategory)
	sb.WriteString(indentLines(headerText, "  "))

	// 根据工具类型生成不同的叙述内容
	var narrativeContent string
	switch toolCategory {
	case "dice":
		narrativeContent = generateDiceNarrative(toolName, args, result, execErr)
	case "character":
		narrativeContent = generateCharacterNarrative(toolName, args, result, execErr)
	case "item":
		narrativeContent = generateItemNarrative(toolName, args, result, execErr)
	case "world":
		narrativeContent = generateWorldNarrative(toolName, args, result, execErr)
	default:
		narrativeContent = generateGenericNarrative(toolName, args, result, execErr)
	}

	sb.WriteString(indentLines(narrativeContent, "  "))

	sb.WriteString("╚════════════════════════════════════════════════════════════════╝\n")

	return sb.String()
}

// getToolCategory 获取工具类型分类
func getToolCategory(toolName string) string {
	switch toolName {
	case "roll_dice", "skill_check", "saving_throw":
		return "dice"
	case "deal_damage", "heal_character", "add_condition", "remove_condition":
		return "character"
	case "add_item", "remove_item", "spend_gold", "gain_gold":
		return "item"
	case "move_to_scene", "spawn_npc", "remove_npc", "set_flag", "get_flag":
		return "world"
	default:
		return "generic"
	}
}

// getToolNarrativeHeader 获取工具叙述标题
func getToolNarrativeHeader(toolName, category string) string {
	headers := map[string]string{
		"roll_dice":        "🎲 骰子滚动",
		"skill_check":      "🎯 技能检定",
		"saving_throw":     "🛡️ 豁免检定",
		"deal_damage":      "⚔️ 造成伤害",
		"heal_character":   "💚 治疗恢复",
		"add_condition":    "🔮 施加状态",
		"remove_condition": "✨ 移除状态",
		"add_item":         "📦 获得物品",
		"remove_item":      "📤 失去物品",
		"spend_gold":       "💰 消耗金币",
		"gain_gold":        "💎 获得金币",
		"move_to_scene":    "🚶 场景转换",
		"spawn_npc":        "👤 NPC出现",
		"remove_npc":       "👥 NPC消失",
		"set_flag":         "🏁 设置标记",
		"get_flag":         "🔍 查询标记",
	}
	if header, ok := headers[toolName]; ok {
		return header
	}
	return "⚙️ " + toolName
}

// generateDiceNarrative 生成骰子类工具叙述
func generateDiceNarrative(toolName string, args map[string]interface{}, result *tools.ToolResult, execErr error) string {
	var sb strings.Builder

	if execErr != nil {
		sb.WriteString(fmt.Sprintf("  └─ 骰子投掷出现异常，命运之轮暂时停滞... (%s)\n", execErr.Error()))
		return sb.String()
	}

	if result == nil {
		return "  └─ 骰子滚动...\n"
	}

	// 提取骰子结果信息
	if data, ok := result.Data.(map[string]interface{}); ok {
		// 支持 int 和 float64 两种类型
		var total int
		var hasTotal bool
		if totalFloat, ok := data["total"].(float64); ok {
			total = int(totalFloat)
			hasTotal = true
		} else if totalInt, ok := data["total"].(int); ok {
			total = totalInt
			hasTotal = true
		}

		if toolName == "skill_check" || toolName == "saving_throw" {
			var dc int
			var hasDC bool
			if dcFloat, ok := data["dc"].(float64); ok {
				dc = int(dcFloat)
				hasDC = true
			} else if dcInt, ok := data["dc"].(int); ok {
				dc = dcInt
				hasDC = true
			}
			if hasTotal && hasDC {
				if result.Success {
					sb.WriteString(fmt.Sprintf("  └─ 🎉 检定成功！投出 %d (DC %d)\n", total, dc))
					sb.WriteString("     命运眷顾着这位冒险者...\n")
				} else {
					sb.WriteString(fmt.Sprintf("  └─ 💔 检定失败。投出 %d (DC %d)\n", total, dc))
					sb.WriteString("     命运之线似乎在此时断裂...\n")
				}
			} else if result.Narrative != "" {
				sb.WriteString(fmt.Sprintf("  └─ %s\n", result.Narrative))
			}
		} else if hasTotal {
			sb.WriteString(fmt.Sprintf("  └─ 骰子落地：**%d**\n", total))
		} else if result.Narrative != "" {
			sb.WriteString(fmt.Sprintf("  └─ %s\n", result.Narrative))
		}
	} else if result.Narrative != "" {
		sb.WriteString(fmt.Sprintf("  └─ %s\n", result.Narrative))
	}

	return sb.String()
}

// generateCharacterNarrative 生成角色类工具叙述
func generateCharacterNarrative(toolName string, args map[string]interface{}, result *tools.ToolResult, execErr error) string {
	var sb strings.Builder

	if execErr != nil {
		sb.WriteString(fmt.Sprintf("  └─ ⚡ 能量波动，%s 失败...\n", toolName))
		return sb.String()
	}

	if result == nil {
		return "  └─ 角色状态变化中...\n"
	}

	switch toolName {
	case "deal_damage":
		if amount, ok := args["amount"].(float64); ok {
			if result.Success {
				sb.WriteString(fmt.Sprintf("  └─ ⚔️ 造成了 %d 点伤害！\n", int(amount)))
				sb.WriteString("     伤口鲜血直流，痛苦的呻吟回荡在空气中...\n")
			} else {
				sb.WriteString("  └─ 攻击未能造成有效伤害\n")
			}
		}
	case "heal_character":
		if amount, ok := args["amount"].(float64); ok {
			if result.Success {
				sb.WriteString(fmt.Sprintf("  └─ 💚 恢复了 %d 点生命值！\n", int(amount)))
				sb.WriteString("     神圣的光芒笼罩全身，伤口开始愈合...\n")
			} else {
				sb.WriteString("  └─ 治疗效果未能生效\n")
			}
		}
	case "add_condition":
		if condition, ok := args["condition"].(string); ok {
			if result.Success {
				sb.WriteString(fmt.Sprintf("  └─ 🔮 被施加了【%s】状态！\n", condition))
				sb.WriteString("     诡异的力量缠绕着身体，状态发生了改变...\n")
			}
		}
	case "remove_condition":
		if condition, ok := args["condition"].(string); ok {
			if result.Success {
				sb.WriteString(fmt.Sprintf("  └─ ✨ 【%s】状态已被移除！\n", condition))
				sb.WriteString("     压抑的感觉逐渐消散，力量重新涌动...\n")
			}
		}
	default:
		if result.Narrative != "" {
			sb.WriteString(fmt.Sprintf("  └─ %s\n", result.Narrative))
		}
	}

	return sb.String()
}

// generateItemNarrative 生成物品类工具叙述
func generateItemNarrative(toolName string, args map[string]interface{}, result *tools.ToolResult, execErr error) string {
	var sb strings.Builder

	if execErr != nil {
		sb.WriteString("  └─ 物品操作失败...\n")
		return sb.String()
	}

	if result == nil {
		return "  └─ 物品状态变化中...\n"
	}

	switch toolName {
	case "add_item":
		if itemName, ok := args["item_name"].(string); ok {
			if result.Success {
				sb.WriteString(fmt.Sprintf("  └─ 📦 获得了【%s】！\n", itemName))
				sb.WriteString("     物品被小心翼翼地收入囊中...\n")
			}
		}
	case "remove_item":
		if itemName, ok := args["item_name"].(string); ok {
			if result.Success {
				sb.WriteString(fmt.Sprintf("  └─ 📤 失去了【%s】\n", itemName))
				sb.WriteString("     物品从背包中消失了...\n")
			}
		}
	case "spend_gold":
		if amount, ok := args["amount"].(float64); ok {
			if result.Success {
				sb.WriteString(fmt.Sprintf("  └─ 💰 支付了 %d 枚金币\n", int(amount)))
				sb.WriteString("     金币叮当作响，交易完成...\n")
			} else {
				sb.WriteString("  └─ 💰 金币不足，交易失败\n")
			}
		}
	case "gain_gold":
		if amount, ok := args["amount"].(float64); ok {
			if result.Success {
				sb.WriteString(fmt.Sprintf("  └─ 💎 获得了 %d 枚金币！\n", int(amount)))
				sb.WriteString("     金币落入钱袋，发出悦耳的声音...\n")
			}
		}
	default:
		if result.Narrative != "" {
			sb.WriteString(fmt.Sprintf("  └─ %s\n", result.Narrative))
		}
	}

	return sb.String()
}

// generateWorldNarrative 生成世界类工具叙述
func generateWorldNarrative(toolName string, args map[string]interface{}, result *tools.ToolResult, execErr error) string {
	var sb strings.Builder

	if execErr != nil {
		sb.WriteString("  └─ 世界发生了某种奇异的扰动...\n")
		return sb.String()
	}

	if result == nil {
		return "  └─ 世界正在变化...\n"
	}

	switch toolName {
	case "move_to_scene":
		if sceneID, ok := args["scene_id"].(string); ok {
			if result.Success {
				sb.WriteString(fmt.Sprintf("  └─ 🚶 进入场景【%s】\n", sceneID))
				sb.WriteString("     环境开始模糊，新的景象逐渐显现...\n")
			}
		}
	case "spawn_npc":
		if npcName, ok := args["npc_name"].(string); ok {
			if result.Success {
				sb.WriteString(fmt.Sprintf("  └─ 👤 【%s】出现了！\n", npcName))
				sb.WriteString("     一个身影从阴影中显现...\n")
			}
		}
	case "remove_npc":
		if npcID, ok := args["npc_id"].(string); ok {
			if result.Success {
				sb.WriteString(fmt.Sprintf("  └─ 👥 【%s】离开了\n", npcID))
				sb.WriteString("     身影渐渐消失在远处...\n")
			}
		}
	case "set_flag":
		if flagKey, ok := args["flag_key"].(string); ok {
			if result.Success {
				sb.WriteString(fmt.Sprintf("  └─ 🏁 标记【%s】已设置\n", flagKey))
				sb.WriteString("     世界记住了这个改变...\n")
			}
		}
	case "get_flag":
		if flagKey, ok := args["flag_key"].(string); ok {
			if result.Success {
				sb.WriteString(fmt.Sprintf("  └─ 🔍 查询标记【%s】\n", flagKey))
			}
		}
	default:
		if result.Narrative != "" {
			sb.WriteString(fmt.Sprintf("  └─ %s\n", result.Narrative))
		}
	}

	return sb.String()
}

// generateGenericNarrative 生成通用工具叙述
func generateGenericNarrative(toolName string, args map[string]interface{}, result *tools.ToolResult, execErr error) string {
	var sb strings.Builder

	if execErr != nil {
		sb.WriteString(fmt.Sprintf("  └─ %s 执行失败\n", toolName))
		return sb.String()
	}

	if result != nil && result.Narrative != "" {
		sb.WriteString(fmt.Sprintf("  └─ %s\n", result.Narrative))
	} else {
		sb.WriteString(fmt.Sprintf("  └─ %s 执行完成\n", toolName))
	}

	return sb.String()
}
