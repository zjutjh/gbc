package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zjutjh/gbc/comm"
	"github.com/zjutjh/gbc/template"
)

var Body bool
var Query bool
var Header bool
var Uri bool

var apiCreateCmd = &cobra.Command{
	Use:   "api",
	Short: "创建API模版",
	Long:  `创建API模版`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化模板
		apiTemplate := template.APITemplate

		if !Body {
			comm.UI("接口是否存在body参数? (y|n(default)):", "n", func(b bool) {
				Body = b
			})
		}
		if !Query {
			comm.UI("接口是否存在query参数? (y|n(default)):", "n", func(b bool) {
				Query = b
			})
		}
		if !Header {
			comm.UI("接口是否存在header参数? (y|n(default)):", "n", func(b bool) {
				Header = b
			})
		}
		if !Uri {
			comm.UI("接口是否存在uri参数? (y|n(default)):", "n", func(b bool) {
				Uri = b
			})
		}

		if Body {
			apiTemplate = strings.Replace(apiTemplate, "{$RequestBody}", "\n\tBody struct {}", 1)
			apiTemplate = strings.Replace(apiTemplate, "{$RequestBodyInit}", "\n\terr = ctx.ShouldBindJSON(&{$Receiver}.Request.Body)\n\tif err != nil {\n\t\treturn err\n\t}", 1)
		} else {
			apiTemplate = strings.Replace(apiTemplate, "{$RequestBody}", "", 1)
			apiTemplate = strings.Replace(apiTemplate, "{$RequestBodyInit}", "", 1)
		}
		if Query {
			apiTemplate = strings.Replace(apiTemplate, "{$RequestQuery}", "\n\tQuery struct {}", 1)
			apiTemplate = strings.Replace(apiTemplate, "{$RequestQueryInit}", "\n\terr = ctx.ShouldBindQuery(&{$Receiver}.Request.Query)\n\tif err != nil {\n\t\treturn err\n\t}", 1)
		} else {
			apiTemplate = strings.Replace(apiTemplate, "{$RequestQuery}", "", 1)
			apiTemplate = strings.Replace(apiTemplate, "{$RequestQueryInit}", "", 1)
		}
		if Header {
			apiTemplate = strings.Replace(apiTemplate, "{$RequestHeader}", "\n\tHeader struct {}", 1)
			apiTemplate = strings.Replace(apiTemplate, "{$RequestHeaderInit}", "\n\terr = ctx.ShouldBindHeader(&{$Receiver}.Request.Header)\n\tif err != nil {\n\t\treturn err\n\t}", 1)
		} else {
			apiTemplate = strings.Replace(apiTemplate, "{$RequestHeader}", "", 1)
			apiTemplate = strings.Replace(apiTemplate, "{$RequestHeaderInit}", "", 1)
		}
		if Uri {
			apiTemplate = strings.Replace(apiTemplate, "{$RequestUri}", "\n\tUri struct {}", 1)
			apiTemplate = strings.Replace(apiTemplate, "{$RequestUriInit}", "\n\terr = ctx.ShouldBindUri(&{$Receiver}.Request.Uri)\n\tif err != nil {\n\t\treturn err\n\t}", 1)
		} else {
			apiTemplate = strings.Replace(apiTemplate, "{$RequestUri}", "", 1)
			apiTemplate = strings.Replace(apiTemplate, "{$RequestUriInit}", "", 1)
		}
		path, apiName, packageName, err := comm.ParseKey(args[0], "api", "./api/", ".go")
		if err != nil {
			comm.OutputError("创建API错误: %s", err.Error())
			return
		}
		receiverName := strings.ToLower(string(apiName[0]))

		// 替换api模板
		apiApiContent := strings.ReplaceAll(apiTemplate, "{$PackageName}", packageName)
		apiApiContent = strings.ReplaceAll(apiApiContent, "{$ApiStruct}", apiName)
		apiApiContent = strings.ReplaceAll(apiApiContent, "{$Receiver}", receiverName)
		apiApiContent = strings.ReplaceAll(apiApiContent, "{$ApiInfo}", "Info     struct{}        `name:\"API名称\" desc:\"API描述\"`")

		// 创建api文件
		err = os.WriteFile(path, []byte(apiApiContent), 0644)
		if err != nil {
			comm.OutputError("创建API错误: %s", err.Error())
		} else {
			comm.OutputLook("创建API[%s]成功, 请记得前往./router/router.go中进行必要的API注册", path)
		}
	},
}

func init() {
	apiCreateCmd.Flags().BoolVarP(&Body, "body", "", false, "With Request Body")
	apiCreateCmd.Flags().BoolVarP(&Query, "query", "", false, "With Request Query")
	apiCreateCmd.Flags().BoolVarP(&Header, "header", "", false, "With Request Header")
	apiCreateCmd.Flags().BoolVarP(&Uri, "uri", "", false, "With Request Uri")
	rootCmd.AddCommand(apiCreateCmd)
}
