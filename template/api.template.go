package template

var APITemplate = `package {$PackageName}

import (
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

// {$ApiStruct}Handler API router注册点
func {$ApiStruct}Handler() gin.HandlerFunc {
	api := {$ApiStruct}Api{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hf{$ApiStruct}).Pointer()).Name()] = api
	return hf{$ApiStruct}
}

type {$ApiStruct}Api struct {
	{$ApiInfo}
	Request  {$ApiStruct}ApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response {$ApiStruct}ApiResponse // API响应数据 (Body中的Data部分)
}

type {$ApiStruct}ApiRequest struct {{$RequestUri}{$RequestHeader}{$RequestQuery}{$RequestBody}
}

type {$ApiStruct}ApiResponse struct {}

// Run Api业务逻辑执行点
func ({$Receiver} *{$ApiStruct}Api) Run(ctx *gin.Context) kit.Code {
	// 在此编写具体接口业务逻辑
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func ({$Receiver} *{$ApiStruct}Api) Init(ctx *gin.Context) (err error) {{$RequestUriInit}{$RequestHeaderInit}{$RequestQueryInit}{$RequestBodyInit}
	return err
}

//  hf{$ApiStruct} API执行入口
func hf{$ApiStruct}(ctx *gin.Context) {
	api := &{$ApiStruct}Api{}
	err := api.Init(ctx)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("参数绑定校验错误")
		reply.Fail(ctx, comm.CodeParameterInvalid)
		return
	}
	code := api.Run(ctx)
	if !ctx.IsAborted() {
		if code == comm.CodeOK {
			reply.Success(ctx, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
`
