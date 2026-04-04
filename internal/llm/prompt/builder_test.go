package prompt

import (
	"strings"
	"testing"
)

func TestParseColorMarkers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string // 检查输出中是否包含这些子串（因为包含ANSI转义序列，所以只检查内容部分）
	}{
		{
			name:     "number marker",
			input:    "造成 {{number:15}} 点伤害",
			contains: []string{"15"},
		},
		{
			name:     "keyword marker",
			input:    "使用 {{keyword:火球术}} 攻击",
			contains: []string{"火球术"},
		},
		{
			name:     "status marker",
			input:    "目标 {{status:中毒}} 了",
			contains: []string{"中毒"},
		},
		{
			name:     "combat marker",
			input:    "{{combat:挥剑攻击}}！",
			contains: []string{"挥剑攻击"},
		},
		{
			name:     "success marker",
			input:    "{{success:命中！}}",
			contains: []string{"命中！"},
		},
		{
			name:     "danger marker",
			input:    "{{danger:攻击未命中}}",
			contains: []string{"攻击未命中"},
		},
		{
			name:     "quote marker",
			input:    "{{quote:你好，冒险者}}",
			contains: []string{"你好，冒险者"},
		},
		{
			name:     "multiple markers",
			input:    "{{success:命中！}} 造成 {{number:8}} 点伤害，目标 {{status:流血}}",
			contains: []string{"命中！", "8", "流血"},
		},
		{
			name:     "no markers",
			input:    "这是一段普通文本",
			contains: []string{"这是一段普通文本"},
		},
		{
			name:     "unknown marker type",
			input:    "{{unknown:内容}}",
			contains: []string{"内容"},
		},
		{
			name:     "empty content",
			input:    "{{number:}}",
			contains: []string{},
		},
		{
			name:     "nested brackets not allowed",
			input:    "{{number:{{keyword:5}}}}",
			contains: []string{"{{keyword:5"}, // 正则应该只匹配到第一个 }}
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseColorMarkers(tt.input)

			// 检查是否包含预期的内容
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("ParseColorMarkers() output = %q, should contain %q", result, expected)
				}
			}

			// 验证原始标记已被替换（不应该再包含 {{ 或 }}）
			if strings.Contains(result, "{{") || strings.Contains(result, "}}") {
				// 除非是测试未知标记类型或特殊情况
				if tt.name != "unknown marker type" && tt.name != "nested brackets not allowed" && tt.name != "empty content" {
					t.Errorf("ParseColorMarkers() output still contains markers: %q", result)
				}
			}
		})
	}
}

func TestParseColorMarkersPreservesText(t *testing.T) {
	input := "你攻击了 {{keyword:哥布林}}，造成 {{number:12}} 点伤害！"
	result := ParseColorMarkers(input)

	// 检查文本的其他部分是否被保留
	if !strings.Contains(result, "你攻击了") {
		t.Error("ParseColorMarkers() should preserve text before markers")
	}
	if !strings.Contains(result, "，造成") {
		t.Error("ParseColorMarkers() should preserve text between markers")
	}
	if !strings.Contains(result, "点伤害！") {
		t.Error("ParseColorMarkers() should preserve text after markers")
	}
}

func TestColorMarkerStyles(t *testing.T) {
	// 验证所有预期的标记类型都有对应的样式
	expectedTypes := []string{"number", "keyword", "status", "combat", "success", "danger", "quote"}

	for _, markerType := range expectedTypes {
		if _, exists := ColorMarkerStyles[markerType]; !exists {
			t.Errorf("ColorMarkerStyles missing style for marker type: %s", markerType)
		}
	}
}
