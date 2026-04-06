package tools

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/zwh8800/cdnd/domain/character"
	"github.com/zwh8800/cdnd/domain/combat"
	dice2 "github.com/zwh8800/cdnd/domain/dice"
	"github.com/zwh8800/cdnd/domain/monster"
)

// StartCombatTool 开始战斗工具
type StartCombatTool struct {
	BaseTool
	state StateAccessor
}

// NewStartCombatTool 创建开始战斗工具
func NewStartCombatTool(state StateAccessor) *StartCombatTool {
	return &StartCombatTool{
		BaseTool: NewBaseTool("start_combat", "开始一场战斗。参数: enemies (敌人列表，包含monster_id和可选name_override)"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *StartCombatTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"enemies": map[string]interface{}{
				"type":        "array",
				"description": "敌人列表",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"monster_id": map[string]interface{}{
							"type":        "string",
							"description": "怪物模板ID",
						},
						"name_override": map[string]interface{}{
							"type":        "string",
							"description": "自定义名称（可选）",
						},
					},
					"required": []string{"monster_id"},
				},
			},
		},
		"required": []string{"enemies"},
	}
}

// Execute 执行开始战斗
func (t *StartCombatTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if t.state == nil || t.state.GetCharacter() == nil {
		return nil, ErrStateNotAvailable
	}

	// 检查是否已在战斗中
	if t.state.GetCombat() != nil && t.state.GetCombat().Active {
		return &ToolResult{
			Success:   false,
			Error:     "战斗已经在进行中",
			Narrative: "⚠️ 战斗已经在进行中！",
		}, nil
	}

	enemiesData, ok := args["enemies"].([]interface{})
	if !ok || len(enemiesData) == 0 {
		return nil, ErrInvalidArguments
	}

	// 创建怪物管理器
	monsterMgr := monster.NewManager()

	var participants []*combat.Combatant
	var enemyNames []string

	// 生成敌人
	for _, enemyData := range enemiesData {
		enemyMap, ok := enemyData.(map[string]interface{})
		if !ok {
			continue
		}

		monsterID, ok := enemyMap["monster_id"].(string)
		if !ok {
			continue
		}

		nameOverride := ""
		if name, ok := enemyMap["name_override"].(string); ok {
			nameOverride = name
		}

		combatant, err := monsterMgr.SpawnFromTemplate(monsterID, nameOverride)
		if err != nil {
			return &ToolResult{
				Success:   false,
				Error:     fmt.Sprintf("生成怪物失败: %v", err),
				Narrative: fmt.Sprintf("❌ 无法生成怪物: %s", monsterID),
			}, nil
		}

		// 投先攻
		initiativeRoll, _ := dice2.ParseAndRoll("1d20")
		combatant.Initiative = initiativeRoll.Total + combatant.Initiative

		participants = append(participants, combatant)
		enemyNames = append(enemyNames, combatant.Name)
	}

	// 创建玩家战斗参与者
	char := t.state.GetCharacter()
	playerCombatant := &combat.Combatant{
		ID:           "player",
		Name:         char.Name,
		IsPlayer:     true,
		IsNPC:        false,
		HP:           char.HitPoints.Current,
		MaxHP:        char.HitPoints.Max,
		AC:           char.ArmorClass,
		Initiative:   char.Attributes.Modifier(character.Dexterity),
		Conditions:   char.Conditions,
		Abilities:    char.Attributes,
		SavingThrows: make(map[string]int),
	}

	// 计算玩家豁免加值
	for _, ability := range []character.Ability{
		character.Strength, character.Dexterity, character.Constitution,
		character.Intelligence, character.Wisdom, character.Charisma,
	} {
		modifier := char.Attributes.Modifier(ability)
		// 检查是否熟练
		if savingThrow, ok := char.SavingThrows[ability]; ok && savingThrow.Proficient {
			modifier += char.ProficiencyBonus
		}
		playerCombatant.SavingThrows[string(ability)] = modifier
	}

	// 投玩家先攻
	playerInitiativeRoll, _ := dice2.ParseAndRoll("1d20")
	playerCombatant.Initiative = playerInitiativeRoll.Total + playerCombatant.Initiative

	participants = append(participants, playerCombatant)

	// 开始战斗
	t.state.StartCombat(participants)

	// 构建先攻顺序描述
	initiativeOrder := t.state.GetCombat().Initiative
	initiativeDesc := "\n  ├─ 先攻顺序:"
	for i, entry := range initiativeOrder {
		combatant := t.state.GetCombatant(entry.EntityID)
		if combatant != nil {
			marker := ""
			if i == 0 {
				marker = " ← 先攻"
			}
			initiativeDesc += fmt.Sprintf("\n  │  %d. %s (%d)%s", i+1, combatant.Name, entry.Initiative, marker)
		}
	}

	narrative := fmt.Sprintf("⚔️ 战斗开始！\n  ├─ 敌人: %v%s\n  └─ 回合: 1", enemyNames, initiativeDesc)

	return &ToolResult{
		Success:   true,
		Narrative: narrative,
		Data: map[string]interface{}{
			"enemies":           enemyNames,
			"participant_count": len(participants),
			"initiative_order":  initiativeOrder,
			"round":             1,
		},
	}, nil
}

// AttackTool 攻击工具
type AttackTool struct {
	BaseTool
	state StateAccessor
}

// NewAttackTool 创建攻击工具
func NewAttackTool(state StateAccessor) *AttackTool {
	return &AttackTool{
		BaseTool: NewBaseTool("attack", "进行攻击检定。参数: attacker (攻击者ID), target (目标ID), attack_type (攻击类型), advantage (是否优势), disadvantage (是否劣势)"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *AttackTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"attacker": map[string]interface{}{
				"type":        "string",
				"description": "攻击者ID（player 或 敌人ID）",
			},
			"target": map[string]interface{}{
				"type":        "string",
				"description": "目标ID（player 或 敌人ID）",
			},
			"attack_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"melee", "ranged", "spell"},
				"description": "攻击类型: melee(近战), ranged(远程), spell(法术)",
			},
			"advantage": map[string]interface{}{
				"type":        "boolean",
				"description": "是否优势（可选，默认false）",
				"default":     false,
			},
			"disadvantage": map[string]interface{}{
				"type":        "boolean",
				"description": "是否劣势（可选，默认false）",
				"default":     false,
			},
		},
		"required": []string{"attacker", "target", "attack_type"},
	}
}

// Execute 执行攻击
func (t *AttackTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if t.state == nil || t.state.GetCombat() == nil {
		return nil, ErrStateNotAvailable
	}

	attackerID, _ := args["attacker"].(string)
	targetID, _ := args["target"].(string)
	attackType, _ := args["attack_type"].(string)
	advantage, _ := args["advantage"].(bool)
	disadvantage, _ := args["disadvantage"].(bool)

	// 获取攻击者和目标
	var attacker, target *combat.Combatant
	if attackerID == "player" {
		for _, p := range t.state.GetCombat().Participants {
			if p.IsPlayer {
				attacker = p
				break
			}
		}
	} else {
		attacker = t.state.GetCombatant(attackerID)
	}

	if targetID == "player" {
		for _, p := range t.state.GetCombat().Participants {
			if p.IsPlayer {
				target = p
				break
			}
		}
	} else {
		target = t.state.GetCombatant(targetID)
	}

	if attacker == nil || target == nil {
		return &ToolResult{
			Success:   false,
			Error:     "攻击者或目标不存在",
			Narrative: "❌ 攻击者或目标不存在！",
		}, nil
	}

	// 获取攻击加值
	attackBonus := 0
	if attacker.IsPlayer {
		// 玩家：根据攻击类型使用力量或敏捷
		if attackType == "ranged" {
			attackBonus = attacker.Abilities.Modifier(character.Dexterity)
		} else {
			attackBonus = attacker.Abilities.Modifier(character.Strength)
		}
		// 假设玩家熟练武器（简化处理）
		attackBonus += t.state.GetCharacter().ProficiencyBonus
	} else {
		// 敌人：使用怪物的攻击加值（简化处理，使用力量调整值+2）
		attackBonus = attacker.Abilities.Modifier(character.Strength) + 2
	}

	// 投攻击骰
	rollType := dice2.NormalRoll
	if advantage {
		rollType = dice2.AdvantageRoll
	} else if disadvantage {
		rollType = dice2.DisadvantageRoll
	}
	roll := dice2.RollDice(1, 20, attackBonus, rollType)

	// 判断是否命中
	isHit := roll.Total >= target.AC
	isCritical := roll.Critical == dice2.CritSuccess
	isFumble := roll.Critical == dice2.CritFail

	// 大成功自动命中，大失败自动未命中
	if isCritical {
		isHit = true
	}
	if isFumble {
		isHit = false
	}

	// 构建结果叙述
	narrative := fmt.Sprintf("⚔️ %s 攻击 %s\n", attacker.Name, target.Name)
	narrative += fmt.Sprintf("  ├─ 攻击检定: %d (1d20%+d vs AC %d)", roll.Total, attackBonus, target.AC)

	if isCritical {
		narrative += " 💥 暴击！\n"
	} else if isFumble {
		narrative += " ❌ 大失败！\n"
	} else if isHit {
		narrative += " ✅ 命中！\n"
	} else {
		narrative += " ❌ 未命中\n"
	}

	// 如果命中，计算伤害
	if isHit || isCritical {
		damage := 0
		damageType := "挥砍"

		if attacker.IsPlayer {
			// 玩家伤害：1d8 + 力量/敏捷调整值
			weapon := "1d8"
			if attackType == "ranged" {
				damageType = "穿刺"
			}
			damageRoll, _ := dice2.ParseAndRoll(weapon)
			damage = damageRoll.Total
			if attackType == "ranged" {
				damage += attacker.Abilities.Modifier(character.Dexterity)
			} else {
				damage += attacker.Abilities.Modifier(character.Strength)
			}
		} else {
			// 敌人伤害：简化处理，1d6 + 力量调整值
			damageRoll, _ := dice2.ParseAndRoll("1d6")
			damage = damageRoll.Total + attacker.Abilities.Modifier(character.Strength)
			if damage < 1 {
				damage = 1
			}
		}

		// 暴击时额外伤害
		if isCritical {
			extraRoll, _ := dice2.ParseAndRoll("1d8")
			damage += extraRoll.Total
			narrative += fmt.Sprintf("  ├─ 暴击伤害: %d 点%s！\n", damage, damageType)
		} else {
			narrative += fmt.Sprintf("  ├─ 造成 %d 点%s伤害\n", damage, damageType)
		}

		// 更新目标HP
		oldHP := target.HP
		target.HP -= damage
		if target.HP < 0 {
			target.HP = 0
		}

		narrative += fmt.Sprintf("  └─ %s 生命值: %d → %d/%d", target.Name, oldHP, target.HP, target.MaxHP)

		// 检查目标是否死亡
		if target.HP == 0 && !target.IsPlayer {
			narrative += " ☠️ 倒地！"
		}

		return &ToolResult{
			Success:   true,
			Narrative: narrative,
			Data: map[string]interface{}{
				"hit":           true,
				"critical":      isCritical,
				"attack_roll":   roll.Total,
				"total":         roll.Total,
				"target_ac":     target.AC,
				"damage":        damage,
				"damage_type":   damageType,
				"target_hp":     target.HP,
				"target_max_hp": target.MaxHP,
			},
		}, nil
	}

	return &ToolResult{
		Success:   isHit,
		Narrative: narrative,
		Data: map[string]interface{}{
			"hit":         isHit,
			"critical":    roll.Critical,
			"attack_roll": roll.Total,
			"total":       roll.Total,
			"target_ac":   target.AC,
		},
	}, nil
}

// NextTurnTool 下一回合工具
type NextTurnTool struct {
	BaseTool
	state StateAccessor
}

// NewNextTurnTool 创建下一回合工具
func NewNextTurnTool(state StateAccessor) *NextTurnTool {
	return &NextTurnTool{
		BaseTool: NewBaseTool("next_turn", "推进到下一回合。参数: 无"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *NextTurnTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

// Execute 执行下一回合
func (t *NextTurnTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if t.state == nil || t.state.GetCombat() == nil {
		return nil, ErrStateNotAvailable
	}

	combat := t.state.GetCombat()
	oldRound := combat.Round

	// 推进回合
	nextCombatant := t.state.NextTurn()

	if nextCombatant == nil {
		return &ToolResult{
			Success:   false,
			Error:     "无法推进回合",
			Narrative: "⚠️ 无法推进回合",
		}, nil
	}

	// 重置玩家行动标记（如果是玩家回合）
	if nextCombatant.IsPlayer {
		combat.PlayerActionUsed = false
		combat.PlayerBonusActionUsed = false
	}

	// 构建叙述
	narrative := "🔄 回合推进\n"
	if combat.Round != oldRound {
		narrative += fmt.Sprintf("  ├─ 第 %d 轮开始\n", combat.Round)
	}
	narrative += fmt.Sprintf("  └─ 现在是 %s 的回合", nextCombatant.Name)

	return &ToolResult{
		Success:   true,
		Narrative: narrative,
		Data: map[string]interface{}{
			"round":             combat.Round,
			"turn":              combat.CurrentTurn,
			"current_combatant": nextCombatant.Name,
			"is_player_turn":    nextCombatant.IsPlayer,
		},
	}, nil
}

// EndCombatTool 结束战斗工具
type EndCombatTool struct {
	BaseTool
	state StateAccessor
}

// NewEndCombatTool 创建结束战斗工具
func NewEndCombatTool(state StateAccessor) *EndCombatTool {
	return &EndCombatTool{
		BaseTool: NewBaseTool("end_combat", "结束战斗。参数: reason (结束原因: victory/defeat/flee)"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *EndCombatTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"reason": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"victory", "defeat", "flee", "negotiate"},
				"description": "结束原因: victory(胜利), defeat(失败), flee(逃跑), negotiate(谈判)",
			},
		},
		"required": []string{"reason"},
	}
}

// Execute 执行结束战斗
func (t *EndCombatTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if t.state == nil || t.state.GetCombat() == nil {
		return nil, ErrStateNotAvailable
	}

	reason, _ := args["reason"].(string)
	combat := t.state.GetCombat()

	// 计算战斗统计
	rounds := combat.Round
	duration := time.Since(combat.StartedAt).Minutes()

	// 计算获得的经验值
	xpAwarded := 0
	if reason == "victory" {
		for _, p := range combat.Participants {
			if !p.IsPlayer && p.HP <= 0 {
				// 简化处理：每个敌人50 XP
				xpAwarded += 50
			}
		}
	}

	// 更新玩家经验值
	if xpAwarded > 0 && t.state.GetCharacter() != nil {
		t.state.GetCharacter().Experience += xpAwarded
	}

	// 构建结果叙述
	resultText := ""
	switch reason {
	case "victory":
		resultText = "🏆 胜利！"
	case "defeat":
		resultText = "💀 战败..."
	case "flee":
		resultText = "🏃 成功逃脱"
	case "negotiate":
		resultText = "🤝 通过谈判结束"
	}

	narrative := fmt.Sprintf("⚔️ 战斗结束 - %s\n", resultText)
	narrative += fmt.Sprintf("  ├─ 持续: %d 轮 (%.1f分钟)\n", rounds, duration)
	if xpAwarded > 0 {
		narrative += fmt.Sprintf("  └─ 获得 %d 经验值", xpAwarded)
	}

	// 结束战斗
	t.state.EndCombat()

	return &ToolResult{
		Success:   true,
		Narrative: narrative,
		Data: map[string]interface{}{
			"reason":    reason,
			"rounds":    rounds,
			"duration":  duration,
			"xp_gained": xpAwarded,
		},
	}, nil
}

// SpawnEnemyTool 生成敌人工具（战斗中召唤增援）
type SpawnEnemyTool struct {
	BaseTool
	state StateAccessor
}

// NewSpawnEnemyTool 创建生成敌人工具
func NewSpawnEnemyTool(state StateAccessor) *SpawnEnemyTool {
	return &SpawnEnemyTool{
		BaseTool: NewBaseTool("spawn_enemy", "在战斗中生成新的敌人。参数: monster_id, name_override (可选)"),
		state:    state,
	}
}

// Parameters 返回参数定义
func (t *SpawnEnemyTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"monster_id": map[string]interface{}{
				"type":        "string",
				"description": "怪物模板ID",
			},
			"name_override": map[string]interface{}{
				"type":        "string",
				"description": "自定义名称（可选）",
			},
		},
		"required": []string{"monster_id"},
	}
}

// Execute 执行生成敌人
func (t *SpawnEnemyTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if t.state == nil || t.state.GetCombat() == nil {
		return nil, ErrStateNotAvailable
	}

	monsterID, _ := args["monster_id"].(string)
	nameOverride := ""
	if name, ok := args["name_override"].(string); ok {
		nameOverride = name
	}

	// 生成敌人
	monsterMgr := monster.NewManager()
	combatant, err := monsterMgr.SpawnFromTemplate(monsterID, nameOverride)
	if err != nil {
		return &ToolResult{
			Success:   false,
			Error:     fmt.Sprintf("生成敌人失败: %v", err),
			Narrative: fmt.Sprintf("❌ 无法生成敌人: %s", monsterID),
		}, nil
	}

	// 投先攻
	initiativeRoll, _ := dice2.ParseAndRoll("1d20")
	combatant.Initiative = initiativeRoll.Total + combatant.Initiative

	// 添加到战斗
	cbt := t.state.GetCombat()
	cbt.Participants = append(cbt.Participants, combatant)

	// 添加到先攻列表并重新排序
	cbt.Initiative = append(cbt.Initiative, combat.InitiativeEntry{
		EntityID:   combatant.ID,
		Initiative: combatant.Initiative,
		HasActed:   false,
		IsPlayer:   false,
	})
	sortInitiative(cbt.Initiative)

	narrative := fmt.Sprintf("👹 %s 加入了战斗！\n  └─ 先攻: %d", combatant.Name, combatant.Initiative)

	return &ToolResult{
		Success:   true,
		Narrative: narrative,
		Data: map[string]interface{}{
			"enemy_name": combatant.Name,
			"enemy_id":   combatant.ID,
			"initiative": combatant.Initiative,
			"hp":         combatant.HP,
			"ac":         combatant.AC,
		},
	}, nil
}

// sortInitiative 按先攻值排序（高到低）
func sortInitiative(entries []combat.InitiativeEntry) {
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Initiative > entries[i].Initiative {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
}

// init 初始化随机种子
func init() {
	rand.Seed(time.Now().UnixNano())
}
