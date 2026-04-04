# Phase 2 游戏引擎核心实现计划

## Context

本计划旨在实现 D&D CLI 游戏的 Phase 2 核心功能，包括游戏引擎基础架构、TUI 交互式角色创建流程、以及基于 LLM Tool Call 的对话交互循环。项目基础框架已完成（角色系统、骰子系统、LLM 集成、配置系统、TUI 框架），需要在此基础上构建可运行的游戏循环。

**用户选择确认：**
- 角色创建：TUI 交互式创建
- 存档系统：内存 + JSON 文件
- LLM 角色：DM + NPC 扮演
- 优先级：基础游戏循环 + LLM 对话

**核心设计原则：**
1. 严格按照 D&D 5e 官方规则设计种族和职业系统
2. 采用 Tool Call 机制让 LLM 调用预定义的游戏控制函数

---

## D&D 5e 完整种族和职业系统

> **参考文档**: `docs/DnD_5E_新手套组_基础入门规则CN.pdf`
> 所有术语使用官方中文翻译

### 种族系统设计 (严格按照官方规则)

**现有实现** (`internal/character/race.go`)：
- 已有 4 个种族：人类、精灵、矮人、半身人
- 缺少：子种族系统、年龄/身高/体重范围、完整的 D&D 5e 种族

**扩展目标：**

```go
// Race 种族结构 (扩展版)
type Race struct {
    ID              string              `json:"id"`
    Name            string              `json:"name"`            // 中文名称
    Description     string              `json:"description"`     // 中文描述
    Size            Size                `json:"size"`            // 体型
    Speed           int                 `json:"speed"`           // 速度（尺）
    AbilityBonuses  map[Ability]int     `json:"ability_bonuses"` // 属性加值
    Traits          []Trait             `json:"traits"`          // 种族特性
    Languages       []string            `json:"languages"`       // 语言
    // 新增字段
    SubRaces        []SubRace           `json:"sub_races,omitempty"`     // 子种族
    AgeRange        AgeRange            `json:"age_range"`               // 成年年龄和寿命
    HeightRange     HeightRange         `json:"height_range"`            // 身高范围
    WeightRange     WeightRange         `json:"weight_range"`            // 体重范围
    WeaponTraining  []string            `json:"weapon_training,omitempty"` // 武器熟练
    Cantrips        []string            `json:"cantrips,omitempty"`      // 天生法术（戏法）
}

// SubRace 子种族
type SubRace struct {
    ID             string          `json:"id"`
    Name           string          `json:"name"`           // 中文子种族名
    Description    string          `json:"description"`    // 中文描述
    AbilityBonuses map[Ability]int `json:"ability_bonuses"`
    Traits         []Trait         `json:"traits"`
}

// Size 体型（官方中文）
type Size string
const (
    SizeTiny       Size = "微型"    // Tiny
    SizeSmall      Size = "小型"    // Small
    SizeMedium     Size = "中型"    // Medium
    SizeLarge      Size = "大型"    // Large
    SizeHuge       Size = "巨型"    // Huge
    SizeGargantuan Size = "超巨型"  // Gargantuan
)
```

**完整种族列表 (《玩家手册》官方)：**

| 种族 | 主属性加值 | 速度 | 体型 | 子种族 |
|------|-----------|------|------|--------|
| 人类 | 全属性+1 | 30尺 | 中型 | - |
| 精灵 | 敏捷+2 | 30尺 | 中型 | 高等精灵、木精灵、卓尔精灵 |
| 矮人 | 体质+2 | 25尺 | 中型 | 山丘矮人、山地矮人 |
| 半身人 | 敏捷+2 | 25尺 | 小型 | 轻足半身人、壮实半身人 |
| 龙裔 | 力量+2，魅力+1 | 30尺 | 中型 | 按龙种区分 |
| 侏儒 | 智力+2 | 25尺 | 小型 | 森林侏儒、岩石侏儒 |
| 半精灵 | 魅力+2，任选两项+1 | 30尺 | 中型 | - |
| 半兽人 | 力量+2，体质+1 | 30尺 | 中型 | - |
| 提夫林 | 智力+1，魅力+2 | 30尺 | 中型 | - |

### 职业系统设计 (严格按照官方规则)

**现有实现** (`internal/character/class.go`)：
- 已有 4 个职业：战士、法师、游荡者、牧师
- 缺少：完整 12 职业、子职业系统、法术槽计算

**扩展目标：**

```go
// Class 职业结构 (扩展版)
type Class struct {
    ID                  string           `json:"id"`
    Name                string           `json:"name"`           // 中文名称
    Type                ClassType        `json:"type"`
    Description         string           `json:"description"`    // 中文描述
    HitDice             HitDice          `json:"hit_dice"`       // 生命骰
    PrimaryAbility      []Ability        `json:"primary_ability"` // 主要属性
    SavingThrows        []Ability        `json:"saving_throws"`  // 豁免检定熟练
    SkillCount          int              `json:"skill_count"`    // 可选技能数量
    SkillOptions        []SkillType      `json:"skill_options"`  // 可选技能列表
    ArmorProficiencies  []string         `json:"armor_proficiencies"`  // 护甲熟练
    WeaponProficiencies []string         `json:"weapon_proficiencies"` // 武器熟练
    Tools               []string         `json:"tools"`          // 工具熟练
    Features            []Feature        `json:"features"`       // 职业特性
    Spellcasting        bool             `json:"spellcasting"`   // 是否施法者
    // 新增字段
    SubClasses          []SubClass       `json:"sub_classes,omitempty"`    // 子职业
    SpellcastingAbility Ability          `json:"spellcasting_ability,omitempty"` // 施法属性
    SpellSlotProgression SpellSlotTable  `json:"spell_slot_progression,omitempty"` // 法术槽成长表
    RitualCaster        bool             `json:"ritual_caster,omitempty"`  // 仪式施法
    CantripCount        int              `json:"cantrip_count,omitempty"`  // 戏法数量
}

// SubClass 子职业 (1级/2级/3级选择)
type SubClass struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`        // 中文子职业名
    Description string    `json:"description"` // 中文描述
    Level       int       `json:"level"`       // 获得等级
    Features    []Feature `json:"features"`
}

// HitDice 生命骰（官方中文）
type HitDice string
const (
    D6  HitDice = "d6"
    D8  HitDice = "d8"
    D10 HitDice = "d10"
    D12 HitDice = "d12"
)
```

**完整职业列表 (《玩家手册》官方)：**

| 职业 | 生命骰 | 主属性 | 豁免 | 技能数 | 施法 | 子职业 (3级) |
|------|-------|--------|------|-------|------|--------------|
| 野蛮人 | d12 | 力量 | 力量/体质 | 2 | 否 | 狂战士、图腾武士 |
| 吟游诗人 | d8 | 魅力 | 敏捷/魅力 | 3 | 是 (魅力) | 逸闻学院、勇武学院 |
| 牧师 | d8 | 感知 | 感知/魅力 | 2 | 是 (感知) | 7个神祗领域 |
| 德鲁伊 | d8 | 感知 | 智力/感知 | 2 | 是 (感知) | 大地结社、月亮结社 |
| 战士 | d10 | 力量/敏捷 | 力量/体质 | 2 | 否* | 冠军勇士、战斗大师、奥法骑士 |
| 武僧 | d8 | 敏捷/感知 | 力量/敏捷 | 2 | 否 | 敞开心门、暗影之道、四象之道 |
| 圣武士 | d10 | 力量/魅力 | 感知/魅力 | 2 | 是 (魅力) | 虔诚之誓、远古之誓、复仇之誓 |
| 游侠 | d10 | 敏捷/感知 | 力量/敏捷 | 3 | 是 (感知) | 猎人、兽王 |
| 游荡者 | d8 | 敏捷 | 敏捷/智力 | 4 | 否* | 盗贼、刺客、诡术师 |
| 术士 | d6 | 魅力 | 体质/魅力 | 2 | 是 (魅力) | 龙族血统、狂野魔法 |
| 邪术师 | d8 | 魅力 | 感知/魅力 | 2 | 是 (魅力) | 妖精宗主、邪魔宗主、旧日支配者 |
| 法师 | d6 | 智力 | 智力/感知 | 2 | 是 (智力) | 8个奥术传承 |

*注：奥法骑士和诡术师有施法能力

### 六项属性与技能（官方中文）

```go
// Ability 六项属性
type Ability string
const (
    Strength     Ability = "力量"      // Strength - 体能的量化
    Dexterity    Ability = "敏捷"      // Dexterity - 灵活度的量化
    Constitution Ability = "体质"      // Constitution - 耐受力的量化
    Intelligence Ability = "智力"      // Intelligence - 记忆与思维能力的量化
    Wisdom       Ability = "感知"      // Wisdom - 直觉与感受能力的量化
    Charisma     Ability = "魅力"      // Charisma - 个性气质的量化
)

// SkillType 技能类型（官方中文）
type SkillType string
const (
    // 力量技能
    Athletics SkillType = "运动"      // Athletics
    // 敏捷技能
    Acrobatics     SkillType = "体操"      // Acrobatics
    SleightOfHand  SkillType = "手法"      // Sleight of Hand
    Stealth        SkillType = "隐匿"      // Stealth
    // 智力技能
    Arcana        SkillType = "奥秘"      // Arcana
    History       SkillType = "历史"      // History
    Investigation SkillType = "调查"      // Investigation
    Nature        SkillType = "自然"      // Nature
    Religion      SkillType = "宗教"      // Religion
    // 感知技能
    AnimalHandling SkillType = "驯兽"     // Animal Handling
    Insight        SkillType = "洞察"     // Insight
    Medicine       SkillType = "医药"     // Medicine
    Perception     SkillType = "察觉"     // Perception
    Survival       SkillType = "求生"     // Survival
    // 魅力技能
    Deception    SkillType = "欺瞒"      // Deception
    Intimidation SkillType = "威吓"      // Intimidation
    Performance  SkillType = "表演"      // Performance
    Persuasion   SkillType = "说服"      // Persuasion
)
```

### 伤害类型（官方中文）

```go
// DamageType 伤害类型
type DamageType string
const (
    Acid       DamageType = "强酸"      // Acid
    Bludgeoning DamageType = "钝击"     // Bludgeoning
    Cold       DamageType = "冷冻"      // Cold
    Fire       DamageType = "火焰"      // Fire
    Force      DamageType = "力场"      // Force
    Lightning  DamageType = "闪电"      // Lightning
    Necrotic   DamageType = "黯蚀"      // Necrotic
    Piercing   DamageType = "穿刺"      // Piercing
    Poison     DamageType = "毒素"      // Poison
    Psychic    DamageType = "心灵"      // Psychic
    Radiant    DamageType = "光耀"      // Radiant
    Slashing   DamageType = "挥砍"      // Slashing
    Thunder    DamageType = "雷鸣"      // Thunder
)
```

### 魔法学派（官方中文）

```go
// SpellSchool 魔法学派
type SpellSchool string
const (
    SchoolAbjuration    SpellSchool = "防护"    // Abjuration - 守护、屏障
    SchoolConjuration   SpellSchool = "咒法"    // Conjuration - 传送物件和生物
    SchoolDivination    SpellSchool = "预言"    // Divination - 得知未来和秘密
    SchoolEnchantment   SpellSchool = "附魔"    // Enchantment - 影响心智
    SchoolEvocation     SpellSchool = "塑能"    // Evocation - 能量塑形
    SchoolIllusion      SpellSchool = "幻术"    // Illusion - 瞒骗感官
    SchoolNecromancy    SpellSchool = "死灵"    // Necromancy - 操纵生死能量
    SchoolTransmutation SpellSchool = "变化"    // Transmutation - 改变属性
)
```

### 需修改的文件

| 文件 | 修改内容 |
|------|----------|
| `internal/character/race.go` | 扩展 Race 结构，添加子种族，补全 9 个种族 |
| `internal/character/class.go` | 扩展 Class 结构，添加子职业，补全 12 个职业 |
| `internal/character/race_data.go` | **新建** - 完整的种族数据定义 |
| `internal/character/class_data.go` | **新建** - 完整的职业数据定义 |
| `internal/character/spell_slots.go` | **新建** - 法术槽成长表 |

---

## Tool Call 机制架构

### 设计理念

使用 OpenAI Function Calling / Tool Call 机制，让 LLM (DM) 通过调用预定义的工具函数来控制游戏世界，而不是依赖自然语言解析。这提供了：

1. **可靠性**：工具调用结果确定，不依赖文本解析
2. **安全性**：工具可做参数验证和权限控制
3. **可扩展性**：新增功能只需添加新工具
4. **一致性**：游戏规则由代码强制执行

### 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         LLM (DM)                                │
│  - 分析玩家行动                                                 │
│  - 决定调用哪些工具                                             │
│  - 生成叙述文本                                                 │
└───────────────────────────┬─────────────────────────────────────┘
                            │ Tool Call Request
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Tool Registry                                │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌───────────┐ │
│  │ roll_dice   │ │ skill_check │ │ deal_damage │ │ heal_char │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └───────────┘ │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌───────────┐ │
│  │ add_item    │ │ remove_item │ │ spawn_npc   │ │ move_scene│ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └───────────┘ │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌───────────┐ │
│  │ set_cond    │ │ start_combat│ │ end_combat  │ │ advance   │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └───────────┘ │
└───────────────────────────┬─────────────────────────────────────┘
                            │ Tool Execution
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Game State                                   │
│  - Character: HP, Inventory, Conditions                        │
│  - World: Scene, NPCs, Items                                   │
│  - Combat: Initiative, Turn Order                              │
└─────────────────────────────────────────────────────────────────┘
```

### Tool 接口定义

```go
// internal/game/tools/types.go

// Tool 工具接口
type Tool interface {
    // Name 工具名称
    Name() string
    // Description 工具描述 (LLM 可见)
    Description() string
    // Parameters JSON Schema 参数定义
    Parameters() map[string]interface{}
    // Execute 执行工具
    Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
}

// ToolResult 工具执行结果
type ToolResult struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data"`
    Narrative string      `json:"narrative"` // 用于叙述的文本
    Error     string      `json:"error,omitempty"`
}
```

### DM 工具集定义

**分类：骰子和检定**

| 工具名 | 用途 | 参数 |
|--------|------|------|
| `roll_dice` | 投骰子 | `notation: string` (如 "1d20+5") |
| `skill_check` | 技能检定 | `skill: string` (技能名), `dc: int` (难度等级), `advantage: bool` (优势) |
| `saving_throw` | 豁免检定 | `ability: string` (属性), `dc: int` (难度等级) |
| `attack_roll` | 攻击检定 | `target: string` (目标), `advantage: bool` (优势) |

**分类：角色状态**

| 工具名 | 用途 | 参数 |
|--------|------|------|
| `deal_damage` | 造成伤害 | `target: string` (目标), `amount: int` (数值), `type: string` (伤害类型) |
| `heal_character` | 恢复生命值 | `target: string` (目标), `amount: int` (数值) |
| `add_condition` | 添加状态 | `target: string` (目标), `condition: string` (状态名), `duration: int` (持续时间) |
| `remove_condition` | 移除状态 | `target: string` (目标), `condition: string` (状态名) |
| `modify_hp` | 修改生命值 | `target: string` (目标), `current: int` (当前值), `temp: int` (临时值) |

**分类：物品和装备**

| 工具名 | 用途 | 参数 |
|--------|------|------|
| `add_item` | 获得物品 | `item_id: string` (物品ID), `quantity: int` (数量) |
| `remove_item` | 失去物品 | `item_id: string` (物品ID), `quantity: int` (数量) |
| `equip_item` | 装备物品 | `item_id: string` (物品ID), `slot: string` (装备栏位) |
| `spend_gold` | 花费金币 | `amount: int` (数量) |
| `gain_gold` | 获得金币 | `amount: int` (数量) |

**分类：世界和场景**

| 工具名 | 用途 | 参数 |
|--------|------|------|
| `spawn_npc` | 生成NPC | `npc_id: string` (NPC ID), `location: string` (位置) |
| `remove_npc` | 移除NPC | `npc_id: string` (NPC ID) |
| `move_to_scene` | 切换场景 | `scene_id: string` (场景ID) |
| `add_exit` | 添加出口 | `direction: string` (方向), `scene_id: string` (目标场景) |
| `set_scene_property` | 设置场景属性 | `key: string` (键), `value: string` (值) |

**分类：战斗系统**

| 工具名 | 用途 | 参数 |
|--------|------|------|
| `start_combat` | 开始战斗 | `enemies: []string` (敌人列表) |
| `end_combat` | 结束战斗 | `victory: bool` (是否胜利) |
| `roll_initiative` | 投先攻 | `participants: []string` (参与者列表) |
| `next_turn` | 下一回合 | - |

**分类：剧情和任务**

| 工具名 | 用途 | 参数 |
|--------|------|------|
| `add_quest` | 添加任务 | `quest_id: string`, `title: string` (标题), `description: string` (描述) |
| `update_quest` | 更新任务 | `quest_id: string`, `progress: int` (进度), `status: string` (状态) |
| `set_flag` | 设置标记 | `key: string` (键), `value: bool` (值) |
| `get_flag` | 获取标记 | `key: string` (键) |

### 工具实现示例

```go
// internal/game/tools/skill_check.go

type SkillCheckTool struct {
    rules *rules.Engine
    state *game.State
}

func (t *SkillCheckTool) Name() string {
    return "skill_check"
}

func (t *SkillCheckTool) Description() string {
    return "为玩家角色进行一次技能检定。当玩家尝试一个可能失败的动作时使用此工具。"
}

func (t *SkillCheckTool) Parameters() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "skill": map[string]interface{}{
                "type":        "string",
                "enum":        []string{"运动", "体操", "手法", "隐匿", "奥秘", "历史", "调查", "自然", "宗教", "驯兽", "洞察", "医药", "察觉", "求生", "欺瞒", "威吓", "表演", "说服"},
                "description": "要检定的技能名称",
            },
            "dc": map[string]interface{}{
                "type":        "integer",
                "minimum":     1,
                "maximum":     30,
                "description": "难度等级 (DC)",
            },
            "advantage": map[string]interface{}{
                "type":        "boolean",
                "description": "是否具有优势",
                "default":     false,
            },
        },
        "required": []string{"skill", "dc"},
    }
}

func (t *SkillCheckTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
    skillName := args["skill"].(string)  // 中文技能名，如 "运动"
    dc := int(args["dc"].(float64))
    advantage := false
    if v, ok := args["advantage"]; ok {
        advantage = v.(bool)
    }

    // 获取角色技能
    skill, ok := t.state.Character.Skills[character.SkillType(skillName)]
    if !ok {
        return nil, fmt.Errorf("未知技能: %s", skillName)
    }

    // 执行检定
    rollType := dice.NormalRoll
    if advantage {
        rollType = dice.AdvantageRoll
    }

    result := t.rules.SkillCheck(t.state.Character, skill, dc, rollType)

    resultText := "失败"
    if result.Success {
        resultText = "成功"
    }

    return &ToolResult{
        Success: result.Success,
        Data: map[string]interface{}{
            "roll":      result.Roll.Total,
            "modifier":  result.Roll.Modifier,
            "dc":        dc,
            "margin":    result.Margin,
            "critical":  result.Critical,
        },
        Narrative: fmt.Sprintf("投骰结果: %d (DC %d) - %s", result.Roll.Total, dc, resultText),
    }, nil
}
```

### LLM 调用流程

```go
// internal/game/engine.go

func (e *Engine) ProcessWithTools(ctx context.Context, playerInput string) tea.Cmd {
    return func() tea.Msg {
        // 1. 构建消息
        messages := e.buildMessages(playerInput)

        // 2. 调用 LLM (带工具定义)
        resp, err := e.llmProvider.GenerateWithTools(ctx, &llm.Request{
            Messages: messages,
            Tools:    e.toolRegistry.GetToolDefinitions(),
        })
        if err != nil {
            return ErrorResponseMsg{Err: err}
        }

        // 3. 处理工具调用
        if resp.ToolCalls != nil {
            for _, call := range resp.ToolCalls {
                result, err := e.toolRegistry.Execute(ctx, call.Name, call.Arguments)
                if err != nil {
                    // 工具执行错误，记录并继续
                    e.logger.Errorf("tool %s failed: %v", call.Name, err)
                    continue
                }

                // 将工具结果添加到消息
                messages = append(messages, llm.Message{
                    Role:    llm.RoleTool,
                    Content: fmt.Sprintf("Tool %s result: %v", call.Name, result),
                })
            }

            // 4. 再次调用 LLM 生成最终叙述
            finalResp, err := e.llmProvider.Generate(ctx, &llm.Request{
                Messages: messages,
            })
            if err != nil {
                return ErrorResponseMsg{Err: err}
            }

            return DMResponseMsg{Content: finalResp.Content}
        }

        // 5. 无工具调用，直接返回叙述
        return DMResponseMsg{Content: resp.Content}
    }
}
```

### 安全性设计

```go
// internal/game/tools/registry.go

type ToolRegistry struct {
    tools      map[string]Tool
    permissions map[string][]string  // 工具 -> 允许的游戏阶段
    state      *game.State
}

// Execute 执行工具 (带权限检查)
func (r *ToolRegistry) Execute(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
    tool, ok := r.tools[name]
    if !ok {
        return nil, fmt.Errorf("unknown tool: %s", name)
    }

    // 权限检查
    if !r.checkPermission(name) {
        return nil, fmt.Errorf("tool %s not allowed in current phase", name)
    }

    // 参数验证
    if err := r.validateArgs(tool, args); err != nil {
        return nil, fmt.Errorf("invalid arguments: %v", err)
    }

    // 执行工具
    return tool.Execute(ctx, args)
}

// checkPermission 检查工具在当前阶段是否可用
func (r *ToolRegistry) checkPermission(name string) bool {
    allowedPhases, ok := r.permissions[name]
    if !ok {
        return true  // 未配置则默认允许
    }

    currentPhase := r.state.Phase
    for _, phase := range allowedPhases {
        if string(currentPhase) == phase {
            return true
        }
    }
    return false
}
```

### 文件结构

```
internal/game/tools/
├── types.go           # Tool 接口定义
├── registry.go        # 工具注册表
├── validator.go       # 参数验证器
├── dice_tools.go      # 骰子和检定工具
├── character_tools.go # 角色状态工具
├── item_tools.go      # 物品和装备工具
├── world_tools.go     # 世界和场景工具
├── combat_tools.go    # 战斗系统工具
└── quest_tools.go     # 剧情和任务工具
```

---

## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                    CLI Layer (cmd/)                             │
│    start.go ──► Game Engine ──► TUI App                        │
└─────────────────────────────────────────────────────────────────┘
                                 │
┌────────────────────────────────▼────────────────────────────────┐
│                   Game Engine (internal/game/)                  │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────────────┐    │
│  │   Engine    │  │    State    │  │   EventDispatcher    │    │
│  │  - Run()    │  │ - Character │  │   - Subscribe()      │    │
│  │  - Process()│  │ - Scene     │  │   - Publish()        │    │
│  └─────────────┘  └─────────────┘  └──────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
         │               │                   │
         ▼               ▼                   ▼
┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐
│ llm/prompt  │  │   rules/    │  │       save/             │
│ - Builder   │  │ - Engine    │  │ - Manager               │
│ - Templates │  │ - Checks    │  │ - JSONStore             │
└─────────────┘  └─────────────┘  └─────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    World System (internal/world/)               │
│  - Scene    - NPC    - Location    - Manager                    │
└─────────────────────────────────────────────────────────────────┘
```

---

## 文件清单

### Phase 2.0: D&D 5e 种族/职业数据扩展

| 文件路径 | 用途 | 状态 |
|----------|------|------|
| `internal/character/race.go` | 扩展 Race 结构，添加 SubRace、AgeRange 等字段 | 修改 |
| `internal/character/class.go` | 扩展 Class 结构，添加 SubClass、SpellSlotProgression | 修改 |
| `internal/character/race_data.go` | **新建** - 完整 9 个种族数据 (含子种族) | 新建 |
| `internal/character/class_data.go` | **新建** - 完整 12 个职业数据 (含子职业) | 新建 |
| `internal/character/spell_slots.go` | **新建** - 法术槽成长表 | 新建 |

### Phase 2.1: 核心基础设施

| 文件路径 | 用途 |
|----------|------|
| `internal/game/types.go` | 游戏类型定义 (EventType, Phase, Action) |
| `internal/game/state.go` | GameState 结构和状态管理 |
| `internal/game/engine.go` | 游戏引擎主结构 |
| `internal/game/events.go` | 事件系统 (EventBus, EventHandler) |
| `internal/game/commands.go` | Bubbletea 命令定义 |
| `internal/save/types.go` | 存档数据类型 (SaveData, SaveSlot) |
| `internal/save/manager.go` | 存档管理器 |
| `internal/save/json_store.go` | JSON 文件存储实现 |
| `internal/rules/engine.go` | 规则引擎主结构 |
| `internal/rules/checks.go` | 技能检定逻辑 |
| `internal/rules/dc.go` | DC 难度常量 |
| `internal/llm/prompt/builder.go` | 提示词构建器 |
| `internal/llm/prompt/templates.go` | 中文提示词模板 |
| `internal/llm/prompt/context.go` | 游戏上下文构建 |
| `internal/world/scene.go` | 场景数据结构 |
| `internal/world/npc.go` | NPC 数据结构 |
| `internal/world/manager.go` | 世界管理器 |

### Phase 2.2: Tool Call 机制

| 文件路径 | 用途 |
|----------|------|
| `internal/game/tools/types.go` | Tool 接口和 ToolResult 定义 |
| `internal/game/tools/registry.go` | 工具注册表和执行器 |
| `internal/game/tools/validator.go` | 参数验证器 |
| `internal/game/tools/dice_tools.go` | roll_dice, skill_check, saving_throw, attack_roll |
| `internal/game/tools/character_tools.go` | deal_damage, heal_character, add_condition, remove_condition |
| `internal/game/tools/item_tools.go` | add_item, remove_item, equip_item, spend_gold, gain_gold |
| `internal/game/tools/world_tools.go` | spawn_npc, remove_npc, move_to_scene, set_scene_property |
| `internal/game/tools/combat_tools.go` | start_combat, end_combat, roll_initiative, next_turn |
| `internal/game/tools/quest_tools.go` | add_quest, update_quest, set_flag, get_flag |
| `internal/llm/tool_provider.go` | **修改** - 扩展 Provider 接口支持 Tool Call |

### Phase 2.3: 角色创建 TUI

| 文件路径 | 用途 |
|----------|------|
| `internal/ui/creation/wizard.go` | 创建向导主模型 |
| `internal/ui/creation/styles.go` | 向导样式定义 |
| `internal/ui/creation/step_race.go` | 种族选择步骤 (含子种族) |
| `internal/ui/creation/step_class.go` | 职业选择步骤 (含子职业) |
| `internal/ui/creation/step_ability.go` | Point Buy 属性分配 |
| `internal/ui/creation/step_skill.go` | 技能选择步骤 |
| `internal/ui/creation/step_background.go` | 背景故事输入 |
| `internal/ui/creation/step_confirm.go` | 确认创建步骤 |

### Phase 2.4: 游戏主界面集成

| 文件路径 | 用途 |
|----------|------|
| `internal/ui/game.go` | 游戏主界面模型 |
| `internal/ui/stream.go` | 流式响应渲染 |
| `cmd/start.go` | **修改** - 集成游戏引擎启动 |
| `cmd/character.go` | **修改** - 添加 `create` 子命令 |
| `cmd/load.go` | **修改** - 实现存档加载逻辑 |
| `internal/ui/app.go` | **修改** - 扩展支持游戏状态 |

---

## 核心类型定义

### 1. GameState (`internal/game/state.go`)

```go
type GamePhase int

const (
    PhaseCharacterCreation GamePhase = iota
    PhaseIntroduction
    PhaseExploration
    PhaseDialogue
    PhaseCombat
    PhaseGameOver
)

type State struct {
    mu            sync.RWMutex
    SessionID     string               `json:"session_id"`
    Phase         GamePhase            `json:"phase"`
    TurnCount     int                  `json:"turn_count"`
    Character     *character.Character `json:"character"`
    CurrentScene  *world.Scene         `json:"current_scene"`
    History       []llm.Message        `json:"history"`
    CreatedAt     time.Time            `json:"created_at"`
    LastSavedAt   time.Time            `json:"last_saved_at"`
}
```

### 2. Engine (`internal/game/engine.go`)

```go
type Engine struct {
    state       *State
    llmProvider llm.Provider
    prompt      *prompt.Builder
    rules       *rules.Engine
    world       *world.Manager
    save        *save.Manager
    events      *EventDispatcher
    config      *config.Config
}

func (e *Engine) Start(character *character.Character) tea.Cmd
func (e *Engine) ProcessPlayerAction(ctx context.Context, action string) tea.Cmd
func (e *Engine) RollDice(notation string) (*dice.Result, error)
func (e *Engine) Save(slot int) error
func (e *Engine) Load(slot int) error
```

### 3. CheckResult (`internal/rules/checks.go`)

```go
type CheckType int

const (
    CheckAbility CheckType = iota
    CheckSkill
    CheckSavingThrow
    CheckAttack
)

type CheckRequest struct {
    Type      CheckType
    Ability   character.Ability
    Skill     character.SkillType
    DC        int
    RollType  dice.RollType
    Character *character.Character
}

type CheckResult struct {
    Request  CheckRequest
    Roll     dice.Result
    Total    int
    Success  bool
    Margin   int
    Critical CriticalType
}
```

### 4. Prompt Builder (`internal/llm/prompt/builder.go`)

```go
type Builder struct {
    language  string
    templates Templates
}

func (b *Builder) BuildSystemPrompt(ctx *GameContext) string
func (b *Builder) BuildSceneContext(scene *world.Scene) string
func (b *Builder) BuildCharacterContext(c *character.Character) string
func (b *Builder) BuildHistoryContext(history []llm.Message, maxTurns int) []llm.Message
```

### 5. Save Manager (`internal/save/manager.go`)

```go
type Manager struct {
    saveDir string  // ~/.cdnd/saves/
    cache   map[int]*SaveData
}

func (m *Manager) Save(slot int, data *SaveData) error
func (m *Manager) Load(slot int) (*SaveData, error)
func (m *Manager) ListSlots() ([]SaveSlot, error)
func (m *Manager) Delete(slot int) error
```

---

## 角色创建向导流程

```
┌─────────────────────────────────────────────────────┐
│  角色创建向导                            [步骤 1/6] │
├─────────────────────────────────────────────────────┤
│                                                     │
│  选择种族                                           │
│  ────────────                                       │
│                                                     │
│  > ◉ 人类            [全属性+1]                    │
│    ○ 精灵            [敏捷+2, 黑暗视觉]            │
│    ○ 矮人            [体质+2, 矮人韧性]            │
│    ○ 半身人          [敏捷+2, 幸运]                │
│                                                     │
├─────────────────────────────────────────────────────┤
│  描述: 人类多才多艺，适应性强...                   │
├─────────────────────────────────────────────────────┤
│  [Enter 确认] [↑↓ 选择] [Esc 返回]                 │
└─────────────────────────────────────────────────────┘
```

**步骤顺序：**
1. **种族选择** - 展示 9 个种族，选择子种族（如有）
2. **职业选择** - 展示 12 个职业，显示主要特性
3. **属性分配** - 点数购买 (27点)
4. **技能选择** - 基于职业选择技能熟练
5. **背景故事** - 输入角色名称、背景、阵营
6. **确认创建** - 展示角色卡片，确认保存

---

## LLM 对话循环设计

### 消息流

```
玩家输入 ──► InputModel ──► PlayerActionMsg
                                    │
                                    ▼
                              Engine.ProcessAction()
                                    │
                    ┌───────────────┼───────────────┐
                    ▼               ▼               ▼
               规则检定      构建Prompt       世界更新
                    │               │               │
                    └───────────────┼───────────────┘
                                    ▼
                          LLM GenerateStream()
                                    │
                                    ▼
                          StreamChunkMsg ──► NarrativeModel
```

### 提示词模板结构

```yaml
system_prompt: |
  你是一位经验丰富的《龙与地下城》地下城主(DM)。

  ## 你的职责
  1. 叙述游戏场景和剧情
  2. 扮演所有 NPC（非玩家角色）
  3. 裁决玩家的行动和技能检定

  ## 输出格式
  - 叙述文字使用普通文本
  - 技能检定: 调用 skill_check 工具
  - NPC对话: "角色名: 对话内容"
  - 战斗: 调用 attack_roll 和 deal_damage 工具

  ## 难度等级参考
  - 非常简单: DC 5
  - 简单: DC 10
  - 中等: DC 15
  - 困难: DC 20
  - 非常困难: DC 25
  - 几乎不可能: DC 30

character_context: |
  角色: {name}
  种族: {race} | 职业: {class} | 等级: {level}
  生命值: {hp}/{max_hp} | 护甲等级: {ac}
  属性: 力量{str} 敏捷{dex} 体质{con} 智力{int} 感知{wis} 魅力{cha}
  熟练技能: {proficient_skills}
```

### 流式响应处理

```go
// Bubbletea 命令
func StreamDMResponseCmd(e *Engine, messages []llm.Message) tea.Cmd {
    return func() tea.Msg {
        stream, err := e.llmProvider.GenerateStream(ctx, &llm.Request{
            Messages: messages,
            Stream:   true,
        })
        if err != nil {
            return StreamErrorMsg{Err: err}
        }

        for chunk := range stream {
            if chunk.Error != nil {
                return StreamErrorMsg{Err: chunk.Error}
            }
            if chunk.Done {
                return StreamEndMsg{FullContent: fullContent}
            }
            // 发送增量更新
            return StreamChunkMsg{Content: chunk.Content}
        }
        return nil
    }
}
```

---

## 实现步骤

### Step 0: D&D 5e 种族/职业数据扩展 (优先级最高)

1. **扩展种族结构**
   - 修改 `internal/character/race.go`：添加 SubRace、AgeRange、HeightRange、WeightRange 字段
   - 创建 `internal/character/race_data.go`：定义完整 9 个种族数据
   - 包含子种族：High Elf/Wood Elf/Drow, Hill Dwarf/Mountain Dwarf, Lightfoot Halfling/Stout Halfling, Forest Gnome/Rock Gnome

2. **扩展职业结构**
   - 修改 `internal/character/class.go`：添加 SubClass、SpellSlotProgression 字段
   - 创建 `internal/character/class_data.go`：定义完整 12 个职业数据
   - 创建 `internal/character/spell_slots.go`：法术槽成长表 (1-20级)

### Step 1: 核心类型定义

1. 创建 `internal/game/types.go` - 定义 EventType, GamePhase, ActionType
2. 创建 `internal/game/state.go` - GameState 结构
3. 创建 `internal/game/events.go` - EventBus 实现

### Step 2: 存档系统

1. 创建 `internal/save/types.go` - SaveData, SaveSlot
2. 创建 `internal/save/json_store.go` - JSON 文件读写
3. 创建 `internal/save/manager.go` - 存档管理器

### Step 3: 规则引擎

1. 创建 `internal/rules/engine.go` - 规则引擎主结构
2. 创建 `internal/rules/checks.go` - SkillCheck, SavingThrow 实现
3. 创建 `internal/rules/dc.go` - DC 难度常量

### Step 4: Tool Call 机制

1. 创建 `internal/game/tools/types.go` - Tool 接口定义
2. 创建 `internal/game/tools/registry.go` - 工具注册表
3. 创建 `internal/game/tools/validator.go` - 参数验证
4. 创建各分类工具文件 (dice_tools.go, character_tools.go, 等)
5. 扩展 `internal/llm/provider.go` - 添加 GenerateWithTools 方法
6. 修改 `internal/llm/openai.go` - 实现 OpenAI Tool Call 支持

### Step 5: 提示词系统

1. 创建 `internal/llm/prompt/templates.go` - 中文模板
2. 创建 `internal/llm/prompt/builder.go` - 提示词构建器
3. 创建 `internal/llm/prompt/context.go` - 上下文构建

### Step 6: 世界系统

1. 创建 `internal/world/scene.go` - Scene 结构
2. 创建 `internal/world/npc.go` - NPC 结构
3. 创建 `internal/world/manager.go` - 世界管理器

### Step 7: 游戏引擎

1. 创建 `internal/game/engine.go` - Engine 主结构
2. 创建 `internal/game/commands.go` - Bubbletea 命令
3. 创建 `internal/ui/game.go` - 游戏主界面
4. 创建 `internal/ui/stream.go` - 流式响应处理

### Step 8: 角色创建 TUI

1. 创建 `internal/ui/creation/wizard.go` - 向导框架
2. 创建各步骤文件 (step_*.go)
3. 修改 `cmd/character.go` - 集成创建命令

### Step 9: CLI 集成

1. 修改 `cmd/start.go` - 启动游戏
2. 修改 `cmd/load.go` - 加载存档

---

## 验证方案

### 1. 编译验证

```bash
go build -o cdnd .
./cdnd --help
```

### 2. 种族/职业数据验证

```bash
# 验证种族数据加载
./cdnd character create
# 检查种族列表是否包含 9 个种族
# 检查子种族是否正确显示

# 验证职业数据加载
# 检查职业列表是否包含 12 个职业
# 检查子职业是否在 3 级时可选
```

### 3. 角色创建验证

```bash
./cdnd character create
# 依次验证：
# - 种族选择界面显示子种族选项
# - 职业选择界面显示子职业选项
# - Point Buy 点数计算正确 (27点)
# - 技能选择限制正确 (基于职业)
# - 法术施法者获得正确的法术槽
# - 角色保存到 ~/.cdnd/characters/
```

### 4. Tool Call 验证

```bash
# 启动游戏
./cdnd start

# 测试技能检定工具
# 玩家输入: "我试图攀爬城墙"
# 验证: LLM 调用 skill_check 工具
# 验证: 骰子结果正确显示

# 测试伤害工具
# 玩家输入: "我攻击哥布林"
# 验证: LLM 调用 attack_roll 和 deal_damage 工具
# 验证: 目标 HP 正确减少
```

### 5. 游戏启动验证

```bash
./cdnd start
# 验证：
# - TUI 正常渲染
# - Header 显示角色信息
# - 输入区域可输入
# - LLM 流式响应正常
# - Tool Call 正确执行
```

### 6. 存档验证

```bash
# 游戏中按 Ctrl+S 保存
# 检查 ~/.cdnd/saves/slot_1.json 存在
./cdnd load --slot 1
# 验证游戏状态恢复正确
# 验证角色属性、物品、位置正确
```

### 7. 单元测试

```bash
# 运行规则引擎测试
go test ./internal/rules/... -v

# 运行工具测试
go test ./internal/game/tools/... -v

# 运行所有测试
go test ./... -v
```

---

## 关键文件路径

**最关键的 10 个文件：**

1. `internal/character/race_data.go` - 完整种族数据，角色创建基础
2. `internal/character/class_data.go` - 完整职业数据，角色创建基础
3. `internal/game/tools/registry.go` - Tool Call 核心，LLM 与游戏交互的桥梁
4. `internal/game/tools/dice_tools.go` - 骰子检定工具，最常用的工具
5. `internal/game/engine.go` - 游戏引擎核心，整合所有模块
6. `internal/ui/creation/wizard.go` - 角色创建向导，TUI 交互核心
7. `internal/llm/prompt/builder.go` - 提示词构建，LLM 对话质量关键
8. `internal/save/manager.go` - 存档管理，游戏持久化核心
9. `internal/ui/game.go` - 游戏主界面，连接引擎和 TUI
10. `internal/llm/openai.go` - OpenAI Tool Call 实现

---

## 风险和注意事项

1. **Tool Call 兼容性**：不同 LLM 提供商的 Tool Call 格式可能不同，需要抽象适配层
2. **流式响应超时**：LLM API 可能超时，需要设置合理的 context timeout
3. **内存管理**：历史对话需要限制长度，避免 token 超限
4. **并发安全**：GameState 的读写需要加锁保护
5. **错误处理**：LLM 调用失败需要优雅降级，不影响游戏状态
6. **编码问题**：JSON 存档需要正确处理 UTF-8 编码
7. **D&D 规则准确性**：种族/职业数据必须严格对照官方规则书验证
8. **工具权限控制**：某些工具只在特定游戏阶段可用，防止 LLM 滥用
