package cmd

import (
	"net/http"
	"os"
	"os/exec"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	"github.com/zjutjh/gbc/comm"
	"github.com/zjutjh/gbc/config"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "检查并升级gbc自身版本",
	Long:  `检查并升级gbc自身版本`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取当前最新版本
		latestVersion := "v0.0.0"
		latestVersionC, _ := version.NewVersion(latestVersion)
		type Tag struct {
			Name string `json:"name"`
		}
		result := make([]Tag, 0)
		resp, err := resty.New().R().
			ForceContentType("application/json").
			SetResult(&result).
			Get(config.GithubGetTagsAPI)
		if err != nil {
			comm.OutputError("读取远程gbc版本错误: %s", err.Error())
		}
		if resp.StatusCode() != http.StatusOK {
			comm.OutputError("读取远程gbc版本错误: Status Code[%d|%s]", resp.StatusCode(), resp.Status())
		}
		for _, tag := range result {
			ver, err := version.NewVersion(tag.Name)
			if err != nil {
				continue
			}
			if ver.GreaterThan(latestVersionC) {
				latestVersion = tag.Name
				latestVersionC = ver
			}
		}

		myVersionC, _ := version.NewVersion(rootCmd.Version)
		if myVersionC.GreaterThanOrEqual(latestVersionC) {
			comm.OutputLook("当前gbc工具版本[%s]为最新版本", rootCmd.Version)
			return
		}

		// 需要版本升级
		comm.OutputLook("发现gbc工具新版本[%s], 当前本地版本[%s], 开始版本升级...", latestVersion, rootCmd.Version)
		c := exec.Command("go", "install", config.GBCForGoInstall)
		if comm.DebugMode {
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
		}
		err = c.Run()
		if err != nil {
			comm.OutputError("升级gbc工具版本失败: %s", err.Error())
			return
		}
		comm.OutputLook("版本升级完成")
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
