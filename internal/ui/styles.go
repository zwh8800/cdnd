package ui

import "github.com/charmbracelet/lipgloss"

// 调色板
var (
	PrimaryColor   = lipgloss.Color("#7D56F4") // 紫色
	SecondaryColor = lipgloss.Color("#04B575") // 绿色
	DangerColor    = lipgloss.Color("#FF6B6B") // 红色
	WarningColor   = lipgloss.Color("#FFD93D") // 黄色
	AccentColor    = lipgloss.Color("#5C5CFF") // 浅紫色
	BorderColor    = lipgloss.Color("#4A4A4A") // 灰色
	TextColor      = lipgloss.Color("#FAFAFA") // 白色
	SubtleColor    = lipgloss.Color("#626262") // 暗灰色
	BgColor        = lipgloss.Color("#1A1A2E") // 深蓝色
)

// 基础样式
var (
	// 头部样式
	HeaderStyle = lipgloss.NewStyle().
			Background(BgColor).
			Padding(0, 1).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(BorderColor)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			Padding(0, 1)

	StatsStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Padding(0, 1)

	// 叙述区样式
	NarrativeStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Foreground(TextColor)

	// 输入区样式
	InputStyle = lipgloss.NewStyle().
			Padding(1, 2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(BorderColor)

	// 菜单样式
	MenuStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Align(lipgloss.Center)

	MenuItemStyle = lipgloss.NewStyle().
			Foreground(SubtleColor).
			Padding(0, 1)

	MenuItemSelectedStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Bold(true).
				Padding(0, 1).
				Underline(true)

	// 状态面板样式
	StatusStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#252535")).
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor)

	// 骰子结果样式
	DiceBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentColor).
			Padding(0, 1).
			Margin(1, 2)

	DiceSuccessStyle = lipgloss.NewStyle().
				Foreground(SecondaryColor).
				Bold(true)

	DiceFailStyle = lipgloss.NewStyle().
			Foreground(DangerColor).
			Bold(true)

	DiceRollStyle = lipgloss.NewStyle().
			Foreground(AccentColor).
			Bold(true)

	// 通用文本样式
	BoldStyle = lipgloss.NewStyle().
			Bold(true)

	ItalicStyle = lipgloss.NewStyle().
			Italic(true)

	DimStyle = lipgloss.NewStyle().
			Foreground(SubtleColor)

	// 游戏元素专用样式
	DMNarrationStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#87CEEB")). // 浅蓝色
				Padding(0, 1)

	PlayerActionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#98FB98")). // 浅绿色
				Padding(0, 1)

	CombatStyle = lipgloss.NewStyle().
			Foreground(DangerColor).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true)

	FailureStyle = lipgloss.NewStyle().
			Foreground(DangerColor).
			Bold(true)

	// GameStyles 样式定义
	GameStyles = struct {
		Title          lipgloss.Style
		StatusBar      lipgloss.Style
		Box            lipgloss.Style
		InputBox       lipgloss.Style
		Highlight      lipgloss.Style
		PanelTitle     lipgloss.Style
		Positive       lipgloss.Style
		Negative       lipgloss.Style
		SpellSlot      lipgloss.Style
		LocationName   lipgloss.Style
		GoldText       lipgloss.Style
		ConditionBadge lipgloss.Style
	}{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7c3aed")).
			Padding(0, 1),
		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color("#1e1e2e")).
			Foreground(lipgloss.Color("#cdd6f4")).
			Padding(0, 1),
		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#45475a")).
			Padding(0, 1),
		InputBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#89b4fa")).
			Padding(0, 1),
		Highlight: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#f9e2af")),
		PanelTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#b197fc")).
			Padding(0, 1),
		Positive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#69db7c")),
		Negative: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff6b6b")),
		SpellSlot: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9775fa")),
		LocationName: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#4dabf7")),
		GoldText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffd700")),
		ConditionBadge: lipgloss.NewStyle().
			Background(lipgloss.Color("#ffd43b")).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1),
	}
)

// FormatDiceRoll 格式化骰子结果以便显示。
func FormatDiceRoll(notation string, total int, success bool) string {
	var style lipgloss.Style
	if success {
		style = DiceSuccessStyle
	} else {
		style = DiceFailStyle
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		DiceRollStyle.Render(notation),
		" = ",
		style.Render(string(rune('0'+total/10))+string(rune('0'+total%10))),
	)
}

// FormatNarration 格式化 DM 叙述文本。
func FormatNarration(text string) string {
	return DMNarrationStyle.Render("DM: " + text)
}

// FormatPlayerAction 格式化玩家行动文本。
func FormatPlayerAction(text string) string {
	return PlayerActionStyle.Render("你: " + text)
}

// FormatCombat 格式化战斗文本。
func FormatCombat(text string) string {
	return CombatStyle.Render(text)
}
