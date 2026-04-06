package character

// StandardClasses 标准 D&D 5e 职业数据（参考《玩家手册》）
// 所有术语使用官方中文翻译
var StandardClasses = map[string]Class{
	// ============== 野蛮人 ==============
	"barbarian": {
		ID:             "barbarian",
		Name:           "野蛮人",
		NameEn:         "Barbarian",
		Type:           Barbarian,
		Description:    "野蛮人是凶猛的战士，以力量和愤怒著称。他们可以在战斗中进入狂暴状态，获得强大的战斗能力。",
		HitDice:        D12,
		PrimaryAbility: []Ability{Strength},
		SavingThrows:   []Ability{Strength, Constitution},
		SkillCount:     2,
		SkillOptions: []SkillType{
			AnimalHandling, Athletics, Intimidation, Nature, Perception, Survival,
		},
		ArmorProficiencies:  []string{"轻甲", "中甲", "盾牌"},
		WeaponProficiencies: []string{"简易武器", "军用武器"},
		Features: []Feature{
			{Name: "狂暴", Level: 1, Description: "以附赠动作进入狂暴，近战武器伤害+2，力量豁免优势，攻击检定优势"},
			{Name: "无甲防御", Level: 1, Description: "不穿护甲时AC=10+敏捷调整值+体质调整值"},
		},
		SubClasses: []SubClass{
			{
				ID:          "berserker",
				Name:        "狂战士",
				Description: "狂战士之路是纯粹暴力的道路，专注于在狂暴中释放更猛烈的攻击。",
				Level:       3,
				Features: []Feature{
					{Name: "狂乱", Level: 3, Description: "狂暴期间可用附赠动作进行近战武器攻击"},
				},
			},
			{
				ID:          "totem_warrior",
				Name:        "图腾武士",
				Description: "图腾武士与灵兽建立精神连接，获得动物的力量。",
				Level:       3,
				Features: []Feature{
					{Name: "图腾灵兽", Level: 3, Description: "选择熊、鹰、狼或麋鹿图腾获得相应能力"},
				},
			},
		},
	},

	// ============== 吟游诗人 ==============
	"bard": {
		ID:                  "bard",
		Name:                "吟游诗人",
		NameEn:              "Bard",
		Type:                Bard,
		Description:         "吟游诗人以音乐和语言为武器，既是艺人也是施法者。他们可以鼓舞盟友，干扰敌人。",
		HitDice:             D8,
		PrimaryAbility:      []Ability{Charisma},
		SavingThrows:        []Ability{Dexterity, Charisma},
		SkillCount:          3,
		SkillOptions:        AllSkillTypes(), // 吟游诗人可以选择任何技能
		ArmorProficiencies:  []string{"轻甲"},
		WeaponProficiencies: []string{"简易武器", "手弩", "长剑", "短剑", "短弓"},
		Spellcasting:        true,
		SpellcastingAbility: Charisma,
		RitualCaster:        true,
		CantripCount:        2,
		Features: []Feature{
			{Name: "诗人激励", Level: 1, Description: "使用附赠动作让一个生物获得d6激励骰，可在d20检定中使用"},
			{Name: "施法", Level: 1, Description: "可以施展吟游诗人法术"},
		},
		SubClasses: []SubClass{
			{
				ID:          "lore_college",
				Name:        "逸闻学院",
				Description: "逸闻学院收集各种知识和秘辛，专注于辅助和增益能力。",
				Level:       3,
				Features: []Feature{
					{Name: "额外熟练", Level: 3, Description: "获得三项自选技能熟练或三项自选工具熟练"},
					{Name: "言语切割", Level: 3, Description: "可以使用反应对敌人的攻击检定施加-2d6减值"},
				},
			},
			{
				ID:          "valor_college",
				Name:        "勇武学院",
				Description: "勇武学院将魔法与武艺结合，是战场上的全能战士。",
				Level:       3,
				Features: []Feature{
					{Name: "战斗激励", Level: 3, Description: "可以用附赠动作施展诗人激励，激励骰可用于攻击伤害"},
					{Name: "中甲熟练", Level: 3, Description: "获得中甲、盾牌、军用武器熟练"},
				},
			},
		},
	},

	// ============== 牧师 ==============
	"cleric": {
		ID:             "cleric",
		Name:           "牧师",
		NameEn:         "Cleric",
		Type:           Cleric,
		Description:    "牧师是神明的使者，使用神圣魔法为神明服务。他们可以治愈盟友，惩罚异端。",
		HitDice:        D8,
		PrimaryAbility: []Ability{Wisdom},
		SavingThrows:   []Ability{Wisdom, Charisma},
		SkillCount:     2,
		SkillOptions: []SkillType{
			History, Insight, Medicine, Persuasion, Religion,
		},
		ArmorProficiencies:  []string{"轻甲", "中甲", "盾牌"},
		WeaponProficiencies: []string{"简易武器"},
		Spellcasting:        true,
		SpellcastingAbility: Wisdom,
		RitualCaster:        true,
		CantripCount:        3,
		Features: []Feature{
			{Name: "施法", Level: 1, Description: "可以施展牧师法术"},
			{Name: "神圣领域", Level: 1, Description: "选择一个神祗领域获得相应能力"},
		},
		SubClasses: []SubClass{
			{
				ID:          "life_domain",
				Name:        "生命领域",
				Description: "生命领域专注于治愈和保护，是神圣治愈者。",
				Level:       1,
				Features: []Feature{
					{Name: "生命门徒", Level: 1, Description: "治疗法术恢复量+2+法术环阶"},
					{Name: "重度治愈", Level: 2, Description: "使用1次引导神力施放群体治愈"},
				},
			},
			{
				ID:          "light_domain",
				Name:        "光明领域",
				Description: "光明领域控制火焰和光芒，是对抗黑暗的力量。",
				Level:       1,
				Features: []Feature{
					{Name: "光明赐福", Level: 1, Description: "知晓光明戏法，感知（察觉）检定具有优势"},
					{Name: "耀光", Level: 2, Description: "使用引导神力让敌人受到光耀伤害"},
				},
			},
			{
				ID:          "war_domain",
				Name:        "战争领域",
				Description: "战争领域是战场上的战士牧师，擅长战斗。",
				Level:       1,
				Features: []Feature{
					{Name: "战争祭司", Level: 1, Description: "可以自己使用战争神祗引导神力"},
					{Name: "引导打击", Level: 2, Description: "使用引导神力让盟友的武器攻击伤害+10"},
				},
			},
		},
	},

	// ============== 德鲁伊 ==============
	"druid": {
		ID:             "druid",
		Name:           "德鲁伊",
		NameEn:         "Druid",
		Type:           Druid,
		Description:    "德鲁伊是自然的守护者，可以变化形态并施展自然魔法。他们与动植物保持着神秘联系。",
		HitDice:        D8,
		PrimaryAbility: []Ability{Wisdom},
		SavingThrows:   []Ability{Intelligence, Wisdom},
		SkillCount:     2,
		SkillOptions: []SkillType{
			Arcana, AnimalHandling, Insight, Medicine, Nature, Perception, Religion, Survival,
		},
		ArmorProficiencies:  []string{"轻甲", "中甲", "盾牌（非金属）"},
		WeaponProficiencies: []string{"短棒", "匕首", "飞镖", "标枪", "硬头锤", "长棍", "弯刀", "镰刀", "短矛", "投石索"},
		Spellcasting:        true,
		SpellcastingAbility: Wisdom,
		RitualCaster:        true,
		CantripCount:        2,
		Features: []Feature{
			{Name: "施法", Level: 1, Description: "可以施展德鲁伊法术"},
			{Name: "德鲁伊语", Level: 1, Description: "可以使用德鲁伊语与其他德鲁伊交流"},
			{Name: "野性形态", Level: 2, Description: "可以变化为挑战等级1/4或更低的野兽"},
		},
		SubClasses: []SubClass{
			{
				ID:          "land_circle",
				Name:        "大地结社",
				Description: "大地结社专注于魔法和自然知识，获得额外的法术能力。",
				Level:       2,
				Features: []Feature{
					{Name: "自然恢复", Level: 2, Description: "短休时可以恢复部分法术位"},
					{Name: "大地行者", Level: 6, Description: "在非魔法困难地形上移动不消耗额外移动力"},
				},
			},
			{
				ID:          "moon_circle",
				Name:        "月亮结社",
				Description: "月亮结社专注于野性形态，可以变化为更强大的野兽。",
				Level:       2,
				Features: []Feature{
					{Name: "战斗变形", Level: 2, Description: "可以使用动作变形，并获得更高的挑战等级形态"},
					{Name: "元素野性", Level: 10, Description: "可以变化为元素生物"},
				},
			},
		},
	},

	// ============== 战士 ==============
	"fighter": {
		ID:             "fighter",
		Name:           "战士",
		NameEn:         "Fighter",
		Type:           Fighter,
		Description:    "战士是战斗专家，精通各种武器和护甲。他们拥有出色的耐力和战斗技巧。",
		HitDice:        D10,
		PrimaryAbility: []Ability{Strength, Dexterity},
		SavingThrows:   []Ability{Strength, Constitution},
		SkillCount:     2,
		SkillOptions: []SkillType{
			Acrobatics, AnimalHandling, Athletics, History, Insight, Intimidation, Perception, Survival,
		},
		ArmorProficiencies:  []string{"所有护甲", "盾牌"},
		WeaponProficiencies: []string{"简易武器", "军用武器"},
		Features: []Feature{
			{Name: "战斗风格", Level: 1, Description: "选择一个战斗风格获得特殊加成"},
			{Name: "回气", Level: 1, Description: "使用附赠动作恢复1d10+战士等级生命值"},
			{Name: "额外攻击", Level: 5, Description: "攻击动作可以攻击两次"},
		},
		SubClasses: []SubClass{
			{
				ID:          "champion",
				Name:        "冠军勇士",
				Description: "冠军勇士是纯粹的身体战士，专注于命中和伤害。",
				Level:       3,
				Features: []Feature{
					{Name: "扩展重击", Level: 3, Description: "攻击检定18-20都算作重击"},
					{Name: "运动健将", Level: 7, Description: "体力检定中未熟练时熟练加值减半"},
				},
			},
			{
				ID:          "battle_master",
				Name:        "战斗大师",
				Description: "战斗大师是战术大师，可以使用战技控制战场。",
				Level:       3,
				Features: []Feature{
					{Name: "战技 superiority", Level: 3, Description: "学习4种战技，获得4个卓越骰（d8）"},
					{Name: "战争学生", Level: 3, Description: "从战士技能列表中选择2项技能获得熟练"},
				},
			},
			{
				ID:          "eldritch_knight",
				Name:        "奥法骑士",
				Description: "奥法骑士将魔法与武艺结合，可以施展法师法术。",
				Level:       3,
				Features: []Feature{
					{Name: "施法", Level: 3, Description: "可以施展法师法术（专注塑能系和防护系）"},
					{Name: "战争施法", Level: 7, Description: "施法时武器攻击检定具有优势"},
				},
			},
		},
	},

	// ============== 武僧 ==============
	"monk": {
		ID:             "monk",
		Name:           "武僧",
		NameEn:         "Monk",
		Type:           Monk,
		Description:    "武僧是习武的苦行者，将身体和意志修炼成武器。他们可以使用气来进行超凡的壮举。",
		HitDice:        D8,
		PrimaryAbility: []Ability{Dexterity, Wisdom},
		SavingThrows:   []Ability{Strength, Dexterity},
		SkillCount:     2,
		SkillOptions: []SkillType{
			Acrobatics, Athletics, History, Insight, Religion, Stealth,
		},
		ArmorProficiencies:  []string{"无"},
		WeaponProficiencies: []string{"简易武器", "短剑"},
		Tools:               []string{"一种工匠工具或一种乐器"},
		Features: []Feature{
			{Name: "无甲防御", Level: 1, Description: "不穿护甲和盾牌时AC=10+敏捷调整值+感知调整值"},
			{Name: "武术", Level: 1, Description: "徒手攻击伤害变为d4，可以用敏捷进行徒手攻击"},
			{Name: "气", Level: 2, Description: "获得气点，可用于特殊能力"},
		},
		SubClasses: []SubClass{
			{
				ID:          "open_hand",
				Name:        "敞开心门",
				Description: "敞开心门是最传统的武僧道路，专注于徒手格斗技巧。",
				Level:       3,
				Features: []Feature{
					{Name: "散打", Level: 3, Description: "疾风连击命中后可以推倒、推后或使敌人无法反应"},
				},
			},
			{
				ID:          "shadow",
				Name:        "暗影之道",
				Description: "暗影之道是潜行者武僧，可以操纵暗影和潜伏。",
				Level:       3,
				Features: []Feature{
					{Name: "暗影技艺", Level: 3, Description: "获得黑暗视觉，可以施展黑暗术、寂静术等法术"},
				},
			},
			{
				ID:          "four_elements",
				Name:        "四象之道",
				Description: "四象之道可以操纵元素力量，施法如武。",
				Level:       3,
				Features: []Feature{
					{Name: "元素技艺", Level: 3, Description: "可以使用气点施展元素法术"},
				},
			},
		},
	},

	// ============== 圣武士 ==============
	"paladin": {
		ID:             "paladin",
		Name:           "圣武士",
		NameEn:         "Paladin",
		Type:           Paladin,
		Description:    "圣武士是神圣誓言的战士，将武艺与神圣魔法结合。他们是正义的化身。",
		HitDice:        D10,
		PrimaryAbility: []Ability{Strength, Charisma},
		SavingThrows:   []Ability{Wisdom, Charisma},
		SkillCount:     2,
		SkillOptions: []SkillType{
			Athletics, Insight, Intimidation, Medicine, Persuasion, Religion,
		},
		ArmorProficiencies:  []string{"所有护甲", "盾牌"},
		WeaponProficiencies: []string{"简易武器", "军用武器"},
		Spellcasting:        true,
		SpellcastingAbility: Charisma,
		Features: []Feature{
			{Name: "神圣感知", Level: 1, Description: "可以使用动作侦测周围的天界生物、邪魔和不死生物"},
			{Name: "圣疗", Level: 1, Description: "可以消耗生命值池治疗盟友"},
			{Name: "战斗风格", Level: 2, Description: "选择一个战斗风格获得特殊加成"},
		},
		SubClasses: []SubClass{
			{
				ID:          "devotion",
				Name:        "虔诚之誓",
				Description: "虔诚之誓是最高尚的圣武士道路，守护正义与善良。",
				Level:       3,
				Features: []Feature{
					{Name: "十诫", Level: 3, Description: "可以施展圣火术等法术"},
					{Name: "神圣武器", Level: 3, Description: "使用引导神力让武器发光并造成额外伤害"},
				},
			},
			{
				ID:          "ancients",
				Name:        "远古之誓",
				Description: "远古之誓守护自然和善良，对抗邪恶势力。",
				Level:       3,
				Features: []Feature{
					{Name: "自然誓言", Level: 3, Description: "可以施展自然相关法术"},
					{Name: "先贤光环", Level: 3, Description: "光环内盟友对豁免检定具有优势"},
				},
			},
			{
				ID:          "vengeance",
				Name:        "复仇之誓",
				Description: "复仇之誓是对抗邪恶的狂战士，誓要肃清一切黑暗。",
				Level:       3,
				Features: []Feature{
					{Name: "复仇誓言", Level: 3, Description: "可以施展猎人印记等法术"},
					{Name: "神圣仇敌", Level: 3, Description: "使用引导神力对敌人攻击检定具有优势"},
				},
			},
		},
	},

	// ============== 游侠 ==============
	"ranger": {
		ID:             "ranger",
		Name:           "游侠",
		NameEn:         "Ranger",
		Type:           Ranger,
		Description:    "游侠是荒野中的猎人和追踪者，擅长远程战斗和潜行。他们与自然保持着紧密联系。",
		HitDice:        D10,
		PrimaryAbility: []Ability{Dexterity, Wisdom},
		SavingThrows:   []Ability{Strength, Dexterity},
		SkillCount:     3,
		SkillOptions: []SkillType{
			AnimalHandling, Athletics, Insight, Investigation, Nature, Perception, Stealth, Survival,
		},
		ArmorProficiencies:  []string{"轻甲", "中甲", "盾牌"},
		WeaponProficiencies: []string{"简易武器", "军用武器"},
		Spellcasting:        true,
		SpellcastingAbility: Wisdom,
		Features: []Feature{
			{Name: "宿敌", Level: 1, Description: "对选定类型的生物获得追踪和记忆优势"},
			{Name: "自然探索者", Level: 1, Description: "对选定地形类型获得探索和感知优势"},
			{Name: "战斗风格", Level: 2, Description: "选择一个战斗风格获得特殊加成"},
			{Name: "施法", Level: 2, Description: "可以施展游侠法术"},
		},
		SubClasses: []SubClass{
			{
				ID:          "hunter",
				Name:        "猎人",
				Description: "猎人是传统的游侠，专注于狩猎和战斗技巧。",
				Level:       3,
				Features: []Feature{
					{Name: "猎人猎物", Level: 3, Description: "选择一种猎物获得额外伤害或能力"},
				},
			},
			{
				ID:          "beast_master",
				Name:        "兽王",
				Description: "兽王与动物伙伴建立心灵链接，共同战斗。",
				Level:       3,
				Features: []Feature{
					{Name: "兽王伙伴", Level: 3, Description: "获得一只动物伙伴，可以指挥其行动"},
				},
			},
		},
	},

	// ============== 游荡者 ==============
	"rogue": {
		ID:             "rogue",
		Name:           "游荡者",
		NameEn:         "Rogue",
		Type:           Rogue,
		Description:    "游荡者是敏捷的潜行者和诡计大师，擅长伏击和偷袭。他们可以造成致命的突袭伤害。",
		HitDice:        D8,
		PrimaryAbility: []Ability{Dexterity},
		SavingThrows:   []Ability{Dexterity, Intelligence},
		SkillCount:     4,
		SkillOptions: []SkillType{
			Acrobatics, Athletics, Deception, Insight, Intimidation, Investigation,
			Perception, Performance, Persuasion, SleightOfHand, Stealth,
		},
		ArmorProficiencies:  []string{"轻甲"},
		WeaponProficiencies: []string{"简易武器", "手弩", "长剑", "刺剑", "短剑"},
		Tools:               []string{"盗贼工具"},
		Features: []Feature{
			{Name: "专精", Level: 1, Description: "选择两项自选技能获得双倍熟练加值"},
			{Name: "偷袭", Level: 1, Description: "有优势攻击时额外造成2d6伤害（每2级+1d6）"},
			{Name: "盗贼黑话", Level: 1, Description: "可以理解和使用盗贼黑话"},
			{Name: "灵巧动作", Level: 2, Description: "可以使用附赠动作进行技巧动作或隐藏"},
		},
		SubClasses: []SubClass{
			{
				ID:          "thief",
				Name:        "盗贼",
				Description: "盗贼是传统游荡者，专注于潜行和技巧。",
				Level:       3,
				Features: []Feature{
					{Name: "灵巧快手", Level: 3, Description: "可以使用附赠动作使用物品"},
					{Name: "精通攀爬", Level: 3, Description: "攀爬不需要消耗额外移动力"},
				},
			},
			{
				ID:          "assassin",
				Name:        "刺客",
				Description: "刺客专注于伏击和致命一击。",
				Level:       3,
				Features: []Feature{
					{Name: "刺杀", Level: 3, Description: "对未行动的目标攻击具有优势，命中时重击"},
					{Name: "渗透专精", Level: 3, Description: "可以进行快速伪装和伪造文书"},
				},
			},
			{
				ID:          "arcane_trickster",
				Name:        "诡术师",
				Description: "诡术师将魔法与诡术结合，可以施展法师法术。",
				Level:       3,
				Features: []Feature{
					{Name: "法师之手", Level: 3, Description: "可以用附赠动作控制法师之手进行技巧动作"},
					{Name: "施法", Level: 3, Description: "可以施展法师法术（专注幻术系和附魔系）"},
				},
			},
		},
	},

	// ============== 术士 ==============
	"sorcerer": {
		ID:             "sorcerer",
		Name:           "术士",
		NameEn:         "Sorcerer",
		Type:           Sorcerer,
		Description:    "术士天生拥有魔法力量，体内流淌着魔力源泉的血统。他们可以操控魔法能量。",
		HitDice:        D6,
		PrimaryAbility: []Ability{Charisma},
		SavingThrows:   []Ability{Constitution, Charisma},
		SkillCount:     2,
		SkillOptions: []SkillType{
			Arcana, Deception, Insight, Intimidation, Persuasion, Religion,
		},
		ArmorProficiencies:  []string{"无"},
		WeaponProficiencies: []string{"匕首", "飞镖", "投石索", "短矛", "轻弩"},
		Spellcasting:        true,
		SpellcastingAbility: Charisma,
		CantripCount:        4,
		Features: []Feature{
			{Name: "施法", Level: 1, Description: "可以施展术士法术"},
			{Name: "术法点", Level: 1, Description: "获得术法点，可用于超魔或恢复法术位"},
			{Name: "超魔", Level: 2, Description: "学习超魔选项来强化法术"},
		},
		SubClasses: []SubClass{
			{
				ID:          "draconic",
				Name:        "龙族血统",
				Description: "龙族血统术士拥有龙族血脉，可以获得龙类特性。",
				Level:       1,
				Features: []Feature{
					{Name: "龙族先祖", Level: 1, Description: "选择一种龙类获得相应伤害抗性"},
					{Name: "龙族恢复力", Level: 1, Description: "HP上限+1，升级时再+1；无甲时AC=13+敏捷调整值"},
				},
			},
			{
				ID:          "wild_magic",
				Name:        "狂野魔法",
				Description: "狂野魔法术士体内涌动着不稳定的魔法能量。",
				Level:       1,
				Features: []Feature{
					{Name: "狂野魔法涌动", Level: 1, Description: "施法后可能触发狂野魔法效果"},
					{Name: "混乱之潮", Level: 1, Description: "可以使用术法点重掷d20"},
				},
			},
		},
	},

	// ============== 邪术师 ==============
	"warlock": {
		ID:             "warlock",
		Name:           "邪术师",
		NameEn:         "Warlock",
		Type:           Warlock,
		Description:    "邪术师与强大的异界存在订立契约，获得魔法力量。他们使用祈唤而非法术位施法。",
		HitDice:        D8,
		PrimaryAbility: []Ability{Charisma},
		SavingThrows:   []Ability{Wisdom, Charisma},
		SkillCount:     2,
		SkillOptions: []SkillType{
			Arcana, Deception, History, Intimidation, Investigation, Nature, Religion,
		},
		ArmorProficiencies:  []string{"轻甲"},
		WeaponProficiencies: []string{"简易武器"},
		Spellcasting:        true,
		SpellcastingAbility: Charisma,
		CantripCount:        2,
		Features: []Feature{
			{Name: "宗主契约", Level: 1, Description: "选择一个异界宗主获得相应能力"},
			{Name: "祈唤", Level: 2, Description: "学习祈唤来获得特殊能力"},
		},
		SubClasses: []SubClass{
			{
				ID:          "archfey",
				Name:        "妖精宗主",
				Description: "妖精宗主是妖精界的强大存在，给予幻术和魅惑能力。",
				Level:       1,
				Features: []Feature{
					{Name: "妖精祝福", Level: 1, Description: "可以用反应魅惑或恐惧敌人"},
				},
			},
			{
				ID:          "fiend",
				Name:        "邪魔宗主",
				Description: "邪魔宗主是地狱或深渊的强大魔鬼或恶魔。",
				Level:       1,
				Features: []Feature{
					{Name: "黑暗祝福", Level: 1, Description: "消灭敌人后获得临时生命值"},
				},
			},
			{
				ID:          "great_old_one",
				Name:        "旧日支配者",
				Description: "旧日支配者是来自遥远世界的疯狂存在。",
				Level:       1,
				Features: []Feature{
					{Name: "旧日祝福", Level: 1, Description: "可以用附赠动作与目标心灵交流"},
				},
			},
		},
	},

	// ============== 法师 ==============
	"wizard": {
		ID:             "wizard",
		Name:           "法师",
		NameEn:         "Wizard",
		Type:           Wizard,
		Description:    "法师是奥术的大师，通过学习和研究掌握魔法。他们拥有最广泛的法术列表。",
		HitDice:        D6,
		PrimaryAbility: []Ability{Intelligence},
		SavingThrows:   []Ability{Intelligence, Wisdom},
		SkillCount:     2,
		SkillOptions: []SkillType{
			Arcana, History, Insight, Investigation, Medicine, Religion,
		},
		ArmorProficiencies:  []string{"无"},
		WeaponProficiencies: []string{"匕首", "飞镖", "投石索", "长棍", "轻弩"},
		Spellcasting:        true,
		SpellcastingAbility: Intelligence,
		RitualCaster:        true,
		CantripCount:        3,
		Features: []Feature{
			{Name: "奥术复苏", Level: 1, Description: "短休时可以恢复部分法术位"},
			{Name: "施法", Level: 1, Description: "可以施展法师法术"},
			{Name: "法术书", Level: 1, Description: "开始时有6个1环法术，可以将发现的法术抄入法术书"},
		},
		SubClasses: []SubClass{
			{
				ID:          "evocation",
				Name:        "塑能学派",
				Description: "塑能学派专注于创造元素能量，如火焰、闪电和冰霜。",
				Level:       2,
				Features: []Feature{
					{Name: "雕文", Level: 2, Description: "塑能法术可以穿过盟友而不伤害他们"},
					{Name: "能量强化", Level: 10, Description: "塑能法术伤害增加智力调整值"},
				},
			},
			{
				ID:          "abjuration",
				Name:        "防护学派",
				Description: "防护学派专注于保护法术，可以抵御各种伤害。",
				Level:       2,
				Features: []Feature{
					{Name: "奥术结界", Level: 2, Description: "获得奥术结界可以吸收伤害"},
				},
			},
			{
				ID:          "divination",
				Name:        "预言学派",
				Description: "预言学派专注于洞察未来，可以改变命运。",
				Level:       2,
				Features: []Feature{
					{Name: "预言专家", Level: 2, Description: "可以预骰并替换未来的d20结果"},
				},
			},
			{
				ID:          "conjuration",
				Name:        "咒法学派",
				Description: "咒法学派专注于传送和召唤，可以操纵空间。",
				Level:       2,
				Features: []Feature{
					{Name: "细微召唤", Level: 2, Description: "施法后可以传送10尺"},
				},
			},
		},
	},
}

// GetClass 根据 ID 获取职业
func GetClass(id string) *Class {
	if class, ok := StandardClasses[id]; ok {
		return &class
	}
	return nil
}

// AllClasses 返回所有职业列表
func AllClasses() []Class {
	classes := make([]Class, 0, len(StandardClasses))
	for _, class := range StandardClasses {
		classes = append(classes, class)
	}
	return classes
}

// FightingStyles 战斗风格选项
var FightingStyles = map[string]string{
	"archery":      "弓术：远程攻击检定+2",
	"defense":      "防御：穿甲时AC+1",
	"dueling":      "单手武器：单手持握近战武器时伤害+2",
	"great_weapon": "双手武器：双手近战武器重击时可再攻击一次",
	"protection":   "保护：可以用反应让攻击邻接盟友的敌人具有劣势",
	"two_weapon":   "双武器：双持武器时可以使用附赠动作攻击另一把武器",
}

// MetamagicOptions 超魔选项（术士）
var MetamagicOptions = map[string]string{
	"careful_spell":    "谨慎法术：让范围内盟友自动成功豁免",
	"distant_spell":    "远距法术：法术距离翻倍",
	"empowered_spell":  "强化法术：重掷伤害骰",
	"extended_spell":   "延时法术：持续时间翻倍",
	"heightened_spell": "高阶法术：让目标首次豁免具有劣势",
	"quickened_spell":  "极速法术：施法时间从动作变为附赠动作",
	"subtle_spell":     "隐秘法术：不需要言语和姿势成分",
	"twinned_spell":    "双生法术：让单体法术影响第二个目标",
}
