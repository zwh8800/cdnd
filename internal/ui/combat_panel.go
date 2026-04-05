package ui

import (
	"fmt"
	"strings"

	"github.com/zwh8800/cdnd/internal/combat"
	"github.com/zwh8800/cdnd/internal/game/state"
)

// renderCombatPanel 渲染战斗面板（GameModel 方法）
func (m *GameModel) renderCombatPanel() string {
	// 检查是否在战斗阶段
	if m.phase != state.PhaseCombat {
		return ""
	}

	// 获取战斗状态
	gameState := m.engine.GetState()
	if gameState == nil || gameState.Combat == nil {
		return ""
	}

	combat := gameState.Combat
	if !combat.Active {
		return ""
	}

	var sections []string

	// 头部信息
	header := CombatPanelStyles.Header.Render(fmt.Sprintf("⚔️ 战斗 - 第 %d 轮", combat.Round))
	sections = append(sections, header)

	// 敌人状态
	enemySection := m.renderEnemyList(combat.Participants)
	if enemySection != "" {
		sections = append(sections, CombatPanelStyles.Divider.Render(strings.Repeat("─", max(0, m.windowWidth-8))))
		sections = append(sections, enemySection)
	}

	// 先攻顺序
	initiativeSection := m.renderInitiativeList(combat.Initiative, combat.CurrentTurn, combat.Participants)
	if initiativeSection != "" {
		sections = append(sections, CombatPanelStyles.Divider.Render(strings.Repeat("─", max(0, m.windowWidth-8))))
		sections = append(sections, initiativeSection)
	}

	content := strings.Join(sections, "\n")
	return CombatPanelStyles.Panel.Width(m.windowWidth - 4).Render(content)
}

// renderEnemyList 渲染敌人列表
func (m *GameModel) renderEnemyList(participants []*combat.Combatant) string {
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
func (m *GameModel) renderInitiativeList(initiative []combat.InitiativeEntry, currentTurn int, participants []*combat.Combatant) string {
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

// getCombatPanelHeight 获取战斗面板高度
func (m *GameModel) getCombatPanelHeight(combat *combat.CombatState) int {
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
