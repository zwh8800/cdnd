# Viewport 滚动功能修复方案

## 背景

当前 UI 界面存在两个主要问题：
1. **滚动功能失效** - 输出内容超出屏幕时无法滚动查看历史内容
2. **自适应宽高调整缺失** - 终端窗口大小改变时界面布局没有相应调整

## bubbletea 库关键发现

### viewport.Model 工作原理

```go
// viewport.go 核心结构
type Model struct {
    Width  int      // 视口宽度
    Height int      // 视口高度
    YOffset int     // 当前滚动位置
    lines []string  // 内容行
    KeyMap KeyMap   // 键盘绑定
}

// Update 方法处理键盘事件
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // 匹配 KeyMap.Up -> ScrollUp(1)
        // 匹配 KeyMap.Down -> ScrollDown(1)
        // 匹配 KeyMap.PageUp -> PageUp()
        // 匹配 KeyMap.PageDown -> PageDown()
    }
}
```

### 默认 KeyMap 绑定
- **Up**: "up", "k" - 向上滚动一行
- **Down**: "down", "j" - 向下滚动一行
- **PageUp**: "pgup", "b" - 向上翻页
- **PageDown**: "pgdown", " ", "f" - 向下翻页

### WindowSizeMsg 机制
```go
// screen.go
type WindowSizeMsg struct {
    Width  int
    Height int
}
```
- 程序启动时自动发送一次
- 终端大小变化时自动发送
- 需要在 Update 中处理并更新组件尺寸

## 问题根因分析

### 问题 1: viewport.Update() 未被正确调用

当前代码在 `handleKeyPress` 中处理 KeyUp/KeyDown：
```go
case tea.KeyUp, tea.KeyDown:
    var cmd tea.Cmd
    m.viewport, cmd = m.viewport.Update(msg)
    return m, cmd
```

**问题**: 这只在 `handleKeyPress` 中调用，而 `handleKeyPress` 会在 `Update` 中被提前返回，导致其他消息处理时无法更新 viewport。

### 问题 2: 指针接收器问题

`updateViewportContent` 已改为指针接收器，但调用它的地方（如 `Update` 方法中的 case 分支）使用的是值接收器的 `GameModel`，这导致修改不会保存。

### 问题 3: viewport 尺寸计算

当前计算：
```go
viewportHeight := m.height - 6
viewportWidth := m.width - 2
```

需要验证这个计算是否正确匹配实际 UI 布局。

### 问题 4: 消息流被阻断

在 `Update` 方法中，`case tea.KeyMsg` 直接返回 `handleKeyPress`，阻止了 viewport 接收到其他必要消息（如鼠标滚轮事件）。

## 修复方案

### 核心修改：重构消息处理流程

**关键原则**: viewport 应该始终接收到消息，而不是在某些条件下被阻断。

#### 修改 internal/ui/game.go

##### 1. 修改 Update 方法结构

```go
func (m *GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    // 始终让 viewport 处理消息（滚动、鼠标滚轮等）
    newViewport, cmd := m.viewport.Update(msg)
    m.viewport = newViewport
    cmds = append(cmds, cmd)

    switch msg := msg.(type) {
    case tea.KeyMsg:
        // 处理特殊按键，但不阻断消息流
        switch msg.Type {
        case tea.KeyCtrlC, tea.KeyEsc:
            return m, tea.Quit
        case tea.KeyEnter:
            // 处理输入...
        }
        // 其他按键交给 textinput 处理
        m.input, cmd = m.input.Update(msg)
        cmds = append(cmds, cmd)

    case tea.WindowSizeMsg:
        // 处理窗口大小变化...
    }

    return m, tea.Batch(cmds...)
}
```

##### 2. 正确处理 WindowSizeMsg

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    m.ready = true

    // 计算 viewport 尺寸
    // 状态栏: 约 2 行（包含 padding）
    // 输入栏: 约 3 行（包含 border）
    // 分隔符: 约 1 行
    headerHeight := 2
    inputHeight := 3
    separatorHeight := 1

    // Box 样式有 Padding(0, 1) 和 Border
    // Border 占用 2 行（上下各 1），Padding 不占额外行
    viewportHeight := m.height - headerHeight - inputHeight - separatorHeight - 2 // -2 for border
    viewportWidth := m.width - 4  // -2 for left/right border and padding

    if viewportHeight < 1 {
        viewportHeight = 1
    }
    if viewportWidth < 1 {
        viewportWidth = 1
    }

    m.viewport.Width = viewportWidth
    m.viewport.Height = viewportHeight
    m.input.Width = m.width - 6

    // 重新设置内容以触发重新计算
    m.updateViewportContent()
```

##### 3. 确保 updateViewportContent 被正确调用

在所有修改 output 的地方调用 `updateViewportContent()`：
- DMResponseMsg 处理后
- StreamChunkMsg 处理后
- KeyEnter 处理后

##### 4. 移除 handleKeyPress 中的滚动处理

将滚动事件交给 viewport 的默认 KeyMap 处理，不再在 handleKeyPress 中特殊处理 KeyUp/KeyDown。

### 验证方法

1. 运行程序：`go run ./cmd/load.go`
2. 输入多行内容使输出超出屏幕
3. 测试滚动：
   - 按 Up/Down 键滚动
   - 按 PageUp/PageDown 翻页
   - 使用鼠标滚轮（如果终端支持）
4. 测试窗口调整：
   - 拖动终端窗口边界
   - 确认内容区域自动调整大小

## 修改文件清单

| 文件 | 修改内容 |
|------|----------|
| `internal/ui/game.go` | 重构 Update 方法，正确处理 viewport 和 WindowSizeMsg |

## 详细代码修改

### game.go 完整修改

#### Update 方法重构

```go
func (m *GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

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
        statusBarHeight := 2   // 状态栏 + 换行
        inputBoxHeight := 3    // 输入框 + border
        separatorHeight := 1   // 组件间换行

        // 计算 viewport 尺寸
        // GameStyles.Box 有 Border(RoundedBorder) 和 Padding(0, 1)
        // Border 占 2 行高度，左右 padding 不占行
        viewportHeight := m.height - statusBarHeight - inputBoxHeight - separatorHeight - 2
        viewportWidth := m.width - 4  // 左右 border(2) + padding(2)

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
            for _, narrative := range msg.ToolNarratives {
                m.output = append(m.output, narrative)
            }
            m.output = append(m.output, msg.Content)
            m.phase = msg.Phase
        }
        m.updateViewportContent()
        m.viewport.GotoBottom()

    case StreamChunkMsg:
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

        m.streamingContent += msg.Content
        m.updateViewportContent()
        m.viewport.GotoBottom()
        return m, tea.Batch(append(cmds, waitForStreamChunks(msg.stream))...)

    case StreamStartMsg:
        m.isStreaming = true
        m.streamingContent = ""
        return m, tea.Batch(append(cmds, waitForStreamChunks(msg.Stream))...)
    }

    return m, tea.Batch(cmds...)
}
```

#### 删除 handleKeyPress 方法

将必要的按键处理逻辑移入 Update 方法后，删除 handleKeyPress 方法。

#### 保持 updateViewportContent 为指针接收器

```go
func (m *GameModel) updateViewportContent() {
    var lines []string
    lines = append(lines, m.output...)
    if m.isStreaming && m.streamingContent != "" {
        lines = append(lines, m.streamingContent)
    }
    content := strings.Join(lines, "\n")
    m.viewport.SetContent(content)
}
```

## 预期效果

1. **滚动正常工作**: Up/Down 键可以滚动内容，PageUp/PageDown 可以翻页
2. **窗口自适应**: 调整终端窗口大小时，UI 组件自动重新布局
3. **新内容自动定位**: 新内容到达时自动滚动到底部
4. **保持现有功能**: 输入、流式输出、状态显示等功能不受影响
