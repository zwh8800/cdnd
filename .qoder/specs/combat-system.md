# Phase 3: 战斗系统实现计划

## 上下文

Phase 2 已完成游戏引擎核心、角色创建和对话循环。Phase 3 目标是实现完整的回合制战斗系统，包括怪物数据、战斗工具和战斗UI。

**用户需求**：
- 战斗模式：LLM驱动 + 工具辅助（玩家自然语言描述行动）
- 怪物数据：预设模板库（15个基础怪物）
- 战斗UI：扩展现有GameModel，通过阶段切换显示战斗信息

**已有基础**：
- `CombatState`/`Combatant` 数据结构已定义 (`internal/save/types.go`)
- `StartCombat()`/`EndCombat()`/`NextTurn()` 方法已存在 (`internal/game/state.go`)
- `DealDamageTool`/`HealCharacterTool` 已存在但对NPC处理简化
- `PhaseCombat` 游戏阶段已定义
- UI三层布局（状态栏 + viewport + 输入框）已完善

**新增需求**：战斗独立历史记录
- 战斗期间LLM交互使用独立历史记录，不污染主历史
- 战斗结束后生成摘要，仅将摘要加入主历史

---

## 实现步骤

### 步骤1: 怪物模板系统

**新建文件**: `internal/monster/`

```
internal/monster/
├── types.go       # MonsterTemplate, MonsterAction 数据结构
├── templates.go   # 15个预设怪物模板（代码内嵌）
└── manager.go     # 怪物管理器（查找、实例化）
```

---

### 步骤1.5: 战斗历史记录系统

**修改文件**: `internal/save/types.go`

在 `CombatState` 中添加战斗历史字段：
```go
type CombatState struct {
    // ... 现有字段 ...
    History []llm.Message `json:"history,omitempty"` // 战斗期间的历史记录
}
```

**修改文件**: `internal/game/state.go`

新增方法：
```go
// AddCombatHistory 添加战斗历史消息
func (s *State) AddCombatHistory(msg llm.Message)

// GetCombatHistory 获取战斗历史（用于构建LLM请求）
func (s *State) GetCombatHistory() []llm.Message

// ClearCombatHistory 清空战斗历史
func (s *State) ClearCombatHistory()

// GenerateCombatSummary 请求LLM生成战斗摘要
func (s *State) GenerateCombatSummary(ctx context.Context, provider llm.Provider) (string, error)
```

**设计说明**：
- 战斗开始时，保存当前主历史记录的"断点"
- 战斗期间所有LLM交互使用 `CombatState.History`
- 战斗结束时，调用LLM生成摘要（基于战斗历史）
- 将摘要作为单条消息添加到主历史记录
- 清空 `CombatState.History` 并结束战斗

**MonsterTemplate 核心字段**：
```go
type MonsterTemplate struct {
    ID, Name, Size, Type, Alignment string
    CR float64
    XP int
    HP, AC, Speed int
    Abilities character.Attributes
    Actions []MonsterAction
    DamageResistances, DamageImmunities []string
}

type MonsterAction struct {
    Name, Type string
    AttackBonus int
    Damage, DamageType, Range string
}
```

**15个预设怪物**（CR递增）：
- CR 1/4: 哥布林、骷髅、僵尸、巨鼠、狗头人
- CR 1/2: 熊地精、哥布林首领、骷髅战士、恐狼
- CR 1-2: 兽人、食人魔、食尸鬼
- CR 3-5: 巨魔、枭熊、幽灵

---

### 步骤2: 战斗工具扩展

**新建文件**: `internal/tools/combat_tools.go`

| 工具 | 功能 | 关键参数 |
|------|------|---------|
| `StartCombatTool` | 初始化战斗 | enemies: []{monster_id, name_override} |
| `AttackTool` | 攻击检定 | attacker, target, attack_type, advantage |
| `NextTurnTool` | 推进回合 | 无 |
| `EndCombatTool` | 结束战斗 | reason: victory/defeat/flee |
| `SpawnEnemyTool` | 战斗中生成敌人 | monster_id, position |

**AttackTool 核心流程**：
1. 确定攻击者（玩家/敌人）
2. 调用 `rules.AttackRoll()` 执行检定
3. 判断命中（对比AC，处理暴击）
4. 计算伤害（投骰 + 属性加值）
5. 更新目标HP
6. 返回结果叙述

---

### 步骤3: 战斗状态管理增强

**修改文件**: `internal/game/state.go`

新增方法：
- `GetCombatant(id string) *Combatant` - 按ID查找参与者
- `RemoveCombatant(id string)` - 移除死亡敌人
- `IsPlayerTurn() bool` - 判断是否玩家回合
- `GetEnemies() []*Combatant` - 获取所有敌人列表

**修改文件**: `internal/game/engine.go`
- 注册新的战斗工具
- 在 `generateToolNarrative()` 添加战斗工具叙述
- 修改 `Process()` 方法支持战斗历史记录：
  - 战斗阶段使用 `state.GetCombatHistory()` 构建消息
  - 战斗结束后调用 `GenerateCombatSummary()` 并添加到主历史

**Engine.Process() 战斗历史处理逻辑**：
```go
// 构建LLM请求消息时
var messages []llm.Message
if state.GetPhase() == PhaseCombat {
    // 战斗阶段：使用系统提示 + 战斗历史 + 当前输入
    messages = append(messages, llm.Message{Role: RoleSystem, Content: combatSystemPrompt})
    messages = append(messages, state.GetCombatHistory()...)
    messages = append(messages, llm.Message{Role: RoleUser, Content: input})
} else {
    // 非战斗阶段：使用原有逻辑（系统提示 + 主历史 + 当前输入）
    messages = append(messages, llm.Message{Role: RoleSystem, Content: systemPrompt})
    messages = append(messages, prompt.BuildHistoryContext(state.GetHistory(), 20)...)
    messages = append(messages, llm.Message{Role: RoleUser, Content: input})
}
```

---

### 步骤4: LLM提示词扩展

**修改文件**: `internal/llm/prompt/templates.go`

新增战斗阶段系统提示：
```
你正在主持一场D&D战斗。当前状态：
- 回合: {round}, 当前行动: {current_combatant}
- 玩家HP: {player_hp}/{player_max_hp}
- 敌人状态: {enemy_list}

可用工具：
- attack: 进行攻击检定
- next_turn: 结束当前回合
- end_combat: 结束战斗

玩家行动描述: {player_input}
请解析玩家意图并调用相应工具...
```

---

### 步骤5: 战斗UI扩展

**新建文件**: `internal/ui/combat_panel.go`

```go
func RenderCombatPanel(combat *save.CombatState, width int) string
func RenderEnemyStatus(enemies []*save.Combatant, width int) string
func RenderInitiativeOrder(initiative []save.InitiativeEntry, current int) string
```

**修改文件**: `internal/ui/game.go`
- 新增字段: `combatPanelExpanded bool`
- 修改 `View()`: 战斗阶段时渲染战斗面板
- 修改 `recalculateViewport()`: 战斗时调整高度

**修改文件**: `internal/ui/statusbar.go`
- 战斗阶段展开模式显示敌人状态
- 添加先攻顺序指示器

---

### 步骤6: 战斗流程集成（含独立历史记录）

**战斗开始流程**：
```
LLM识别战斗 → 调用 start_combat →
  → 创建Combatant实例 → 投先攻 → 排序 →
  → state.StartCombat() → PhaseCombat →
  → 初始化 CombatState.History（空）→
  → 添加系统消息到战斗历史 →
  → 返回战斗状态给LLM
```

**战斗回合流程（使用独立历史）**：
```
[玩家回合]
1. 玩家输入 → Engine.Process()
2. 构建消息：[系统提示] + [CombatState.History] + [玩家输入]
3. LLM生成响应（可能包含工具调用）
4. 执行工具 → 更新状态
5. 将 assistant + tool 消息添加到 CombatState.History
6. 返回响应给UI（但不添加到主History）

[敌人回合]
1. Engine.Process() 自动触发（或LLM调用 next_turn）
2. 构建消息使用 CombatState.History
3. LLM根据怪物模板决策行动
4. 执行 attack 等工具
5. 更新 CombatState.History
6. 继续下一回合
```

**战斗结束流程（历史合并）**：
```
调用 end_combat →
  1. 收集战斗统计数据：
     - 回合数、战斗时长
     - 击杀敌人列表
     - 玩家受到的总伤害
     - 使用的关键技能/法术
  
  2. 调用 GenerateCombatSummary()：
     - 将 CombatState.History 发送给LLM
     - 提示词："请总结这场战斗的关键事件，2-3句话"
     - 返回摘要文本
  
  3. 添加摘要到主历史：
     - 格式："【战斗】vs {敌人列表} - {结果}\n{摘要}"
     - 添加到 state.History（主历史）
  
  4. 清理：
     - 清空 CombatState.History
     - state.EndCombat() → PhaseExploration
```

**示例战斗摘要**：
```
【战斗】vs 哥布林 x3 - 胜利
经过3回合的激战，你成功击败了三个哥布林。战斗中你使用了剑击和火焰箭，
虽然受到了12点伤害，但最终凭借精准的攻击取得了胜利。获得150经验值。
```

---

## 关键文件清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/monster/types.go` | 新建 | 怪物数据结构 |
| `internal/monster/templates.go` | 新建 | 15个预设怪物 |
| `internal/monster/manager.go` | 新建 | 怪物管理器 |
| `internal/tools/combat_tools.go` | 新建 | 5个战斗工具 |
| `internal/ui/combat_panel.go` | 新建 | 战斗面板渲染 |
| `internal/game/state.go` | 修改 | 添加战斗辅助方法 |
| `internal/game/engine.go` | 修改 | 注册工具、叙述生成、战斗历史处理 |
| `internal/llm/prompt/templates.go` | 修改 | 战斗提示词 |
| `internal/ui/game.go` | 修改 | 战斗UI集成 |
| `internal/ui/statusbar.go` | 修改 | 战斗状态显示 |

---

## 验证计划

### 单元测试
```bash
go test ./internal/monster/... -v
go test ./internal/tools/... -v -run Combat
```

### 集成测试
1. 启动游戏: `go run main.go start`
2. 完成角色创建
3. 在对话中触发战斗（输入"攻击这个哥布林"）
4. 验证:
   - 战斗面板正确显示
   - 先攻顺序排序正确
   - 攻击检定计算正确
   - HP更新正确
   - 战斗结束流程正常

### 手动测试场景
1. **单敌人战斗**: 玩家 vs 1个哥布林
2. **多敌人战斗**: 玩家 vs 3个骷髅
3. **战斗逃跑**: 使用"逃跑"选项
4. **Boss战斗**: 玩家 vs 食人魔（高CR敌人）

### 战斗历史记录验证
1. **历史隔离测试**：
   - 开始战斗前记录主历史消息数
   - 进行3-4回合战斗
   - 验证主历史消息数未增加
   - 验证 CombatState.History 有消息

2. **摘要生成测试**：
   - 结束战斗后检查主历史
   - 验证只有一条战斗摘要消息
   - 验证摘要包含战斗结果和关键事件
   - 验证摘要不包含具体骰子数值

3. **上下文连贯性测试**：
   - 战斗前进行一段对话
   - 进行战斗
   - 战斗后继续对话
   - 验证LLM能正确理解战斗前后的上下文

---

## 实现优先级

1. **P0 - 核心基础**
   - 怪物模板系统
   - StartCombatTool / AttackTool / EndCombatTool

2. **P1 - UI集成**
   - 战斗面板渲染
   - 状态栏增强

3. **P2 - 完善功能**
   - NextTurnTool / SpawnEnemyTool
   - 战斗提示词优化
   - 战利品/经验值处理

4. **P3 - 高级功能**（可选）
   - 法术战斗
   - 伤害抗性/易伤
   - 传奇动作

---

## 战斗历史记录设计要点

### 为什么需要独立历史？

1. **避免污染主叙事**：战斗中的技术性对话（"投骰：15命中，造成8点伤害"）会打断故事连贯性
2. **减少Token消耗**：长战斗可能产生大量消息，独立历史可以控制上下文大小
3. **更好的摘要质量**：LLM基于完整战斗历史生成摘要，比逐条记录更连贯

### 实现关键点

**数据隔离**：
- `state.History` - 主历史（探索、对话阶段使用）
- `state.Combat.History` - 战斗历史（战斗阶段专用）

**上下文切换**：
```go
// Engine.Process() 中的消息构建
if state.GetPhase() == PhaseCombat {
    history = state.GetCombatHistory()  // 使用战斗历史
} else {
    history = prompt.BuildHistoryContext(state.GetHistory(), 20)  // 使用主历史
}
```

**摘要生成提示词**：
```
请根据以下战斗记录，生成一段2-3句话的简洁摘要。
重点包括：战斗结果、关键行动、值得注意的事件。
不要包含具体的骰子数值或伤害计算细节。

战斗记录：
{CombatState.History}

请生成摘要：
```

**合并到主历史**：
```go
summary := state.GenerateCombatSummary(ctx, provider)
state.AddHistory(llm.Message{
    Role: RoleAssistant,
    Content: fmt.Sprintf("【战斗】vs %s - %s\n%s", 
        enemyList, result, summary),
})
```
