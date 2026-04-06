package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// versionCmd 表示版本命令。
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Long:  `显示 cdnd 的版本号、Git 提交和构建信息。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("cdnd %s\n", Version)
		fmt.Printf("  Git 提交: %s\n", GitCommit)
		fmt.Printf("  构建日期: %s\n", BuildDate)
		fmt.Printf("  Go 版本: %s\n", runtime.Version())
		fmt.Printf("  系统/架构: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
