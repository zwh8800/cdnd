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
	windowWidth  int
	windowHeight int

	// 上：状态栏相关数据
	phase           save.GamePhase
	expanded        bool // 是否展开状态栏
	statusBarHeight int  // 状态栏实际高度

	// 中：剧情框相关数据
	viewport viewport.Model
	lines    []string

	// 下：输入框相关数据
	inputBox           textinput.Model
	loadingFrame       int // Braille 动画帧索引 (0-5)
	loadingTimer       int // 进度点动画帧 (0-3)
	loading            bool
	loadingPhraseCount int // 加载文案计数器
}

// NewGameModel 创建游戏模型
func NewGameModel(engine *game.Engine) *GameModel {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Focus()

	vp := viewport.New(0, 0)
	vp.SetHorizontalStep(10)

	return &GameModel{
		engine:          engine,
		ctx:             context.Background(),
		lines:           make([]string, 0),
		inputBox:        ti,
		viewport:        vp,
		statusBarHeight: 2,
	}
}

// Init 初始化
func (m *GameModel) Init() tea.Cmd {
	// 根据历史是否为空判断是新游戏还是载入存档
	if len(m.engine.GetState().History) == 0 {
		// 新游戏：分两阶段初始化
		// 阶段 1：立即显示欢迎消息（快速）
		m.showWelcomeMessage()
		// 阶段 2：异步触发 LLM 对话
		// 添加初始输入到输出
		const initPrompt = "我是谁 我在哪"
		m.lines = append(m.lines, fmt.Sprintf(strings.Repeat("-", 50)+"\n> %s", initPrompt))
		m.updateViewportContent()
		m.loading = true
		return tea.Batch(m.processInput(initPrompt), m.startLoadingAnimation())
	} else {
		// 载入存档：恢复历史对话
		m.restoreHistory()
		return nil
	}
}

// Update 更新
func (m *GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.inputBox.Value() == "" {
				return m, nil
			}

			// 添加玩家输入到输出
			m.lines = append(m.lines, fmt.Sprintf(strings.Repeat("-", m.windowWidth)+"\n> %s", m.inputBox.Value()))
			m.updateViewportContent()
			m.viewport.PageDown()

			// 发送到引擎
			input := m.inputBox.Value()
			m.inputBox.SetValue("")
			m.loading = true
			m.loadingFrame = 0
			m.loadingTimer = 0
			return m, tea.Batch(append(cmds, m.processInput(input), m.startLoadingAnimation())...)
		}

		// 其他按键交给 textinput 处理
		m.inputBox, cmd = m.inputBox.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		m.recalculateViewport()

	case DMResponseMsg:
		m.loading = false
		if msg.Err != nil {
			m.lines = append(m.lines, fmt.Sprintf("错误: %v", msg.Err))
		} else {
			// 先显示工具叙述
			for _, narrative := range msg.ToolNarratives {
				m.lines = append(m.lines, narrative)
			}
			// 再显示DM响应内容
			m.lines = append(m.lines, msg.Content)
			m.phase = msg.Phase
		}
		m.updateViewportContent()
		m.viewport.GotoBottom()

	case LoadingTickMsg:
		// 更新加载动画帧
		if m.loading {
			m.loadingFrame = (m.loadingFrame + 1) % 6
			m.loadingTimer = (m.loadingTimer + 1) % 4
			m.loadingPhraseCount++ // 每次tick增加计数器
			return m, tea.Batch(append(cmds, m.startLoadingAnimation())...)
		}
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// showWelcomeMessage 立即显示欢迎消息（快速，无 LLM 调用）
func (m *GameModel) showWelcomeMessage() {
	welcomeMsg, err := m.engine.ShowWelcomeMessage()
	if err != nil {
		m.lines = append(m.lines, fmt.Sprintf("错误: %v", err))
		return
	}

	// 立即显示欢迎消息
	m.lines = append(m.lines, welcomeMsg)
	m.updateViewportContent()
}

// restoreHistory 恢复存档的对话历史到UI
func (m *GameModel) restoreHistory() {
	history := m.engine.GetState().History
	if len(history) == 0 {
		return
	}

	// 遍历历史消息，格式化后添加到 lines
	for _, msg := range history {
		switch msg.Role {
		case "user":
			// 玩家输入，添加分隔线和 "> " 前缀
			separator := strings.Repeat("-", 50)
			m.lines = append(m.lines, separator)
			m.lines = append(m.lines, "> "+msg.Content)
		case "assistant":
			// DM响应，直接添加（已包含ANSI颜色代码）
			m.lines = append(m.lines, msg.Content)
		case "tool":
			// 工具消息，跳过（不应出现在最终显示中）
			continue
		}
	}

	// 更新游戏阶段和视口
	m.phase = m.engine.GetState().Phase
	m.updateViewportContent()
	m.viewport.GotoBottom()
}

// DMResponseMsg DM响应消息
type DMResponseMsg struct {
	Content        string
	Phase          save.GamePhase
	ToolNarratives []string
	Err            error
}

// processInput 处理输入命令（使用 Tool Call 版本）
func (m *GameModel) processInput(input string) tea.Cmd {
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

// LoadingTickMsg 加载动画计时消息
type LoadingTickMsg time.Time

// startLoadingAnimation 启动加载动画计时器
func (m *GameModel) startLoadingAnimation() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return LoadingTickMsg(t)
	})
}

// recalculateViewport 重新计算视口尺寸
func (m *GameModel) recalculateViewport() {
	// 计算 UI 组件高度
	inputBoxHeight := 3
	separatorHeight := 1

	// GameStyles.Box 有 Border(RoundedBorder) 和 Padding(0, 1)
	// Border 占 2 行高度（上下各 1）
	viewportHeight := m.windowHeight - m.statusBarHeight - inputBoxHeight - separatorHeight - 2
	viewportWidth := m.windowWidth - 4 // 左右 border(2) + padding(2)

	if viewportHeight < 1 {
		viewportHeight = 1
	}
	if viewportWidth < 1 {
		viewportWidth = 1
	}

	m.viewport.Width = viewportWidth
	m.viewport.Height = viewportHeight
	m.inputBox.Width = m.windowWidth - 7

	m.updateViewportContent()
}

// updateViewportContent 更新 viewport 内容
func (m *GameModel) updateViewportContent() {
	content := strings.Join(m.lines, "\n")
	m.viewport.SetContent(content)
}

// View 渲染
func (m *GameModel) View() string {
	var b strings.Builder

	// 状态栏
	b.WriteString(m.renderStatusBar())
	b.WriteString("\n")

	// 主输出区域
	outputHeight := m.windowHeight - m.statusBarHeight - 4 // 预留输入栏、边框等
	b.WriteString(m.renderStoryBox(outputHeight))
	b.WriteString("\n")

	// 输入栏
	b.WriteString(m.renderInputBox())

	return b.String()
}

// renderStatusBar 渲染状态栏
func (m *GameModel) renderStatusBar() string {
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

// renderStoryBox 渲染剧情框
func (m *GameModel) renderStoryBox(height int) string {
	// 直接使用 viewport 的 View() 方法
	// 注意：内容已经在 Update 中通过 updateViewportContent 设置
	return GameStyles.Box.Height(height).Render(m.viewport.View())
}

// renderInputBox 渲染输入栏
func (m *GameModel) renderInputBox() string {
	if m.loading {
		// Braille 旋转器
		braille := brailleFrames[m.loadingFrame]

		// 进度点动画
		progressDots := strings.Repeat("·", m.loadingTimer) + strings.Repeat(" ", max(0, 3-m.loadingTimer))

		// D&D风格的加载文案池
		loadingPhrases := []string{
			"DM正在构思剧情",
			"骰子正在滚动",
			"魔法正在生效",
			"命运之轮在转动",
			"地下城主在思考",
			"冒险即将展开",
			"神秘力量在涌动",
			"故事正在编织",
			"龙息正在酝酿",
			"传送门正在开启",
			"卷轴正在解读",
			"预言水晶在闪烁",
			"地下城迷雾在散去",
			"英雄的命运在召唤",
			"古老符文在发光",
		}

		// 每5帧切换一次文案，充分利用所有15个文案
		phraseIndex := (m.loadingPhraseCount / 6) % len(loadingPhrases)
		loadingText := fmt.Sprintf("%s %s %s", braille, loadingPhrases[phraseIndex], progressDots)
		return GameStyles.InputBox.Render(loadingText)
	}
	return GameStyles.InputBox.Render(m.inputBox.View())
}
