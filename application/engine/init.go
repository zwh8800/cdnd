package engine

import (
	"fmt"
	"strings"

	"github.com/zwh8800/cdnd/domain"
	"github.com/zwh8800/cdnd/domain/character"
	"github.com/zwh8800/cdnd/domain/llm"
	"github.com/zwh8800/cdnd/infrastructure/prompt"
)

// IntroSequence 游戏初始化对话序列
type IntroSequence struct {
	Lines []string         // 按显示顺序排列的消息行
	Phase domain.GamePhase // 当前游戏阶段
}

// GenerateWelcomeMessage 根据角色信息生成个性化欢迎消息
func GenerateWelcomeMessage(c *character.Character) string {
	var sb strings.Builder

	raceName := "未知种族"
	if c.Race.Name != "" {
		raceName = c.Race.Name
	}

	className := "未知职业"
	if c.Class.Name != "" {
		className = c.Class.Name
	}

	background := "冒险者"
	if c.Background != "" {
		background = c.Background
	}

	sb.WriteString(fmt.Sprintf("🎲 欢迎来到 D&D 世界，{{keyword:%s}}！", c.Name))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("你是一位 {{keyword:%s}} 的 {{keyword:%s}}，等级 {{number:%d}}。", raceName, className, c.Level))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("作为一名 %s，你即将踏上一段充满未知与危险的冒险旅程。", background))
	sb.WriteString("\n\n")
	sb.WriteString("命运之轮已经开始，冒险的篇章等待着你来书写...")

	return sb.String()
}

// ShowWelcomeMessage 生成并显示欢迎消息（快速，无 LLM 调用）
func (e *Engine) ShowWelcomeMessage() (string, error) {
	character := e.GetCharacter()
	if character == nil {
		return "", fmt.Errorf("角色数据为空")
	}

	// 生成欢迎消息并添加到历史
	welcomeMsg := GenerateWelcomeMessage(character)
	coloredWelcome := prompt.ParseColorMarkers(welcomeMsg)
	e.state.AddHistory(llm.Message{
		Role:    llm.RoleAssistant,
		Content: welcomeMsg,
	})

	return coloredWelcome, nil
}
