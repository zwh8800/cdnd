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
	Short: "D&D CLI game powered by LLM",
	Long: `cdnd is a command-line Dungeons & Dragons role-playing game
powered by Large Language Models (LLM).

Experience an interactive D&D adventure with AI as your Dungeon Master.
Supports multiple LLM providers including OpenAI, Anthropic Claude, and Ollama.`,
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

	// 全局标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cdnd/config.yaml)")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug mode")
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
