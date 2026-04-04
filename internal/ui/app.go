// Package ui 使用 Bubbletea 提供 TUI 组件。
package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AppModel 表示主应用程序模型。
type AppModel struct {
	// 尺寸
	width  int
	height int

	// 组件
	header    HeaderModel
	narrative NarrativeModel
	input     InputModel
	menu      MenuModel
	status    StatusModel

	// 状态
	ready   bool
	focused FocusArea
}

// FocusArea 表示当前聚焦的组件。
type FocusArea int

const (
	FocusInput FocusArea = iota
	FocusMenu
	FocusStatus
)

// NewAppModel 创建一个新的应用程序模型。
func NewAppModel() AppModel {
	return AppModel{
		header:    NewHeaderModel(),
		narrative: NewNarrativeModel(),
		input:     NewInputModel(),
		menu:      NewMenuModel(),
		status:    NewStatusModel(),
		focused:   FocusInput,
	}
}

// Init 初始化模型。
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.header.Init(),
		m.narrative.Init(),
		m.input.Init(),
		m.menu.Init(),
		m.status.Init(),
	)
}

// Update 处理消息。
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab":
			// 循环切换焦点
			m.focused = (m.focused + 1) % 3
		case "shift+tab":
			// 反向循环切换焦点
			if m.focused == 0 {
				m.focused = FocusStatus
			} else {
				m.focused--
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// 计算组件尺寸
		headerHeight := 3
		inputHeight := 5
		menuHeight := 3

		m.header.width = m.width
		m.header.height = headerHeight

		m.narrative.width = m.width
		m.narrative.height = m.height - headerHeight - inputHeight - menuHeight - 2

		m.input.width = m.width
		m.input.height = inputHeight

		m.menu.width = m.width
		m.menu.height = menuHeight

		m.status.width = m.width / 3
		m.status.height = m.height - headerHeight

		m.ready = true
	}

	// 更新聚焦的组件
	switch m.focused {
	case FocusInput:
		newInput, cmd := m.input.Update(msg)
		m.input = newInput.(InputModel)
		cmds = append(cmds, cmd)
	case FocusMenu:
		newMenu, cmd := m.menu.Update(msg)
		m.menu = newMenu.(MenuModel)
		cmds = append(cmds, cmd)
	}

	// 始终更新头部和叙述区
	newHeader, cmd := m.header.Update(msg)
	m.header = newHeader.(HeaderModel)
	cmds = append(cmds, cmd)

	newNarrative, cmd := m.narrative.Update(msg)
	m.narrative = newNarrative.(NarrativeModel)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View 渲染模型。
func (m AppModel) View() string {
	if !m.ready {
		return "\n  初始化中..."
	}

	// 布局：头部在上方，叙述区在中间，输入+菜单在底部
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.header.View(),
		m.narrative.View(),
		m.input.View(),
		m.menu.View(),
	)
}

// HeaderModel 表示顶部状态栏。
type HeaderModel struct {
	width, height int
	characterName string
	level         int
	hp            string
}

// NewHeaderModel 创建一个新的头部模型。
func NewHeaderModel() HeaderModel {
	return HeaderModel{
		characterName: "冒险者",
		level:         1,
		hp:            "10/10",
	}
}

// Init 初始化头部。
func (m HeaderModel) Init() tea.Cmd {
	return nil
}

// Update 处理消息。
func (m HeaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// View 渲染头部。
func (m HeaderModel) View() string {
	title := TitleStyle.Render("D&D 冒险")
	stats := StatsStyle.Render("HP: " + m.hp + "  Lv." + string(rune('0'+m.level)))

	// 组合标题和状态
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		strings.Repeat(" ", max(0, m.width-lipgloss.Width(title)-lipgloss.Width(stats)-2)),
		stats,
	)

	return HeaderStyle.Width(m.width).Render(content)
}

// NarrativeModel 表示主叙述显示区。
type NarrativeModel struct {
	width, height int
	viewport      viewport.Model
	content       []string
}

// NewNarrativeModel 创建一个新的叙述模型。
func NewNarrativeModel() NarrativeModel {
	return NarrativeModel{
		content: []string{},
	}
}

// Init 初始化叙述区。
func (m NarrativeModel) Init() tea.Cmd {
	return nil
}

// Update 处理消息。
func (m NarrativeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	newViewport, cmd := m.viewport.Update(msg)
	m.viewport = newViewport
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View 渲染叙述区。
func (m NarrativeModel) View() string {
	if len(m.content) == 0 {
		return NarrativeStyle.Width(m.width).Height(m.height).Render(
			"欢迎来到你的冒险！\n\n按回车键开始...",
		)
	}

	m.viewport.SetContent(strings.Join(m.content, "\n\n"))
	return NarrativeStyle.Width(m.width).Height(m.height).Render(m.viewport.View())
}

// AddContent 向叙述区添加内容。
func (m *NarrativeModel) AddContent(text string) {
	m.content = append(m.content, text)
}

// InputModel 表示用户输入区。
type InputModel struct {
	width, height int
	textarea      textarea.Model
}

// NewInputModel 创建一个新的输入模型。
func NewInputModel() InputModel {
	ti := textarea.New()
	ti.SetWidth(60)
	ti.SetHeight(1)
	ti.SetValue("")
	ti.Focus()

	return InputModel{
		textarea: ti,
	}
}

// Init 初始化输入区。
func (m InputModel) Init() tea.Cmd {
	return nil
}

// Update 处理消息。
func (m InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	newTextarea, cmd := m.textarea.Update(msg)
	m.textarea = newTextarea
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View 渲染输入区。
func (m InputModel) View() string {
	return InputStyle.Width(m.width).Render(
		"> " + m.textarea.View(),
	)
}

// MenuModel 表示操作菜单。
type MenuModel struct {
	width, height int
	choices       []string
	selected      int
}

// NewMenuModel 创建一个新的菜单模型。
func NewMenuModel() MenuModel {
	return MenuModel{
		choices: []string{"[攻击]", "[技能]", "[物品]", "[交谈]", "[逃跑]"},
	}
}

// Init 初始化菜单。
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update 处理消息。
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.selected > 0 {
				m.selected--
			}
		case "right", "l":
			if m.selected < len(m.choices)-1 {
				m.selected++
			}
		case "enter":
			// 处理选择
			return m, nil
		}
	}

	return m, nil
}

// View 渲染菜单。
func (m MenuModel) View() string {
	items := make([]string, len(m.choices))
	for i, choice := range m.choices {
		if i == m.selected {
			items[i] = MenuItemSelectedStyle.Render(choice)
		} else {
			items[i] = MenuItemStyle.Render(choice)
		}
	}

	return MenuStyle.Width(m.width).Render(
		strings.Join(items, "  "),
	)
}

// StatusModel 表示角色状态面板。
type StatusModel struct {
	width, height int
	visible       bool
}

// NewStatusModel 创建一个新的状态模型。
func NewStatusModel() StatusModel {
	return StatusModel{
		visible: false,
	}
}

// Init 初始化状态面板。
func (m StatusModel) Init() tea.Cmd {
	return nil
}

// Update 处理消息。
func (m StatusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// View 渲染状态面板。
func (m StatusModel) View() string {
	if !m.visible {
		return ""
	}

	return StatusStyle.Width(m.width).Height(m.height).Render(
		"角色状态\n\n即将推出...",
	)
}

// max 辅助函数，返回最大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
