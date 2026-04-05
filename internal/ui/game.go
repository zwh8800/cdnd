package ui

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zwh8800/cdnd/internal/game"
	"github.com/zwh8800/cdnd/internal/game/state"
	"github.com/zwh8800/cdnd/internal/llm/prompt"
)

// 加载动画常量
var brailleFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠦"}

// 输入模式常量
const (
	inputModeText   = "text"   // 文本输入模式
	inputModeSelect = "select" // 选择模式
)

// 切换到文本输入的选项标签
const otherOptionLabel = "其他行动..."

// optionItem 选项列表项
type optionItem string

func (i optionItem) FilterValue() string { return "" }

// optionDelegate 自定义选项列表委托
type optionDelegate struct {
	normalStyle   lipgloss.Style
	selectedStyle lipgloss.Style
}

func (d optionDelegate) Height() int                             { return 1 }
func (d optionDelegate) Spacing() int                            { return 0 }
func (d optionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d optionDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(optionItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, string(i))

	if index == m.Index() {
		fmt.Fprint(w, d.selectedStyle.Render("> "+str))
	} else {
		fmt.Fprint(w, d.normalStyle.Render(str))
	}
}

// GameModel 游戏主界面模型
type GameModel struct {
	engine *game.Engine
	ctx    context.Context

	// UI状态
	windowWidth  int
	windowHeight int

	// 上：状态栏相关数据
	phase           state.GamePhase
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
	loadingPhraseCount int    // 加载文案计数器
	inputMode          string // "text" 或 "select"
	optionsList        list.Model
	currentOptions     []string
}

// NewGameModel 创建游戏模型
func NewGameModel(engine *game.Engine) *GameModel {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Focus()

	vp := viewport.New(0, 0)
	vp.SetHorizontalStep(10)

	// 初始化选项列表（使用自定义delegate支持翻页）
	opts := []list.Item{}
	delegate := optionDelegate{
		normalStyle:   lipgloss.NewStyle().PaddingLeft(4),
		selectedStyle: lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170")),
	}
	optionsList := list.New(opts, delegate, 0, 1)
	optionsList.SetShowStatusBar(false)
	optionsList.SetFilteringEnabled(false)
	optionsList.SetShowHelp(false)
	optionsList.SetShowTitle(false)

	return &GameModel{
		engine:          engine,
		ctx:             context.Background(),
		lines:           make([]string, 0),
		inputBox:        ti,
		viewport:        vp,
		statusBarHeight: 2,
		inputMode:       inputModeText,
		optionsList:     optionsList,
		currentOptions:  []string{},
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

// getInputHeight 根据输入模式获取输入区域高度
func (m *GameModel) getInputHeight() int {
	if m.inputMode == inputModeSelect {
		return 5
	} else {
		return 3
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
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			// 如果在文本输入模式且有可用选项，切换到选择模式
			if m.inputMode == inputModeText && len(m.currentOptions) > 0 && !m.loading {
				m.inputMode = inputModeSelect
				m.recalculateViewport()
				return m, nil
			}
			// 否则退出程序
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

			// 选择模式下的Enter处理
			if m.inputMode == inputModeSelect {
				selected := m.optionsList.SelectedItem()
				if selected != nil {
					item := selected.(optionItem)
					if string(item) == otherOptionLabel {
						// 切换到文本输入模式
						m.inputMode = inputModeText
						m.recalculateViewport()
						return m, nil
					}
					// 发送选中选项作为输入
					return m.handleInput(string(item))
				}
				return m, nil
			}

			// 文本输入模式下的Enter处理
			if m.inputMode == inputModeText {
				if m.inputBox.Value() == "" {
					return m, nil
				}
				return m.handleInput(m.inputBox.Value())
			}

		case tea.KeyLeft, tea.KeyRight:
			// 选择模式下，上下键由optionsList处理（支持翻页）
			if m.inputMode == inputModeSelect && !m.loading && len(m.currentOptions) > 0 {
				m.optionsList, cmd = m.optionsList.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				return m, tea.Batch(cmds...)
			}
			// 其他情况由viewport处理
		}

		// 文本输入模式下，其他按键交给 textinput 处理
		if m.inputMode == inputModeText {
			m.inputBox, cmd = m.inputBox.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
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
			m.lines = append(m.lines, prompt.ParseColorMarkers(msg.Content))
			m.phase = msg.Phase
			// 更新选项
			m.updateOptions(msg.Options)
		}
		m.updateViewportContent()
		m.viewport.GotoBottom()

	case LoadingTickMsg:
		// 更新加载动画帧
		if m.loading {
			m.inputMode = inputModeText
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
			// DM响应，直接添加（未包含ANSI颜色代码）
			m.lines = append(m.lines, prompt.ParseColorMarkers(msg.Content))
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
	Phase          state.GamePhase
	ToolNarratives []string
	Options        []string
	Err            error
}

// processInput 处理输入命令（使用 Tool Call 版本）
func (m *GameModel) processInput(input string) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.engine.Process(m.ctx, input)
		if err != nil {
			return DMResponseMsg{Err: err}
		}
		return DMResponseMsg{
			Content:        resp.Content,
			Phase:          resp.Phase,
			ToolNarratives: resp.ToolNarratives,
			Options:        resp.Options,
		}
	}
}

// handleInput 统一处理输入（无论是选择还是文本输入）
func (m *GameModel) handleInput(input string) (tea.Model, tea.Cmd) {
	// 添加玩家输入到输出
	m.lines = append(m.lines, fmt.Sprintf(strings.Repeat("-", m.windowWidth-4)+"\n> %s", input))
	m.updateViewportContent()
	m.viewport.PageDown()

	// 重置输入状态
	m.inputBox.SetValue("")
	m.loading = true
	m.loadingFrame = 0
	m.loadingTimer = 0

	return m, tea.Batch(m.processInput(input), m.startLoadingAnimation())
}

// updateOptions 更新当前选项
func (m *GameModel) updateOptions(options []string) {
	m.currentOptions = options

	if len(options) > 0 {
		// 有选项时切换到选择模式
		m.inputMode = inputModeSelect

		// 构建列表项（添加"其他行动..."选项）
		items := make([]list.Item, 0, len(options)+1)
		for _, opt := range options {
			items = append(items, optionItem(prompt.ParseColorMarkers(opt)))
		}
		items = append(items, optionItem(otherOptionLabel))

		m.optionsList.SetItems(items)
		m.optionsList.ResetSelected()
	} else {
		// 无选项时切换到文本模式
		m.inputMode = inputModeText
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
	// 根据输入模式动态计算高度
	inputHeight := m.getInputHeight()
	separatorHeight := 1

	// GameStyles.Box 有 Border(RoundedBorder) 和 Padding(0, 1)
	// Border 占 2 行高度（上下各 1）
	viewportHeight := m.windowHeight - m.statusBarHeight - inputHeight - separatorHeight - 2
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

	// 更新选项列表尺寸
	m.optionsList.SetWidth(m.windowWidth - 4)

	m.updateViewportContent()
}

// updateViewportContent 更新 viewport 内容
func (m *GameModel) updateViewportContent() {
	// 计算 viewport 的内容宽度（减去 border 和 padding）
	viewportWidth := m.viewport.Width

	// 如果宽度小于等于 0，直接连接文本
	if viewportWidth <= 0 {
		content := strings.Join(m.lines, "\n")
		m.viewport.SetContent(content)
		return
	}

	// 使用 lipgloss 对每一行进行自动换行
	var wrappedLines []string
	wrapStyle := lipgloss.NewStyle().Width(viewportWidth)

	for _, line := range m.lines {
		// 对每一行应用宽度样式，lipgloss 会自动换行
		wrappedLines = append(wrappedLines, wrapStyle.Render(line))
	}

	content := strings.Join(wrappedLines, "\n")
	m.viewport.SetContent(content)
}

// View 渲染
func (m *GameModel) View() string {
	var b strings.Builder

	// 状态栏
	b.WriteString(m.renderStatusBar())
	b.WriteString("\n")

	// 主输出区域 - 根据输入模式动态计算高度
	inputHeight := m.getInputHeight()
	outputHeight := m.windowHeight - m.statusBarHeight - inputHeight - 1 // 1行分隔符
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
	if height < 1 {
		height = 1
	}
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

	// 选择模式：显示选项列表（支持翻页）
	if m.inputMode == inputModeSelect {
		return GameStyles.InputBox.Render(m.optionsList.View())
	} else {
		return GameStyles.InputBox.Render(m.inputBox.View())
	}
}
