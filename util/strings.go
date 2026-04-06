package util

import "strings"

// IndentLines 给每一行文本添加指定的前缀缩进
func IndentLines(text string, prefix string) string {
	lines := strings.Split(text, "\n")
	var result strings.Builder
	for _, line := range lines {
		if line != "" {
			result.WriteString(prefix + line + "\n")
		}
	}
	return result.String()
}
