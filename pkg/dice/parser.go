package dice

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Expression 表示解析后的骰子表达式。
type Expression struct {
	Count    int      // 骰子数量
	Sides    int      // 面数
	Modifier int      // 调整值
	RollType RollType // 掷骰类型
}

// Parse 解析骰子表达式字符串，如 "2d6+3"、"1d20"、"d8-1"。
// 支持：
//   - 基础表达式：XdY、dY（默认为 1 个骰子）
//   - 调整值：+N、-N
//   - 优势：Xd20adv、Xd20a
//   - 劣势：Xd20dis、Xd20d
//
// 示例：
//
//	"2d6"     -> 2 个六面骰
//	"1d20+5"  -> 1 个二十面骰 + 5 调整值
//	"d8-1"    -> 1 个八面骰 - 1 调整值
//	"1d20adv" -> 1 个 d20 优势掷骰
//	"2d20dis" -> 2 个 d20 劣势掷骰
func Parse(expr string) (*Expression, error) {
	// 规范化表达式
	expr = strings.ToLower(strings.TrimSpace(expr))

	// 匹配骰子表达式的正则表达式
	// 分组：数量？、面数、优势/劣势？、符号？、调整值？
	re := regexp.MustCompile(`^(\d*)d(\d+)(adv|dis|a|d)?([+-]\d+)?$`)

	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return nil, fmt.Errorf("无效的骰子表达式: %s", expr)
	}

	e := &Expression{}

	// 解析数量（默认为 1）
	if matches[1] != "" {
		count, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("无效的骰子数量: %s", matches[1])
		}
		e.Count = count
	} else {
		e.Count = 1
	}

	// 解析面数
	sides, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("无效的面数: %s", matches[2])
	}
	e.Sides = sides

	// 解析优势/劣势
	if matches[3] != "" {
		switch matches[3] {
		case "adv", "a":
			e.RollType = AdvantageRoll
		case "dis", "d":
			e.RollType = DisadvantageRoll
		}
	}

	// 解析调整值
	if matches[4] != "" {
		modifier, err := strconv.Atoi(matches[4])
		if err != nil {
			return nil, fmt.Errorf("无效的调整值: %s", matches[4])
		}
		e.Modifier = modifier
	}

	return e, nil
}

// Roll 执行骰子表达式并返回结果。
func (e *Expression) Roll() Result {
	return RollDice(e.Count, e.Sides, e.Modifier, e.RollType)
}

// String 返回表达式的字符串表示。
func (e *Expression) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%dd%d", e.Count, e.Sides))

	switch e.RollType {
	case AdvantageRoll:
		sb.WriteString("adv")
	case DisadvantageRoll:
		sb.WriteString("dis")
	}

	if e.Modifier > 0 {
		sb.WriteString(fmt.Sprintf("+%d", e.Modifier))
	} else if e.Modifier < 0 {
		sb.WriteString(fmt.Sprintf("%d", e.Modifier))
	}

	return sb.String()
}

// ParseAndRoll 是一个便捷函数，解析并掷骰一步完成。
func ParseAndRoll(expr string) (Result, error) {
	e, err := Parse(expr)
	if err != nil {
		return Result{}, err
	}
	return e.Roll(), nil
}

// MustParse 解析骰子表达式，出错时 panic。
func MustParse(expr string) *Expression {
	e, err := Parse(expr)
	if err != nil {
		panic(err)
	}
	return e
}
