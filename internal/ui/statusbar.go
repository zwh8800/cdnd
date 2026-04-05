package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/zwh8800/cdnd/internal/character"
	"github.com/zwh8800/cdnd/internal/game/state"
)

// renderStatusBarCompact 渲染单行紧凑状态栏
func (m *GameModel) renderStatusBarCompact(c *character.Character) string {
	// 左侧：角色信息
	var leftParts []string
	leftParts = append(leftParts, c.Name)
	if c.Race.Name != "" {
		leftParts = append(leftParts, c.Race.Name)
	}
	leftParts = append(leftParts, fmt.Sprintf("%d级", c.Level))
	if c.HasClass() {
		leftParts = append(leftParts, c.Class.Name)
	}
	left := strings.Join(leftParts, " - ")

	// 右侧：核心 D&D 数据
	var rightParts []string

	// HP (彩色)
	hpPercent := float64(c.HitPoints.Current) / float64(c.HitPoints.Max)
	hpColor := "#00ff00"
	if hpPercent <= 0 {
		hpColor = "#888888"
	} else if hpPercent <= 0.25 {
		hpColor = "#ff0000"
	} else if hpPercent <= 0.5 {
		hpColor = "#ffaa00"
	}
	hpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(hpColor))
	hpText := fmt.Sprintf("HP:%d/%d", c.HitPoints.Current, c.HitPoints.Max)
	rightParts = append(rightParts, hpStyle.Render(hpText))

	// 战斗阶段特有信息
	if m.phase == state.PhaseCombat {
		// 动作指示器
		actionUsed := false
		bonusActionUsed := false
		if state := m.engine.GetState(); state != nil && state.Combat != nil {
			actionUsed = state.Combat.PlayerActionUsed
			bonusActionUsed = state.Combat.PlayerBonusActionUsed
		}
		actionIcon := "●"
		if actionUsed {
			actionIcon = "○"
		}
		bonusActionIcon := "▲"
		if bonusActionUsed {
			bonusActionIcon = "△"
		}
		rightParts = append(rightParts, actionIcon+" "+bonusActionIcon)

		// 先攻
		initMod := c.Attributes.Modifier(character.Dexterity)
		if initMod >= 0 {
			initStr := GameStyles.Positive.Render(fmt.Sprintf("先攻:+%d", initMod))
			rightParts = append(rightParts, initStr)
		} else {
			initStr := GameStyles.Negative.Render(fmt.Sprintf("先攻:%d", initMod))
			rightParts = append(rightParts, initStr)
		}

		// 法术环 (Unicode 分数)
		if !c.SpellSlots.IsEmpty() {
			spellFraction := m.getTopSpellSlotsFraction(c)
			if spellFraction != "" {
				rightParts = append(rightParts, GameStyles.SpellSlot.Render(spellFraction))
			}
		}
	} else {
		// 非战斗阶段显示 AC
		rightParts = append(rightParts, fmt.Sprintf("AC:%d", c.ArmorClass))

		// 探索阶段显示位置
		if m.phase == state.PhaseExploration {
			if scene := m.engine.GetCurrentScene(); scene != nil {
				locName := m.abbreviateLocation(scene.Name, 6)
				rightParts = append(rightParts, GameStyles.LocationName.Render(locName))
			}
		}
	}

	// 金币 (非战斗阶段显示)
	if m.phase != state.PhaseCombat {
		rightParts = append(rightParts, GameStyles.GoldText.Render(fmt.Sprintf("G:%d", c.Gold)))
	}

	// 阶段名称
	phaseColor := "#cdd6f4"
	switch m.phase {
	case state.PhaseCombat:
		phaseColor = "#ff6b6b"
	case state.PhaseExploration:
		phaseColor = "#69db7c"
	case state.PhaseRest:
		phaseColor = "#74c0fc"
	case state.PhaseDialogue:
		phaseColor = "#ffd43b"
	}
	phaseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(phaseColor)).Bold(true)
	rightParts = append(rightParts, phaseStyle.Render(m.phase.String()))

	// 回合数
	if state := m.engine.GetState(); state != nil && state.TurnCount > 0 {
		rightParts = append(rightParts, fmt.Sprintf("T:%d", state.TurnCount))
	}

	// 组合右侧
	right := strings.Join(rightParts, " | ")

	// 计算间距
	leftRendered := GameStyles.StatusBar.Render(left)
	rightRendered := GameStyles.StatusBar.Render(right)
	leftWidth := lipgloss.Width(leftRendered)
	rightWidth := lipgloss.Width(rightRendered)
	padding := max(0, m.windowWidth-leftWidth-rightWidth-2)

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		leftRendered,
		strings.Repeat(" ", padding),
		rightRendered,
	)

	return bar
}

// abbreviateLocation 缩写位置名
func (m *GameModel) abbreviateLocation(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	return name[:maxLen-3] + "..."
}

// getTopSpellSlotsFraction 获取最高可用两个环阶的法术环分数 (Unicode)
func (m *GameModel) getTopSpellSlotsFraction(c *character.Character) string {
	if c.SpellSlots.IsEmpty() {
		return ""
	}

	// 获取最大法术槽 (根据职业和等级)
	casterType := character.GetCasterType(c.Class.ID)
	maxSlots := character.GetSpellSlotsByType(casterType, c.Level)

	// 找到最高两个有法术槽的环阶
	var levels []int
	for level := 9; level >= 1; level-- {
		maximum := maxSlots.GetSlotsByLevel(level)
		if maximum > 0 {
			levels = append(levels, level)
			if len(levels) >= 2 {
				break
			}
		}
	}

	if len(levels) == 0 {
		return ""
	}

	// 构建分数显示
	var parts []string
	for _, level := range levels {
		current := c.SpellSlots.GetSlotsByLevel(level)
		maximum := maxSlots.GetSlotsByLevel(level)
		fraction := m.formatSpellFraction(current, maximum)
		levelName := character.GetSpellLevelName(level)
		parts = append(parts, fmt.Sprintf("%s%s", levelName, fraction))
	}

	return strings.Join(parts, " ")
}

// formatSpellFraction 格式化法术分数为 Unicode
func (m *GameModel) formatSpellFraction(current, max int) string {
	// Unicode 分数映射
	fractions := map[string]string{
		"0/1": "0", "0/2": "0", "0/3": "0", "0/4": "0",
		"1/2": "½", "1/3": "⅓", "2/3": "⅔",
		"1/4": "¼", "3/4": "¾", "2/4": "½",
		"1/5": "⅕", "2/5": "⅖", "3/5": "⅗", "4/5": "⅘",
		"1/6": "⅙", "5/6": "⅚",
		"1/8": "⅛", "3/8": "⅜", "5/8": "⅝", "7/8": "⅞",
	}

	key := fmt.Sprintf("%d/%d", current, max)
	if unicodeFraction, ok := fractions[key]; ok {
		return unicodeFraction
	}
	return key
}

// renderStatusBarExpanded 渲染多行展开状态栏
func (m *GameModel) renderStatusBarExpanded(c *character.Character) string {
	// 顶部栏 (紧凑信息)
	topBar := m.renderStatusBarCompact(c)

	// 构建面板
	panels := []string{
		m.renderStatPanel(c),
		m.renderEquipmentPanel(c),
		m.renderSpellPanel(c),
		m.renderConditionsPanel(c),
		m.renderLocationPanel(c),
	}

	// 根据终端宽度决定布局
	if m.windowWidth >= 100 {
		// 宽终端：水平排列
		content := lipgloss.JoinHorizontal(lipgloss.Top, panels...)
		return topBar + "\n" + content
	}
	// 窄终端：垂直排列
	content := strings.Join(panels, "\n")
	return topBar + "\n" + content
}

// renderStatPanel 渲染属性面板
func (m *GameModel) renderStatPanel(c *character.Character) string {
	abilities := []struct {
		name  string
		value int
		color string
	}{
		{"STR", c.Attributes.Strength, "#ff6b6b"},
		{"DEX", c.Attributes.Dexterity, "#69db7c"},
		{"CON", c.Attributes.Constitution, "#ffa94d"},
		{"INT", c.Attributes.Intelligence, "#74c0fc"},
		{"WIS", c.Attributes.Wisdom, "#b197fc"},
		{"CHA", c.Attributes.Charisma, "#ffd43b"},
	}

	var lines []string
	lines = append(lines, GameStyles.PanelTitle.Render("── 属性 ──"))

	for _, ab := range abilities {
		mod := c.Attributes.Modifier(character.Ability(ab.name))
		modStr := fmt.Sprintf("%+d", mod)
		modStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ab.color))
		line := fmt.Sprintf("%s %d %s", ab.name, ab.value, modStyle.Render(modStr))
		lines = append(lines, GameStyles.StatusBar.Render(line))
	}

	// AC, 先攻, 速度, 熟练
	statsLine := fmt.Sprintf("AC:%d 先攻:%+d", c.ArmorClass, c.Attributes.Modifier(character.Dexterity))
	statsLine2 := fmt.Sprintf("速度:%d 熟练:%+d", c.Speed, c.ProficiencyBonus)
	lines = append(lines, GameStyles.StatusBar.Render(statsLine))
	lines = append(lines, GameStyles.StatusBar.Render(statsLine2))

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7c3aed")).
		Padding(0, 1)

	return borderStyle.Render(strings.Join(lines, "\n"))
}

// renderEquipmentPanel 渲染装备面板
func (m *GameModel) renderEquipmentPanel(c *character.Character) string {
	var lines []string
	lines = append(lines, GameStyles.PanelTitle.Render("── 装备 ──"))

	// 查找武器和护甲
	weapon := "无"
	armor := "无"
	for _, item := range c.Equipment {
		if item.Type == "weapon" {
			weapon = item.Name
		} else if item.Type == "armor" {
			armor = item.Name
		}
	}

	lines = append(lines, GameStyles.StatusBar.Render(fmt.Sprintf("武器: %s", weapon)))
	lines = append(lines, GameStyles.StatusBar.Render(fmt.Sprintf("护甲: %s", armor)))
	lines = append(lines, GameStyles.StatusBar.Render(GameStyles.GoldText.Render(fmt.Sprintf("金币: %d", c.Gold))))
	lines = append(lines, GameStyles.StatusBar.Render(fmt.Sprintf("背包: %d 件", len(c.Inventory))))

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7c3aed")).
		Padding(0, 1)

	return borderStyle.Render(strings.Join(lines, "\n"))
}

// renderSpellPanel 渲染法术面板
func (m *GameModel) renderSpellPanel(c *character.Character) string {
	var lines []string
	lines = append(lines, GameStyles.PanelTitle.Render("── 法术 ──"))

	casterType := character.GetCasterType(c.Class.ID)
	if casterType == character.SpellcastingNone {
		lines = append(lines, GameStyles.StatusBar.Foreground(lipgloss.Color("#888888")).Render("非施法者"))
	} else {
		// 施法属性
		abilityName := ""
		if c.SpellcastingAbility != "" {
			abilityName = string(c.SpellcastingAbility)
		}
		lines = append(lines, GameStyles.StatusBar.Render(fmt.Sprintf("施法: %s", abilityName)))

		// 戏法数量
		cantripCount := 0
		for _, spell := range c.Spells {
			if spell.Level == 0 {
				cantripCount++
			}
		}
		lines = append(lines, GameStyles.StatusBar.Render(fmt.Sprintf("戏法: %d", cantripCount)))

		// 法术槽
		maxSlots := character.GetSpellSlotsByType(casterType, c.Level)
		for level := 1; level <= 9; level++ {
			maximum := maxSlots.GetSlotsByLevel(level)
			if maximum == 0 {
				continue
			}
			current := c.SpellSlots.GetSlotsByLevel(level)
			levelName := character.GetSpellLevelName(level)

			// 构建法术槽指示器
			slots := ""
			for i := 0; i < maximum; i++ {
				if i < current {
					slots += "█"
				} else {
					slots += "░"
				}
			}
			lines = append(lines, GameStyles.StatusBar.Render(
				fmt.Sprintf("%s: %s %d/%d", levelName, slots, current, maximum)))
		}
	}

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7c3aed")).
		Padding(0, 1)

	return borderStyle.Render(strings.Join(lines, "\n"))
}

// renderConditionsPanel 渲染状态效果面板
func (m *GameModel) renderConditionsPanel(c *character.Character) string {
	var lines []string
	lines = append(lines, GameStyles.PanelTitle.Render("── 状态 ──"))

	conditions := c.GetConditions()
	if len(conditions) == 0 {
		lines = append(lines, GameStyles.StatusBar.
			Foreground(lipgloss.Color("#888888")).
			Render("无状态效果"))
	} else {
		for _, cond := range conditions {
			lines = append(lines, GameStyles.ConditionBadge.Render("⚠ "+cond))
		}
	}

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7c3aed")).
		Padding(0, 1)

	return borderStyle.Render(strings.Join(lines, "\n"))
}

// renderLocationPanel 渲染位置/时间面板
func (m *GameModel) renderLocationPanel(c *character.Character) string {
	var lines []string
	lines = append(lines, GameStyles.PanelTitle.Render("── 位置/时间 ──"))

	scene := m.engine.GetCurrentScene()
	if scene != nil {
		lines = append(lines, GameStyles.LocationName.Render(scene.Name))
		lines = append(lines, GameStyles.StatusBar.Render(fmt.Sprintf("类型: %s", scene.Type)))
		lines = append(lines, GameStyles.StatusBar.Render(fmt.Sprintf("光照: %s", scene.LightLevel)))
		lines = append(lines, GameStyles.StatusBar.Render(fmt.Sprintf("地形: %s", scene.Terrain)))
	} else {
		lines = append(lines, GameStyles.StatusBar.
			Foreground(lipgloss.Color("#888888")).
			Render("未知位置"))
	}

	// 回合和时间
	if state := m.engine.GetState(); state != nil {
		lines = append(lines, GameStyles.StatusBar.Render(fmt.Sprintf("回合: %d", state.TurnCount)))
		lines = append(lines, GameStyles.StatusBar.Render(fmt.Sprintf("用时: %s", m.formatPlayTime(state.PlayedTime))))
	}

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7c3aed")).
		Padding(0, 1)

	return borderStyle.Render(strings.Join(lines, "\n"))
}

// formatPlayTime 格式化游戏时间
func (m *GameModel) formatPlayTime(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
