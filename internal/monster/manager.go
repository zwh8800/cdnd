package monster

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/zwh8800/cdnd/internal/character"
	"github.com/zwh8800/cdnd/internal/combat"
	"github.com/zwh8800/cdnd/pkg/dice"
)

// Manager 怪物管理器
type Manager struct {
	rng *rand.Rand
}

// NewManager 创建怪物管理器
func NewManager() *Manager {
	return &Manager{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SpawnFromTemplate 从模板生成战斗参与者
func (m *Manager) SpawnFromTemplate(templateID string, nameOverride string) (*combat.Combatant, error) {
	template, ok := GetTemplate(templateID)
	if !ok {
		return nil, fmt.Errorf("怪物模板不存在: %s", templateID)
	}

	// 生成唯一ID
	id := uuid.New().String()

	// 确定名称
	name := template.Name
	if nameOverride != "" {
		name = nameOverride
	}

	// 投生命值
	maxHP := m.rollHP(template.HP)

	// 计算先攻（敏捷调整值）
	initiative := template.Abilities.Modifier(character.Dexterity)

	// 构建豁免加值映射
	savingThrows := make(map[string]int)
	if template.SavingThrows != nil {
		for ability, bonus := range template.SavingThrows {
			savingThrows[ability] = bonus
		}
	}
	// 为未指定的豁免使用属性调整值
	for _, ability := range []character.Ability{
		character.Strength, character.Dexterity, character.Constitution,
		character.Intelligence, character.Wisdom, character.Charisma,
	} {
		abilityName := string(ability)
		if _, ok := savingThrows[abilityName]; !ok {
			savingThrows[abilityName] = template.Abilities.Modifier(ability)
		}
	}

	combatant := &combat.Combatant{
		ID:           id,
		Name:         name,
		IsPlayer:     false,
		IsNPC:        false, // 标记为怪物而非NPC
		HP:           maxHP,
		MaxHP:        maxHP,
		AC:           template.AC,
		Initiative:   initiative,
		Conditions:   []string{},
		Abilities:    template.Abilities,
		SavingThrows: savingThrows,
	}

	return combatant, nil
}

// SpawnMultiple 生成多个相同类型的怪物
func (m *Manager) SpawnMultiple(templateID string, count int) ([]*combat.Combatant, error) {
	var combatants []*combat.Combatant

	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s %d", GetTemplateName(templateID), i+1)
		combatant, err := m.SpawnFromTemplate(templateID, name)
		if err != nil {
			return nil, err
		}
		combatants = append(combatants, combatant)
	}

	return combatants, nil
}

// SpawnGroup 生成混合怪物群
func (m *Manager) SpawnGroup(group map[string]int) ([]*combat.Combatant, error) {
	var combatants []*combat.Combatant

	for templateID, count := range group {
		groupCombatants, err := m.SpawnMultiple(templateID, count)
		if err != nil {
			return nil, err
		}
		combatants = append(combatants, groupCombatants...)
	}

	return combatants, nil
}

// rollHP 投生命值
func (m *Manager) rollHP(expression string) int {
	result, err := dice.ParseAndRoll(expression)
	if err != nil {
		// 如果解析失败，使用默认值
		return 10
	}
	return result.Total
}

// GetTemplateName 获取模板名称（辅助函数）
func GetTemplateName(templateID string) string {
	template, ok := GetTemplate(templateID)
	if !ok {
		return "未知怪物"
	}
	return template.Name
}

// GetMonsterXP 获取怪物的经验值
func GetMonsterXP(templateID string) int {
	template, ok := GetTemplate(templateID)
	if !ok {
		return 0
	}
	return template.XP
}

// CalculateEncounterXP 计算遭遇战总经验值
func CalculateEncounterXP(templateIDs []string) int {
	total := 0
	for _, id := range templateIDs {
		total += GetMonsterXP(id)
	}
	return total
}

// GetRecommendedEncounter 根据玩家等级推荐遭遇战
func (m *Manager) GetRecommendedEncounter(partyLevel int, partySize int) map[string]int {
	// 简单的遭遇战生成逻辑
	// 根据队伍等级和规模推荐合适的怪物组合

	encounter := make(map[string]int)

	switch {
	case partyLevel <= 2:
		// 低等级：哥布林、狗头人、巨鼠
		encounter["goblin"] = m.rng.Intn(3) + 1
		if m.rng.Float32() < 0.3 {
			encounter["giant_rat"] = m.rng.Intn(2) + 1
		}

	case partyLevel <= 4:
		// 中等级：骷髅、僵尸、兽人
		encounter["skeleton"] = m.rng.Intn(3) + 2
		if m.rng.Float32() < 0.4 {
			encounter["zombie"] = 1
		}

	case partyLevel <= 6:
		// 较高等级：食人魔、食尸鬼
		encounter["ghoul"] = m.rng.Intn(2) + 1
		encounter["skeleton_warrior"] = m.rng.Intn(2) + 1

	default:
		// 高等级：巨魔、幽灵
		encounter["troll"] = 1
		encounter["goblin"] = m.rng.Intn(3) + 2
	}

	return encounter
}

// GetMonsterAction 获取怪物的默认攻击动作
func GetMonsterAction(templateID string) (*MonsterAction, error) {
	template, ok := GetTemplate(templateID)
	if !ok {
		return nil, fmt.Errorf("怪物模板不存在: %s", templateID)
	}

	if len(template.Actions) == 0 {
		return nil, fmt.Errorf("怪物没有定义动作: %s", templateID)
	}

	// 返回第一个动作作为默认动作
	return &template.Actions[0], nil
}

// GetMonsterActionByName 获取指定名称的怪物动作
func GetMonsterActionByName(templateID, actionName string) (*MonsterAction, error) {
	template, ok := GetTemplate(templateID)
	if !ok {
		return nil, fmt.Errorf("怪物模板不存在: %s", templateID)
	}

	for _, action := range template.Actions {
		if action.Name == actionName {
			return &action, nil
		}
	}

	return nil, fmt.Errorf("怪物 %s 没有动作: %s", templateID, actionName)
}

// SelectRandomAction 随机选择怪物的动作（用于AI决策）
func (m *Manager) SelectRandomAction(templateID string) (*MonsterAction, error) {
	template, ok := GetTemplate(templateID)
	if !ok {
		return nil, fmt.Errorf("怪物模板不存在: %s", templateID)
	}

	if len(template.Actions) == 0 {
		return nil, fmt.Errorf("怪物没有定义动作: %s", templateID)
	}

	// 随机选择一个动作
	index := m.rng.Intn(len(template.Actions))
	return &template.Actions[index], nil
}
