package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zjutjh/gbc/comm"
	"github.com/zjutjh/gbc/template"
)

var createCronCmd = &cobra.Command{
	Use:   "cron",
	Short: "创建Cron模版",
	Long:  `创建Cron模版`,
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化模板
		cronTemplate := template.CronTemplate

		path, cronName, packageName, err := comm.ParseKey(args[0], "cron", "./cron/", ".go")
		if err != nil {
			comm.OutputError("创建cron错误: %s", err.Error())
			return
		}

		// 替换cron模板
		cronTemplate = strings.ReplaceAll(cronTemplate, "{$PackageName}", packageName)
		cronTemplate = strings.ReplaceAll(cronTemplate, "{$CronName}", cronName)

		// 创建cron文件
		err = os.WriteFile(path, []byte(cronTemplate), 0644)
		if err != nil {
			comm.OutputError("创建cron错误: %s", err.Error())
			return
		} else {
			comm.OutputLook("创建cron[%s]成功, 请记得前往./register/cron.go中进行必要的定时任务注册", path)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCronCmd)
}
