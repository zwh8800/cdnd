package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zwh8800/cdnd/internal/config"
	"github.com/zwh8800/cdnd/internal/llm"
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

		fmt.Println("可用的 LLM 提供者:")
		fmt.Println("========================")

		for name, provider := range cfg.LLM.Providers {
			defaultMarker := ""
			if name == cfg.LLM.DefaultProvider {
				defaultMarker = " (默认)"
			}
			fmt.Printf("\n  %s%s\n", name, defaultMarker)
			fmt.Printf("    模型: %s\n", provider.Model)
			if provider.BaseURL != "" {
				fmt.Printf("    基础 URL: %s\n", provider.BaseURL)
			}

			// 检查 API 密钥是否已配置
			if provider.APIKey != "" {
				fmt.Println("    API 密钥: 已配置")
			} else if name != "ollama" {
				fmt.Println("    API 密钥: 未设置")
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
		providerConfig, exists := cfg.LLM.Providers[providerName]
		if !exists {
			fmt.Fprintf(os.Stderr, "未知的提供者: %s\n", providerName)
			fmt.Fprintln(os.Stderr, "可用的提供者: openai, anthropic, ollama")
			os.Exit(1)
		}

		fmt.Printf("正在测试与 %s 的连接...\n", providerName)
		fmt.Printf("模型: %s\n", providerConfig.Model)

		// 创建 LLM 提供者实例
		provider, err := llm.NewProvider(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating LLM provider: %v\n", err)
			os.Exit(1)
		}

		// 发送测试请求
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
			fmt.Fprintf(os.Stderr, "未知的提供者: %s\n", providerName)
			os.Exit(1)
		}

		cfg.LLM.DefaultProvider = providerName

		if err := config.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("默认提供者已设置为: %s\n", providerName)
	},
}

func init() {
	rootCmd.AddCommand(providerCmd)
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerTestCmd)
	providerCmd.AddCommand(providerSetDefaultCmd)
}
