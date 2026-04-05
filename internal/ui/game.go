package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zwh8800/cdnd/internal/game"
	"github.com/zwh8800/cdnd/internal/save"
)

// 加载动画常量
var brailleFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠦"}

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

	// 加载动画状态
	loadingFrame int // Braille 动画帧索引 (0-5)
	loadingTimer int // 进度点动画帧 (0-3)

	// 状态栏模式
	expanded        bool // 是否展开状态栏
	statusBarHeight int  // 状态栏实际高度
}

// NewGameModel 创建游戏模型
func NewGameModel(engine *game.Engine) GameModel {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Focus()

	vp := viewport.New(0, 0)
	vp.MouseWheelEnabled = true

	return GameModel{
		engine:          engine,
		ctx:             context.Background(),
		output:          make([]string, 0),
		input:           ti,
		viewport:        vp,
		statusBarHeight: 2,
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

		case tea.KeyTab:
			m.expanded = !m.expanded
			if m.expanded {
				m.statusBarHeight = 12
			} else {
				m.statusBarHeight = 2
			}
			m.recalculateViewport()
			m.viewport.GotoBottom()
			return m, nil

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
			m.loadingFrame = 0
			m.loadingTimer = 0
			return m, tea.Batch(append(cmds, m.processInput(input), startLoadingAnimation())...)
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
		m.recalculateViewport()

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

	case LoadingTickMsg:
		// 更新加载动画帧
		if m.loading {
			m.loadingFrame = (m.loadingFrame + 1) % 6
			m.loadingTimer = (m.loadingTimer + 1) % 4
			return m, tea.Batch(append(cmds, startLoadingAnimation())...)
		}
		// loading 已为 false，停止动画
		return m, nil
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

// recalculateViewport 重新计算视口尺寸
func (m *GameModel) recalculateViewport() {
	// 计算 UI 组件高度
	inputBoxHeight := 3
	separatorHeight := 1

	// GameStyles.Box 有 Border(RoundedBorder) 和 Padding(0, 1)
	// Border 占 2 行高度（上下各 1）
	viewportHeight := m.height - m.statusBarHeight - inputBoxHeight - separatorHeight - 2
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
	outputHeight := m.height - m.statusBarHeight - 4 // 预留输入栏、边框等
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

	if m.expanded {
		return m.renderStatusBarExpanded(c)
	} else {
		return m.renderStatusBarCompact(c)
	}
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
		// Braille 旋转器
		braille := brailleFrames[m.loadingFrame]

		// 进度点动画
		progressDots := strings.Repeat("·", m.loadingTimer) + strings.Repeat(" ", maxInt(0, 3-m.loadingTimer))

		loadingText := fmt.Sprintf("%s 处理中 %s", braille, progressDots)
		return GameStyles.InputBox.Render(loadingText)
	}
	return GameStyles.InputBox.Render(m.input.View())
}

// LoadingTickMsg 加载动画计时消息
type LoadingTickMsg time.Time

// DMResponseMsg DM响应消息
type DMResponseMsg struct {
	Content        string
	Phase          save.GamePhase
	ToolNarratives []string
	Err            error
}

// startLoadingAnimation 启动加载动画计时器
func startLoadingAnimation() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return LoadingTickMsg(t)
	})
}

// maxInt 辅助函数
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
