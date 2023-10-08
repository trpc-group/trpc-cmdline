# `rpc option` 测试用例

## option alias 

### pb 的写法

- 在 `rpc` 的 `body` 中定义 `option`，形如

```protobuf
service helloworld_svr {
    rpc Hello(HelloReq) returns(HelloRsp) {
        option(trpc.alias) = "/api/v1/helloworld";
    };
}
```

- `trpc.alias` 重命名 `rpc` 的方法名。

```protobuf
service helloworld_svr {
    //@alias= "/api/helloworld"
    rpc Hello(HelloReq) returns(HelloRsp);
}
```


### 使用方法

1. pb 中加入 `option` 定义，或者使用 @alias 注解，推荐使用前者，注解的方式后续会移除。
2. 运行命令的时候：

```bash
# option的方式
trpc create -p=./cmd/testcase.option/helloworld.proto

# 注解的方式
trpc create -p=./cmd/testcase.option/helloworld.proto --alias
```

## option swagger 

### pb 的写法

- 在 `rpc` 的 `body` 中定义 option，形如

```go
service helloworld_svr {
    rpc Hello(HelloReq) returns(HelloRsp) {
        option(trpc.swagger) = {
            title : "你好世界"
            method: "get"
            description:
                "入参：msg\n"
                "作用：用于演示 helloword\n"
        };
    };
}
```

- `trpc.swagger` 的 `title` 为该 `rpc` 的方法名，`method` 为 `http` 的请求方法（如果该接口用于 `http`，
由于 `swagger-ui` 会识别一个 `method`，如果该字段不填，默认为 `post`），`description` 用于描述此接口。

### 使用方法

运行命令：

```bash
trpc create -p=./cmd/testcase.option/helloworld.swagger.proto --swagger --alias
```

- 在当前目录下会生成 `apidocs.swagger.json` 
- 下载 `swagger-ui` (https://github.com/swagger-api/swagger-ui)
- 进入到仓库下的 `dist` 目录，将 `apidocs.swagger.json` 拷贝至此，并修改 `index.html` 文件中的 `url` 为 `apidocs.swagger.json`。
- `npm install -g http-server`，直接运行 `http-server` 后可以通过 bash 显示的 url 对 swagger 页面进行访问。

## option go_tag

### pb 的写法

```pb
import "trpc/proto/trpc_options.proto";

message Req{
    string msg = 1 [ (trpc.go_tag)='gorm:"any_msg"' ];
}
```

### 使用方法

运行命令：

```bash
trpc create -p=./cmd/testcase.option/helloworld.swagger.proto --go_tag
```
