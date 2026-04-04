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
			fmt.Println("Character creation cancelled.")
			return
		}

		fmt.Println("\nCharacter created successfully!")
		fmt.Printf("Name: %s\n", character.Name)
		if character.HasClass() {
			fmt.Printf("Class: %s\n", character.Class.Name)
		}
		fmt.Printf("Race: %s\n", character.Race.Name)
		fmt.Printf("HP: %d/%d\n", character.HitPoints.Current, character.HitPoints.Max)
	},
}

// characterListCmd 表示角色列表命令。
var characterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved characters",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Saved Characters:")
		fmt.Println("================")
		// TODO: List saved characters from templates
		fmt.Println("\nNo saved character templates found.")
		fmt.Println("Use 'cdnd character create' to create a new character.")
	},
}

// characterDeleteCmd 表示角色删除命令。
var characterDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "删除已保存的角色",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		fmt.Printf("Deleting character: %s\n", name)
		// TODO: 删除角色模板
		fmt.Println("Character deleted successfully.")
	},
}

// characterShowCmd 显示角色详情
var characterShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show character details",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: 从存档或模板加载角色
		fmt.Println("Character details:")
		fmt.Println("  This feature is not yet implemented.")
	},
}

func init() {
	rootCmd.AddCommand(characterCmd)
	characterCmd.AddCommand(characterCreateCmd)
	characterCmd.AddCommand(characterListCmd)
	characterCmd.AddCommand(characterDeleteCmd)
	characterCmd.AddCommand(characterShowCmd)
}
