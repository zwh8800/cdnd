package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zwh8800/cdnd/internal/game"
	"github.com/zwh8800/cdnd/internal/save"
)

// GameModel 游戏主界面模型
type GameModel struct {
	engine *game.Engine
	ctx    context.Context

	// UI状态
	width  int
	height int
	ready  bool

	// 输入
	input textinput.Model

	// 输出
	output      []string
	outputIndex int

	// 状态
	phase   save.GamePhase
	loading bool
	err     error
}

// NewGameModel 创建游戏模型
func NewGameModel(engine *game.Engine) GameModel {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Focus()

	return GameModel{
		engine: engine,
		ctx:    context.Background(),
		output: make([]string, 0),
		input:  ti,
	}
}

// Init 初始化
func (m GameModel) Init() tea.Cmd {
	return nil
}

// Update 更新
func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.input.Width = msg.Width - 6
		return m, nil

	case DMResponseMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			m.output = append(m.output, fmt.Sprintf("错误: %v", msg.Err))
		} else {
			m.output = append(m.output, msg.Content)
			m.phase = msg.Phase
		}
		return m, nil
	}

	// 更新 textinput
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// handleKeyPress 处理按键
func (m GameModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return m, tea.Quit

	case tea.KeyEnter:
		if m.loading {
			return m, nil
		}
		if m.input.Value() == "" {
			return m, nil
		}

		// 添加玩家输入到输出
		m.output = append(m.output, fmt.Sprintf("> %s", m.input.Value()))

		// 发送到引擎
		input := m.input.Value()
		m.input.SetValue("")
		m.loading = true
		return m, m.processInput(input)

	case tea.KeyUp:
		// 滚动输出
		if m.outputIndex > 0 {
			m.outputIndex--
		}
		return m, nil

	case tea.KeyDown:
		if m.outputIndex < len(m.output)-1 {
			m.outputIndex++
		}
		return m, nil

	default:
		// 其他按键交给 textinput 处理
		m.input, _ = m.input.Update(msg)
		return m, nil
	}
}

// processInput 处理输入命令
func (m GameModel) processInput(input string) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.engine.ProcessPlayerInput(m.ctx, input)
		if err != nil {
			return DMResponseMsg{Err: err}
		}
		return DMResponseMsg{
			Content: resp.Content,
			Phase:   resp.Phase,
		}
	}
}

// View 渲染
func (m GameModel) View() string {
	if !m.ready {
		return "正在加载..."
	}

	var b strings.Builder

	// 状态栏
	b.WriteString(m.renderStatusBar())
	b.WriteString("\n")

	// 主输出区域
	outputHeight := m.height - 6 // 预留状态栏、输入栏、边框
	b.WriteString(m.renderOutput(outputHeight))
	b.WriteString("\n")

	// 输入栏
	b.WriteString(m.renderInput())

	return b.String()
}

// renderStatusBar 渲染状态栏
func (m GameModel) renderStatusBar() string {
	c := m.engine.GetCharacter()
	if c == nil {
		return GameStyles.Title.Render("D&D CLI - 无角色")
	}

	hpStyle := lipgloss.NewStyle()
	if c.HitPoints.Current <= c.HitPoints.Max/4 {
		hpStyle = hpStyle.Foreground(lipgloss.Color("#ff0000"))
	} else if c.HitPoints.Current <= c.HitPoints.Max/2 {
		hpStyle = hpStyle.Foreground(lipgloss.Color("#ffaa00"))
	} else {
		hpStyle = hpStyle.Foreground(lipgloss.Color("#00ff00"))
	}

	hpText := fmt.Sprintf("HP: %d/%d", c.HitPoints.Current, c.HitPoints.Max)
	left := fmt.Sprintf("%s - %s %d级", c.Name, c.Race.Name, c.Level)
	if c.HasClass() {
		left = fmt.Sprintf("%s - %s %d级 %s", c.Name, c.Race.Name, c.Level, c.Class.Name)
	}
	right := fmt.Sprintf("%s | %s", hpStyle.Render(hpText), m.phase.String())

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		GameStyles.StatusBar.Render(left),
		strings.Repeat(" ", maxInt(0, m.width-lipgloss.Width(GameStyles.StatusBar.Render(left))-lipgloss.Width(right)-2)),
		GameStyles.StatusBar.Render(right),
	)

	return bar
}

// renderOutput 渲染输出区域
func (m GameModel) renderOutput(height int) string {
	if len(m.output) == 0 {
		return GameStyles.Box.Render("欢迎来到D&D冒险！输入你的行动开始游戏。")
	}

	// 计算显示范围
	start := 0
	if len(m.output) > height {
		start = len(m.output) - height
	}

	lines := m.output[start:]
	if len(lines) > height {
		lines = lines[len(lines)-height:]
	}

	content := strings.Join(lines, "\n")
	return GameStyles.Box.Height(height).Render(content)
}

// renderInput 渲染输入栏
func (m GameModel) renderInput() string {
	if m.loading {
		return GameStyles.InputBox.Render("处理中... ")
	}
	return GameStyles.InputBox.Render(m.input.View())
}

// DMResponseMsg DM响应消息
type DMResponseMsg struct {
	Content string
	Phase   save.GamePhase
	Err     error
}

// maxInt 辅助函数
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GameStyles 样式定义
var GameStyles = struct {
	Title     lipgloss.Style
	StatusBar lipgloss.Style
	Box       lipgloss.Style
	InputBox  lipgloss.Style
	Highlight lipgloss.Style
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
}
