package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
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
	output []string

	// 视口（用于滚动显示）
	viewport viewport.Model

	// 流式输出缓冲区
	streamingContent string
	isStreaming      bool

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

	vp := viewport.New(0, 0)
	vp.MouseWheelEnabled = true

	return GameModel{
		engine:   engine,
		ctx:      context.Background(),
		output:   make([]string, 0),
		input:    ti,
		viewport: vp,
	}
}

// Init 初始化
func (m GameModel) Init() tea.Cmd {
	return nil
}

// Update 更新
func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// 始终让 viewport 处理消息（滚动、鼠标滚轮等）
	m.viewport, cmd = m.viewport.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 处理特殊按键
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
			m.updateViewportContent()
			m.viewport.GotoBottom()

			// 发送到引擎
			input := m.input.Value()
			m.input.SetValue("")
			m.loading = true
			return m, tea.Batch(append(cmds, m.processInput(input))...)
		}

		// 其他按键交给 textinput 处理
		m.input, cmd = m.input.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// 计算 UI 组件高度
		// 状态栏: 约 2 行（内容 + 换行）
		// 输入框: 约 3 行（内容 + border）
		// 组件间换行: 约 1 行
		statusBarHeight := 2
		inputBoxHeight := 3
		separatorHeight := 1

		// GameStyles.Box 有 Border(RoundedBorder) 和 Padding(0, 1)
		// Border 占 2 行高度（上下各 1）
		viewportHeight := m.height - statusBarHeight - inputBoxHeight - separatorHeight - 2
		viewportWidth := m.width - 4 // 左右 border(2) + padding(2)

		if viewportHeight < 1 {
			viewportHeight = 1
		}
		if viewportWidth < 1 {
			viewportWidth = 1
		}

		m.viewport.Width = viewportWidth
		m.viewport.Height = viewportHeight
		m.input.Width = m.width - 6

		m.updateViewportContent()

	case DMResponseMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			m.output = append(m.output, fmt.Sprintf("错误: %v", msg.Err))
		} else {
			// 先显示工具叙述
			for _, narrative := range msg.ToolNarratives {
				m.output = append(m.output, narrative)
			}
			// 再显示DM响应内容
			m.output = append(m.output, msg.Content)
			m.phase = msg.Phase
		}
		m.updateViewportContent()
		m.viewport.GotoBottom()

	case StreamChunkMsg:
		// 处理流式数据块
		if msg.Error != nil {
			m.loading = false
			m.isStreaming = false
			m.err = msg.Error
			m.output = append(m.output, fmt.Sprintf("错误: %v", msg.Error))
			m.updateViewportContent()
			m.viewport.GotoBottom()
			return m, tea.Batch(cmds...)
		}

		if msg.Done {
			// 流式完成
			m.loading = false
			m.isStreaming = false
			if m.streamingContent != "" {
				m.output = append(m.output, m.streamingContent)
			}
			m.streamingContent = ""
			m.updateViewportContent()
			m.viewport.GotoBottom()
			return m, tea.Batch(cmds...)
		}

		// 累积流式内容并继续等待下一个数据块
		m.streamingContent += msg.Content
		m.updateViewportContent()
		m.viewport.GotoBottom()
		return m, tea.Batch(append(cmds, waitForStreamChunks(msg.stream))...)

	case StreamStartMsg:
		// 开始流式输出
		m.isStreaming = true
		m.streamingContent = ""
		return m, tea.Batch(append(cmds, waitForStreamChunks(msg.Stream))...)
	}

	return m, tea.Batch(cmds...)
}

// processInput 处理输入命令（使用 Tool Call 版本）
func (m GameModel) processInput(input string) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.engine.ProcessWithTools(m.ctx, input)
		if err != nil {
			return DMResponseMsg{Err: err}
		}
		return DMResponseMsg{
			Content:        resp.Content,
			Phase:          resp.Phase,
			ToolNarratives: resp.ToolNarratives,
		}
	}
}

// waitForStreamChunks 等待流式数据块
func waitForStreamChunks(stream <-chan game.StreamChunk) tea.Cmd {
	return func() tea.Msg {
		chunk, ok := <-stream
		if !ok {
			return StreamChunkMsg{Done: true}
		}
		return StreamChunkMsg{
			Content: chunk.Content,
			Done:    chunk.Done,
			Error:   chunk.Error,
			stream:  stream,
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
	if len(m.output) == 0 && !m.isStreaming {
		return GameStyles.Box.Height(height).Render("欢迎来到D&D冒险！输入你的行动开始游戏。")
	}

	// 直接使用 viewport 的 View() 方法
	// 注意：内容已经在 Update 中通过 updateViewportContent 设置
	return GameStyles.Box.Height(height).Render(m.viewport.View())
}

// updateViewportContent 更新 viewport 内容
func (m *GameModel) updateViewportContent() {
	// 构建输出内容，在每条消息之间添加分隔线
	var lines []string

	for _, output := range m.output {
		lines = append(lines, output)
	}

	// 如果正在流式输出，添加当前流式内容
	if m.isStreaming && m.streamingContent != "" {
		lines = append(lines, m.streamingContent)
	}

	content := strings.Join(lines, "\n")
	m.viewport.SetContent(content)
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
	Content        string
	Phase          save.GamePhase
	ToolNarratives []string
	Err            error
}

// StreamStartMsg 流式开始消息
type StreamStartMsg struct {
	Stream <-chan game.StreamChunk
}

// StreamChunkMsg 流式数据块消息
type StreamChunkMsg struct {
	Content string
	Done    bool
	Error   error
	stream  <-chan game.StreamChunk
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
