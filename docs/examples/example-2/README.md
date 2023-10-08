README
==============================================================================

为了支持更多的应用场景，trpc 也在`trpc-go/trpc`自定义了一些 pb 扩展，用来支持：

# 1. 支持接口别名 trpc.alias

## pb 的写法

- 在 `rpc` 的 `body` 中定义 `option`，形如

```protobuf
service helloworld_svr {
    rpc Hello(HelloReq) returns(HelloRsp) { option(trpc.alias) = "/api/v1/helloworld"; };
}
```

这种方式适用于需要对接口进行重命名的操作，比如一个 trpc 接口的 rpcName 通常是
/package.service/method 这种形式的，如果想提供一个 http 接口的话可能根据 api 网关配
置要求需要配置成符合指定样式的，如/api/service/method。

pb 大致就按照方式进行编写，如果是因为存量协议问题，如 ilive 协议中只有数值型大写、小写命令字：

```
service helloworld_svr {
    rpc Hello(HelloReq) returns(HelloRsp) { option (trpc.alias) = "0x100_0x200";};
}
```

再或者，为了兼容存量协议，有些协议头设计是没有字符串形式的 cmd 或者 method 字段的，
可能只有一个数值型的 cmd 字段，这种就需要将字符串形式的 rpcName 转换成对应数值的字
符串形式，以方便后续编码时将其转换成数值设置到协议头。
   
对应协议的 Encode 函数会将其 split 为大小写命令字并分别设置到协议头字段中去。
 
- `trpc.alias` 重命名 `rpc` 的方法名 

```protobuf
service helloworld_svr {
    //@alias= "/api/helloworld"
    rpc Hello(HelloReq) returns(HelloRsp);
}
```

## 使用方法

1. pb 中加入 `option` 定义，或者使用 @alias 注解，推荐使用前者，注解的方式后续会移除。
2. 运行命令的时候：

```bash
# option 的方式
trpc create -p=./cmd/testcase.option/helloworld.proto

# 注解的方式
trpc create -p=./cmd/testcase.option/helloworld.proto --alias
```

引用这里的扩展选项，需要引入 trpc.proto，在 trpc 安装阶段该 pb 文件会自动安装到 ~/.trpc-cmdline-assets/ 下面，

上述 ~/.trpc-cmdline-assets 搜索路径会自动设置到 pb 的搜索路径中，以保证能够正确解析。

执行命令 `trpc create -p hello.proto` 来完成工程的创建，如果您是使用的//@alias
来指定别名，则需要显示指定--alias 选项。

ps: 

建议使用 option 的方式，因为下面你会看到 swagger、gotag 都是 option 方式，希望保持一致的操作风格。
后续也不排除会 drop --alias 这个选项以及//@alias 注解的方式。

# 2. 支持自定义 struct tag trpc.go_tag

## pb 的写法

```pb
import "trpc/proto/trpc_options.proto";

message Req{
    string msg = 1 [ (trpc.go_tag)='gorm:"any_msg"' ];
}
```

## 使用方法

运行命令：

```bash
trpc create -p=./cmd/testcase.option/helloworld.swagger.proto --go_tag
```

# 3. 支持 swagger api 文档，比较多，略

## pb 的写法

- 在 `rpc` 的 `body` 中定义 option，形如

```protobuf
service helloworld_svr {
    rpc Hello(HelloReq) returns(HelloRsp) {
        option(trpc.swagger) = {
            title : "你好世界"
            method: "get"
            description:
                "入参：msg\n"
                "作用：用于演示 helloword\n"
            params: {
                name: "msg"
                required: true
                default: "hello"
            }
            params: {
                name: "cnt"
                required: true
                default: "1"
            }
        };
    };
}
```

- `trpc.swagger` 的 `title` 为该 `rpc` 的方法名，`method` 为 `http` 的请求方法（如果该接口用于 `http`，
  由于 `swagger-ui` 会识别一个 `method`，如果该字段不填，默认为 `post`），`description` 用于描述此接口。
- `params` 则用于指定 `request` 各字段的一些属性，比如是否为必要字段 (`required`), 默认值 (`default`), `params` 可以写多次从而描述多个字段。
- 假如期望生成的 swagger 描述的 "parameters" 显示为 `"in": "body"`, 需要额外指定 `--swagger-json-param` 标志 (`--swagger` 标志仍需要提供）。

## 使用方法

运行命令：

```bash
trpc create -p=./cmd/testcase.option/helloworld.swagger.proto --swagger --alias
```

- 在当前目录下会生成 `apidocs.swagger.json`
- 下载 `swagger-ui` (https://github.com/swagger-api/swagger-ui)
- 进入到仓库下的 `dist` 目录，将 `apidocs.swagger.json` 拷贝至此，并修改 `index.html` 文件中的 `url` 为 `apidocs.swagger.json`。
- `npm install -g http-server`，直接运行 `http-server` 后可以通过 bash 显示的 url 对 swagger 页面进行访问。
