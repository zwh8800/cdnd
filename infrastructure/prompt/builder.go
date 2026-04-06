package prompt

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/zwh8800/cdnd/domain/character"
	"github.com/zwh8800/cdnd/domain/combat"
	"github.com/zwh8800/cdnd/domain/llm"
	"github.com/zwh8800/cdnd/domain/world"
)

// Builder 提示词构建器
type Builder struct {
	templates *Templates
}

// NewBuilder 创建新的提示词构建器
func NewBuilder() *Builder {
	return &Builder{
		templates: DefaultTemplates(),
	}
}

// GameContext 游戏上下文
type GameContext struct {
	Phase         string // 游戏阶段名称
	Character     *character.Character
	CurrentScene  *world.Scene
	DMContext     string
	History       []llm.Message
	TurnCount     int
	WorldFlags    map[string]bool
	WorldCounters map[string]int
}

// BuildSystemPrompt 构建系统提示词
func (b *Builder) BuildSystemPrompt(ctx *GameContext) string {
	var sb strings.Builder

	// 基础DM角色设定
	sb.WriteString(b.templates.DMRole)
	sb.WriteString("\n\n")

	// 添加游戏规则
	sb.WriteString(b.templates.GameRules)
	sb.WriteString("\n\n")

	// 添加角色信息
	if ctx.Character != nil {
		sb.WriteString("## 玩家角色\n")
		sb.WriteString(b.BuildCharacterContext(ctx.Character))
		sb.WriteString("\n\n")
	}

	// 添加场景信息
	if ctx.CurrentScene != nil {
		sb.WriteString("## 当前场景\n")
		sb.WriteString(b.BuildSceneContext(ctx.CurrentScene))
		sb.WriteString("\n\n")
	}

	// 添加DM上下文
	if ctx.DMContext != "" {
		sb.WriteString("## DM上下文\n")
		sb.WriteString(ctx.DMContext)
		sb.WriteString("\n\n")
	}

	// 添加工具调用说明
	sb.WriteString(b.templates.ToolInstructions)

	return sb.String()
}

// BuildCharacterContext 构建角色上下文
func (b *Builder) BuildCharacterContext(c *character.Character) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**%s** - ", c.Name))
	if c.HasClass() {
		sb.WriteString(fmt.Sprintf("%s %d级 %s", c.Race.Name, c.Level, c.Class.Name))
	} else {
		sb.WriteString(fmt.Sprintf("%s %d级", c.Race.Name, c.Level))
	}
	sb.WriteString("\n")

	// 属性
	sb.WriteString("**属性**: ")
	abilities := []struct {
		name  string
		value int
		mod   int
	}{
		{"力量", c.Attributes.Strength, c.Attributes.Modifier(character.Strength)},
		{"敏捷", c.Attributes.Dexterity, c.Attributes.Modifier(character.Dexterity)},
		{"体质", c.Attributes.Constitution, c.Attributes.Modifier(character.Constitution)},
		{"智力", c.Attributes.Intelligence, c.Attributes.Modifier(character.Intelligence)},
		{"感知", c.Attributes.Wisdom, c.Attributes.Modifier(character.Wisdom)},
		{"魅力", c.Attributes.Charisma, c.Attributes.Modifier(character.Charisma)},
	}
	for i, a := range abilities {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%s %d(%+d)", a.name, a.value, a.mod))
	}
	sb.WriteString("\n")

	// 生命值
	sb.WriteString(fmt.Sprintf("**生命值**: %d/%d\n", c.HitPoints.Current, c.HitPoints.Max))
	if c.HitPoints.Temp > 0 {
		sb.WriteString(fmt.Sprintf(" (临时生命值: %d)", c.HitPoints.Temp))
	}

	// 护甲等级
	sb.WriteString(fmt.Sprintf("**护甲等级**: %d\n", c.ArmorClass))

	// 速度
	sb.WriteString(fmt.Sprintf("**速度**: %d尺\n", c.Speed))

	// 熟练技能
	sb.WriteString("**熟练技能**: ")
	skillNames := make([]string, 0)
	for _, skillType := range character.AllSkillTypes() {
		if c.HasSkillProficiency(skillType) {
			skillNames = append(skillNames, character.GetSkillName(skillType))
		}
	}
	if len(skillNames) > 0 {
		sb.WriteString(strings.Join(skillNames, ", "))
	} else {
		sb.WriteString("无")
	}
	sb.WriteString("\n")

	// 金币
	sb.WriteString(fmt.Sprintf("**金币**: %d gp\n", c.Gold))

	return sb.String()
}

// BuildSceneContext 构建场景上下文
func (b *Builder) BuildSceneContext(scene *world.Scene) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**%s**\n", scene.Name))
	sb.WriteString(scene.Description)
	sb.WriteString("\n")

	// 出口
	if len(scene.Exits) > 0 {
		sb.WriteString("**出口**: ")
		exitNames := make([]string, 0)
		for _, exit := range scene.Exits {
			if exit.Visible {
				exitNames = append(exitNames, exit.Name)
			}
		}
		sb.WriteString(strings.Join(exitNames, ", "))
		sb.WriteString("\n")
	}

	// NPC
	if len(scene.NPCs) > 0 {
		sb.WriteString(fmt.Sprintf("**在场NPC**: %d个\n", len(scene.NPCs)))
	}

	// 光照
	sb.WriteString(fmt.Sprintf("**光照**: %s\n", scene.LightLevel.String()))

	return sb.String()
}

// BuildHistoryContext 构建历史上下文
func (b *Builder) BuildHistoryContext(history []llm.Message, maxTurns int) []llm.Message {
	if len(history) <= maxTurns {
		return history
	}

	// 保留最近的对话
	return history[len(history)-maxTurns:]
}

// BuildIntroPrompt 构建开场提示词
func (b *Builder) BuildIntroPrompt(c *character.Character) string {
	var sb strings.Builder

	sb.WriteString(b.templates.IntroPrompt)
	sb.WriteString("\n\n")

	if c != nil {
		sb.WriteString("玩家角色信息:\n")
		sb.WriteString(b.BuildCharacterContext(c))
	}

	return sb.String()
}

// BuildCombatPrompt 构建战斗提示词
func (b *Builder) BuildCombatPrompt(ctx *GameContext, combatInfo string) string {
	var sb strings.Builder

	sb.WriteString(b.templates.CombatPrompt)
	sb.WriteString("\n\n")
	sb.WriteString(combatInfo)

	return sb.String()
}

// BuildDialoguePrompt 构建对话提示词
func (b *Builder) BuildDialoguePrompt(npcName string, disposition string) string {
	return fmt.Sprintf(b.templates.DialoguePrompt, npcName, disposition)
}

// BuildRestPrompt 构建休息提示词
func (b *Builder) BuildRestPrompt() string {
	return b.templates.RestPrompt
}

// BuildPlayerActionPrompt 构建玩家行动提示词
func (b *Builder) BuildPlayerActionPrompt(action string) string {
	return fmt.Sprintf("玩家行动: %s\n\n请描述玩家的行动结果，并在需要时调用相应的工具函数。", action)
}

// CombatContext 战斗上下文
type CombatContext struct {
	Round            int    // 当前回合
	CurrentCombatant string // 当前行动者
	PlayerName       string // 玩家名称
	PlayerHP         int    // 玩家当前HP
	PlayerMaxHP      int    // 玩家最大HP
	PlayerAC         int    // 玩家AC
	EnemyList        string // 敌人列表
	InitiativeOrder  string // 先攻顺序
}

// BuildCombatSystemPrompt 构建战斗阶段系统提示
func (b *Builder) BuildCombatSystemPrompt(combat *combat.CombatState, character *character.Character, participants []*combat.Combatant) string {
	if combat == nil || character == nil {
		return ""
	}

	// 构建敌人状态列表
	var enemyParts []string
	for _, p := range combat.Participants {
		if !p.IsPlayer && p.HP > 0 {
			enemyParts = append(enemyParts, fmt.Sprintf("%s(HP:%d/%d)", p.Name, p.HP, p.MaxHP))
		}
	}
	enemyList := strings.Join(enemyParts, ", ")

	// 构建先攻顺序
	var initiativeLines []string
	for i, entry := range combat.Initiative {
		var name string
		for _, p := range participants {
			if p.ID == entry.EntityID {
				name = p.Name
				break
			}
		}
		if name == "" {
			continue
		}
		marker := ""
		if i == combat.CurrentTurn {
			marker = " ← 当前"
		}
		initiativeLines = append(initiativeLines, fmt.Sprintf("\n  %d. %s%s", i+1, name, marker))
	}
	initiativeOrder := strings.Join(initiativeLines, "")

	// 获取当前行动者名称
	currentCombatant := "未知"
	if combat.CurrentTurn < len(combat.Participants) {
		currentCombatant = combat.Participants[combat.CurrentTurn].Name
	}

	ctx := CombatContext{
		Round:            combat.Round,
		CurrentCombatant: currentCombatant,
		PlayerName:       character.Name,
		PlayerHP:         character.HitPoints.Current,
		PlayerMaxHP:      character.HitPoints.Max,
		PlayerAC:         character.ArmorClass,
		EnemyList:        enemyList,
		InitiativeOrder:  initiativeOrder,
	}

	// 使用模板渲染
	tmpl, err := template.New("combat_system").Parse(b.templates.CombatSystemPrompt)
	if err != nil {
		// 如果模板解析失败，降级为原始格式
		return b.buildCombatSystemPromptFallback(&ctx)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return b.buildCombatSystemPromptFallback(&ctx)
	}

	return buf.String()
}

// buildCombatSystemPromptFallback 降级方案（直接格式化）
func (b *Builder) buildCombatSystemPromptFallback(ctx *CombatContext) string {
	return fmt.Sprintf(b.templates.CombatSystemPrompt,
		ctx.Round,
		ctx.CurrentCombatant,
		ctx.PlayerName,
		ctx.PlayerHP,
		ctx.PlayerMaxHP,
		ctx.PlayerAC,
		ctx.EnemyList,
		ctx.InitiativeOrder,
	)
}

// TruncateHistory 截断历史记录以适应上下文限制
func (b *Builder) TruncateHistory(history []llm.Message, maxTokens int) []llm.Message {
	// 简化实现：保留最近的对话
	// 实际实现应该计算token数量
	if len(history) > 20 {
		return history[len(history)-20:]
	}
	return history
}
