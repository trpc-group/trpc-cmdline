# generate *.pb.go

生成 ****.pb.go*** 的做法，通过大家习惯的方式，直接借助 ***protoc*** 命令来生成：
```bash
protoc --go_out=. helloworld.proto
```

# generate server stub

通过 jhump/protoreflect 来解析 *.proto 文件，解析后得到 ***FileDescriptorProto*** 对象，该对象中记录了我们关心的 *.proto 描述信息，然后将这些信息填充到模板中，完成 server、client stub 的代码生成。

对于 server stub，这里需要注意与框架代码的整合：
- 一方面要符合业务开发人员的习惯，使业务开发工作尽可能简单，如只填充业务函数即可；
- 一方面要将业务代码和框架代码进行粘合，组合框架能力、默认插件实现、整合 wrapper 方法等实现开箱即用；

server 要处理的请求命令字或者 rpc 接口，都已经在*.proto 中定义：
- `trpc` 根据 service 中定义的 rpc，建立映射关系 rpcName->Handler；
- `trpc` 提供方法，使得 server 能够注册上述映射关系，并在 router 中根据 req 找到 Handler；

可以参考 `protoc --go_out=plugins=grpc:.` 与 `protoc --go_out=.` 两种方式生成的代码的区别，来参考下 grpc 是如何建立映射关系的。

## demo

```protobuf
syntax = "proto3";
package helloworld;

// Hello
message HelloRequest {
    string from     = 1;    // say hello from
    string to       = 2;    // say hello to
    string words    = 3;    // hello words
}

message HelloResponse {
    uint32 errcode  = 1;    // error code
    string errmsg   = 2;    // error msg
}

// Bye
message ByeRequest {
    string from     = 1;    // say bye from
    string to       = 2;    // say bye to
    string words    = 3;    // bye words
}

message ByeResponse {
    uint32 errcode  = 1;    // error code
    string errmsg   = 2;    // error msg
}

// service: greeter
service greeter {
    rpc SayHello ( HelloRequest )    returns ( HelloResponse );
    rpc SayBye   ( ByeRequest   )    returns ( ByeResponse   );
}
```

运行 trpc 来生成 server 工程：`trpc create -protofile=helloworld.proto -protocol=trpc`，需要生成的内容应该包括：

- helloworld.pb.go
- service interface

```go
import pb "helloworld.proto"

type GreeterServer interface {
    SayHello(ctx context.Context, req *pb.HelloRequest) (rsp *pb.HelloResponse, error)
    SayBye(ctx context.Context, req *pb.ByeRequest) (rsp *pb.ByeResponse, error)
}
```

- service rpcName->rpcMethod register 

```go
func RegisterGreeterServer(s *server.Server, svr GreeterServer) {
    s.RegisterService(&_Greeter_serviceDesc, svr)
}

func _Greeter_SayHello_Handler(svr interface{}, ctx context.Context) error {
    reqCtx := ctx.Value(ctxkey)

    req := new(HelloRequest)
    if err := reqCtx.Decode(req); err != nil {
        return err
    }

    if rsp, err := svr.SayHello(ctx, req); err != nil {
        return err
    }

    reqCtx.rspChan <- rsp
    return nil
}

var _Greeter_serviceDesc = trpc.ServiceDesc{
    ServiceName: "helloworld.greeter",
    HandlerType: (*GreeterServer)(nil),
    Methods: []trpc.MethodDesc{
        {
            MethodName: "SayHello",
            Handler: _Greeter_SayHello_Handler,
        },
        {
            MethodName: "SayBye",
            Handler: _Greeter_SayBye_Handler,
        }
    },
    Streams: []trpc.StreamDesc{},
    MetaData: "helloworld.proto",
}

```

当前 server 端提供了一个 server.WithHandler(....) 来封装所有的请求处理、tracing、拦截器、监控、logging 等逻辑，这个没问题，Handler 内部会请求 Dispatcher 来完成 rpc 请求到 rpc 处理函数的分发。所以这里还需要提供一个 Dispatcher？

- 方法 1：trpc 可以显示提供一个 dispatcher 出来，server 端使用的时候 WithDispatcher(GreeterServer.Dispatcher) 就可以；
- 方法 2：新增一个方法 server.RegisterService(trpc.ServiceDesc)，server 自己注册；
- 方法 3：Dispatcher 接口提供 Add 等方法，支持直接注册到 dispatcher；
- 方法 4：server 提供 Dispatch(rpcName, rpcMethod) 直接进行注册，内部注册到 server.Opts.Dispatcher 上；

有多种方式，需要综合考虑下，对于后面支持进程多 server 实例有用处！

# generate client stub

# generate config 

# generate others
