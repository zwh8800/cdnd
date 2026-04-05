package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/zwh8800/cdnd/internal/config"
	"github.com/zwh8800/cdnd/internal/game"
	"github.com/zwh8800/cdnd/internal/llm"
	"github.com/zwh8800/cdnd/internal/ui"
)

var (
	startSaveSlot int
	startScenario string
	skipCreation  bool
	noAutosave    bool
)

// startCmd 表示开始命令。
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "开始新游戏",
	Long: `开始一场新的 D&D 冒险游戏。

你可以指定存档槽位和剧本。如果未指定，游戏将使用默认设置。
使用 --skip-creation 可跳过角色创建（用于测试）。`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()

		// 命令行选项覆盖自动保存设置
		if noAutosave {
			cfg.Game.Autosave = false
		}

		// 获取 LLM 提供者
		provider, err := llm.NewProvider(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating LLM provider: %v\n", err)
			os.Exit(1)
		}

		// 创建游戏引擎
		engine, err := game.NewEngine(cfg, provider)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating game engine: %v\n", err)
			os.Exit(1)
		}

		// 确保游戏退出时清理自动保存资源
		defer engine.StopAutosave()

		// 如果跳过角色创建，使用测试角色
		if skipCreation {
			// TODO: 创建测试角色
			fmt.Println("跳过创建模式 - 尚未实现")
			return
		}

		// 启动角色创建 TUI
		creationModel := ui.NewCharacterCreationModel()
		p := tea.NewProgram(creationModel, tea.WithAltScreen())

		finalModel, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running character creation: %v\n", err)
			os.Exit(1)
		}

		// 获取创建的角色
		character := finalModel.(*ui.CharacterCreationModel).GetCharacter()
		if character == nil {
			fmt.Println("Character creation cancelled.")
			return
		}

		// 开始游戏
		if err := engine.Start(character); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting game: %v\n", err)
			os.Exit(1)
		}

		// 保存初始存档
		if err := engine.SaveGame(startSaveSlot); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to save game: %v\n", err)
		}

		// 启动游戏 TUI
		gameModel := ui.NewGameModel(engine)
		gameP := tea.NewProgram(gameModel, tea.WithAltScreen())

		if _, err := gameP.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running game: %v\n", err)
			os.Exit(1)
		}

		// 游戏结束后保存
		if err := engine.SaveGame(startSaveSlot); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to save game: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().IntVarP(&startSaveSlot, "save-slot", "s", 1, "存档槽位编号（1-10）")
	startCmd.Flags().StringVarP(&startScenario, "scenario", "S", "default", "要游玩的剧本")
	startCmd.Flags().BoolVar(&skipCreation, "skip-creation", false, "跳过角色创建（用于测试）")
	startCmd.Flags().BoolVar(&noAutosave, "no-autosave", false, "禁用自动保存")
}
