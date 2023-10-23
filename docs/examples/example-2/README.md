English | [中文](README.zh_CN.md)

# Extended Features

## Custom Interface Aliases

In most cases, the trpc interface's rpc name is in the form of `/trpc.app.server.service/Method`. When using the HTTP RPC feature, it is highly likely that you would want the interface name to conform to a given specification, such as `/v2/api/service`. In this case, you need to use the interface alias customization feature.

There are two ways to use it:

1. Use `trpc.alias`, in which case you need to `import "trpc/proto/trpc_options.proto";` as follows:

```protobuf
import "trpc/proto/trpc_options.proto";
service HelloWorldService {
  rpc Hello(HelloReq) returns(HelloRsp) { option(trpc.alias) = "/api/v1/helloworld"; };
}
```

Note that the "trpc/proto/trpc_options.proto" file does not require the user to specify `-d` to import it. Instead, it naturally exists in the `~/.trpc-cmdline-assets/submodules/trpc-protocol` path and is automatically added to the search path. Users can directly import it. If users want to resolve protobuf files in the editor, they only need to find the editor's settings and add `~/.trpc-cmdline-assets/submodules/trpc-protocol` to the editor's protobuf plugin search path.

2. Use the `//@alias=` annotation, in which case no `import` is required, but you need to append the `--alias` option when executing the `trpc create` command, as follows:

```protobuf
service HelloWorldService {
  //@alias="/api/helloworld"
  rpc Hello(HelloReq) returns(HelloRsp);
}
```

```shell
trpc create -p helloworld.proto -o out --alias
```

## Custom Field Tags

By default, the data structure definitions in the generated `pb.go` file contain `protobuf` and `json` tags, so these fields can be serialized using these tag names as identifiers, such as:

```go
type HelloRequest struct {
    // ...
    Msg string `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
}
```

In some cases, users may want to customize more go tags. You can use `trpc.go_tag` to specify them and `import "trpc/proto/trpc_options.proto";`, as follows:

```protobuf
import "trpc/proto/trpc_options.proto";
message Req{
  string msg = 1 [ (trpc.go_tag)='gorm:"any_msg"' ];
}
```

And when executing `trpc create`, you need to specify the `--gotag` option:

```shell
trpc create -p helloworld.proto -o out --gotag
```

## Generating validate.pb.go File

For a complete example, see [/testcase/create/10-validate-pgv/helloworld.proto](/testcase/create/10-validate-pgv/helloworld.proto).

This feature requires the installation of [protoc-gen-validate](https://github.com/bufbuild/protoc-gen-validate). Typically, `trpc setup` can automatically install these dependencies.

Here is an example of a proto file:

```proto
syntax = "proto3";
package helloworld;

option go_package="trpc.group/some-example/helloworld";

import "validate/validate.proto";

service HelloSvr {
    rpc Hello(HelloRequest) returns(HelloResponse);
}

message HelloRequest {
    string msg = 1 [(validate.rules).string.email=true];
}

message HelloResponse {
    int32 err_code = 1; 
    string err_msg = 2; 
}
```

Points to note:

* The reference is `import "validate/validate.proto";`. You can download this file from the [protoc-gen-validate repository](https://github.com/bufbuild/protoc-gen-validate/blob/main/validate/validate.proto) and specify the path (download this file to `somedir/validate/validate.proto` and specify `-d somedir`). The trpc-cmdline tool has a built-in copy of this file.
* The syntax for the validation rule `[(validate.rules).string.email=true]` can be found in the [protoc-gen-validate documentation](https://github.com/bufbuild/protoc-gen-validate/blob/v1.0.2/README.md).
* When generating code, you need to add `--validate=true`, like this:
  ```shell
  trpc create -p helloworld.proto -o out --validate
  ```
* The generated stub code will include the `xxx.validate.pb.go` file.
* In the generated project code, the following two locations will automatically add validate-related plugin information:
  * All `main.go` will add an anonymous reference `import _ "trpc.group/trpc-go/trpc-filter/validation"`
  * In `trpc_go.yaml`, the client/server filter configuration will include `- validation`

## Generating swagger/openapi Documentation

The trpc-cmdline tool provides the `trpc apidocs` subcommand to generate documentation. Users can execute `trpc apidocs -h` to view all supported command options.

Use `trpc.swagger` in the `proto` file and `import "trpc/swagger/swagger.proto";`, as follows:

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

You can execute the output file name with `--swagger-out=file.json`, such as:

```shell
trpc apidocs -p helloworld.proto --swagger --swagger-out=output.swagger.json
```

The commands related to openapi are similar, such as:

```shell
trpc apidocs -p helloworld.proto --openapi --openapi-out=output.openapi.json
```

Some additional options:

* `--swagger-json-param`: Can make the generated "parameters" description display as `"in": "body"`
* `--order-by-pbname`: In the generated document, the definition of data structure and service interface is displayed in the order of the original `proto` file, instead of being sorted by the first letter (default is `false`, that is, sorted by the first letter)
* `-d`: Specifies the search path for `proto` file dependencies, the same as `-d` in the `trpc create` command
* `--alias`: Display the interface name with alias in the document
* `--keep-orig-rpcname`: When `--alias=true`, by default, both the original rpc name and the name after alias will be displayed. Users can specify `--keep-orig-rpcname=false` to make the document only display the name after alias, and not display the original rpc name
