package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/zwh8800/cdnd/application/engine"
	"github.com/zwh8800/cdnd/infrastructure/config"
	llm2 "github.com/zwh8800/cdnd/infrastructure/llm"
	"github.com/zwh8800/cdnd/interface/ui"
)

var loadSlot int
var loadNoAutosave bool

// loadCmd 表示加载命令。
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "加载已保存的游戏",
	Long: `从存档槽位加载之前保存的游戏。

示例:
  cdnd load --slot 1
  cdnd load -s 3`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()

		// 命令行选项覆盖自动保存设置
		if loadNoAutosave {
			cfg.Game.Autosave = false
		}

		// 获取 LLM 提供者
		provider, err := llm2.NewProvider(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating LLM provider: %v\n", err)
			os.Exit(1)
		}

		// 创建游戏引擎
		engine, err := engine.NewEngine(cfg, provider)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating game engine: %v\n", err)
			os.Exit(1)
		}

		// 确保游戏退出时清理自动保存资源
		defer engine.StopAutosave()

		// 加载存档
		fmt.Printf("正在从槽位 %d 加载游戏...\n", loadSlot)
		if err := engine.LoadGame(loadSlot); err != nil {
			fmt.Fprintf(os.Stderr, "加载游戏失败: %v\n", err)
			os.Exit(1)
		}

		// 显示存档信息
		c := engine.GetCharacter()
		if c != nil {
			fmt.Printf("已加载角色: %s\n", c.Name)
		}

		// 启动游戏 TUI
		gameModel := ui.NewGameModel(engine)
		gameP := tea.NewProgram(gameModel, tea.WithAltScreen())

		if _, err := gameP.Run(); err != nil {
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
	Short: "列出所有存档槽位",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()

		provider, err := llm2.NewProvider(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		engine, err := engine.NewEngine(cfg, provider)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		slots, err := engine.GetSaveSlots()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing saves: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("存档槽位:")
		fmt.Println("===========")
		for _, slot := range slots {
			if slot.CharacterName != "" {
				fmt.Printf("\n槽位 %d:\n", slot.Slot)
				fmt.Printf("  角色: %s (等级 %d %s)\n", slot.CharacterName, slot.CharacterLevel, slot.CharacterClass)
				fmt.Printf("  阶段: %s\n", slot.Phase.String())
				fmt.Printf("  位置: %s\n", slot.Preview)
				fmt.Printf("  游玩时间: %d 分钟\n", slot.PlayTime/60)
				fmt.Printf("  最后保存: %s\n", slot.UpdatedAt.Format("2006-01-02 15:04"))
			} else {
				fmt.Printf("\n槽位 %d: [空]\n", slot.Slot)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(savesCmd)

	loadCmd.Flags().IntVarP(&loadSlot, "slot", "s", 1, "存档槽位编号")
	loadCmd.Flags().BoolVar(&loadNoAutosave, "no-autosave", false, "禁用自动保存")
}
