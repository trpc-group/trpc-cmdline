[English](README.md) | 中文

# 文档

## 设计实现

`trpc-cmdline` 工具依赖于 [Protobuf](https://protobuf.dev/)，以 [`proto`](https://github.com/protocolbuffers/protobuf) 文件为中间媒介，借助 [protoc](https://grpc.io/docs/protoc-installation/) 以生成含有数据结构定义的桩代码，借助于 [Go 模板](https://pkg.go.dev/text/template)以生成含有服务定义的桩代码。

其中：

* 含有数据结构定义的桩代码带有 `pb.go` 后缀
* 含有服务定义的桩代码带有 `trpc.go` 后缀

## 代码模板

`trpc-cmdline` 将指定的 `proto` 文件进行解析之后，会根据相应语言的代码模板进行桩代码生成。

这些代码模板位于 `install/protobuf/asset_${language}` 目录下，如 [install/protobuf/asset_go](/install/protobuf/asset_go/), [install/protobuf/asset_cpp](/install/protobuf/asset_cpp/) 等。

代码模板中可以引用自定义的变量以及函数，比如：

* 可以在模板文件中通过 `{{.PackageName}}` 来引用 `FileDescriptor.PackageName` 的值
* 可以使用自定义函数 `title` 等: `{{hello | title}}` => `Hello`

推荐通过阅读已有的模板文件进行学习仿写。

通过指定 `--assetdir` 可以将生成时使用的模板文件夹替换为你所指定的路径，比如：

```bash
trpc create -p hello.proto --assetdir=~/.trpc-cmdline-assets/protobuf/asset_go
```

注意：这里的 `--assetdir` 需要指定一个绝对路径。

[install](/install/) 目录下的文件会跟随二进制在执行前自动解压到用户的 `~/.trpc-cmdline-assets/` 目录下，因此默认的模板路径为 `~/.trpc-cmdline-assets/protobuf/asset_go`

其他语言如 C++ 同理，需要额外指定 `--lang=cpp`，其默认的模板路径为 `~/.trpc-cmdline-assets/protobuf/asset_cpp`

## 使用示例

* [example1](examples/example-1/README.zh_CN.md) 展示了生成工程及桩代码的更多选项及细节，例如
  * 如何指定含有依赖的 `proto` 文件的代码生成
  * 如何指定工程的 go module name
  * 如何指定输出文件路径
  * 如何只生成桩代码
  * 如何生成其他协议（HTTP 等）桩代码
  * 如何生成 flatbuffers 桩代码
  * ...
* [example2](examples/example-2/README.zh_CN.md) 展示了如何使用 pb option 扩展功能，例如
  * 如何为服务名添加别名
  * 如何为字段添加自定义 tag
  * 如何生成 validate.pb.go 文件
  * 如何生成 swagger/openapi 文档
  * ...
