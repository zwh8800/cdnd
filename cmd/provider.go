package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zwh8800/cdnd/internal/config"
)

// providerCmd 表示提供者命令组。
var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "管理 LLM 提供者",
	Long: `列出、测试和配置 LLM 提供者。

支持的提供者：openai、anthropic、ollama`,
}

// providerListCmd 表示提供者列表命令。
var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出可用的 LLM 提供者",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()

		fmt.Println("Available LLM Providers:")
		fmt.Println("========================")

		for name, provider := range cfg.LLM.Providers {
			defaultMarker := ""
			if name == cfg.LLM.DefaultProvider {
				defaultMarker = " (default)"
			}
			fmt.Printf("\n  %s%s\n", name, defaultMarker)
			fmt.Printf("    Model: %s\n", provider.Model)
			if provider.BaseURL != "" {
				fmt.Printf("    Base URL: %s\n", provider.BaseURL)
			}

			// 检查 API 密钥是否已配置
			if provider.APIKey != "" {
				fmt.Println("    API Key: configured")
			} else if name != "ollama" {
				fmt.Println("    API Key: not set")
			}
		}
	},
}

// providerTestCmd 表示提供者测试命令。
var providerTestCmd = &cobra.Command{
	Use:   "test <provider>",
	Short: "测试 LLM 提供者连接",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]
		cfg := config.Get()

		// 检查提供者是否存在
		provider, exists := cfg.LLM.Providers[providerName]
		if !exists {
			fmt.Fprintf(os.Stderr, "Unknown provider: %s\n", providerName)
			fmt.Fprintln(os.Stderr, "Available providers: openai, anthropic, ollama")
			os.Exit(1)
		}

		fmt.Printf("Testing connection to %s...\n", providerName)
		fmt.Printf("Model: %s\n", provider.Model)

		// TODO: 实现实际连接测试
		fmt.Println("\nConnection test not yet implemented. Coming soon!")
	},
}

// providerSetDefaultCmd 表示设置默认提供者命令。
var providerSetDefaultCmd = &cobra.Command{
	Use:   "set-default <provider>",
	Short: "设置默认 LLM 提供者",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]
		cfg := config.Get()

		// 检查提供者是否存在
		if _, exists := cfg.LLM.Providers[providerName]; !exists {
			fmt.Fprintf(os.Stderr, "Unknown provider: %s\n", providerName)
			os.Exit(1)
		}

		cfg.LLM.DefaultProvider = providerName

		if err := config.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Default provider set to: %s\n", providerName)
	},
}

func init() {
	rootCmd.AddCommand(providerCmd)
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerTestCmd)
	providerCmd.AddCommand(providerSetDefaultCmd)
}
