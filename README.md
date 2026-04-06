# CDND

由大语言模型（LLM）驱动的命令行龙与地下城（D&D 5e）角色扮演游戏。

## 简介

CDND 是一款基于 Go 语言开发的 D&D 5e 桌面角色扮演游戏，采用终端用户界面（TUI），由大语言模型担任地下城主（DM），为玩家提供沉浸式的互动冒险体验。游戏支持多种 LLM 提供商，包括 OpenAI、Anthropic Claude 和 Ollama（本地模型）。

## 主要特性

- **LLM 驱动的地下城主**：由 AI 担任 DM，动态生成故事情节、NPC 对话和游戏世界
- **多 LLM 提供商支持**：
  - OpenAI API（兼容 OpenAI 格式的 API，如阿里云 DashScope）
  - Anthropic Claude API
  - Ollama 本地模型
- **完整的 D&D 5e 角色系统**：
  - 种族选择（人类、精灵、矮人、半身人等）
  - 职业系统（战士、法师、游侠、牧师等）
  - 属性分配（力量、敏捷、体质、智力、感知、魅力）
  - 技能、法术槽、 inventory 管理
- **战斗系统**：
  - 先攻顺序与回合制战斗
  - 骰子检定（d4、d6、d8、d10、d12、d20、d100）
  - 技能检定与豁免检定
  - 状态效果（Condition）管理
- **世界与任务系统**：
  - 场景管理与移动
  - NPC 生成与交互
  - 任务状态追踪
- **存档功能**：
  - 多槽位存档（支持 10 个存档槽）
  - 自动保存（可配置间隔）
  - 存档预览与管理
- **终端用户界面（TUI）**：
  - 基于 Bubble Tea 构建的交互式界面
  - 角色创建向导
  - 状态栏实时显示角色信息
  - 战斗面板
  - 打字机效果与彩色输出（可配置）
- **DM 工具系统**：LLM 可调用的工具集，用于管理游戏状态（掷骰、伤害、治疗、物品、移动等）

## 技术栈

- **Go 1.24+**：主要编程语言
- **Cobra**：CLI 命令框架
- **Viper**：配置文件管理
- **Bubble Tea / Bubbles**：TUI 框架（Charm 生态）
- **Lipgloss**：终端样式与布局
- **OpenAI SDK / Anthropic SDK**：LLM API 客户端

## 安装

### 前提条件

- Go 1.24 或更高版本
- 一个可用的 LLM 提供商 API 密钥（或本地 Ollama 实例）

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/zwh8800/cdnd.git
cd cdnd

# 下载依赖
make deps

# 构建
make build

# 可选：安装到 $GOPATH/bin
make install
```

构建后的二进制文件位于 `bin/cdnd`。

### 使用 Go 安装

```bash
go install github.com/zwh8800/cdnd@latest
```

## 配置

### 初始化配置

运行以下命令初始化配置文件：

```bash
cdnd config init
```

配置文件默认位于 `~/.cdnd/config.yaml`。你也可以通过 `--config` 标志指定自定义路径：

```bash
cdnd --config /path/to/config.yaml start
```

### 配置文件示例

```yaml
llm:
  default_provider: openai
  providers:
    openai:
      api_key: ""  # 也可通过 OPENAI_API_KEY 环境变量设置
      model: "gpt-4"
      base_url: "https://api.openai.com/v1"
      max_tokens: 65536
      temperature: 1.0
    anthropic:
      api_key: ""  # 也可通过 ANTHROPIC_API_KEY 环境变量设置
      model: "claude-3-opus-20240229"
      max_tokens: 4096
      temperature: 0.7
    ollama:
      model: "llama2"
      base_url: "http://localhost:11434"
      max_tokens: 4096
      temperature: 0.7

game:
  autosave: true
  autosave_interval: 5m
  max_history_turns: 100
  language: "zh-CN"

display:
  typewriter_effect: true
  typing_speed: 50ms
  color_output: true
  show_tokens: false
```

### 配置 LLM 提供商

#### OpenAI

1. 在 [OpenAI 平台](https://platform.openai.com/) 获取 API 密钥
2. 设置环境变量 `OPENAI_API_KEY=your-api-key`，或在配置文件中填写 `api_key`
3. 如需使用兼容 API（如阿里云 DashScope），修改 `base_url` 和 `model`

#### Anthropic Claude

1. 在 [Anthropic 控制台](https://console.anthropic.com/) 获取 API 密钥
2. 设置环境变量 `ANTHROPIC_API_KEY=your-api-key`，或在配置文件中填写 `api_key`

#### Ollama（本地）

1. 安装 [Ollama](https://ollama.ai/)
2. 拉取所需模型：`ollama pull llama2`
3. 确保 Ollama 服务正在运行（默认 `http://localhost:11434`）

### 切换 LLM 提供商

```bash
cdnd provider set anthropic
cdnd provider set ollama
cdnd provider set openai
```

## 游戏玩法

### 开始新游戏

```bash
cdnd start
```

可选参数：

| 参数 | 简写 | 说明 | 默认值 |
|------|------|------|--------|
| `--save-slot` | `-s` | 存档槽位编号（1-10） | 1 |
| `--scenario` | `-S` | 剧本名称 | default |
| `--skip-creation` | | 跳过角色创建（测试用） | false |
| `--no-autosave` | | 禁用自动保存 | false |

### 加载存档

```bash
cdnd load --slot 1
cdnd load -s 3
```

### 查看存档

```bash
cdnd saves
```

### 角色管理

```bash
# 创建角色
cdnd character create

# 列出角色
cdnd character list

# 显示角色详情
cdnd character show [name]

# 删除角色
cdnd character delete [name]
```

### 其他命令

```bash
# 查看配置
cdnd config view

# 初始化配置
cdnd config init

# 查看版本
cdnd version
```

### 游戏内交互

游戏启动后，你将进入 TUI 界面：

- **角色创建阶段**：通过交互式表单选择种族、职业、分配属性点
- **游戏阶段**：输入自然语言描述你的行动，AI DM 将响应并推进剧情
- **战斗阶段**：系统自动进入回合制战斗，你可以描述攻击、施法、移动等行动
- **状态栏**：屏幕底部实时显示 HP、AC、等级、位置等关键信息

## 项目结构

```
cdnd/
├── application/
│   ├── engine/          # 游戏引擎核心（事件循环、工具注册、LLM 交互）
│   ├── state/           # 游戏状态管理
│   └── tools/           # DM 工具实现（骰子、战斗、物品、世界等）
├── domain/
│   ├── character/       # 角色领域模型（属性、种族、职业、技能、法术、物品栏）
│   ├── combat/          # 战斗状态与逻辑
│   ├── dice/            # 骰子系统（解析与投掷）
│   ├── events/          # 事件分发系统
│   ├── llm/             # LLM 类型定义
│   ├── monster/         # 怪物管理
│   ├── quest/           # 任务状态
│   ├── rules/           # D&D 5e 规则引擎
│   └── world/           # 世界管理（场景、NPC）
├── infrastructure/
│   ├── config/          # 配置加载与管理
│   ├── llm/             # LLM 提供商实现（OpenAI、Anthropic、Ollama）
│   ├── prompt/          # 提示词构建与模板
│   └── storage/         # 存档存储管理
├── interface/
│   ├── cmd/             # CLI 命令（Cobra）
│   └── ui/              # TUI 界面（Bubble Tea）
├── docs/                # 文档
├── Makefile             # 构建脚本
├── config.example.yaml  # 配置示例
└── main.go              # 入口文件
```

## 开发

### 常用命令

```bash
# 构建
make build

# 运行
make run

# 运行测试
make test

# 生成覆盖率报告
make test-cover

# 格式化代码
make fmt

# 代码检查
make lint

# 清理构建产物
make clean

# 下载依赖
make deps

# 整理依赖
make tidy

# 热重载开发（需要 entr）
make watch
```

### 跨平台构建

```bash
make build-all      # 构建所有平台
make build-linux    # Linux amd64
make build-darwin   # macOS amd64 + arm64
make build-windows  # Windows amd64
```

## 许可证

本项目采用 [GNU General Public License v3.0](LICENSE)。

## 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建你的特性分支（`git checkout -b feature/amazing-feature`）
3. 提交你的改动（`git commit -m 'Add some amazing feature'`）
4. 推送到分支（`git push origin feature/amazing-feature`）
5. 提交 Pull Request

## 致谢

- D&D 5e 规则由 Wizards of the Coast 提供
- [Charm](https://github.com/charmbracelet) 生态提供的 TUI 组件
- OpenAI、Anthropic、Ollama 提供的 LLM 能力
