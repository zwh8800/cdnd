# Tool Call 处理机制实现计划

## Context

### 问题背景
当前游戏引擎虽然实现了完整的工具系统（15个D&D游戏工具），但LLM无法实际调用这些工具。用户尝试让DM执行操作（如掷骰子、技能检定）时，系统只是将伪XML格式的工具调用请求当作普通文本输出到屏幕，而不是执行实际的工具函数。

### 根本原因
整个 tool_call 调用链路是**断开的**：
1. LLM Request 结构缺少 `Tools` 字段 → 工具定义从未发送给LLM
2. LLM Response 结构缺少 `ToolCalls` 字段 → 无法捕获LLM返回的工具调用
3. Provider 实现没有处理工具调用 → OpenAI/Anthropic/Ollama 都只处理纯文本
4. ProcessWithTools 方法没有实现工具执行循环 → 方法存在但功能缺失

### 预期结果
实现完整的 tool_call 处理机制：
- LLM 能够识别可用的工具定义
- LLM 返回工具调用请求时能够被正确解析
- 工具函数能够被执行并修改游戏状态
- 工具执行结果能够反馈给LLM生成最终叙述
- UI 能够显示工具调用过程和结果

---

## Implementation Plan

### Phase 1: 数据结构层修改

**文件**: `/Users/wastecat/code/go/cdnd/internal/llm/provider.go`

#### 1.1 扩展 Request 结构
```go
type Request struct {
    Messages    []Message         `json:"messages"`
    Model       string            `json:"model,omitempty"`
    MaxTokens   int               `json:"max_tokens,omitempty"`
    Temperature float64           `json:"temperature,omitempty"`
    Stream      bool              `json:"stream,omitempty"`
    Tools       []ToolDefinition  `json:"tools,omitempty"`        // 新增
    ToolChoice  interface{}       `json:"tool_choice,omitempty"`  // 新增: "auto"|"none"|"required"|具体工具
}
```

#### 1.2 扩展 Response 结构
```go
type Response struct {
    ID           string      `json:"id"`
    Content      string      `json:"content"`
    Model        string      `json:"model"`
    Usage        Usage       `json:"usage"`
    ToolCalls    []ToolCall  `json:"tool_calls,omitempty"`    // 新增
    FinishReason string      `json:"finish_reason,omitempty"` // 新增: "stop"|"tool_calls"
}
```

#### 1.3 扩展 Message 结构
```go
type Message struct {
    Role       MessageRole `json:"role"`
    Content    string      `json:"content"`
    ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`    // 新增: assistant 消息包含工具调用
    ToolCallID string      `json:"tool_call_id,omitempty"`  // 新增: tool 角色消息关联调用ID
    Name       string      `json:"name,omitempty"`          // 新增: tool 角色消息的工具名称
}
```

#### 1.4 扩展 StreamChunk 结构
```go
type StreamChunk struct {
    Content      string      `json:"content"`
    Done         bool        `json:"done"`
    Error        error       `json:"error,omitempty"`
    ToolCalls    []ToolCall  `json:"tool_calls,omitempty"`   // 新增
    FinishReason string      `json:"finish_reason,omitempty"`// 新增
}
```

#### 1.5 新增工具相关类型定义
```go
// ToolDefinition 工具定义（与 tools.ToolDefinition 结构一致）
type ToolDefinition struct {
    Type     string                 `json:"type"`
    Function ToolFunctionDefinition `json:"function"`
}

type ToolFunctionDefinition struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCall 工具调用请求
type ToolCall struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Arguments string `json:"arguments"` // JSON string
}
```

---

### Phase 2: Provider 层修改

#### 2.1 OpenAI Provider 修改

**文件**: `/Users/wastecat/code/go/cdnd/internal/llm/openai.go`

**修改 Generate 方法** (第42-90行):
1. 添加工具定义转换逻辑
2. 在 ChatCompletionRequest 中添加 Tools 字段
3. 解析响应中的 ToolCalls

```go
// 转换工具定义
var openaiTools []openai.Tool
if len(req.Tools) > 0 {
    openaiTools = make([]openai.Tool, len(req.Tools))
    for i, tool := range req.Tools {
        openaiTools[i] = openai.Tool{
            Type: openai.ToolTypeFunction,
            Function: &openai.FunctionDefinition{
                Name:        tool.Function.Name,
                Description: tool.Function.Description,
                Parameters:  tool.Function.Parameters,
            },
        }
    }
}

// 构建请求时添加工具
resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
    Model:       model,
    Messages:    messages,
    MaxTokens:   maxTokens,
    Temperature: float32(temp),
    Tools:       openaiTools,           // 新增
    ToolChoice:  req.ToolChoice,        // 新增
})

// 解析响应中的 ToolCalls
var toolCalls []ToolCall
if len(resp.Choices[0].Message.ToolCalls) > 0 {
    toolCalls = make([]ToolCall, len(resp.Choices[0].Message.ToolCalls))
    for i, tc := range resp.Choices[0].Message.ToolCalls {
        toolCalls[i] = ToolCall{
            ID:        tc.ID,
            Name:      tc.Function.Name,
            Arguments: tc.Function.Arguments,
        }
    }
}

return &Response{
    ID:           resp.ID,
    Content:      resp.Choices[0].Message.Content,
    Model:        resp.Model,
    ToolCalls:    toolCalls,
    FinishReason: string(resp.Choices[0].FinishReason),
    Usage:        Usage{...},
}, nil
```

**修改 GenerateStream 方法** (第92-154行):
- 添加工具定义到请求
- 累积流式传输的工具调用（OpenAI 流式工具调用是增量传输的）
- 当 FinishReason 为 "tool_calls" 时发送完整的工具调用

#### 2.2 Anthropic Provider 修改

**文件**: `/Users/wastecat/code/go/cdnd/internal/llm/anthropic.go`

**修改 Generate 方法** (第43-101行):
1. 转换工具定义为 Anthropic 格式
2. 解析响应中的 ToolUseBlock

```go
// 转换工具定义
var anthropicTools []anthropic.ToolParam
if len(req.Tools) > 0 {
    anthropicTools = make([]anthropic.ToolParam, len(req.Tools))
    for i, tool := range req.Tools {
        anthropicTools[i] = anthropic.ToolParam{
            Name:        anthropic.F(tool.Function.Name),
            Description: anthropic.F(tool.Function.Description),
            InputSchema: anthropic.F(tool.Function.Parameters),
        }
    }
}

// 解析响应中的 ToolUse blocks
var toolCalls []ToolCall
for _, block := range message.Content {
    if block.Type == anthropic.ContentBlockTypeToolUse {
        argsJSON, _ := json.Marshal(block.Input)
        toolCalls = append(toolCalls, ToolCall{
            ID:        block.ID,
            Name:      block.Name,
            Arguments: string(argsJSON),
        })
    }
}
```

#### 2.3 Ollama Provider 修改

**文件**: `/Users/wastecat/code/go/cdnd/internal/llm/ollama.go`

Ollama 使用 OpenAI 兼容 API，修改方式与 OpenAI 相同。

---

### Phase 3: Engine 层修改

**文件**: `/Users/wastecat/code/go/cdnd/internal/game/engine.go`

#### 3.1 重写 ProcessWithTools 方法 (第191-214行)

实现完整的 Agentic Loop：

```go
func (e *Engine) ProcessWithTools(ctx context.Context, input string) (*DMResponse, error) {
    e.mu.Lock()
    defer e.mu.Unlock()
    e.state.IncrementTurn()

    // 1. 获取工具定义
    toolDefs := e.toolRegistry.GetToolDefinitions()
    llmToolDefs := convertToLLMToolDefs(toolDefs)

    // 2. 构建初始消息
    messages := e.buildMessagesWithTools(input)

    // 3. Agentic Loop (最多10次迭代)
    const maxIterations = 10
    var allToolCalls []tools.ToolCall

    for i := 0; i < maxIterations; i++ {
        // 3.1 调用 LLM
        resp, err := e.llmProvider.Generate(ctx, &llm.Request{
            Messages: messages,
            Tools:    llmToolDefs,
        })
        if err != nil {
            return nil, fmt.Errorf("LLM调用失败: %w", err)
        }

        // 3.2 检查是否有工具调用
        if len(resp.ToolCalls) == 0 {
            // 没有工具调用，返回最终响应
            e.saveHistory(input, resp.Content, allToolCalls)
            return &DMResponse{
                Content:   resp.Content,
                Phase:     e.state.GetPhase(),
                ToolCalls: allToolCalls,
            }, nil
        }

        // 3.3 执行所有工具调用
        assistantMsg := llm.Message{
            Role:      llm.RoleAssistant,
            Content:   resp.Content,
            ToolCalls: resp.ToolCalls,
        }
        messages = append(messages, assistantMsg)

        for _, tc := range resp.ToolCalls {
            // 解析参数
            var args map[string]interface{}
            json.Unmarshal([]byte(tc.Arguments), &args)

            // 执行工具
            result, err := e.toolRegistry.Execute(ctx, tc.Name, args)

            // 记录工具调用
            allToolCalls = append(allToolCalls, tools.ToolCall{
                ID:        tc.ID,
                Name:      tc.Name,
                Arguments: args,
            })

            // 添加工具结果消息
            toolMsg := llm.Message{
                Role:       llm.RoleTool,
                Content:    formatToolResult(result, err),
                ToolCallID: tc.ID,
                Name:       tc.Name,
            }
            messages = append(messages, toolMsg)
        }
    }

    // 超过最大迭代次数
    return nil, fmt.Errorf("工具调用超过最大迭代次数")
}
```

#### 3.2 添加辅助方法

```go
// convertToLLMToolDefs 转换工具定义格式
func convertToLLMToolDefs(toolDefs []*tools.ToolDefinition) []llm.ToolDefinition

// formatToolResult 格式化工具结果为消息
func formatToolResult(result *tools.ToolResult, err error) string

// saveHistory 保存对话历史
func (e *Engine) saveHistory(input, content string, toolCalls []tools.ToolCall)
```

#### 3.3 流式支持（后续优化）

流式工具调用需要处理增量传输的工具参数，实现复杂度较高。作为后续优化项：

```go
type ToolStreamChunk struct {
    Type       string             // "text"|"tool_call_start"|"tool_result"|"done"|"error"
    Content    string
    ToolCall   *ToolCallInfo
    ToolResult *ToolResult
    Error      error
}

func (e *Engine) ProcessWithToolsStream(ctx context.Context, input string) (<-chan ToolStreamChunk, error)
```

**注意**：当前阶段 UI 调用 ProcessPlayerInputStream 时会退化为纯文本响应，工具调用仅在 ProcessWithTools 中生效。后续可将 ProcessPlayerInputStream 替换为 ProcessWithToolsStream 以获得完整的流式体验。

---

### Phase 4: UI 层集成

**文件**: `/Users/wastecat/code/go/cdnd/internal/ui/game.go`

#### 4.1 添加工具调用状态字段

```go
type GameModel struct {
    // 现有字段...

    // 工具调用相关
    toolCallsInFlight []ToolCallDisplay
    toolResults       []ToolResultDisplay
}
```

#### 4.2 添加工具调用渲染

```go
// renderToolCall 渲染工具调用状态
func renderToolCall(tc ToolCallDisplay) string {
    return fmt.Sprintf("🔧 %s %s...", tc.Name, tc.Params)
}

// renderToolResult 渲染工具执行结果
func renderToolResult(tr ToolResultDisplay) string {
    if tr.Success {
        return fmt.Sprintf("✅ %s: %s", tr.Name, tr.Narrative)
    }
    return fmt.Sprintf("❌ %s: %s", tr.Name, tr.Error)
}
```

#### 4.3 更新 View 方法

在主输出区域显示工具调用和结果。

---

### Phase 5: 测试和验证

#### 5.1 单元测试
- 测试工具定义转换
- 测试工具调用解析
- 测试工具执行流程

#### 5.2 集成测试
- 启动游戏，输入 "我尝试施放法术"
- 验证 LLM 返回 skill_check 工具调用
- 验证工具被执行，角色状态可能改变
- 验证 LLM 根据工具结果生成叙述

#### 5.3 端到端测试命令
```bash
# 构建并运行
go build -o cdnd ./cmd && ./cdnd

# 测试场景
# 1. 开始新游戏
# 2. 输入 "我检查自己的状态"
# 3. 验证系统执行 get_character_info 工具
# 4. 输入 "我尝试撬开铁门（力量运动检定）"
# 5. 验证系统执行 skill_check 工具并显示结果
```

---

## 实现策略

**当前阶段优先实现**：
- 非流式的 `ProcessWithTools` 方法（完整的 Agentic Loop）
- UI 暂时使用非流式调用，后续优化

**后续优化项**：
- 流式工具调用支持（ProcessWithToolsStream）
- UI 实时显示工具执行进度

---

## 关键文件清单

| 文件路径 | 修改内容 | 优先级 |
|---------|---------|--------|
| `internal/llm/provider.go` | 扩展 Request/Response/Message/StreamChunk，新增 ToolCall/ToolDefinition 类型 | P0 |
| `internal/llm/openai.go` | Generate 和 GenerateStream 添加工具支持 | P0 |
| `internal/llm/anthropic.go` | Generate 和 GenerateStream 添加工具支持 | P0 |
| `internal/llm/ollama.go` | Generate 和 GenerateStream 添加工具支持 | P1 |
| `internal/game/engine.go` | 重写 ProcessWithTools，新增 ProcessWithToolsStream | P0 |
| `internal/ui/game.go` | 添加工具调用状态显示 | P1 |

---

## 验证方法

1. **编译检查**: `go build ./...`
2. **类型检查**: `go vet ./...`
3. **单元测试**: `go test ./internal/llm/... ./internal/game/...`
4. **功能测试**: 运行游戏并测试工具调用场景

预期测试场景：
- 用户输入 "我攻击哥布林" → LLM 调用 `attack` 工具 → 显示攻击结果
- 用户输入 "我搜索房间" → LLM 调用 `skill_check` 工具 → 显示检定结果
- 用户输入 "我治疗自己" → LLM 调用 `heal_character` 工具 → 显示治疗结果
