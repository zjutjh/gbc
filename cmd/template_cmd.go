package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zjutjh/gbc/comm"
	"github.com/zjutjh/gbc/template"
)

var createCMDCmd = &cobra.Command{
	Use:   "cmd",
	Short: "创建Command模板",
	Long:  "创建Command模板",
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化模板
		cmdTemplate := template.CMDTemplate

		path, cmdName, packageName, err := comm.ParseKey(args[0], "cmd", "./cmd/", ".go")
		if err != nil {
			comm.OutputError("创建command错误: %s", err.Error())
			return
		}

		// 替换cmd模板
		cmdTemplate = strings.ReplaceAll(cmdTemplate, "{$PackageName}", packageName)
		cmdTemplate = strings.ReplaceAll(cmdTemplate, "{$CMDName}", cmdName)

		// 创建cmd文件
		err = os.WriteFile(path, []byte(cmdTemplate), 0644)
		if err != nil {
			comm.OutputError("创建command错误: %s", err.Error())
			return
		} else {
			comm.OutputLook("创建command[%s]成功, 请记得前往./register/cmd.go中进行必要的命令注册", path)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCMDCmd)
}
