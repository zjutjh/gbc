package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/zjutjh/gbc/comm"
)

var rootCmd = &cobra.Command{
	Use:   "gbc",
	Short: "精弘网络本地开发者工具",
	Long:  "精弘网络本地开发者工具",
	Run: func(cmd *cobra.Command, args []string) {
		comm.OutputInfo("当前gbc工具本地版本号: %s", cmd.Version)
		_ = cmd.Help()
	},
	Version: "v1.2.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		comm.OutputError("执行发生错误: %s", err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&comm.DebugMode, "debug", "d", false, "展示更多过程信息进行调试")
}
