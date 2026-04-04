package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zwh8800/cdnd/internal/config"
)

// configCmd 表示配置命令组。
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "管理配置",
	Long: `查看和管理 cdnd 配置设置。

配置存储在 ~/.cdnd/config.yaml`,
}

// configInitCmd 表示配置初始化命令。
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化配置文件",
	Long: `创建带有默认设置的新配置文件。

如果配置文件不存在，将创建 ~/.cdnd/config.yaml。`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.InitConfigFile(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
			os.Exit(1)
		}
		configPath, _ := config.GetConfigPath()
		fmt.Printf("配置文件已创建: %s\n", configPath)
	},
}

// configGetCmd 表示配置获取命令。
var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "获取配置值",
	Long: `根据键获取配置值。

如果未指定键，则显示所有配置。`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()

		if len(args) == 0 {
			// 显示所有配置
			printAllConfig(cfg)
			return
		}

		// 显示指定键的值
		key := args[0]
		value := viper.Get(key)
		if value == nil {
			fmt.Fprintf(os.Stderr, "未找到该键: %s\n", key)
			os.Exit(1)
		}
		fmt.Printf("%s: %v\n", key, value)
	},
}

// configSetCmd 表示配置设置命令。
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "设置配置值",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		viper.Set(key, value)

		if err := config.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("已设置 %s = %s\n", key, value)
	},
}

func printAllConfig(cfg *config.Config) {
	fmt.Println("当前配置:")
	fmt.Println("======================")
	fmt.Println("\nLLM 设置:")
	fmt.Printf("  默认提供者: %s\n", cfg.LLM.DefaultProvider)
	for name, provider := range cfg.LLM.Providers {
		fmt.Printf("\n  [%s]\n", name)
		fmt.Printf("    模型: %s\n", provider.Model)
		if provider.BaseURL != "" {
			fmt.Printf("    基础 URL: %s\n", provider.BaseURL)
		}
		fmt.Printf("    最大 Token: %d\n", provider.MaxTokens)
		fmt.Printf("    温度: %.2f\n", provider.Temperature)
	}

	fmt.Println("\n游戏设置:")
	fmt.Printf("  自动保存: %v\n", cfg.Game.Autosave)
	fmt.Printf("  自动保存间隔: %v\n", cfg.Game.AutosaveInterval)
	fmt.Printf("  最大历史轮数: %d\n", cfg.Game.MaxHistoryTurns)
	fmt.Printf("  语言: %s\n", cfg.Game.Language)

	fmt.Println("\n显示设置:")
	fmt.Printf("  打字机效果: %v\n", cfg.Display.TypewriterEffect)
	fmt.Printf("  打字速度: %v\n", cfg.Display.TypingSpeed)
	fmt.Printf("  彩色输出: %v\n", cfg.Display.ColorOutput)

	fmt.Println("\n高级设置:")
	fmt.Printf("  缓存启用: %v\n", cfg.Advanced.CacheEnabled)
	fmt.Printf("  缓存 TTL: %v\n", cfg.Advanced.CacheTTL)
	fmt.Printf("  日志级别: %s\n", cfg.Advanced.LogLevel)
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
}
