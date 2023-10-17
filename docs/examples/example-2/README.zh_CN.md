[English](README.md) | 中文

# 扩展功能

## 自定义接口别名

在通常情况下，trpc 接口的 rpc name 形式为 `/trpc.app.server.service/Method`，当使用了 HTTP RPC 功能时，大概率期望这个接口名能够符合一个给定的规范，比如 `/v2/api/service`，此时就需要用到接口别名自定义功能。

有两种用法：

1. 使用 `trpc.alias`，此时需要 `import "trpc/proto/trpc_options.proto";` 如：

```protobuf
import "trpc/proto/trpc_options.proto";
service HelloWorldService {
  rpc Hello(HelloReq) returns(HelloRsp) { option(trpc.alias) = "/api/v1/helloworld"; };
}
```

注意这里的 "trpc/proto/trpc_options.proto" 文件不需要用户自己指定 `-d` 来引入，而是自然存在于 `~/.trpc-cmdline-assets/submodules/trpc-protocol` 路径下并自动添加到搜索路径中，用户直接直接 `import` 即可。假如用户希望在编辑器下能够解析 protobuf 文件，只需找到编辑器的设置，将 `~/.trpc-cmdline-assets/submodules/trpc-protocol` 添加到编辑器的 protobuf 插件的搜索路径下即可。

2. 使用 `//@alias=` 注解，此时不需要任何 `import`，但是需要在执行 `trpc create` 命令时追加 `--alias` 选项，如：

```protobuf
service HelloWorldService {
  //@alias="/api/helloworld"
  rpc Hello(HelloReq) returns(HelloRsp);
}
```

```shell
trpc create -p helloworld.proto -o out --alias
```

## 自定义字段 tag

在默认情况下，生成的 `pb.go` 文件中的数据结构定义中存在 `protobuf` 以及 `json` 的 tag，从而这些字段能够以这些 tag name 作为标识进行序列化，如：

```go
type HelloRequest struct {
    // ...
    Msg string `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
}
```

在一些情况下用户期望能够自定义更多的 go tag，可以使用 `trpc.go_tag` 进行指定，并 `import "trpc/proto/trpc_options.proto";`，如：

```protobuf
import "trpc/proto/trpc_options.proto";
message Req{
  string msg = 1 [ (trpc.go_tag)='gorm:"any_msg"' ];
}
```

并且在执行 `trpc create` 的时候需要指定 `--gotag` 选项：

```shell
trpc create -p helloworld.proto -o out --gotag
```

## 生成 swagger/openapi 文档

trpc-cmdline 工具提供了 `trpc apidocs` 子命令以生成文档，用户可以执行 `trpc apidocs -h` 以查看所有支持的命令选项。

在 `proto`` 文件中使用 `trpc.swagger`，并 `import "trpc/swagger/swagger.proto";`，如：

```protobuf
service HelloWorldService {
    rpc Hello(HelloReq) returns(HelloRsp) {
        option(trpc.swagger) = {
            title : "helloworld"
            method: "get"
            description: "some description"
            params: {
                name: "msg"
                required: true
                default: "hello"
            }
        };
    };
}
```

可以通过 `--swagger-out=file.json` 来执行输出文件名，如：

```shell
trpc apidocs -p helloworld.proto --swagger --swagger-out=output.swagger.json
```

openapi 相关的命令也是类似的，如：

```shell
trpc apidocs -p helloworld.proto --openapi --openapi-out=output.openapi.json
```

一些额外的选项：

* `--swagger-json-param`：可以使生成的 "parameters" 描述显示未 `"in": "body"`
* `--order-by-pbname`：在生成的文档中将数据结构及服务接口的定义按照原始 `proto` 文件中的顺序进行展示，而不是按照首字母进行排序（默认为 `false`，即按照首字母进行排序）
* `-d`：指定 `proto` 文件依赖的搜索路径，和 `trpc create` 命令中的 `-d` 含义相同
* `--alias`：在文档中显示带有 alias 的接口名
* `--keep-orig-rpcname`：在 `--alias=true` 的时候，默认情况下原始的 rpc name 以及 alias 之后的名称都会显示，用户可以指定 `--keep-orig-rpcname=false` 以使文档只显示 alias 之后的名称，不显示原始的 rpc name
