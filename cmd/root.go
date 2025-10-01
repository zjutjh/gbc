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
		_ = cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		comm.OutputError("执行发生错误: %s", err.Error())
		os.Exit(1)
	}
}
