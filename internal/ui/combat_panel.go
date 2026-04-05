package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/zwh8800/cdnd/internal/combat"
)

// CombatPanelStyles 战斗面板样式
var CombatPanelStyles = struct {
	Panel       lipgloss.Style
	Header      lipgloss.Style
	EnemyName   lipgloss.Style
	EnemyHP     lipgloss.Style
	EnemyHPLow  lipgloss.Style
	Initiative  lipgloss.Style
	CurrentTurn lipgloss.Style
	RoundInfo   lipgloss.Style
	Divider     lipgloss.Style
}{
	Panel: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(0, 1),
	Header: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF6B6B")),
	EnemyName: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")),
	EnemyHP: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#69db7c")),
	EnemyHPLow: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")),
	Initiative: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD93D")),
	CurrentTurn: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")),
	RoundInfo: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")),
	Divider: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#45475a")),
}

// RenderCombatPanel 渲染战斗面板
func RenderCombatPanel(combat *combat.CombatState, width int) string {
	if combat == nil || !combat.Active {
		return ""
	}

	var sections []string

	// 头部信息
	header := CombatPanelStyles.Header.Render(fmt.Sprintf("⚔️ 战斗 - 第 %d 轮", combat.Round))
	sections = append(sections, header)

	// 敌人状态
	enemySection := renderEnemyList(combat.Participants, width-4)
	if enemySection != "" {
		sections = append(sections, CombatPanelStyles.Divider.Render(strings.Repeat("─", width-4)))
		sections = append(sections, enemySection)
	}

	// 先攻顺序
	initiativeSection := renderInitiativeList(combat.Initiative, combat.CurrentTurn, combat.Participants, width-4)
	if initiativeSection != "" {
		sections = append(sections, CombatPanelStyles.Divider.Render(strings.Repeat("─", width-4)))
		sections = append(sections, initiativeSection)
	}

	content := strings.Join(sections, "\n")
	return CombatPanelStyles.Panel.Width(width).Render(content)
}

// renderEnemyList 渲染敌人列表
func renderEnemyList(participants []*combat.Combatant, width int) string {
	var enemies []*combat.Combatant
	for _, p := range participants {
		if !p.IsPlayer && p.HP > 0 {
			enemies = append(enemies, p)
		}
	}
	if len(enemies) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, "👹 敌人:")

	for _, enemy := range enemies {
		hpPercent := float64(enemy.HP) / float64(enemy.MaxHP)
		hpStyle := CombatPanelStyles.EnemyHP
		if hpPercent < 0.3 {
			hpStyle = CombatPanelStyles.EnemyHPLow
		}

		hpBar := renderHPBar(enemy.HP, enemy.MaxHP, 8)
		line := fmt.Sprintf("  %s %s (AC:%d)",
			CombatPanelStyles.EnemyName.Render(enemy.Name),
			hpStyle.Render(fmt.Sprintf("[%s %d/%d]", hpBar, enemy.HP, enemy.MaxHP)),
			enemy.AC,
		)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderInitiativeList 渲染先攻顺序
func renderInitiativeList(initiative []combat.InitiativeEntry, currentTurn int, participants []*combat.Combatant, width int) string {
	if len(initiative) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, "⚡ 先攻顺序:")

	for i, entry := range initiative {
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

		marker := "  "
		if i == currentTurn {
			marker = "▶ "
			name = CombatPanelStyles.CurrentTurn.Render(name)
		}

		line := fmt.Sprintf("%s%d. %s (%d)", marker, i+1, name, entry.Initiative)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderHPBar 渲染生命值条
func renderHPBar(current, max int, length int) string {
	if max <= 0 {
		return strings.Repeat("░", length)
	}

	filled := int(float64(current) / float64(max) * float64(length))
	if filled < 0 {
		filled = 0
	}
	if filled > length {
		filled = length
	}

	empty := length - filled
	return strings.Repeat("█", filled) + strings.Repeat("░", empty)
}

// GetCombatPanelHeight 获取战斗面板高度
func GetCombatPanelHeight(combat *combat.CombatState) int {
	if combat == nil || !combat.Active {
		return 0
	}

	height := 3

	enemyCount := 0
	for _, p := range combat.Participants {
		if !p.IsPlayer && p.HP > 0 {
			enemyCount++
		}
	}
	if enemyCount > 0 {
		height += 2 + enemyCount
	}

	if len(combat.Initiative) > 0 {
		height += 2 + len(combat.Initiative)
	}

	return height
}
