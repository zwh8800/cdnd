# 增强状态栏渲染 - 实现计划

## Context

当前游戏的状态栏（`renderStatusBar`）只显示基本的角色名称、等级、HP 和游戏阶段信息。为了提升玩家的游戏体验，需要扩展状态栏功能，使其能够显示更丰富的角色信息，包括属性、装备、法术位、状态效果、位置和时间等。采用**默认单行 + Tab 键切换多行面板**的设计，并且根据游戏阶段动态调整信息优先级。

## 修改文件

1. `internal/character/character.go` - 添加 Conditions 字段和方法
2. `internal/ui/game.go` - 核心状态栏实现（主要修改）
3. `internal/ui/styles.go` - 新增样式定义

## 实现步骤

### Step 1: 添加 Conditions 字段到 Character 结构体

**文件**: `internal/character/character.go`

- 在 Character 结构体中添加 `Conditions []string` 字段
- 添加辅助方法：`HasCondition()`, `AddCondition()`, `RemoveCondition()`, `GetConditions()`

### Step 2: 扩展 GameModel 结构体

**文件**: `internal/ui/game.go`

- 添加字段：
  - `expanded bool` - 是否展开状态栏
  - `statusBarHeight int` - 状态栏实际高度
- 在 `NewGameModel()` 中初始化 `statusBarHeight = 2`

### Step 3: 添加 Tab 键切换逻辑

**文件**: `internal/ui/game.go` 的 `Update()` 方法

- 在 `tea.KeyMsg` 处理中添加 `case tea.KeyTab`
- 切换 `m.expanded` 状态
- 调用 `recalculateViewport()` 重新计算视口高度
- 添加 `recalculateViewport()` 辅助方法
- 更新 `tea.WindowSizeMsg` 使用 `m.statusBarHeight` 而非硬编码值

### Step 4: 重构单行模式（核心 D&D 信息增强）

**文件**: `internal/ui/game.go`

- 将现有 `renderStatusBar()` 逻辑提取为 `renderStatusBarCompact()`
- 增强单行显示，左右分区设计：

**左侧（角色信息）:**
```
{角色名} - {种族} {等级}级 {职业}
```
示例: `瑟恩 - 人类 5级 法师`

**右侧（核心 D&D 数据，用分隔符分割）:**
```
HP:{当前}/{最大} | AC:{护甲等级} | 先攻:{修正} | G:{金币} | {阶段} | {位置名}
```
示例: `HP:28/32 | AC:12 | 先攻:+2 | G:45 | 探索 | 迷雾森林`

**完整单行效果 (终端宽度 >= 100):**
```
瑟恩 - 人类 5级 法师 ───────────────── HP:28/32 | AC:12 | 先攻:+2 | G:45 | 探索 | 迷雾森林 ── [Tab]详情
```

**窄终端自适应 (宽度 < 90):**
- 省略职业名称: `瑟恩 - 人类 5级`
- 省略位置: `HP:28/32 | AC:12 | 先攻:+2 | G:45 | 探索`

**极窄终端 (宽度 < 70):**
- 只显示最核心信息: `瑟恩 Lv5 | HP:28/32 AC:12 | 探索`

**信息优先级和颜色编码:**
- HP: 四级颜色（绿 >50%、橙 25-50%、红 <25%、灰 0）
- AC: 白色常规显示
- 先攻修正: 正数绿色、负数红色
- 金币: 金色 (#ffd700)
- 阶段: 根据阶段变色（战斗红、探索绿、休息蓝）
- 位置名: 青色，超过 6 字符截断为 `xxx...`
- `[Tab]详情`: 灰色提示文本，宽度充足时显示

**根据游戏阶段动态调整右侧信息:**
| 阶段 | 右侧显示内容 |
|------|-------------|
| 战斗 | `HP:28/32 ● ▲ AC:12 先攻:+2 ⅗ 战斗 回合:12` |
| 探索 | `HP:28/32 AC:12 G:45 探索 迷雾森林` |
| 休息 | `HP:28/32 AC:12 G:45 休息 回合:45` |
| 对话 | `HP:28/32 AC:12 G:45 对话 酒馆` |

**阶段特定信息详解:**

- **战斗阶段独有**:
  - **动作指示器**: `●` (动作未使用) / `○` (动作已使用)
  - **附赠动作指示器**: `▲` (附赠动作未使用) / `△` (附赠动作已使用)
  - **动作状态判断**: 
    - 需要追踪玩家是否在当前回合使用了动作/附赠动作
    - 方案：在 `CombatState` 中添加 `PlayerActionUsed bool` 和 `PlayerBonusActionUsed bool` 字段
    - 当玩家选择攻击/施法等动作时设置为 `true`，回合结束时重置为 `false`
  - **先攻信息**: `先攻:+2` (从 `c.Initiative` 获取)
  - **法术环分数**: 使用 Unicode 分数符号，如 `⅔` (2/3)、`⅗` (3/5)、`⅜` (3/8)
    - 只显示最高可用的两个环阶
    - 通过对比 `c.SpellSlots` (当前) 和 `GetSpellSlotsByType` (最大) 计算分数
    - Unicode 分数映射表: `0/1`→`0`, `1/2`→`½`, `1/3`→`⅓`, `2/3`→`⅔`, `1/4`→`¼`, `3/4`→`¾`, `1/5`→`⅕`, `2/5`→`⅖`, `3/5`→`⅗`, `4/5`→`⅘`, `1/6`→`⅙`, `5/6`→`⅚`, `1/8`→`⅛`, `3/8`→`⅜`, `5/8`→`⅝`, `7/8`→`⅞`
    - 若无对应 Unicode 分数则回退显示 `当前/最大` 格式

- **探索阶段**:
  - **不显示先攻信息** (进入探索时先攻顺序已被清除)
  - 显示位置名和金币

- **非战斗阶段**:
  - 不显示动作指示器和先攻

- `renderStatusBar()` 改为根据 `m.expanded` 分发到不同渲染函数
- 新增辅助函数:
  - `abbreviateLocation(name string, maxLen int) string` - 智能截断位置名
  - `formatModifier(mod int) string` - 格式化属性修正（+2/-1）

### Step 5: 实现多行展开模式

**文件**: `internal/ui/game.go`

新增以下函数：

1. **`renderStatusBarExpanded()`** - 多行面板容器
   - 宽终端（>=100 字符）：5 个面板水平排列
   - 窄终端（<100 字符）：2-3 行紧凑布局

2. **`renderStatPanel()`** - 属性面板
   ```
   ┌──── 属性 ────┐
   │ STR 15 (+2)  │
   │ DEX 14 (+2)  │
   │ CON 13 (+1)  │
   │ INT 16 (+3)  │
   │ WIS 12 (+1)  │
   │ CHA 10 (+0)  │
   │ AC:12 先攻:+2│
   │ 速度:30 熟:+3│
   └──────────────┘
   ```

3. **`renderEquipmentPanel()`** - 装备面板
   ```
   ┌─── 装备 ─────┐
   │ 武器: 长剑   │
   │ 护甲: 皮甲   │
   │ 盾牌: 无     │
   │ 金币: 45     │
   │ 背包: 12 件   │
   └──────────────┘
   ```

4. **`renderSpellPanel()`** - 法术面板
   ```
   ┌─── 法术 ─────┐
   │ 施法属性: 智力│
   │ 戏法: 3      │
   │ 1 环: ███░ 3/4│
   │ 2 环: ██░░ 2/2│
   │ 3 环: ░░░░ 0/2│
   └──────────────┘
   ```
   - 非施法者显示"非施法者"

5. **`renderConditionsPanel()`** - 状态效果面板
   ```
   ┌─── 状态 ────┐
   │ 无状态效果   │
   │ 或           │
   │ ⚠ 中毒       │
   │ ⚠ 眩晕       │
   └──────────────┘
   ```

6. **`renderLocationPanel()`** - 位置/时间面板
   ```
   ┌── 位置/时间 ─┐
   │ 迷雾森林     │
   │ 类型: 荒野   │
   │ 光照: 明亮   │
   │ 地形: 普通   │
   │ 回合: 45     │
   │ 用时: 1h23m  │
   └──────────────┘
   ```

7. **辅助函数**:
   - `formatPlayTime(seconds int) string` - 格式化游戏时间
   - `abbreviateLocation(name string) string` - 缩写位置名

### Step 6: 添加样式定义

**文件**: `internal/ui/styles.go` 或 `game.go` 的 `GameStyles`

新增样式：
- `PanelBorder` - 面板边框（紫色边框）
- `PanelTitle` - 面板标题（粗体、居中、紫色）
- `StatLabel` / `StatValue` - 属性标签和值
- `StatModPositive/Negative/Neutral` - 修正值颜色（绿/红/灰）
- `ConditionBadge` - 状态效果徽章（黄底黑字）
- `SpellSlotFilled/Empty` - 法术位指示器
- `LocationName` - 位置名（青色粗体）
- `GoldText` - 金币（金色）

颜色编码：
- 每个属性用不同颜色：STR 红、DEX 绿、CON 橙、INT 蓝、WIS 紫、CHA 黄
- HP 保持现有四级颜色（绿/橙/红/灰）
- 条件效果使用黄色警告色

## 验证方案

1. **编译测试**: `go build ./...` 确保无编译错误
2. **运行测试**: `go run cmd/cdnd/main.go` 启动游戏
3. **功能验证**:
   - 默认状态栏显示增强信息（位置、回合数）
   - 按 Tab 键切换到多行面板模式
   - 再次按 Tab 键恢复单行模式
   - 检查所有 5 个面板的信息正确性
   - 调整终端宽度测试响应式布局
   - 测试非施法者角色的法术面板显示
   - 测试添加/移除 Conditions 的显示效果
4. **边界测试**:
   - 窄终端（80 字符以下）布局
   - 宽终端（120 字符以上）布局
   - 角色创建阶段（无角色时）的状态栏
   - 无场景信息时的位置显示