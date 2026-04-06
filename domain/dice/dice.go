// Package dice 提供 D&D 5e 的掷骰子工具。
package dice

import (
	"crypto/rand"
	"math/big"
)

// RollType 定义掷骰类型。
type RollType int

const (
	// NormalRoll 是标准掷骰。
	NormalRoll RollType = iota
	// AdvantageRoll 掷两次取高值。
	AdvantageRoll
	// DisadvantageRoll 掷两次取低值。
	DisadvantageRoll
)

// CriticalType 定义暴击结果类型。
type CriticalType int

const (
	// CritNone 表示无暴击。
	CritNone CriticalType = iota
	// CritSuccess 表示自然 20（暴击/成功）。
	CritSuccess
	// CritFail 表示自然 1（暴击失败）。
	CritFail
)

// Result 表示掷骰结果。
type Result struct {
	Dice     []int        `json:"dice"`      // 单个骰子结果
	Modifier int          `json:"modifier"`  // 应用的调整值
	Total    int          `json:"total"`     // 骰子总和 + 调整值
	Critical CriticalType `json:"critical"`  // 暴击结果
	RollType RollType     `json:"roll_type"` // 掷骰类型
	Dropped  []int        `json:"dropped"`   // 丢弃的骰子（用于优势/劣势）
}

// Roll 掷 n 个 s 面的骰子。
func Roll(n, s int) []int {
	results := make([]int, n)
	for i := 0; i < n; i++ {
		results[i] = rollDie(s)
	}
	return results
}

// rollDie 使用 crypto/rand 掷单个 s 面骰子以获得真正的随机性。
func rollDie(sides int) int {
	if sides <= 0 {
		return 0
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(sides)))
	if err != nil {
		// 如果 crypto/rand 失败，回退到 math/rand
		return 1 // 安全回退
	}
	return int(n.Int64()) + 1
}

// D20 掷一个 d20 并返回结果。
func D20() int {
	return rollDie(20)
}

// D20WithModifier 掷一个带调整值的 d20 并返回 Result。
func D20WithModifier(modifier int, rollType RollType) Result {
	result := Result{
		Modifier: modifier,
		RollType: rollType,
	}

	switch rollType {
	case AdvantageRoll:
		r1, r2 := rollDie(20), rollDie(20)
		if r1 >= r2 {
			result.Dice = []int{r1}
			result.Dropped = []int{r2}
		} else {
			result.Dice = []int{r2}
			result.Dropped = []int{r1}
		}
	case DisadvantageRoll:
		r1, r2 := rollDie(20), rollDie(20)
		if r1 <= r2 {
			result.Dice = []int{r1}
			result.Dropped = []int{r2}
		} else {
			result.Dice = []int{r2}
			result.Dropped = []int{r1}
		}
	default:
		result.Dice = []int{rollDie(20)}
	}

	// 检查暴击
	if len(result.Dice) > 0 {
		switch result.Dice[0] {
		case 20:
			result.Critical = CritSuccess
		case 1:
			result.Critical = CritFail
		}
	}

	result.Total = result.Dice[0] + modifier
	return result
}

// RollDice 根据表达式掷骰并返回 Result。
// 支持 d20 掷骰的优势/劣势。
func RollDice(n, sides, modifier int, rollType RollType) Result {
	result := Result{
		Modifier: modifier,
		RollType: rollType,
	}

	// 仅对单个 d20 处理优势/劣势
	if n == 1 && sides == 20 && rollType != NormalRoll {
		return D20WithModifier(modifier, rollType)
	}

	// 标准掷骰
	result.Dice = Roll(n, sides)
	result.Total = sum(result.Dice) + modifier

	// 检查 d20 的暴击
	if n == 1 && sides == 20 && len(result.Dice) > 0 {
		switch result.Dice[0] {
		case 20:
			result.Critical = CritSuccess
		case 1:
			result.Critical = CritFail
		}
	}

	return result
}

// sum 计算整数切片的总和。
func sum(dice []int) int {
	total := 0
	for _, d := range dice {
		total += d
	}
	return total
}

// IsSuccess 检查掷骰是否达到 DC。
func (r *Result) IsSuccess(dc int) bool {
	return r.Total >= dc
}
