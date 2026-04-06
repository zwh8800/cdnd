package prompt

import (
	"reflect"
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

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		wantOptions []string
		wantContent string
	}{
		{
			name: "标准格式解析",
			text: `你站在城堡的大门前，守卫警惕地看着你。

==========
你的选择是：
  1. 与守卫交谈
  2. 尝试潜入
  3. 转身离开`,
			wantOptions: []string{"与守卫交谈", "尝试潜入", "转身离开"},
			wantContent: "你站在城堡的大门前，守卫警惕地看着你。",
		},
		{
			name: "无选项块",
			text: `你站在城堡的大门前，守卫警惕地看着你。

你想怎么做？`,
			wantOptions: nil,
			wantContent: "你站在城堡的大门前，守卫警惕地看着你。\n\n你想怎么做？",
		},
		{
			name: "空选项列表",
			text: `你站在城堡的大门前。

==========
你的选择是：
`,
			wantOptions: nil,
			wantContent: "你站在城堡的大门前。\n\n==========\n你的选择是：",
		},
		{
			name: "选项数量过多",
			text: `场景描述。

==========
你的选择是：
  1. 选项1
  2. 选项2
  3. 选项3
  4. 选项4
  5. 选项5
  6. 选项6
  7. 选项7
  8. 选项8
  9. 选项9
  10. 选项10
  11. 选项11
  12. 选项12`,
			wantOptions: []string{"选项1", "选项2", "选项3", "选项4", "选项5", "选项6", "选项7", "选项8", "选项9", "选项10"},
			wantContent: "场景描述。",
		},
		{
			name: "特殊字符选项",
			text: `战斗场景。

==========
你的选择是：
  1. 攻击（造成1d8伤害）
  2. 施放"火球术"！
  3. 逃跑...`,
			wantOptions: []string{"攻击（造成1d8伤害）", "施放\"火球术\"！", "逃跑..."},
			wantContent: "战斗场景。",
		},
		{
			name: "等号数量变体",
			text: `场景描述。

=====
你的选择是：
  1. 选项A
  2. 选项B`,
			wantOptions: []string{"选项A", "选项B"},
			wantContent: "场景描述。",
		},
		{
			name: "更多等号变体",
			text: `场景描述。

==================
你的选择是：
  1. 选项A`,
			wantOptions: []string{"选项A"},
			wantContent: "场景描述。",
		},
		{
			name: "选项块后有内容",
			text: `场景描述。

==========
你的选择是：
  1. 选项A
  2. 选项B

这是选项块后的内容。`,
			wantOptions: []string{"选项A", "选项B"},
			wantContent: "场景描述。",
		},
		{
			name: "单行选项",
			text: `场景描述。

==========
你的选择是：
  1. 只有一个选项`,
			wantOptions: []string{"只有一个选项"},
			wantContent: "场景描述。",
		},
		{
			name:        "空文本",
			text:        "",
			wantOptions: nil,
			wantContent: "",
		},
		{
			name: "只有选项块",
			text: `==========
你的选择是：
  1. 选项A
  2. 选项B`,
			wantOptions: []string{"选项A", "选项B"},
			wantContent: "",
		},
		{
			name: "选项带前导空格",
			text: `场景描述。

==========
你的选择是：
1. 选项A
  2. 选项B
    3. 选项C`,
			wantOptions: []string{"选项A", "选项B", "选项C"},
			wantContent: "场景描述。",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOptions, gotContent := ParseOptions(tt.text)

			if !reflect.DeepEqual(gotOptions, tt.wantOptions) {
				t.Errorf("ParseOptions() gotOptions = %v, want %v", gotOptions, tt.wantOptions)
			}

			if gotContent != tt.wantContent {
				t.Errorf("ParseOptions() gotContent = %q, want %q", gotContent, tt.wantContent)
			}
		})
	}
}

func TestParseOptionsBlock(t *testing.T) {
	tests := []struct {
		name  string
		block string
		want  []string
	}{
		{
			name: "标准选项块",
			block: `  1. 选项A
  2. 选项B
  3. 选项C`,
			want: []string{"选项A", "选项B", "选项C"},
		},
		{
			name: "带空行的选项块",
			block: `  1. 选项A

  2. 选项B

  3. 选项C`,
			want: []string{"选项A", "选项B", "选项C"},
		},
		{
			name: "无编号行",
			block: `这是普通文本
没有编号`,
			want: nil,
		},
		{
			name: "空选项行",
			block: `1. 选项A
2.
3. 选项C`,
			want: []string{"选项A", "选项C"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOptionsBlock(tt.block)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseOptionsBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}
