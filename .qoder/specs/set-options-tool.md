# Set Options Tool 实施计划

## 背景与目标

当前游戏UI使用纯文本输入框让玩家输入行动指令。为了提升用户体验，我们希望引入选择/输入混合模式：
- DM通过`set_options`工具提供可选操作列表
- UI优先显示选择框，允许玩家快速选择预设选项
- 同时保留自由输入的能力

## 实施步骤

### 步骤1: 修改游戏状态 (internal/game/state.go)

在游戏状态中添加当前选项字段：

```go
// State 游戏状态
type State struct {
    // ... 现有字段 ...
    
    // 当前可用的操作选项（由DM通过set_options工具设置）
    CurrentOptions []string `json:"current_options,omitempty"`
}
```

添加访问方法：
```go
// SetCurrentOptions 设置当前选项
func (s *State) SetCurrentOptions(options []string) {
    s.CurrentOptions = options
}

// GetCurrentOptions 获取当前选项
func (s *State) GetCurrentOptions() []string {
    return s.CurrentOptions
}

// ClearCurrentOptions 清除当前选项
func (s *State) ClearCurrentOptions() {
    s.CurrentOptions = nil
}
```

### 步骤2: 创建 set_options 工具 (internal/tools/options_tool.go)

创建新文件，实现SetOptionsTool：

```go
package tools

import (
    "context"
)

// SetOptionsTool 设置选项工具
type SetOptionsTool struct {
    BaseTool
    state StateAccessor
}

// NewSetOptionsTool 创建设置选项工具
func NewSetOptionsTool(state StateAccessor) *SetOptionsTool {
    return &SetOptionsTool{
        BaseTool: NewBaseTool("set_options", "设置玩家当前可用的操作选项。每次响应时必须调用此工具提供可选操作列表。参数: options (字符串数组，每个元素是一个可选操作)"),
        state:    state,
    }
}

// Parameters 返回参数定义
func (t *SetOptionsTool) Parameters() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "options": map[string]interface{}{
                "type":        "array",
                "description": "玩家可用的操作选项列表",
                "items": map[string]interface{}{
                    "type": "string",
                },
            },
        },
        "required": []string{"options"},
    }
}

// Execute 执行设置选项
func (t *SetOptionsTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
    if t.state == nil {
        return nil, ErrStateNotAvailable
    }

    optionsInterface, ok := args["options"].([]interface{})
    if !ok {
        return nil, ErrInvalidArguments
    }

    options := make([]string, 0, len(optionsInterface))
    for _, opt := range optionsInterface {
        if str, ok := opt.(string); ok {
            options = append(options, str)
        }
    }

    // 将选项存储到状态中
    // 注意：这里需要StateAccessor接口支持SetCurrentOptions方法
    // 需要在StateAccessor接口中添加此方法

    return &ToolResult{
        Success:   true,
        Narrative: "",
        Data: map[string]interface{}{
            "options": options,
        },
    }, nil
}
```

### 步骤3: 扩展 StateAccessor 接口 (internal/tools/types.go)

在StateAccessor接口中添加选项相关方法：

```go
// StateAccessor 状态访问接口（解耦 tools 和 game 包）
type StateAccessor interface {
    // ... 现有方法 ...
    
    // 当前选项
    SetCurrentOptions(options []string)
    GetCurrentOptions() []string
}
```

### 步骤4: 在游戏引擎中注册工具 (internal/game/engine.go)

在`registerTools`方法中添加：

```go
func (e *Engine) registerTools() {
    // ... 现有工具注册 ...
    e.toolRegistry.Register(tools.NewSetOptionsTool(e.state))
}
```

修改`Process`方法，在返回DMResponse前提取选项：

```go
// DMResponse DM响应
type DMResponse struct {
    Content        string           `json:"content"`
    Phase          save.GamePhase   `json:"phase"`
    ToolCalls      []tools.ToolCall `json:"tool_calls,omitempty"`
    ToolNarratives []string         `json:"tool_narratives,omitempty"`
    Options        []string         `json:"options,omitempty"` // 新增：当前可用选项
}

// 在Process方法返回前，从state中获取选项
return &DMResponse{
    Content:        coloredContent,
    Phase:          e.state.GetPhase(),
    ToolCalls:      allToolCalls,
    ToolNarratives: allNarratives,
    Options:        e.state.GetCurrentOptions(), // 添加选项
}, nil
```

### 步骤5: 修改提示词模板 (internal/llm/prompt/templates.go)

更新ToolInstructions，强调必须调用set_options：

```go
ToolInstructions: `工具调用说明：
可用工具：roll_dice, skill_check, saving_throw, deal_damage, heal_character, add_condition, remove_condition, add_item, remove_item, spend_gold, gain_gold, move_to_scene, spawn_npc, remove_npc, set_flag, get_flag, set_options

使用规则：
1. 需要确定成功/失败时，必须使用工具函数进行检定
2. 工具调用的结果将决定游戏世界的变化
3. 根据工具返回的叙述生成描述文本，使用适当的样式标记突出关键结果
4. **重要：每次响应时必须调用 set_options 工具**，提供玩家当前可用的操作选项列表（3-5个选项）
5. set_options的选项应该具体、可操作，反映当前情境下的合理选择

样式标记示例：
- 检定结果：{{success:成功}} 或 {{danger:失败}}
- 伤害数值：造成 {{number:12}} 点伤害
- 状态效果：目标 {{status:中毒}}
- 恢复生命：恢复 {{number:8}} 点生命值`,
```

在DMRole中添加：

```go
重要原则：
- 始终使用中文回复
- 使用工具函数（Tool Call）来执行骰子检定、伤害计算等规则相关操作
- 不要替玩家做决定，而是描述情况并询问玩家的行动
- 保持中立，不偏向任何一方
- **每次响应必须通过set_options工具提供可选操作列表**，让玩家有明确的选择
```

### 步骤6: 修改UI组件 (internal/ui/game.go)

#### 6.1 添加新依赖和常量

```go
import (
    // ... 现有导入 ...
    "github.com/charmbracelet/bubbles/list"
)

const (
    inputModeText  = "text"   // 文本输入模式
    inputModeSelect = "select" // 选择模式
    otherOptionLabel = "其他行动..." // 切换到文本输入的选项
)
```

#### 6.2 修改GameModel结构

```go
type GameModel struct {
    // ... 现有字段 ...
    
    // 输入模式
    inputMode      string       // "text" 或 "select"
    optionsList    list.Model   // 选项列表组件
    currentOptions []string     // 当前可用选项
}
```

#### 6.3 初始化选项列表

```go
func NewGameModel(engine *game.Engine) *GameModel {
    // ... 现有初始化代码 ...
    
    // 初始化选项列表（初始为空）
    opts := []list.Item{}
    optionsList := list.New(opts, list.NewDefaultDelegate(), 0, 0)
    optionsList.SetShowHelp(false)
    optionsList.SetShowStatusBar(false)
    optionsList.SetFilteringEnabled(false)
    
    return &GameModel{
        // ... 现有字段 ...
        inputMode:      inputModeText, // 默认文本模式，直到收到选项
        optionsList:    optionsList,
        currentOptions: []string{},
    }
}
```

#### 6.4 添加选项更新方法

```go
// updateOptions 更新当前选项
func (m *GameModel) updateOptions(options []string) {
    m.currentOptions = options
    
    if len(options) > 0 {
        // 有选项时切换到选择模式
        m.inputMode = inputModeSelect
        
        // 构建列表项（添加"其他行动..."选项）
        items := make([]list.Item, 0, len(options)+1)
        for _, opt := range options {
            items = append(items, optionItem{title: opt})
        }
        items = append(items, optionItem{title: otherOptionLabel, isOther: true})
        
        m.optionsList.SetItems(items)
    } else {
        // 无选项时切换到文本模式
        m.inputMode = inputModeText
    }
}

// optionItem 列表项
type optionItem struct {
    title   string
    isOther bool
}

func (i optionItem) Title() string       { return i.title }
func (i optionItem) Description() string { return "" }
func (i optionItem) FilterValue() string { return i.title }
```

#### 6.5 修改Update方法处理选择模式

```go
func (m *GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // ... 现有代码 ...
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // 处理ESC键切换模式
        if msg.Type == tea.KeyEsc {
            if m.inputMode == inputModeText && len(m.currentOptions) > 0 {
                // 从文本模式切回选择模式
                m.inputMode = inputModeSelect
                m.inputBox.SetValue("")
                return m, nil
            }
            // 其他情况保持原有的退出行为或根据需要处理
        }
        
        // 选择模式下的处理
        if m.inputMode == inputModeSelect && !m.loading {
            switch msg.Type {
            case tea.KeyEnter:
                // 获取选中项
                selected := m.optionsList.SelectedItem()
                if selected != nil {
                    item := selected.(optionItem)
                    if item.isOther {
                        // 切换到文本输入模式
                        m.inputMode = inputModeText
                        return m, nil
                    }
                    // 发送选中选项作为输入
                    return m.handleInput(item.title)
                }
            case tea.KeyUp, tea.KeyDown:
                // 让列表处理上下键
                var cmd tea.Cmd
                m.optionsList, cmd = m.optionsList.Update(msg)
                return m, cmd
            }
        }
        
        // 文本输入模式下的Enter处理
        if m.inputMode == inputModeText && msg.Type == tea.KeyEnter {
            if !m.loading && m.inputBox.Value() != "" {
                return m.handleInput(m.inputBox.Value())
            }
        }
        
        // ... 其他按键处理 ...
    
    case DMResponseMsg:
        m.loading = false
        if msg.Err != nil {
            m.lines = append(m.lines, fmt.Sprintf("错误: %v", msg.Err))
        } else {
            // ... 现有处理 ...
            m.phase = msg.Phase
            
            // 更新选项（关键：从响应中获取选项）
            m.updateOptions(msg.Options)
        }
        // ... 现有代码 ...
    }
    
    // ... 现有代码 ...
}

// handleInput 统一处理输入
func (m *GameModel) handleInput(input string) (tea.Model, tea.Cmd) {
    m.lines = append(m.lines, fmt.Sprintf(strings.Repeat("-", m.windowWidth)+"\n> %s", input))
    m.updateViewportContent()
    m.viewport.PageDown()
    
    m.inputBox.SetValue("")
    m.loading = true
    m.loadingFrame = 0
    m.loadingTimer = 0
    
    return m, tea.Batch(m.processInput(input), m.startLoadingAnimation())
}
```

#### 6.6 修改View方法渲染选择框

```go
// renderInputBox 渲染输入栏
func (m *GameModel) renderInputBox() string {
    if m.loading {
        // ... 现有加载动画代码 ...
    }
    
    if m.inputMode == inputModeSelect {
        // 渲染选项列表
        return GameStyles.InputBox.Render(m.optionsList.View())
    }
    
    // 文本输入模式
    return GameStyles.InputBox.Render(m.inputBox.View())
}
```

#### 6.7 调整尺寸计算

```go
func (m *GameModel) recalculateViewport() {
    // 计算 UI 组件高度
    inputBoxHeight := 3
    separatorHeight := 1
    
    viewportHeight := m.windowHeight - m.statusBarHeight - inputBoxHeight - separatorHeight - 2
    viewportWidth := m.windowWidth - 4
    
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
    m.optionsList.SetHeight(3) // 固定高度显示选项
    
    m.updateViewportContent()
}
```

#### 6.8 更新DMResponseMsg结构

```go
// DMResponseMsg DM响应消息
type DMResponseMsg struct {
    Content        string
    Phase          save.GamePhase
    ToolNarratives []string
    Options        []string  // 新增：选项列表
    Err            error
}
```

#### 6.9 更新processInput中的响应处理

```go
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
            Options:        resp.Options,  // 传递选项
        }
    }
}
```

## 关键设计决策

1. **选项存储位置**：选项存储在游戏状态中，通过StateAccessor接口暴露给工具访问
2. **模式切换逻辑**：
   - 有选项时默认显示选择框
   - 选择"其他行动..."切换到文本输入
   - 按ESC从文本输入返回选择框
3. **向后兼容**：如果LLM没有调用set_options，Options为空，UI自动回退到文本输入模式
4. **强制调用**：通过提示词模板强调set_options的强制性，但不强制技术层面的约束

## 验证步骤

1. 编译检查：`go build ./...`
2. 运行游戏，观察：
   - 新游戏开始后是否显示选项选择框
   - 选择选项后是否正常发送给DM
   - 选择"其他行动..."后是否切换到文本输入
   - 按ESC是否能返回选择框
   - 加载旧存档时是否正常回退到文本输入模式

## 文件修改清单

1. `internal/game/state.go` - 添加CurrentOptions字段和访问方法
2. `internal/tools/types.go` - 扩展StateAccessor接口
3. `internal/tools/options_tool.go` - 新建set_options工具
4. `internal/game/engine.go` - 注册工具，修改DMResponse
5. `internal/llm/prompt/templates.go` - 更新提示词
6. `internal/ui/game.go` - 实现选择/输入混合UI
