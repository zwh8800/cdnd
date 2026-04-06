package prompt

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

// ColorMarkerStyles 定义颜色标记与lipgloss样式的映射
var ColorMarkerStyles = map[string]lipgloss.Style{
	"number":  lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")),              // 绿色
	"keyword": lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")),              // 紫色
	"status":  lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D")),              // 黄色
	"combat":  lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true),   // 红色加粗
	"success": lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true),   // 绿色加粗
	"danger":  lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true),   // 红色加粗
	"quote":   lipgloss.NewStyle().Foreground(lipgloss.Color("#5C5CFF")).Italic(true), // 浅紫色斜体
}

// colorMarkerRegex 匹配颜色标记的正则表达式
var colorMarkerRegex = regexp.MustCompile(`\{\{(\w+):([^}]+)\}\}`)

// ParseColorMarkers 将文本中的颜色标记转换为带样式的文本
// 支持的标记：{{number:值}}, {{keyword:值}}, {{status:值}}, {{combat:值}}, {{success:值}}, {{danger:值}}, {{quote:值}}
func ParseColorMarkers(text string) string {
	return colorMarkerRegex.ReplaceAllStringFunc(text, func(match string) string {
		// 提取标记类型和内容
		submatches := colorMarkerRegex.FindStringSubmatch(match)
		if len(submatches) != 3 {
			return match
		}

		markerType := submatches[1]
		content := submatches[2]

		// 查找对应的样式
		if style, exists := ColorMarkerStyles[markerType]; exists {
			return style.Render(content)
		}

		// 未知标记类型，返回原始内容
		return content
	})
}
