package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zwh8800/cdnd/internal/character"
)

// CreationStep 角色创建步骤
type CreationStep int

const (
	StepName CreationStep = iota
	StepRace
	StepSubrace
	StepClass
	StepSubclass
	StepAbilityScores
	StepSkills
	StepConfirm
)

// String 返回步骤名称
func (s CreationStep) String() string {
	switch s {
	case StepName:
		return "角色名称"
	case StepRace:
		return "选择种族"
	case StepSubrace:
		return "选择子种族"
	case StepClass:
		return "选择职业"
	case StepSubclass:
		return "选择子职业"
	case StepAbilityScores:
		return "分配属性"
	case StepSkills:
		return "选择技能"
	case StepConfirm:
		return "确认角色"
	default:
		return "未知"
	}
}

// CharacterCreationModel 角色创建模型
type CharacterCreationModel struct {
	step      CreationStep
	name      string
	race      *character.Race
	subrace   *character.SubRace
	class     *character.Class
	subclass  *character.SubClass
	abilities character.Attributes
	skills    []character.SkillType

	availableRaces   []*character.Race
	availableClasses []*character.Class
	availableSkills  []character.SkillType
	abilityPoints    int

	width  int
	height int
	cursor int
	ready  bool
	err    string
}

// NewCharacterCreationModel 创建角色创建模型
func NewCharacterCreationModel() CharacterCreationModel {
	return CharacterCreationModel{
		step:             StepName,
		availableRaces:   character.GetAllRaces(),
		availableClasses: character.GetAllClasses(),
		abilityPoints:    72,
		skills:           make([]character.SkillType, 0),
	}
}

// Init 初始化
func (m CharacterCreationModel) Init() tea.Cmd {
	return nil
}

// Update 更新
func (m CharacterCreationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil
	}
	return m, nil
}

// handleKeyPress 处理按键
func (m CharacterCreationModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEnter:
		return m.handleEnter()
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown:
		m.cursor = m.handleDown()
	case tea.KeyBackspace:
		if m.step == StepName && len(m.name) > 0 {
			m.name = m.name[:len(m.name)-1]
		}
	default:
		if msg.Type == tea.KeyRunes && m.step == StepName {
			m.name += string(msg.Runes)
		}
	}
	return m, nil
}

// handleEnter 处理回车
func (m CharacterCreationModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case StepName:
		if m.name == "" {
			m.err = "请输入角色名称"
			return m, nil
		}
		m.step = StepRace
		m.cursor = 0
		m.err = ""

	case StepRace:
		if m.cursor < len(m.availableRaces) {
			m.race = m.availableRaces[m.cursor]
			if m.race.HasSubRaces() {
				m.step = StepSubrace
			} else {
				m.step = StepClass
			}
			m.cursor = 0
		}

	case StepSubrace:
		if m.race != nil && m.cursor < len(m.race.SubRaces) {
			m.subrace = &m.race.SubRaces[m.cursor]
			m.step = StepClass
			m.cursor = 0
		}

	case StepClass:
		if m.cursor < len(m.availableClasses) {
			m.class = m.availableClasses[m.cursor]
			if m.class.HasSubClasses() {
				m.step = StepSubclass
			} else {
				m.step = StepAbilityScores
				m.initAbilityScores()
			}
			m.cursor = 0
		}

	case StepSubclass:
		if m.class != nil && m.cursor < len(m.class.SubClasses) {
			m.subclass = &m.class.SubClasses[m.cursor]
			m.step = StepAbilityScores
			m.initAbilityScores()
			m.cursor = 0
		}

	case StepAbilityScores:
		m.step = StepSkills
		m.initAvailableSkills()
		m.cursor = 0

	case StepSkills:
		m.step = StepConfirm

	case StepConfirm:
		return m, tea.Quit
	}
	return m, nil
}

// handleDown 处理向下键
func (m CharacterCreationModel) handleDown() int {
	var maxCursor int
	switch m.step {
	case StepRace:
		maxCursor = len(m.availableRaces) - 1
	case StepSubrace:
		if m.race != nil {
			maxCursor = len(m.race.SubRaces) - 1
		}
	case StepClass:
		maxCursor = len(m.availableClasses) - 1
	case StepSubclass:
		if m.class != nil {
			maxCursor = len(m.class.SubClasses) - 1
		}
	case StepSkills:
		maxCursor = len(m.availableSkills) - 1
	default:
		return m.cursor
	}
	if m.cursor < maxCursor {
		return m.cursor + 1
	}
	return m.cursor
}

// initAbilityScores 初始化属性值
func (m *CharacterCreationModel) initAbilityScores() {
	m.abilities = character.Attributes{
		Strength:     10,
		Dexterity:    10,
		Constitution: 10,
		Intelligence: 10,
		Wisdom:       10,
		Charisma:     10,
	}
	if m.race != nil {
		for ability, bonus := range m.race.AbilityBonuses {
			switch ability {
			case character.Strength:
				m.abilities.Strength += bonus
			case character.Dexterity:
				m.abilities.Dexterity += bonus
			case character.Constitution:
				m.abilities.Constitution += bonus
			case character.Intelligence:
				m.abilities.Intelligence += bonus
			case character.Wisdom:
				m.abilities.Wisdom += bonus
			case character.Charisma:
				m.abilities.Charisma += bonus
			}
		}
	}
}

// initAvailableSkills 初始化可用技能
func (m *CharacterCreationModel) initAvailableSkills() {
	m.availableSkills = character.AllSkillTypes()
}

// View 渲染
func (m CharacterCreationModel) View() string {
	if !m.ready {
		return "正在加载..."
	}
	var b strings.Builder
	b.WriteString(CreationStyles.Title.Render("创建角色"))
	b.WriteString("\n\n")
	b.WriteString(m.renderProgress())
	b.WriteString("\n\n")
	b.WriteString(m.renderStep())
	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(CreationStyles.Error.Render(m.err))
	}
	return b.String()
}

// renderProgress 渲染进度
func (m CharacterCreationModel) renderProgress() string {
	steps := []CreationStep{StepName, StepRace, StepClass, StepAbilityScores, StepSkills, StepConfirm}
	var items []string
	for _, s := range steps {
		if s == m.step {
			items = append(items, CreationStyles.ProgressActive.Render("●"))
		} else if s < m.step {
			items = append(items, CreationStyles.ProgressDone.Render("●"))
		} else {
			items = append(items, CreationStyles.ProgressFuture.Render("○"))
		}
	}
	return strings.Join(items, " ")
}

// renderStep 渲染当前步骤
func (m CharacterCreationModel) renderStep() string {
	switch m.step {
	case StepName:
		return m.renderNameStep()
	case StepRace:
		return m.renderRaceStep()
	case StepSubrace:
		return m.renderSubraceStep()
	case StepClass:
		return m.renderClassStep()
	case StepSubclass:
		return m.renderSubclassStep()
	case StepAbilityScores:
		return m.renderAbilityStep()
	case StepSkills:
		return m.renderSkillsStep()
	case StepConfirm:
		return m.renderConfirmStep()
	default:
		return "未知步骤"
	}
}

// renderNameStep 渲染名称步骤
func (m CharacterCreationModel) renderNameStep() string {
	var b strings.Builder
	b.WriteString(CreationStyles.Label.Render("请输入角色名称:"))
	b.WriteString("\n\n")
	b.WriteString(CreationStyles.Input.Render(m.name + "█"))
	b.WriteString("\n\n")
	b.WriteString(CreationStyles.Hint.Render("按回车确认"))
	return b.String()
}

// renderRaceStep 渲染种族步骤
func (m CharacterCreationModel) renderRaceStep() string {
	var b strings.Builder
	b.WriteString(CreationStyles.Label.Render("选择种族:"))
	b.WriteString("\n\n")
	for i, race := range m.availableRaces {
		line := fmt.Sprintf("%s - %s", race.Name, race.Description)
		if i == m.cursor {
			b.WriteString(CreationStyles.SelectedItem.Render("▶ " + line))
		} else {
			b.WriteString(CreationStyles.Item.Render("  " + line))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// renderSubraceStep 渲染子种族步骤
func (m CharacterCreationModel) renderSubraceStep() string {
	if m.race == nil {
		return "请先选择种族"
	}
	var b strings.Builder
	b.WriteString(CreationStyles.Label.Render(fmt.Sprintf("选择%s的子种族:", m.race.Name)))
	b.WriteString("\n\n")
	for i, subrace := range m.race.SubRaces {
		line := fmt.Sprintf("%s - %s", subrace.Name, subrace.Description)
		if i == m.cursor {
			b.WriteString(CreationStyles.SelectedItem.Render("▶ " + line))
		} else {
			b.WriteString(CreationStyles.Item.Render("  " + line))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// renderClassStep 渲染职业步骤
func (m CharacterCreationModel) renderClassStep() string {
	var b strings.Builder
	b.WriteString(CreationStyles.Label.Render("选择职业:"))
	b.WriteString("\n\n")
	for i, class := range m.availableClasses {
		hitDie := string(class.HitDice)
		line := fmt.Sprintf("%s - %s (生命骰: %s)", class.Name, class.Description, hitDie)
		if i == m.cursor {
			b.WriteString(CreationStyles.SelectedItem.Render("▶ " + line))
		} else {
			b.WriteString(CreationStyles.Item.Render("  " + line))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// renderSubclassStep 渲染子职业步骤
func (m CharacterCreationModel) renderSubclassStep() string {
	if m.class == nil {
		return "请先选择职业"
	}
	var b strings.Builder
	b.WriteString(CreationStyles.Label.Render(fmt.Sprintf("选择%s的子职业:", m.class.Name)))
	b.WriteString("\n\n")
	for i, subclass := range m.class.SubClasses {
		line := fmt.Sprintf("%s - %s", subclass.Name, subclass.Description)
		if i == m.cursor {
			b.WriteString(CreationStyles.SelectedItem.Render("▶ " + line))
		} else {
			b.WriteString(CreationStyles.Item.Render("  " + line))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// renderAbilityStep 渲染属性步骤
func (m CharacterCreationModel) renderAbilityStep() string {
	var b strings.Builder
	b.WriteString(CreationStyles.Label.Render("分配属性点:"))
	b.WriteString("\n\n")
	abilities := []struct {
		name  string
		value int
	}{
		{"力量", m.abilities.Strength},
		{"敏捷", m.abilities.Dexterity},
		{"体质", m.abilities.Constitution},
		{"智力", m.abilities.Intelligence},
		{"感知", m.abilities.Wisdom},
		{"魅力", m.abilities.Charisma},
	}
	for _, a := range abilities {
		mod := (a.value - 10) / 2
		line := fmt.Sprintf("%s: %d (%+d)", a.name, a.value, mod)
		b.WriteString(CreationStyles.Item.Render(line))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(CreationStyles.Hint.Render("按回车继续"))
	return b.String()
}

// renderSkillsStep 渲染技能步骤
func (m CharacterCreationModel) renderSkillsStep() string {
	var b strings.Builder
	b.WriteString(CreationStyles.Label.Render("选择熟练技能:"))
	b.WriteString("\n\n")
	for i, skill := range m.availableSkills {
		name := character.GetSkillName(skill)
		var line string
		if m.isSkillSelected(skill) {
			line = fmt.Sprintf("✓ %s", name)
		} else {
			line = fmt.Sprintf("○ %s", name)
		}
		if i == m.cursor {
			b.WriteString(CreationStyles.SelectedItem.Render("▶ " + line))
		} else {
			b.WriteString(CreationStyles.Item.Render("  " + line))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// renderConfirmStep 渲染确认步骤
func (m CharacterCreationModel) renderConfirmStep() string {
	var b strings.Builder
	b.WriteString(CreationStyles.Label.Render("确认角色信息:"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("名称: %s\n", m.name))
	if m.race != nil {
		raceName := m.race.Name
		if m.subrace != nil {
			raceName = m.subrace.Name
		}
		b.WriteString(fmt.Sprintf("种族: %s\n", raceName))
	}
	if m.class != nil {
		className := m.class.Name
		if m.subclass != nil {
			className = fmt.Sprintf("%s (%s)", m.class.Name, m.subclass.Name)
		}
		b.WriteString(fmt.Sprintf("职业: %s\n", className))
	}
	b.WriteString("\n属性:\n")
	b.WriteString(fmt.Sprintf("  力量 %d, 敏捷 %d, 体质 %d\n", m.abilities.Strength, m.abilities.Dexterity, m.abilities.Constitution))
	b.WriteString(fmt.Sprintf("  智力 %d, 感知 %d, 魅力 %d\n", m.abilities.Intelligence, m.abilities.Wisdom, m.abilities.Charisma))
	b.WriteString("\n")
	b.WriteString(CreationStyles.Hint.Render("按回车确认创建角色"))
	return b.String()
}

// isSkillSelected 检查技能是否已选择
func (m CharacterCreationModel) isSkillSelected(skill character.SkillType) bool {
	for _, s := range m.skills {
		if s == skill {
			return true
		}
	}
	return false
}

// CreationStyles 创建角色样式
var CreationStyles = struct {
	Title          lipgloss.Style
	Label          lipgloss.Style
	Input          lipgloss.Style
	Item           lipgloss.Style
	SelectedItem   lipgloss.Style
	ProgressActive lipgloss.Style
	ProgressDone   lipgloss.Style
	ProgressFuture lipgloss.Style
	Hint           lipgloss.Style
	Error          lipgloss.Style
}{
	Title:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7c3aed")).Padding(1, 2),
	Label:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#89b4fa")),
	Input:          lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")).Padding(0, 1),
	Item:           lipgloss.NewStyle().Foreground(lipgloss.Color("#a6adc8")),
	SelectedItem:   lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Bold(true),
	ProgressActive: lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")),
	ProgressDone:   lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")),
	ProgressFuture: lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")),
	Hint:           lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")).Italic(true),
	Error:          lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")),
}

// GetCharacter 获取创建的角色
func (m CharacterCreationModel) GetCharacter() *character.Character {
	if m.name == "" || m.race == nil || m.class == nil {
		return nil
	}
	c := character.NewCharacter(m.name, *m.race, *m.class)
	c.Attributes = m.abilities
	conMod := c.Attributes.Modifier(character.Constitution)
	hitDie := c.Class.HitDice.GetHitDiceValue()
	c.HitPoints.Max = hitDie + conMod
	c.HitPoints.Current = c.HitPoints.Max
	return c
}
