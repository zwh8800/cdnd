package prompt

import (
	"regexp"
	"strings"
)

var (
	// optionsBlockRegex 匹配选项块
	// 匹配模式: ==========\n你的选择是：\n(选项行)
	// 支持等号数量 >= 5 的变体
	optionsBlockRegex = regexp.MustCompile(`(?s)(={5,}\s*\n你的选择是：\s*\n((?:\s*\d+\.\s*[^\n]*\n?)+))`)

	// optionLineRegex 匹配单个选项行
	// 匹配模式: 1. 选项内容
	optionLineRegex = regexp.MustCompile(`^\s*(\d+)\.\s*(.+)$`)
)

// ParseOptions 从文本中解析选项列表
// 返回: options - 解析出的选项列表, content - 移除选项块后的纯净内容
func ParseOptions(text string) ([]string, string) {
	if text == "" {
		return nil, ""
	}

	// 查找选项块
	loc := optionsBlockRegex.FindStringSubmatchIndex(text)
	if loc == nil {
		// 未找到选项块，返回原样内容
		return nil, strings.TrimSpace(text)
	}

	// 提取选项块之前的纯净内容
	content := strings.TrimSpace(text[:loc[0]])

	// 提取选项块内容
	optionsBlock := text[loc[4]:loc[5]]

	// 解析选项列表
	options := parseOptionsBlock(optionsBlock)

	// 限制选项数量（最多10个）
	if len(options) > 10 {
		options = options[:10]
	}

	return options, content
}

// parseOptionsBlock 解析选项块内容，提取选项列表
func parseOptionsBlock(block string) []string {
	var options []string

	lines := strings.Split(block, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 匹配选项行
		matches := optionLineRegex.FindStringSubmatch(line)
		if len(matches) >= 3 {
			optionText := strings.TrimSpace(matches[2])
			if optionText != "" {
				options = append(options, optionText)
			}
		}
	}

	return options
}
