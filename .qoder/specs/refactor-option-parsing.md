# 重构计划：选项解析机制

## 背景

当前项目使用 `set_options` 工具让 LLM 通过工具调用设置玩家可选操作。这种方式存在设计问题，需要重构为通过解析 LLM 输出的结构化文本来提取选项。

## 目标

将选项管理机制从工具调用改为文本解析：
- 减少 LLM 工具调用次数，降低延迟
- 更自然地融入叙事流程
- 简化 LLM 输出逻辑

## 文件修改清单

### 删除文件
| 文件 | 说明 |
|------|------|
| `internal/tools/options_tool.go` | 删除整个文件 |

### 修改文件
| 文件 | 修改内容 |
|------|----------|
| `internal/game/engine.go` | 移除工具注册，集成选项解析器 |
| `internal/llm/prompt/templates.go` | 修改提示词，添加选项输出格式说明 |

**注意**: `internal/tools/types.go` 中的 StateAccessor 接口**保留不变**，SetCurrentOptions/GetCurrentOptions 方法仍需保留，因为解析后的选项仍通过这些方法设置到状态中。

### 新增文件
| 文件 | 说明 |
|------|------|
| `internal/llm/prompt/options_parser.go` | 选项解析器实现 |
| `internal/llm/prompt/options_parser_test.go` | 解析器单元测试 |

---

## 详细实施步骤

### 步骤 1: 创建选项解析器

**文件**: `internal/llm/prompt/options_parser.go`

目标格式：
```
==========
你的选择是：
  1. 选项A
  2. 选项B
  3. 选项C
```

核心函数签名：
```go
// ParseOptions 从文本中解析选项列表
// 返回: options - 解析出的选项列表, content - 移除选项块后的纯净内容
func ParseOptions(text string) (options []string, content string)
```

解析逻辑：
1. 使用正则匹配选项块 `==========\n你的选择是：`
2. 提取编号选项行 `^\s*\d+\.\s*(.+)$`
3. 返回选项列表和清理后的内容

边缘情况处理：
- 未找到选项块 → 返回空选项，原样返回内容
- 选项数量 > 10 → 只保留前 10 个
- 格式变体 → 支持等号数量 >= 5 的宽松匹配

### 步骤 2: 修改提示词模板

**文件**: `internal/llm/prompt/templates.go`

#### DMRole 模板 (第30行附近)
删除：
```
- **每次响应必须通过set_options工具提供可选操作列表**
```

新增（在"重要原则"列表末尾）：
```
- **每次响应末尾必须提供可选操作列表**，格式如下：
  ==========
  你的选择是：
    1. 第一个选项
    2. 第二个选项
    3. 第三个选项
  （3-5个选项，选项应具体、可操作）
```

#### ToolInstructions 模板 (第57行附近)
删除工具列表中的 `set_options`

删除第63-65行关于 set_options 的说明（无需新增，格式说明已在 DMRole 中）

### 步骤 3: 修改引擎集成

**文件**: `internal/game/engine.go`

#### 3.1 移除工具注册 (第76行)
```go
// 删除此行
e.toolRegistry.Register(tools.NewSetOptionsTool(e.state))
```

#### 3.2 集成选项解析器 (第250-261行区域)
修改 Process() 方法中的响应处理逻辑：

```go
// 解析选项并提取纯净内容
options, cleanContent := prompt.ParseOptions(resp.Content)

// 应用颜色标记到纯净内容
coloredContent := prompt.ParseColorMarkers(cleanContent)

// 更新状态中的选项
e.state.SetCurrentOptions(options)

return &DMResponse{
    Content:        coloredContent,
    Options:        options,
    // ...
}
```

#### 3.3 保留清除选项逻辑 (第200行)
```go
e.state.ClearCurrentOptions()  // 每次处理前清空
```

### 步骤 4: 删除选项工具文件

```bash
rm internal/tools/options_tool.go
```

---

## 测试验证

### 单元测试用例
- 标准格式解析 → 正确提取
- 无选项块 → 返回空选项
- 格式变体 → 宽松匹配
- 特殊字符选项 → 正确保留
- 选项数量过多 → 截断处理

### 集成验证
1. 启动游戏，验证开场选项正确显示
2. 测试探索、战斗、对话场景的选项
3. 验证选项解析失败时回退到文本输入模式
4. 确认历史记录中不包含选项块文本

---

## 风险缓解

| 风险 | 缓解措施 |
|------|----------|
| LLM 不遵循格式 | 解析器返回空选项，UI 回退文本模式 |
| 格式变体 | 正则宽松匹配 |
| 存档兼容性 | 选项格式不变，无需迁移 |
