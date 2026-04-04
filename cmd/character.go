package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/zwh8800/cdnd/internal/ui"
)

// characterCmd 表示角色命令组。
var characterCmd = &cobra.Command{
	Use:   "character",
	Short: "管理角色",
	Long: `创建、列出和管理你的 D&D 角色。

角色可以保存为模板，用于开始新游戏。`,
}

// characterCreateCmd 表示角色创建命令。
var characterCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建新角色",
	Long: `交互式创建新的 D&D 角色。

这将引导你选择种族、职业、能力等。`,
	Run: func(cmd *cobra.Command, args []string) {
		model := ui.NewCharacterCreationModel()
		p := tea.NewProgram(model, tea.WithAltScreen())

		finalModel, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		character := finalModel.(ui.CharacterCreationModel).GetCharacter()
		if character == nil {
			fmt.Println("角色创建已取消。")
			return
		}

		fmt.Println("\n角色创建成功！")
		fmt.Printf("姓名: %s\n", character.Name)
		if character.HasClass() {
			fmt.Printf("职业: %s\n", character.Class.Name)
		}
		fmt.Printf("种族: %s\n", character.Race.Name)
		fmt.Printf("HP: %d/%d\n", character.HitPoints.Current, character.HitPoints.Max)
	},
}

// characterListCmd 表示角色列表命令。
var characterListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有已保存的角色",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("已保存的角色:")
		fmt.Println("================")
		// TODO: List saved characters from templates
		fmt.Println("\n未找到已保存的角色模板。")
		fmt.Println("使用 'cdnd character create' 创建新角色。")
	},
}

// characterDeleteCmd 表示角色删除命令。
var characterDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "删除已保存的角色",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		fmt.Printf("正在删除角色: %s\n", name)
		// TODO: 删除角色模板
		fmt.Println("角色已删除。")
	},
}

// characterShowCmd 显示角色详情
var characterShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "显示角色详情",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: 从存档或模板加载角色
		fmt.Println("角色详情:")
		fmt.Println("  此功能尚未实现。")
	},
}

func init() {
	rootCmd.AddCommand(characterCmd)
	characterCmd.AddCommand(characterCreateCmd)
	characterCmd.AddCommand(characterListCmd)
	characterCmd.AddCommand(characterDeleteCmd)
	characterCmd.AddCommand(characterShowCmd)
}
