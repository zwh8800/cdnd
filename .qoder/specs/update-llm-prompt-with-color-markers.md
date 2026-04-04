# LLM提示词模板颜色标记更新计划

## 背景与目标

### 问题/需求
当前DM生成的文本是纯文本格式，关键信息（如角色状态、数值、重要名词等）在终端中无法通过颜色突出显示。用户希望在保持D&D叙事风格的同时，让LLM生成的文本能够包含ANSI颜色代码或样式标记，以便在终端中以不同颜色显示关键信息。

### 预期结果
- DM生成的文本自动包含颜色标记
- 关键信息（数值、状态、重要名词）在终端中高亮显示
- 与现有UI样式系统（styles.go）保持一致
- 保持叙事流畅性，避免过度使用颜色

## 关键文件

| 文件路径 | 作用 |
|---------|-----|
| `/internal/llm/prompt/templates.go` | 提示词模板定义（主要修改目标） |
| `/internal/llm/prompt/builder.go` | 提示词构建器（可能需要修改） |
| `/internal/ui/styles.go` | UI样式系统参考（定义颜色方案） |

## 实施方案

### 1. 颜色标记语法设计

采用Markdown风格的自定义标记语法，便于LLM理解和生成：

| 标记类型 | 语法 | 对应样式 | 使用场景 |
|---------|------|---------|---------|
| 数值/数字 | `{{number:数值}}` | SecondaryColor (绿色) | 生命值、伤害、金币、DC值 |
| 重要名词 | `{{keyword:名词}}` | PrimaryColor (紫色) | 武器、法术、地点、NPC名称 |
| 状态效果 | `{{status:状态}}` | WarningColor (黄色) | 中毒、眩晕、隐身等 |
| 战斗信息 | `{{combat:文本}}` | DangerColor (红色) | 攻击、伤害、战斗动作 |
| 成功提示 | `{{success:文本}}` | SecondaryColor (绿色) | 成功检定、正面结果 |
| 失败/危险 | `{{danger:文本}}` | DangerColor (红色) | 失败检定、危险警告 |
| 引用/NPC对话 | `{{quote:文本}}` | AccentColor (浅紫) | NPC说话内容 |

### 2. 模板修改内容

#### 2.1 DMRole模板修改

在`DMRole`模板中添加颜色标记使用说明：

```go
DMRole: `你是一位经验丰富的龙与地下城（D&D 5e）地下城主（DM）。你的职责是：

1. 讲述故事 - 用生动的中文描述场景、NPC和事件
2. 扮演NPC - 为每个NPC赋予独特的性格、说话方式和动机
3. 执行规则 - 公正地应用D&D 5e规则，调用工具函数进行检定
4. 引导冒险 - 提供有趣的选择和挑战，但让玩家自己做决定
5. 保持节奏 - 控制叙事节奏，在关键时刻制造紧张感

重要原则：
- 始终使用中文回复
- 使用工具函数（Tool Call）来执行骰子检定、伤害计算等规则相关操作
- 不要替玩家做决定，而是描述情况并询问玩家的行动
- 保持中立，不偏向任何一方

文本样式标记（用于突出关键信息）：
- 数值类信息使用 {{number:数值}}，如：{{number:15}}点伤害、DC {{number:15}}
- 重要名词使用 {{keyword:名词}}，如：{{keyword:长剑}}、{{keyword:火球术}}、{{keyword:暗影城堡}}
- 状态效果使用 {{status:状态}}，如：{{status:中毒}}、{{status:眩晕}}
- 战斗动作使用 {{combat:动作}}，如：{{combat:挥剑攻击}}
- 成功结果使用 {{success:描述}}，如：{{success:命中！}}
- 危险/失败使用 {{danger:描述}}，如：{{danger:攻击未命中}}
- NPC对话使用 {{quote:对话内容}}

注意：仅在关键信息处使用标记，保持叙事流畅自然。`,
```

#### 2.2 GameRules模板修改

在`GameRules`模板中添加颜色标记示例：

```go
GameRules: `D&D 5e 基础规则参考

难度等级（DC）：
- 非常简单（DC {{number:5}}）：几乎不可能失败
- 简单（DC {{number:10}}）：稍有挑战但通常成功
- 中等（DC {{number:15}}）：需要一定能力才能成功
- 困难（DC {{number:20}}）：需要高水平能力
- 非常困难（DC {{number:25}}）：只有专家才能成功

优势/劣势：投两次d20，取较高/较低值
大成功：自然{{number:20}}自动成功
大失败：自然{{number:1}}自动失败`,
```

#### 2.3 ToolInstructions模板修改

```go
ToolInstructions: `工具调用说明：
可用工具：roll_dice, skill_check, saving_throw, deal_damage, heal_character, add_condition, remove_condition, add_item, remove_item, spend_gold, gain_gold, move_to_scene, spawn_npc, remove_npc, set_flag, get_flag

使用规则：
1. 需要确定成功/失败时，必须使用工具函数进行检定
2. 工具调用的结果将决定游戏世界的变化
3. 根据工具返回的叙述生成描述文本，使用适当的样式标记突出关键结果`,
```

#### 2.4 场景提示词模板修改

修改`IntroPrompt`、`CombatPrompt`、`DialoguePrompt`、`RestPrompt`，添加颜色标记使用示例：

```go
IntroPrompt: `新的冒险即将开始！请为玩家角色创造一个引人入胜的开场场景。

在描述中，请使用以下样式标记突出关键信息：
- 地点名称：{{keyword:地点名}}
- 重要物品：{{keyword:物品名}}
- 关键数值：{{number:数值}}

记住不要一次性透露所有信息，为玩家留下探索和选择的空间。`,

CombatPrompt: `战斗进行中。当前处于战斗回合。

请使用以下样式标记：
- 攻击动作：{{combat:动作描述}}
- 伤害数值：{{number:伤害值}}
- 命中/未命中：{{success:命中！}} / {{danger:未命中}}
- 状态效果：{{status:状态名}}

需要攻击或检定时，必须使用工具函数。攻击需要命中检定，命中后投伤害骰。`,

DialoguePrompt: `玩家正在与 {{keyword:%s}} 交谈。NPC态度: %s。

请用NPC的声音回应玩家，保持NPC的性格一致性。
NPC的直接对话使用 {{quote:对话内容}} 标记。`,

RestPrompt: `玩家选择休息。请描述休息地点和期间发生的事情。

使用样式标记：
- 地点：{{keyword:地点名}}
- 恢复数值：{{number:恢复量}}

长休息恢复全部生命值和法术槽。`,
```

### 3. 颜色标记解析器实现

在`builder.go`或新建文件中添加颜色标记解析函数：

```go
// ParseColorMarkers 将模板中的颜色标记转换为ANSI颜色代码
func ParseColorMarkers(text string) string {
    // 定义标记到ANSI代码的映射
    markers := map[string]string{
        "number":  "\033[32m",   // 绿色
        "keyword": "\033[35m",   // 紫色
        "status":  "\033[33m",   // 黄色
        "combat":  "\033[31m",   // 红色
        "success": "\033[32m",   // 绿色
        "danger":  "\033[31m",   // 红色
        "quote":   "\033[36m",   // 青色
    }
    reset := "\033[0m"

    result := text
    for marker, color := range markers {
        // 替换开始标记
        startPattern := fmt.Sprintf("{{%s:", marker)
        endPattern := "}}"
        
        // 使用正则或字符串替换
        // 简化实现：直接替换
        for strings.Contains(result, startPattern) {
            startIdx := strings.Index(result, startPattern)
            if startIdx == -1 {
                break
            }
            endIdx := strings.Index(result[startIdx:], endPattern)
            if endIdx == -1 {
                break
            }
            endIdx += startIdx
            
            // 提取内容
            contentStart := startIdx + len(startPattern)
            content := result[contentStart:endIdx]
            
            // 替换为带颜色的文本
            colored := color + content + reset
            result = result[:startIdx] + colored + result[endIdx+len(endPattern):]
        }
    }
    
    return result
}
```

### 4. 与UI样式系统的集成

为了保持与`styles.go`中定义的样式一致，创建一个映射表：

```go
// ColorMarkerStyles 颜色标记与UI样式的对应关系
var ColorMarkerStyles = map[string]lipgloss.Color{
    "number":  SecondaryColor,  // #04B575 - 绿色
    "keyword": PrimaryColor,    // #7D56F4 - 紫色
    "status":  WarningColor,    // #FFD93D - 黄色
    "combat":  DangerColor,     // #FF6B6B - 红色
    "success": SecondaryColor,  // #04B575 - 绿色
    "danger":  DangerColor,     // #FF6B6B - 红色
    "quote":   AccentColor,     // #5C5CFF - 浅紫色
}
```

### 5. 渲染流程修改

在DM响应显示到终端之前，需要解析颜色标记：

```go
// 在game/engine.go或相关显示逻辑中
func (e *Engine) displayDMResponse(response string) {
    // 解析颜色标记
    coloredResponse := prompt.ParseColorMarkers(response)
    // 显示到终端
    fmt.Println(coloredResponse)
}
```

## 验证方案

### 1. 单元测试

编写测试用例验证颜色标记解析：

```go
func TestParseColorMarkers(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {
            input:    "造成 {{number:15}} 点伤害",
            expected: "造成 \033[32m15\033[0m 点伤害",
        },
        {
            input:    "使用 {{keyword:火球术}} 攻击",
            expected: "使用 \033[35m火球术\033[0m 攻击",
        },
        {
            input:    "{{success:命中！}} 造成 {{number:8}} 点伤害",
            expected: "\033[32m命中！\033[0m 造成 \033[32m8\033[0m 点伤害",
        },
    }
    
    for _, tt := range tests {
        result := ParseColorMarkers(tt.input)
        if result != tt.expected {
            t.Errorf("ParseColorMarkers(%q) = %q, want %q", tt.input, result, tt.expected)
        }
    }
}
```

### 2. 集成测试

1. 启动游戏
2. 进入任意场景
3. 观察DM响应是否包含颜色标记
4. 验证终端是否正确显示颜色

### 3. 手动验证清单

- [ ] 数值（生命值、伤害、DC）显示为绿色
- [ ] 重要名词（武器、法术、地点）显示为紫色
- [ ] 状态效果显示为黄色
- [ ] 战斗信息显示为红色
- [ ] 成功提示显示为绿色
- [ ] 失败/危险显示为红色
- [ ] NPC对话显示为浅紫色
- [ ] 叙事流畅性未受影响

## 实施步骤

1. **修改`templates.go`** - 更新所有模板，添加颜色标记使用说明
2. **创建颜色标记解析器** - 在`builder.go`或新文件中添加`ParseColorMarkers`函数
3. **添加样式映射** - 在`styles.go`或`builder.go`中添加`ColorMarkerStyles`
4. **集成到显示逻辑** - 在DM响应显示前调用解析器
5. **编写测试** - 添加单元测试验证解析逻辑
6. **手动验证** - 运行游戏验证颜色显示效果

## 风险与注意事项

1. **LLM理解能力** - 需要确保LLM能够正确理解和使用颜色标记语法
2. **过度标记** - 提示词中强调"仅在关键信息处使用标记"
3. **兼容性** - 确保ANSI代码在不支持颜色的终端中不会破坏显示
4. **性能** - 解析操作应在可接受的时间范围内完成
