English | [中文](README.zh_CN.md)

# Documentation

## Design and Implementation

`trpc-cmdline` depends on [Protobuf](https://protobuf.dev/), using [`proto`](https://github.com/protocolbuffers/protobuf) files as an intermediary, leveraging [protoc](https://grpc.io/docs/protoc-installation/) to generate stub code with data structure definitions, and using [Go templates](https://pkg.go.dev/text/template) to generate stub code with service definitions.

Among them:

* Stub code with data structure definitions has a `pb.go` suffix
* Stub code with service definitions has a `trpc.go` suffix

## Code Template

After `trpc-cmdline` parses the specified `proto` file, it will generate stub code according to the code template of the corresponding language.

These code templates are located in the `install/protobuf/asset_${language}` directory, such as [install/protobuf/asset_go](/install/protobuf/asset_go/), [install/protobuf/asset_cpp](/install/protobuf/asset_cpp/), etc.

You can reference custom variables and functions in the code template, for example:

* You can reference the value of `FileDescriptor.PackageName` in the template file through `{{.PackageName}}`
* You can use custom functions such as `title`: `{{hello | title}}` => `Hello`

It is recommended to learn and imitate by reading existing template files.

By specifying `--assetdir`, you can replace the template folder used during generation with the path you specify, for example:

```bash
trpc create -p hello.proto --assetdir=~/.trpc-cmdline-assets/protobuf/asset_go
```

Note: Here, `--assetdir` needs to specify an absolute path.

The files in the [install](/install/) directory will be automatically decompressed to the user's `~/.trpc-cmdline-assets/` directory before the binary is executed, so the default template path is `~/.trpc-cmdline-assets/protobuf/asset_go`

The same applies to other languages such as C++, you need to specify `--lang=cpp` additionally, and its default template path is `~/.trpc-cmdline-assets/protobuf/asset_cpp`

## Examples

* [example1](examples/example-1/README.md) shows more options and details for generating projects and stub code, such as
  * How to specify code generation for `proto` files with dependencies
  * How to specify the go module name of the project
  * How to specify the output file path
  * How to generate stub code only
  * How to generate stub code for other protocols (such as HTTP)
  * How to generate flatbuffers stub code
  * ...
* [example2](examples/example-2/README.md) shows how to use pb option extension features, such as
  * How to add aliases for service names
  * How to add custom tags for fields
  * How to generate validate.pb.go file
  * How to generate swagger/openapi documentation
