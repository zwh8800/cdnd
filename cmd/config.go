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
		fmt.Printf("Configuration file created: %s\n", configPath)
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
			fmt.Fprintf(os.Stderr, "Key not found: %s\n", key)
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

		fmt.Printf("Set %s = %s\n", key, value)
	},
}

func printAllConfig(cfg *config.Config) {
	fmt.Println("Current Configuration:")
	fmt.Println("======================")
	fmt.Println("\nLLM Settings:")
	fmt.Printf("  Default Provider: %s\n", cfg.LLM.DefaultProvider)
	for name, provider := range cfg.LLM.Providers {
		fmt.Printf("\n  [%s]\n", name)
		fmt.Printf("    Model: %s\n", provider.Model)
		if provider.BaseURL != "" {
			fmt.Printf("    Base URL: %s\n", provider.BaseURL)
		}
		fmt.Printf("    Max Tokens: %d\n", provider.MaxTokens)
		fmt.Printf("    Temperature: %.2f\n", provider.Temperature)
	}

	fmt.Println("\nGame Settings:")
	fmt.Printf("  Autosave: %v\n", cfg.Game.Autosave)
	fmt.Printf("  Autosave Interval: %v\n", cfg.Game.AutosaveInterval)
	fmt.Printf("  Max History Turns: %d\n", cfg.Game.MaxHistoryTurns)
	fmt.Printf("  Language: %s\n", cfg.Game.Language)

	fmt.Println("\nDisplay Settings:")
	fmt.Printf("  Typewriter Effect: %v\n", cfg.Display.TypewriterEffect)
	fmt.Printf("  Typing Speed: %v\n", cfg.Display.TypingSpeed)
	fmt.Printf("  Color Output: %v\n", cfg.Display.ColorOutput)

	fmt.Println("\nAdvanced Settings:")
	fmt.Printf("  Cache Enabled: %v\n", cfg.Advanced.CacheEnabled)
	fmt.Printf("  Cache TTL: %v\n", cfg.Advanced.CacheTTL)
	fmt.Printf("  Log Level: %s\n", cfg.Advanced.LogLevel)
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
}
