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

var loadSlot int

// loadCmd 表示加载命令。
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load a saved game",
	Long: `Load a previously saved game from a save slot.

Example:
  cdnd load --slot 1
  cdnd load -s 3`,
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

		// 加载存档
		fmt.Printf("Loading game from slot %d...\n", loadSlot)
		if err := engine.LoadGame(loadSlot); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading game: %v\n", err)
			os.Exit(1)
		}

		// 显示存档信息
		c := engine.GetCharacter()
		if c != nil {
			fmt.Printf("Loaded character: %s\n", c.Name)
		}

		// 启动游戏 TUI
		gameModel := ui.NewGameModel(engine)
		p := tea.NewProgram(gameModel, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running game: %v\n", err)
			os.Exit(1)
		}

		// 游戏结束后保存
		if err := engine.SaveGame(loadSlot); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to save game: %v\n", err)
		}
	},
}

// savesCmd 列出所有存档
var savesCmd = &cobra.Command{
	Use:   "saves",
	Short: "List all save slots",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()

		provider, err := llm.NewProvider(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		engine, err := game.NewEngine(cfg, provider)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		slots, err := engine.GetSaveSlots()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing saves: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Save Slots:")
		fmt.Println("===========")
		for _, slot := range slots {
			if slot.CharacterName != "" {
				fmt.Printf("\nSlot %d:\n", slot.Slot)
				fmt.Printf("  Character: %s (Level %d %s)\n", slot.CharacterName, slot.CharacterLevel, slot.CharacterClass)
				fmt.Printf("  Phase: %s\n", slot.Phase.String())
				fmt.Printf("  Location: %s\n", slot.Preview)
				fmt.Printf("  Play Time: %d minutes\n", slot.PlayTime/60)
				fmt.Printf("  Last Saved: %s\n", slot.UpdatedAt.Format("2006-01-02 15:04"))
			} else {
				fmt.Printf("\nSlot %d: [Empty]\n", slot.Slot)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(savesCmd)

	loadCmd.Flags().IntVarP(&loadSlot, "slot", "s", 1, "save slot number")
}
