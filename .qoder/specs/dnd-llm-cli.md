# LLM D&D CLI游戏 - 技术选型与实现计划

## 背景

开发一个基于大语言模型的命令行界面（CLI）龙与地下城（D&D）角色扮演游戏。

**用户需求确认：**
- **语言**：Go 1.24.2
- **LLM提供商**：多提供商同时支持（OpenAI、Anthropic Claude、Ollama）
- **UI风格**：交互式TUI（Bubbletea框架），类似Claude Code风格
- **规则系统**：标准D&D 5e规则

**项目现状：**
- Go 1.24.2 环境
- 已添加 Cobra + Viper 依赖
- main.go 为空

---

## 一、技术选型分析

### 1. 编程语言选择

| 语言 | CLI开发 | LLM生态 | 单二进制分发 | 并发性能 | 学习曲线 | 推荐度 |
|------|---------|---------|--------------|----------|----------|--------|
| **Go** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | **推荐** |
| Python | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | 备选 |
| Rust | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ | 不推荐 |
| Node.js | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | 备选 |

**推荐：Go**
- 已有环境和依赖基础
- 单一可执行文件分发，用户体验最佳
- Cobra/Viper是CLI开发黄金组合
- 优秀的并发支持处理流式LLM响应

---

### 2. 大模型API集成方案

**推荐架构：多提供商抽象层**

```
┌─────────────────────────────────────────────────────────┐
│                     Game Engine                          │
├─────────────────────────────────────────────────────────┤
│                  LLM Client Interface                    │
│  ┌─────────────┬─────────────┬─────────────┬──────────┐ │
│  │   OpenAI    │  Anthropic  │   Ollama    │  Custom  │ │
│  │  Provider   │  Provider   │  Provider   │ Provider │ │
│  └─────────────┴─────────────┴─────────────┴──────────┘ │
└─────────────────────────────────────────────────────────┘
```

**依赖选择：**
| 库 | 用途 | 特点 |
|----|------|------|
| `github.com/sashabaranov/go-openai` | OpenAI/Ollama | 社区最活跃，功能完整 |
| `github.com/anthropics/anthropic-sdk-go` | Anthropic Claude | 官方SDK |
| 自定义接口 | 统一抽象 | 避免厂商锁定 |

---

### 3. CLI框架

**推荐：Cobra（已选）**

命令结构设计：
```
cdnd                    # 主命令
├── start              # 开始新游戏
│   ├── --save-slot    # 存档槽位
│   └── --scenario     # 剧本选择
├── load               # 加载存档
│   └── --slot         # 存档槽位
├── character          # 角色管理
│   ├── create         # 创建角色
│   ├── list           # 角色列表
│   └── delete         # 删除角色
├── config             # 配置管理
│   ├── set            # 设置配置
│   ├── get            # 获取配置
│   └── init           # 初始化配置
├── provider           # LLM提供商管理
│   ├── list           # 列出可用提供商
│   ├── test           # 测试连接
│   └── set-default    # 设置默认
└── version            # 版本信息
```

---

### 4. 数据持久化方案

| 方案 | 优点 | 缺点 | 适用阶段 |
|------|------|------|----------|
| JSON文件 | 简单直观、易调试 | 无索引、大文件性能差 | MVP阶段 |
| SQLite (modernc) | 纯Go、SQL支持、跨平台 | 相对复杂 | 完善阶段 |

**存储结构：**
```
~/.cdnd/
├── config.yaml              # 全局配置
├── saves/
│   ├── slot_001/
│   │   ├── metadata.json    # 存档元数据
│   │   ├── character.json   # 角色数据
│   │   ├── world.json       # 世界状态
│   │   └── history.db       # 对话历史
│   └── slot_002/
├── characters/              # 预设角色模板
├── scenarios/               # 剧本模块
└── cache/                   # 缓存目录
```

---

### 5. 配置管理

**推荐：YAML + Viper**

```yaml
# ~/.cdnd/config.yaml
llm:
  default_provider: openai
  providers:
    openai:
      api_key: ${OPENAI_API_KEY}
      model: gpt-4-turbo-preview
      max_tokens: 4096
      temperature: 0.7
    anthropic:
      api_key: ${ANTHROPIC_API_KEY}
      model: claude-3-opus-20240229
    ollama:
      base_url: http://localhost:11434
      model: llama2

game:
  autosave: true
  autosave_interval: 5m
  max_history_turns: 100
  language: zh-CN

display:
  typewriter_effect: true
  typing_speed: 50ms
  color_output: true
```

---

### 6. 项目结构

```
cdnd/
├── cmd/                        # CLI命令入口
│   ├── root.go                 # 根命令，初始化
│   ├── start.go                # start命令
│   ├── load.go                 # load命令
│   ├── character.go            # character命令组
│   ├── config.go               # config命令组
│   ├── provider.go             # provider命令组
│   └── version.go              # version命令
│
├── internal/                   # 私有应用代码
│   ├── game/                   # 游戏核心逻辑
│   │   ├── engine.go           # 游戏引擎
│   │   ├── session.go          # 游戏会话
│   │   ├── turn.go             # 回合管理
│   │   └── combat.go           # 战斗系统
│   │
│   ├── character/              # 角色系统
│   │   ├── character.go        # 角色定义
│   │   ├── attributes.go       # 属性系统
│   │   ├── skills.go           # 技能系统
│   │   ├── inventory.go        # 物品栏
│   │   ├── class.go            # 职业系统
│   │   ├── race.go             # 种族系统
│   │   └── feature.go          # 特性
│   │
│   ├── world/                  # 世界系统
│   │   ├── location.go         # 地点
│   │   ├── npc.go              # NPC管理
│   │   ├── quest.go            # 任务系统
│   │   ├── monster.go          # 怪物定义
│   │   └── events.go           # 事件系统
│   │
│   ├── llm/                    # LLM集成层
│   │   ├── provider.go         # 提供商接口
│   │   ├── registry.go         # 提供商注册中心
│   │   ├── openai.go           # OpenAI实现
│   │   ├── anthropic.go        # Anthropic实现
│   │   ├── ollama.go           # Ollama实现
│   │   ├── stream.go           # 流式响应处理
│   │   └── prompt/             # Prompt模板
│   │       ├── templates.go
│   │       ├── system.go
│   │       ├── combat.go
│   │       └── character.go
│   │
│   ├── rules/                  # D&D 5e规则引擎
│   │   ├── engine.go           # 规则引擎核心
│   │   ├── conditions.go       # 条件判断
│   │   ├── effects.go          # 效果应用
│   │   ├── combat_rules.go     # 战斗规则
│   │   └── skill_rules.go      # 技能规则
│   │
│   ├── save/                   # 存档系统
│   │   ├── manager.go          # 存档管理器
│   │   ├── slot.go             # 槽位管理
│   │   ├── serializer.go       # 序列化
│   │   └── migration.go        # 版本迁移
│   │
│   ├── ui/                     # TUI用户界面
│   │   ├── app.go              # Bubbletea主应用
│   │   ├── header.go           # 顶部状态栏
│   │   ├── narrative.go        # 叙事显示区
│   │   ├── dice.go             # 骰子动画组件
│   │   ├── input.go            # 输入组件
│   │   ├── menu.go             # 行动菜单
│   │   ├── status.go           # 角色状态面板
│   │   ├── styles.go           # 样式定义
│   │   └── markdown.go         # Markdown渲染
│   │
│   └── config/                 # 配置管理
│       ├── config.go           # 配置结构
│       ├── loader.go           # 加载器
│       └── defaults.go         # 默认配置
│
├── pkg/                        # 可导出的库
│   ├── dice/                   # 骰子系统
│   │   ├── dice.go             # 骰子核心
│   │   ├── roll.go             # 掷骰逻辑
│   │   └── parser.go           # 表达式解析
│   └── dnd5e/                  # D&D 5e工具
│       ├── modifiers.go        # 修正值计算
│       └── cr.go               # 挑战等级工具
│
├── data/                       # 游戏数据
│   ├── monsters/               # 怪物数据
│   │   └── srd_monsters.json   # SRD怪物图鉴
│   ├── items/                  # 物品数据
│   │   └── srd_items.json
│   └── spells/                 # 法术数据
│       └── srd_spells.json
│
├── scenarios/                  # 剧本模块
│   └── lost_mine_of_phandelver/
│       ├── scenario.yaml
│       └── ...
│
├── test/                       # 集成测试
│   ├── testdata/
│   └── integration_test.go
│
├── .golangci.yml              # Linter配置
├── Makefile                    # 构建脚本
├── go.mod
├── go.sum
└── README.md
```

---

### 7. 关键依赖

```go
// CLI框架（已有）
github.com/spf13/cobra v1.10.2
github.com/spf13/viper v1.21.0

// LLM客户端
github.com/sashabaranov/go-openai v1.17.9           // OpenAI/Ollama
github.com/anthropics/anthropic-sdk-go v0.2.0       // Anthropic Claude

// TUI框架（核心依赖）
github.com/charmbracelet/bubbletea v0.26.4          // TUI框架核心
github.com/charmbracelet/lipgloss v0.10.0           // 终端样式
github.com/charmbracelet/bubbles v0.18.0            // TUI组件（输入框、列表等）
github.com/charmbracelet/glamour v0.6.0             // Markdown渲染
github.com/charmbracelet/x/ansi                     // ANSI处理

// 工具库
github.com/google/uuid v1.6.0                       // UUID生成
github.com/sirupsen/logrus v1.9.3                   // 结构化日志
golang.org/x/term                                   // 终端处理

// 测试
github.com/stretchr/testify v1.9.0                  // 测试工具
go.uber.org/mock v0.4.0                             // Mock生成
```

---

### 8. TUI界面设计（Bubbletea）

**界面布局：**
```
┌─────────────────────────────────────────────────────────────────┐
│  🐉 D&D Adventure                           [HP: 45/45] [Lv.3]  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  DM: 黑暗的洞穴中，你听到了低沉的咆哮声...                        │
│  一只巨大的地精从阴影中现身，手持生锈的弯刀。                      │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ 🎲 骰子检定: 感知检定 (DC 12)                            │    │
│  │    掷骰: 1d20+3 = [15] + 3 = 18 ✓ 成功                 │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  你敏锐地察觉到地精身后还有两只同伙潜伏着。                       │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│  > [输入你的行动...]                                            │
│                                                                 │
│  [攻击] [技能] [物品] [交谈] [逃跑]                              │
└─────────────────────────────────────────────────────────────────┘
```

**核心组件设计：**

```go
// internal/ui/tui/app.go
type AppModel struct {
    // 窗口尺寸
    width, height int
    
    // 子模型
    header      HeaderModel      // 顶部状态栏
    narrative   NarrativeModel   // 叙事显示区
    dice        DiceRollModel    // 骰子动画
    input       InputModel       // 用户输入
    menu        MenuModel        // 行动菜单
    status      StatusModel      // 角色状态面板（可展开）
    
    // 状态
    session     *game.Session
    focused     FocusArea
}

// 消息类型
type (
    NarrationMsg    string      // 叙事文本
    DiceRollMsg     DiceResult  // 骰子结果
    PlayerActionMsg string      // 玩家行动
    LLMStreamMsg    string      // LLM流式响应
    ErrorMsg        error       // 错误
)
```

**样式系统：**
```go
// internal/ui/styles/styles.go
var (
    // 主题色
    PrimaryColor   = lipgloss.Color("#7D56F4")  // 紫色
    SecondaryColor = lipgloss.Color("#04B575")  // 绿色
    DangerColor    = lipgloss.Color("#FF6B6B")  // 红色
    WarningColor   = lipgloss.Color("#FFD93D")  // 黄色
    
    // 文本样式
    TitleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(PrimaryColor).
        Padding(0, 1)
    
    NarrativeStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FAFAFA")).
        Padding(1, 2)
    
    // 骰子样式
    DiceSuccessStyle = lipgloss.NewStyle().
        Foreground(SecondaryColor).
        Bold(true)
    
    DiceFailStyle = lipgloss.NewStyle().
        Foreground(DangerColor).
        Bold(true)
)
```

**交互功能：**
1. **流式输出**：LLM响应逐字显示，带打字机效果
2. **骰子动画**：骰子滚动动画，显示检定结果
3. **菜单导航**：方向键选择行动，Enter确认
4. **状态面板**：Tab切换显示角色详细属性
5. **历史滚动**：Shift+上/下滚动历史对话

---

### 9. D&D 5e规则系统设计

**核心数据结构：**

```go
// internal/character/character.go
type Character struct {
    ID          string
    Name        string
    Race        Race
    Class       Class
    Level       int
    Background  string
    Alignment   string
    
    // 属性（六维）
    Attributes  Attributes
    
    // 生命值
    HP          HitPoints
    
    // 技能
    Skills      map[SkillType]int
    
    // 装备与物品
    Equipment   []Item
    Inventory   []Item
    Gold        int
    
    // 特性
    Features    []Feature
    Spells      []Spell
    Proficiencies []Proficiency
}

type Attributes struct {
    Strength     int `json:"strength"`
    Dexterity    int `json:"dexterity"`
    Constitution int `json:"constitution"`
    Intelligence int `json:"intelligence"`
    Wisdom       int `json:"wisdom"`
    Charisma     int `json:"charisma"`
}

func (a *Attributes) Modifier(attr string) int {
    // 修正值 = floor((属性值 - 10) / 2)
    val := a.Get(attr)
    return (val - 10) / 2
}
```

**骰子系统：**
```go
// pkg/dice/dice.go
type Dice struct {
    Count int    // 骰子数量
    Sides int    // 骰子面数
    Modifier int // 修正值
}

type RollResult struct {
    Dice     []int  // 每个骰子的结果
    Total    int    // 总计
    Modifier int    // 修正值
    Final    int    // 最终结果
    Critical CriticalType
}

type CriticalType int

const (
    CritNone CriticalType = iota
    CritSuccess  // 自然20
    CritFail     // 自然1
)

func (d *Dice) Roll() RollResult {
    // 支持优势/劣势掷骰
}

// 骰子表达式解析 "2d6+3", "1d20", "4d8"
func Parse(expr string) (*Dice, error)
```

**战斗系统：**
```go
// internal/game/combat.go
type Combat struct {
    Round       int
    Initiative  []Combatant  // 先攻顺序
    CurrentTurn int
    Combatants  map[string]*Combatant
}

type Combatant struct {
    Entity      interface{}  // PlayerCharacter 或 Monster
    HP          int
    AC          int          // 护甲等级
    Initiative  int
    Conditions  []Condition  // 状态效果
    Actions     []Action     // 可用行动
}

func (c *Combat) Attack(attacker, target *Combatant, attack *Attack) AttackResult {
    // 1. 掷攻击骰 (1d20 + 命中修正)
    // 2. 比较目标AC
    // 3. 若命中，掷伤害骰
    // 4. 应用伤害类型和抗性
    // 5. 检查暴击
}

func (c *Combat) SavingThrow(combatant *Combatant, ability string, dc int) bool {
    // 豁免检定
}
```

**规则引擎：**
```go
// internal/rules/engine.go
type RuleEngine struct {
    rules       map[string][]Rule
    conditions  map[string]ConditionFunc
    effects     map[string]EffectFunc
}

type Rule struct {
    ID          string
    Name        string
    Trigger     string           // 触发事件
    Conditions  []Condition      // 条件列表
    Effects     []Effect         // 效果列表
    Priority    int              // 优先级
}

// 内置规则示例
var BuiltInRules = []Rule{
    {
        ID:      "sneak_attack",
        Name:    "偷袭",
        Trigger: "attack_roll",
        Conditions: []Condition{
            {Type: "class", Value: "rogue"},
            {Type: "advantage", Value: true},
        },
        Effects: []Effect{
            {Type: "extra_damage_dice", Value: "level_based"},
        },
    },
    {
        ID:      "critical_hit",
        Name:    "暴击",
        Trigger: "attack_roll",
        Conditions: []Condition{
            {Type: "natural_roll", Value: 20},
        },
        Effects: []Effect{
            {Type: "double_dice", Value: true},
        },
    },
}
```

**怪物数据：**
```go
// internal/world/monster.go
type Monster struct {
    ID          string
    Name        string
    CR          float64    // 挑战等级
    Size        Size
    Type        MonsterType
    
    // 属性
    Attributes  Attributes
    HP          HitPoints
    AC          int
    Speed       int
    
    // 能力
    Actions     []MonsterAction
    Reactions   []MonsterAction
    Legendary   []MonsterAction
    
    // 特性
    Traits      []MonsterTrait
    Immunities  []DamageType
    Resistances []DamageType
    Senses      []Sense
    Languages   []string
}

// 怪物图鉴
type MonsterManual struct {
    monsters map[string]*Monster
}
```

---

### 10. 测试策略

```
测试金字塔
    ┌─────────┐
    │   E2E   │  ← 少量，关键路径
    ├─────────┤
    │ 集成测试 │  ← 中等数量，模块交互
    ├─────────┤
    │ 单元测试 │  ← 大量，核心逻辑
    └─────────┘
```

- 使用 `testify` 进行断言和Mock
- LLM Provider使用接口Mock进行测试
- 覆盖率目标：核心模块 > 80%

---

### 9. 性能与成本优化

**Token优化策略：**
1. 上下文压缩（摘要旧对话）
2. 历史对话滑动窗口
3. 相似请求响应缓存
4. 流式输出提前终止

**成本优化：**
| 策略 | 预期节省 |
|------|----------|
| 使用更小模型处理简单任务 | 50-70% |
| 历史对话压缩 | 30-50% |
| 响应缓存 | 20-40% |
| 本地模型回退 | 可变 |

---

### 10. 可扩展性设计

**插件系统：**
```go
type Plugin interface {
    OnGameStart(session *Session)
    OnTurnStart(turn *Turn)
    OnAction(action Action)
    OnCombatStart(combat *Combat)
    OnGameEnd(session *Session)
}
```

**模组支持：**
```
~/.cdnd/mods/
├── mod_a/
│   ├── mod.yaml
│   ├── scenarios/
│   ├── items/
│   └── rules/
```

---

## 二、实施计划

### Phase 1: 项目骨架初始化（本次实施）

**目标：搭建完整项目架构和核心框架**

#### 1.1 目录结构创建
- [ ] 创建 `cmd/` 目录和所有命令文件
- [ ] 创建 `internal/` 子目录结构
- [ ] 创建 `pkg/dice/` 和 `pkg/dnd5e/`
- [ ] 创建 `data/` 和 `scenarios/` 目录
- [ ] 添加 `.golangci.yml` 和 `Makefile`

#### 1.2 依赖安装
```bash
go get github.com/sashabaranov/go-openai@latest
go get github.com/anthropics/anthropic-sdk-go@latest
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/glamour@latest
go get github.com/google/uuid@latest
go get github.com/sirupsen/logrus@latest
go get github.com/stretchr/testify@latest
```

#### 1.3 CLI命令框架
- [ ] `cmd/root.go` - 根命令，初始化Viper配置
- [ ] `cmd/start.go` - start子命令
- [ ] `cmd/load.go` - load子命令
- [ ] `cmd/character.go` - character命令组
- [ ] `cmd/config.go` - config命令组
- [ ] `cmd/provider.go` - provider命令组
- [ ] `cmd/version.go` - version命令
- [ ] 更新 `main.go` 调用 `cmd.Execute()`

#### 1.4 配置系统
- [ ] `internal/config/config.go` - 配置结构体定义
- [ ] `internal/config/loader.go` - 配置加载逻辑
- [ ] `internal/config/defaults.go` - 默认配置
- [ ] 配置文件路径：`~/.cdnd/config.yaml`

#### 1.5 LLM提供商抽象层
- [ ] `internal/llm/provider.go` - 统一接口定义
  ```go
  type Provider interface {
      Name() string
      Generate(ctx context.Context, req *Request) (*Response, error)
      GenerateStream(ctx context.Context, req *Request) (<-chan StreamChunk, error)
  }
  ```
- [ ] `internal/llm/registry.go` - 提供商注册中心
- [ ] `internal/llm/openai.go` - OpenAI实现
- [ ] `internal/llm/anthropic.go` - Anthropic实现
- [ ] `internal/llm/ollama.go` - Ollama实现

#### 1.6 骰子系统
- [ ] `pkg/dice/dice.go` - 骰子核心结构
- [ ] `pkg/dice/roll.go` - 掷骰逻辑（支持优劣势）
- [ ] `pkg/dice/parser.go` - 表达式解析（"2d6+3"）

#### 1.7 角色系统基础
- [ ] `internal/character/character.go` - 角色结构
- [ ] `internal/character/attributes.go` - 六维属性
- [ ] `internal/character/class.go` - 职业定义
- [ ] `internal/character/race.go` - 种族定义

#### 1.8 TUI框架
- [ ] `internal/ui/app.go` - Bubbletea主模型
- [ ] `internal/ui/styles.go` - 样式定义

#### 1.9 基础测试
- [ ] `pkg/dice/dice_test.go` - 骰子测试
- [ ] `internal/character/attributes_test.go` - 属性测试

---

### Phase 2: 核心游戏功能

1. 游戏引擎核心 (`internal/game/engine.go`)
2. 角色创建流程（TUI交互式创建）
3. 对话循环系统
4. JSON存档系统
5. 叙事显示和流式输出

---

### Phase 3: 战斗系统

1. 完整战斗规则引擎
2. 怪物数据加载
3. 战斗UI界面
4. 先攻轮次管理

---

### Phase 4: 完善与优化

1. SQLite存储迁移
2. 缓存系统
3. 剧本模块
4. 性能优化

---

## 三、关键文件清单

### Phase 1 需要创建/修改的文件：

| 文件路径 | 用途 |
|----------|------|
| `main.go` | 入口点，调用cmd.Execute() |
| `cmd/root.go` | 根命令，Viper初始化 |
| `cmd/start.go` | start子命令 |
| `cmd/load.go` | load子命令 |
| `cmd/character.go` | character命令组 |
| `cmd/config.go` | config命令组 |
| `cmd/provider.go` | provider命令组 |
| `cmd/version.go` | version命令 |
| `internal/config/config.go` | 配置结构体 |
| `internal/config/loader.go` | 配置加载 |
| `internal/config/defaults.go` | 默认配置 |
| `internal/llm/provider.go` | LLM提供商接口 |
| `internal/llm/registry.go` | 提供商注册中心 |
| `internal/llm/openai.go` | OpenAI实现 |
| `internal/llm/anthropic.go` | Anthropic实现 |
| `internal/llm/ollama.go` | Ollama实现 |
| `pkg/dice/dice.go` | 骰子核心 |
| `pkg/dice/roll.go` | 掷骰逻辑 |
| `pkg/dice/parser.go` | 表达式解析 |
| `internal/character/character.go` | 角色结构 |
| `internal/character/attributes.go` | 属性系统 |
| `internal/character/class.go` | 职业定义 |
| `internal/character/race.go` | 种族定义 |
| `internal/ui/app.go` | TUI主模型 |
| `internal/ui/styles.go` | 样式定义 |
| `pkg/dice/dice_test.go` | 骰子测试 |
| `.golangci.yml` | Linter配置 |
| `Makefile` | 构建脚本 |

---

## 四、验证方式

### 构建验证
```bash
go mod tidy
go build ./...
```

### 测试验证
```bash
go test ./...
```

### CLI命令验证
```bash
# 帮助信息
go run . --help

# 版本信息
go run . version

# 配置初始化
go run . config init

# 查看配置
go run . config get

# 提供商列表
go run . provider list

# 角色命令
go run . character --help
```

### LLM连接验证
```bash
# 测试OpenAI连接
go run . provider test openai

# 测试Anthropic连接
go run . provider test anthropic

# 测试Ollama连接
go run . provider test ollama
```

---

## 五、技术风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| LLM API不稳定 | 高 | 多提供商降级、本地模型备用 |
| Token成本过高 | 中 | 缓存、压缩、本地模型 |
| TUI兼容性 | 中 | Bubbletea跨平台测试、降级方案 |
| D&D规则复杂度 | 中 | 分阶段实现，MVP简化 |
| 跨平台编译 | 低 | 纯Go实现，避免CGO |
