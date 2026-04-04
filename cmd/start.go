package cmd

import (
	"context"
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
		character := finalModel.(ui.CharacterCreationModel).GetCharacter()
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
	},
}

// startGameCmd 直接开始游戏（使用已有角色）
var startGameCmd = &cobra.Command{
	Use:   "game",
	Short: "使用已有角色开始游戏",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()

		provider, err := llm.NewProvider(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating LLM provider: %v\n", err)
			os.Exit(1)
		}

		engine, err := game.NewEngine(cfg, provider)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating game engine: %v\n", err)
			os.Exit(1)
		}

		// 尝试加载存档
		if err := engine.LoadGame(startSaveSlot); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading save: %v\n", err)
			fmt.Println("Use 'cdnd start' to create a new character first.")
			os.Exit(1)
		}

		// 启动游戏 TUI
		gameModel := ui.NewGameModel(engine)
		p := tea.NewProgram(gameModel, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running game: %v\n", err)
			os.Exit(1)
		}

		// 游戏结束后保存
		if err := engine.SaveGame(startSaveSlot); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to save game: %v\n", err)
		}
	},
}

// testLLMCmd 测试 LLM 连接
var testLLMCmd = &cobra.Command{
	Use:   "test-llm",
	Short: "测试 LLM 连接",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()

		provider, err := llm.NewProvider(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating LLM provider: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("正在测试与 %s 的连接...\n", cfg.LLM.DefaultProvider)

		resp, err := provider.Generate(context.Background(), &llm.Request{
			Messages: []llm.Message{
				{Role: llm.RoleUser, Content: "Say 'Hello, adventurer!' in Chinese."},
			},
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("响应:")
		fmt.Println(resp.Content)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.AddCommand(startGameCmd)
	startCmd.AddCommand(testLLMCmd)

	startCmd.Flags().IntVarP(&startSaveSlot, "save-slot", "s", 1, "存档槽位编号（1-10）")
	startCmd.Flags().StringVarP(&startScenario, "scenario", "S", "default", "要游玩的剧本")
	startCmd.Flags().BoolVar(&skipCreation, "skip-creation", false, "跳过角色创建（用于测试）")
}
