package template

var CronTemplate = `package {$PackageName}

import (
	"github.com/zjutjh/mygo/config"
	"github.com/zjutjh/mygo/nlog"
)

type {$CronName}Job struct{}

func ({$CronName}Job) Run() {
	// 在此处编写定时任务业务逻辑
	nlog.Pick().WithField("app", config.AppName()).Debug("定时任务运行")
}
`
