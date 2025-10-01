package cmd

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/zjutjh/gbc/comm"
)

const gitGBCTemplate = "git@github.com:zjutjh/gbc-template.git"

var projectCreateCmd = &cobra.Command{
	Use:     "new",
	Short:   "创建新项目模板",
	Long:    "创建新项目模板",
	Example: "gbc new {app} [path]",
	Args:    cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		// 设置默认路径
		path, err := os.Getwd()
		if err != nil {
			comm.OutputError("获取当前工作目录失败: %s", err.Error())
			return
		}
		if len(args) > 1 {
			path, err = filepath.Abs(args[1])
			if err != nil {
				comm.OutputError("获取path[%s]绝对路径失败: %s", args[1], err.Error())
				return
			}
		}

		// 项目创建路径
		projectPath := filepath.Join(path, args[0])
		comm.OutputDebug("创建项目[%s]到目录[%s]开始...", args[0], projectPath)

		// 拉取框架模板
		c := exec.Command("git", "clone", gitGBCTemplate, projectPath)
		err = c.Run()
		if err != nil {
			comm.OutputError("创建项目[%s]失败: 拉取gbc-template错误: %s", args[0], err.Error())
			return
		}

		// 删除.git目录
		os.RemoveAll(filepath.Join(projectPath, ".git"))

		comm.OutputDebug("创建项目[%s]成功", args[0])
	},
}

func init() {
	rootCmd.AddCommand(projectCreateCmd)
}
