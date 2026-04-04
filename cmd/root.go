// Package cmd 包含 cdnd 的 CLI 命令。
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zwh8800/cdnd/internal/config"
)

var (
	// 版本信息（构建时由 ldflags 设置）
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"

	// cfgFile 是配置文件的路径
	cfgFile string
)

// rootCmd 表示不带子命令调用时的基础命令。
var rootCmd = &cobra.Command{
	Use:   "cdnd",
	Short: "由 LLM 驱动的 D&D 命令行游戏",
	Long: `cdnd 是一款由大语言模型（LLM）驱动的命令行龙与地下城角色扮演游戏。

体验与 AI 地下城主互动的互动式 D&D 冒险。
支持多种 LLM 提供商，包括 OpenAI、Anthropic Claude 和 Ollama。`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute 将所有子命令添加到根命令并适当设置标志。
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// 禁用 completion 命令
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// 设置 help 命令描述为中文并隐藏
	helpCmd := &cobra.Command{
		Use:    "help [command]",
		Short:  "查看命令帮助信息",
		Long:   "查看指定命令的详细帮助信息和使用说明。",
		Hidden: true,
	}
	rootCmd.SetHelpCommand(helpCmd)

	// 全局标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径（默认为 $HOME/.cdnd/config.yaml）")
	rootCmd.PersistentFlags().Bool("debug", false, "启用调试模式")
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

// initConfig 读取配置文件和环境变量（如果已设置）。
func initConfig() {
	if cfgFile != "" {
		// 使用标志指定的配置文件
		viper.SetConfigFile(cfgFile)
	} else {
		// 查找主目录
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding home directory: %v\n", err)
			os.Exit(1)
		}

		// 在主目录中搜索配置文件
		viper.AddConfigPath(home + "/.cdnd")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // 读取匹配的环境变量

	// 如果找到配置文件，则读取
	if err := viper.ReadInConfig(); err == nil {
		// 配置文件已找到并成功解析
	}
}
