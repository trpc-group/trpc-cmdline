English | [中文](README.zh_CN.md)

# Project and Stub Code Generation

## Complete Project Generation

The content of the following `helloworld.proto` can be found at /testcase/create/1-without-import/helloworld.proto(/testcase/create/1-without-import/helloworld.proto)

```shell
trpc create -p helloworld.proto
```

By default, a complete project will be generated in the current directory, including stub code and server/client examples.

You can use `-o` to specify the output directory, for example:

```shell
trpc create -p helloworld.proto -o out
```

If the output directory already contains a generated project, you can use `-f` to force overwrite, like:

```shell
trpc create -p helloworld.proto -o out -f
```

## Stub Code Generation Only

By adding `--rpconly`, you can generate only stub code (`pb.go` and `trpc.go`):

```shell
trpc create -p helloworld.proto -o out --rpconly
```

By default, `go.mod` and `_mock.go` files will also be generated. You can disable them with `--nogomod` and `--mock=false`:

```shell
trpc create -p helloworld.proto -o out --rpconly --nogomod --mock=false
```

## Specifying Code Generation for `proto` Files with Dependencies

The situations described above are sufficient for cases where there are no `import` of other `proto` files. For cases where the specified `proto` file has `import` of other `proto` files, you need to use `-d` to specify the search path for dependent `proto` files. Since there are many possible situations when `import`, users can refer to examples 2 to 5 in the [/testcase/create/](/testcase/create/) directory for learning and imitation.

## Generate Stub Code for Dependent `proto` Files while Generating Stub Code Only

When `--rpconly` is specified, if the `proto` file has `import` of other `proto` files, by default, only the stub code corresponding to the directly specified `proto` file will be generated. The intention of this default situation is to allow users to handle each `proto` file by generating stub code separately.

If users expect to generate dependent stub code at the same time, they can add the `--dependencystub` option to enable it.

## Specifying the go module name of the project

When generating a complete project, the go module name of the project will be derived from the `proto package name` of the `proto` file. Users can specify `--mod=xxxx` to change the go module name of the project.

Note: This changes the go module name when generating the project, not the go module name of the stub code itself. The go module name of the stub code itself is determined by the `go_package` in the `proto` file.

## Generate Stub Code for Other Protocols

By specifying `--protocol=http`, you can generate HTTP RPC services. The default is `--protocol=trpc`, which generates trpc protocol services.

## Generate flatbuffers Stub Code

By specifying `--fbs=xxx.fbs`, you can specify the protocol file of flatbuffers, and specify the search path for the file and its dependencies through `--fbsdir`.

These two options are similar to the roles of `-p` and `-d` under protobuf.

For examples of flatbuffers-related writing, you can refer to the files under [/testcase/flatbuffers/](/testcase/flatbuffers/).

## Template Supports Custom Variables and Environment Variables

### Custom Variables --kvrawjson

First, modify the `*.tpl` and reference custom variables as follows:

```tpl
// test json key-value: {{ .KVs.k1 }}
```

Where `k1` is the key name to access. Then specify `--assetdir` and `--kvrawjson`:

```bash
trpc create -p testcase/create/1-without-import/helloworldworld.proto \
            --assetdir ~/.trpc-cmdline-assets/install/protobuf/asset_go \
            --rpconly \
            -o out \
            --mock=false \
            --kvrawjson='{"k1":"v1"}'
```

This way, the corresponding value can be displayed in the generated `tpl.go`:

```go
// test json key-value: v1
```

And this approach naturally supports nested key-value (the type of the first layer key must be string), for example:

```tpl
// test json key-value: {{ .KVs.k1.k2 }}
```

```bash
trpc create -p testcase/create/1-without-import/helloworldworld.proto \
            --assetdir ~/.trpc-cmdline-assets/install/protobuf/asset_go \
            --rpconly \
            -o out \
            --mock=false \
            --kvrawjson='{"k1":{"k2":"v2"}}'
```

```go
// test json key-value: v2
```

### Custom Variables --kvfile

The difference between `--kvfile` and `--kvrawjson` is that the former specifies a json file name. The usage is as follows:

```bash
trpc create -p testcase/create/1-without-import/helloworldworld.proto \
            --assetdir ~/.trpc-cmdline-assets/install/protobuf/asset_go  \
            --rpconly \
            -o out \
            --mock=false \
            --kvfile ./test.json
```

These two flags can be used simultaneously, and the specified key-value pairs will be merged into the same map:

```bash
trpc create -p testcase/create/1-without-import/helloworldworld.proto \
            --assetdir ~/.trpc-cmdline-assets/install/protobuf/asset_go \
            --rpconly \
            -o out \
            --mock=false \
            --kvfile ./test.json \
            --kvrawjson='{"kkk":"xv"}'
```

### Using Environment Variables in Templates

You only need to use `.Envs.XXXX` in the `.tpl` file to access the corresponding environment variables, for example:

```tpl
// get envs: {{ .Envs.USER }}
```

After generation, it corresponds to:

```go
// get envs: root
```

## Customize the `app` and `server` Names in Stub Code

Example:

```shell
trpc create -p helloworld.proto -o out --app="someappname" --server="someservername"
```
