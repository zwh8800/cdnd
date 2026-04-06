package prompt

import (
	"reflect"
	"testing"
)

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
