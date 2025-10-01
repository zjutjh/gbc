# gbc

## What is gbc?

gbc 取自哭泣少女乐队，是精弘网络 HTTP Api命令行生成工具。需和 mygo 搭配使用，可以简化 go Http 开发。

## How to install?

```shell
go install github.com/zjutjh/gbc@latest
```

## How to use?
```shell
gbc api {api}

// 创建带有对应Request参数位的API模板
gbc api {api} --uri
gbc api {api} --header
gbc api {api} --query
gbc api {api} --body

// 叠加使用
gbc api {api} --query --body
```

