package dice

import (
	"testing"
)

func TestRoll(t *testing.T) {
	tests := []struct {
		name  string
		n     int
		sides int
	}{
		{"1d20", 1, 20},
		{"2d6", 2, 6},
		{"4d8", 4, 8},
		{"1d4", 1, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := Roll(tt.n, tt.sides)

			if len(results) != tt.n {
				t.Errorf("Roll() returned %d dice, expected %d", len(results), tt.n)
			}

			for i, r := range results {
				if r < 1 || r > tt.sides {
					t.Errorf("Roll()[%d] = %d, expected between 1 and %d", i, r, tt.sides)
				}
			}
		})
	}
}

func TestD20(t *testing.T) {
	result := D20()
	if result < 1 || result > 20 {
		t.Errorf("D20() = %d, expected between 1 and 20", result)
	}
}

func TestD20WithModifier(t *testing.T) {
	tests := []struct {
		name     string
		modifier int
		rollType RollType
	}{
		{"Normal +5", 5, NormalRoll},
		{"Advantage +3", 3, AdvantageRoll},
		{"Disadvantage -2", -2, DisadvantageRoll},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := D20WithModifier(tt.modifier, tt.rollType)

			if len(result.Dice) != 1 {
				t.Errorf("Expected 1 die, got %d", len(result.Dice))
			}

			if result.Dice[0] < 1 || result.Dice[0] > 20 {
				t.Errorf("Die roll %d out of range [1, 20]", result.Dice[0])
			}

			if result.Modifier != tt.modifier {
				t.Errorf("Modifier = %d, expected %d", result.Modifier, tt.modifier)
			}

			expectedTotal := result.Dice[0] + tt.modifier
			if result.Total != expectedTotal {
				t.Errorf("Total = %d, expected %d", result.Total, expectedTotal)
			}

			// 检查优势/劣势
			if tt.rollType == AdvantageRoll || tt.rollType == DisadvantageRoll {
				if len(result.Dropped) != 1 {
					t.Errorf("Expected 1 dropped die for advantage/disadvantage, got %d", len(result.Dropped))
				}
			}
		})
	}
}

func TestCriticalDetection(t *testing.T) {
	// 测试暴击成功检测
	// 注意：由于随机性，此测试可能有波动，但我们运行多次
	for i := 0; i < 1000; i++ {
		result := D20WithModifier(0, NormalRoll)
		if result.Dice[0] == 20 && result.Critical != CritSuccess {
			t.Error("Natural 20 should be CritSuccess")
		}
		if result.Dice[0] == 1 && result.Critical != CritFail {
			t.Error("Natural 1 should be CritFail")
		}
		if result.Dice[0] > 1 && result.Dice[0] < 20 && result.Critical != CritNone {
			t.Error("Non-extreme rolls should be CritNone")
		}
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		count     int
		sides     int
		modifier  int
		rollType  RollType
		wantError bool
	}{
		{"1d20", "1d20", 1, 20, 0, NormalRoll, false},
		{"2d6", "2d6", 2, 6, 0, NormalRoll, false},
		{"d8", "d8", 1, 8, 0, NormalRoll, false},
		{"1d20+5", "1d20+5", 1, 20, 5, NormalRoll, false},
		{"2d6-1", "2d6-1", 2, 6, -1, NormalRoll, false},
		{"1d20adv", "1d20adv", 1, 20, 0, AdvantageRoll, false},
		{"1d20dis", "1d20dis", 1, 20, 0, DisadvantageRoll, false},
		{"d20a", "d20a", 1, 20, 0, AdvantageRoll, false},
		{"d20d", "d20d", 1, 20, 0, DisadvantageRoll, false},
		{"1d20adv+3", "1d20adv+3", 1, 20, 3, AdvantageRoll, false},
		{"invalid", "abc", 0, 0, 0, NormalRoll, true},
		{"empty", "", 0, 0, 0, NormalRoll, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := Parse(tt.expr)

			if tt.wantError {
				if err == nil {
					t.Errorf("Parse(%q) expected error, got nil", tt.expr)
				}
				return
			}

			if err != nil {
				t.Errorf("Parse(%q) error: %v", tt.expr, err)
				return
			}

			if e.Count != tt.count {
				t.Errorf("Count = %d, expected %d", e.Count, tt.count)
			}
			if e.Sides != tt.sides {
				t.Errorf("Sides = %d, expected %d", e.Sides, tt.sides)
			}
			if e.Modifier != tt.modifier {
				t.Errorf("Modifier = %d, expected %d", e.Modifier, tt.modifier)
			}
			if e.RollType != tt.rollType {
				t.Errorf("RollType = %v, expected %v", e.RollType, tt.rollType)
			}
		})
	}
}

func TestExpressionString(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want string
	}{
		{"simple", "2d6", "2d6"},
		{"with modifier", "1d20+5", "1d20+5"},
		{"negative modifier", "1d20-2", "1d20-2"},
		{"advantage", "1d20adv", "1d20adv"},
		{"disadvantage", "1d20dis", "1d20dis"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			if got := e.String(); got != tt.want {
				t.Errorf("String() = %q, expected %q", got, tt.want)
			}
		})
	}
}

func TestResultIsSuccess(t *testing.T) {
	tests := []struct {
		total   int
		dc      int
		success bool
	}{
		{15, 10, true},
		{10, 10, true},
		{9, 10, false},
		{20, 15, true},
	}

	for _, tt := range tests {
		result := Result{Total: tt.total}
		if got := result.IsSuccess(tt.dc); got != tt.success {
			t.Errorf("IsSuccess(%d) with total %d = %v, expected %v", tt.dc, tt.total, got, tt.success)
		}
	}
}
