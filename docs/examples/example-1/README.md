example-1 工程创建
==============================================================================

# 选项介绍

这里介绍下如何使用 trpc 命令来创建工程，首先查看下`trpc create`创建工程时的常用选项。可以看到有非常多的选项，不要被吓到，绝大多数情况下`trpc create -p <pb 文件>`就足够了。

```bash
$ trpc help create
指定 pb 文件快速创建工程或 rpcstub，

'trpc create' 有两种模式：
- 生成一个完整的服务工程
- 生成被调服务的 rpcstub，需指定'--rpconly'选项。

Usage:
  trpc create [flags]

Flags:
      --alias                  rpcname 采用别名模式
      --assetdir string        允许自定义模板 & 指定模板路径
  -f, --force                  强制覆盖已经存在的代码
      --gotag                  生成自定义 pb struct tag (default true)
  -h, --help                   help for create
      --mock go generate       生成 mock 桩代码（工程下可运行 go generate 更新） (default true)
  -m, --mod string             指定 go module, 默认生成 module 名为：trpc.app.${pb.package}
      --openapi                生成 openapi 3.0 文档
  -o, --output string          指定输出目录（生成完整工程默认输出到名为 pb 文件的目录，rpconly 默认当前目录
      --protocol string        指定使用的协议类型，支持：trpc, http, etc (default "trpc")
  -d, --protodir stringArray   pb 文件（含依赖 pb 文件）的搜索路径，允许指定多次 (default [.])
  -p, --protofile string       指定服务对应的 pb 文件
      --rpconly                只生成桩代码（建议在 stub 目录下执行），可以配合-o 使用
  -s, --split-by-method        是否支持按方法分文件
      --swagger                生成 swagger api 文档
      --swagger-json-param     生成 swagger api 文档请求参数，使用 json body
      --swagger-out string     生成 swagger api 文档 (default "apidocs.swagger.json")
  -v, --verbose                显示详细日志信息
```

# 1 个 pb 文件创建工程

推荐使用 pb 文件作为 IDL 来描述服务接口，这个对公司内同学特别是后台同学来说，应该不陌生。

一个服务的 pb 文件通常包含了 service 定义，rpc 定义，message 定义等，我们先来看一个 pb 文件就可以描述完整服务接口的情形，这种最简单。

file: hello.proto

```
syntax = "proto3";
package hello;

option go_package="trpc.group/anygroup/protocols/hello";

// hello request
message Req{}

// hello response
message Rsp{}

// hello service
service hello_service {
    // Hello method
    rpc Hello(Req) returns(Rsp);
}
```

执行 trpc 命令创建工程：

```bash
$ trpc create -p hello.proto
[create] 创建 trpc 工程 ```hello``` 代码生成成功
[create] 创建 trpc 工程 ```hello``` 后处理成功
```

trpc 命令生成了一个完整的工程，默认输出都和 pb 文件同名的目录中，生成的目录结构如下所示。

- main.go 是服务入口；
- hello_service.go 是服务接口实现，对应单元测试文件也已生成；
- stub 下面是 pb 文件对应的桩代码部分，包括*.pb.go、*.trpc.go、*.mock.go；
- 注意查看下 go.mod 中的 module，一般是按照 trpc.app.$service 来进行命名的；

```bash
hello
├── go.mod
├── go.sum
├── hello_service.go
├── hello_service_test.go
├── main.go
├── stub
│   └── trpc.group
│       └── anygroup
│           └── protocols
│               └── hello
│                   ├── go.mod
│                   ├── hello.pb.go
│                   ├── hello.proto
│                   ├── hello.trpc.go
│                   └── hello_service_mock.go
└── trpc_go.yaml
```

现在可以直接在工程下面进行编译构建`go build`或者测试`go test`。

# 多个 pb 文件创建工程

多个 pb 文件创建工程的情景，指的是我们描述服务接口的 pb 文件中，通过 import 引入了外部其他 pb 文件，并引用了其中的一些定义。

这里的情况根据 pb 目录的组织情况，可以很简单，也可以很复杂：

- 服务 pb 文件，与引入的 pb 文件是否 package 相同；
- 服务 pb 文件，与引入的 pb 文件是否在同一个目录；
- 服务 pb 文件，引入 pb 文件时，是否添加了虚拟路径；
- 被引入的的 pb 文件中，是否又引入了其他的 pb 文件，上面 3 点问题进一步假如问题域；
- ...

对 pb 不熟悉的同学，会想当然认为 pb 很简单嘛，那是因为不了解，pb 作为一个自描述的消息格式，其必须要提供严谨一致的解析才能保证后续消息的正确性。

这里的问题不是很想展开多说，要穷尽大家的使用场景也非常耗费篇幅和时间，如果您对这方面感兴趣，可以参考下 create_test.go 中 create 命令字的测试用例。

我们只描述下相对来说比较简单的情况：
- 服务 pb 为 hello.proto，其中引入了一个外部的 pb 文件 world.proto；
- 它们可能 package 相同，也可能不同；
- 它们可能在相同目录，也可能在不同目录；

anyway，我们需要关注几个点，也只想让开发者关注这几个点：
- 服务 pb 是谁，`-p <protofile>` 来指定
- 被引入的 pb 从哪里可以找到，`-d <protodir>` 来指定，可以多次指定该选项以允许从多个路径中搜索；

所以最终的命令大致这样就可以了：`trpc create -p protofile -d dir1 -d dir2`。

见识过很多 pb 应用比较混乱的案例，这种情况下，我的建议是：
- 简化你的 pb 组织方式，这很重要；
- 编写好构建脚本，如 build.sh，在里面使用 trpc 命令，并写好路径搜索参数，方便后续使用；

其他更好的方法，还是优先考虑简化协议组织方式吧，没有什么比这个更高效的。

# 指定工程 go module

指定工程 go module 的方式有多种：
- 您可以先执行 `go mod init <module>` 初始化之后，然后在当前目录执行`trpc create`，此时将直接在当前目录生成工程文件；
- 您也可以什么都不指定，直接 `trpc create`，将使用 pb 文件名作为输出目录名，对应 go module 将使用 trpc.app.$service 或 package（注：会调整）；
- 您也可以显示指定，直接 `trpc create -mod <module>` 在创建时指定 go module 名称；

以上 3 种方式都可以使用，开发人员可以根据自己习惯来选择。

# 指定工程生成目录

指定工程生成目录，有三种方式：
- 如果您已经初始化过 go module，那么将直接将当前目录作为输出目录；
- 如果执行`trpc create`时没有显示指定输出目录，将默认使用 pb 文件名作为输出目录；
- 如果执行`trpc create`时通过选项`--output|-o`显示指定了输出目录，则使用指定的目录；

以上 3 种方式都可以使用，开发人员可以根据自己习惯来选择。

# 只生成 rpc 桩代码

有时我们并不需要根据 pb 创建一个完整的工程，比如当我们希望调用一个下游服务，我们现在获得了其 pb 文件，现在如果能生成其 rpc 桩代码在项目中引用，我们调用起来就方便多了，怎么办呢？

执行命令 `trpc create -p <protofile> --rpconly` 即可，--rpconly 选项会控制只生成 rpc 对应的桩代码，其他的一概不生成，输出目录默认是当前目录。

您也可以通过选项 `--output|-o`来控制桩代码的输出路径。
