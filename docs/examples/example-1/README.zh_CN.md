[English](README.md) | 中文

# 项目及桩代码生成

## 完整项目生成

以下 `helloworld.proto` 的内容可以参考 [/testcase/create/1-without-import/helloworld.proto](/testcase/create/1-without-import/helloworld.proto)

```shell
trpc create -p helloworld.proto
```

默认会在当前目录生成完整的项目，带有桩代码以及服务端/客户端示例。

可以通过 `-o` 来指定输出目录，比如：

```shell
trpc create -p helloworld.proto -o out
```

如果输出目录已经存在生成的项目，可以使用 `-f` 进行强制覆盖，如：

```shell
trpc create -p helloworld.proto -o out -f
```

## 仅桩代码生成

通过添加 `--rpconly` 可以只生成桩代码（`pb.go` 以及 `trpc.go`）：

```shell
trpc create -p helloworld.proto -o out --rpconly
```

在默认情况下还会生成 `go.mod` 以及 `_mock.go` 文件，可以通过 `--nogomod` 以及 `--mock=false` 来禁用：

```shell
trpc create -p helloworld.proto -o out --rpconly --nogomod --mock=false
```

## 指定含有依赖的 `proto` 文件的代码生成

之前所描述的情况对于不存在 `import` 其他 `proto` 的案例来说已经够用了，对于所指定的 `proto` 文件中存在对其他 `proto` 文件 `import` 了的情况，则需要用 `-d` 来指定对依赖 `proto` 文件的搜索路径，由于在 `import` 时存在非常多的可能情况，用户可以参考 [/testcase/create/](/testcase/create/) 目录下的 2 到 5 的例子进行模仿学习。

## 在仅桩代码生成时同时生成依赖 `proto` 的桩代码

在指定了 `--rpconly` 的时候，假如 `proto` 文件存在对其他 `proto` 文件的 `import`，在默认情况下，只会生成直接指定的这个 `proto` 文件所对应的桩代码，这种默认情况的本意是为了使用户采用对每个 `proto` 文件进行单独桩代码生成而处理的。

假如用户期望能够同时生成依赖的桩代码，可以添加 `--dependencystub` 选项来启用。

## 指定工程的 go module name

在生成完整项目时，项目的 go module name 会从 `proto` 文件的 `proto package name` 中推出来，用户可以指定 `--mod=xxxx` 来改变项目的 go module name

注意：这个改变的是生成项目时的 go module name，而不是桩代码本身的 go module name，桩代码本身的 go module name 由 `proto` 文件中的 `go_package` 决定。

## 生成其他协议桩代码

通过指定 `--protocol=http` 可以生成 HTTP RPC 服务，默认为 `--protocol=trpc` 即生成 trpc 协议服务。

## 生成 flatbuffers 桩代码

通过指定 `--fbs=xxx.fbs` 指定 flatbuffers 的协议文件，通过 `--fbsdir` 来指定文件及依赖的搜索路径。

这两个选项类似于 `-p` 以及 `-d` 在 protobuf 下的作用。

flatbuffers 相关的写法用例可以参考 [/testcase/flatbuffers/](/testcase/flatbuffers/) 下的文件进行学习。

## 模版支持自定义变量及环境变量

### 自定义变量 --kvrawjson

首先修改 `*.tpl`，在其中引用自定义变量如下：

```tpl
// test json key-value: {{ .KVs.k1 }}
```

其中 `k1` 为要访问的 key 名，然后指定 `--assetdir` 以及 `--kvrawjson`：

```bash
trpc create -p testcase/create/1-without-import/helloworldworld.proto \
            --assetdir ~/.trpc-cmdline-assets/install/protobuf/asset_go \
            --rpconly \
            -o out \
            --mock=false \
            --kvrawjson='{"k1":"v1"}'
```

这样最终生成的 `tpl.go` 中可以显示对应的 value：

```go
// test json key-value: v1
```

并且这种做法天然支持嵌套 key-value（第一层 key 的类型必须为 string），比如：

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

### 自定义变量 --kvfile

`--kvfile` 和 `--kvrawjson` 的区别在于：前者指定的是一个 json 文件名，用法如下：

```bash
trpc create -p testcase/create/1-without-import/helloworldworld.proto \
            --assetdir ~/.trpc-cmdline-assets/install/protobuf/asset_go  \
            --rpconly \
            -o out \
            --mock=false \
            --kvfile ./test.json
```

并且这两个 flag 可以同时使用，指定的 key-value 会融合到同一个 map 中

```bash
trpc create -p testcase/create/1-without-import/helloworldworld.proto \
            --assetdir ~/.trpc-cmdline-assets/install/protobuf/asset_go \
            --rpconly \
            -o out \
            --mock=false \
            --kvfile ./test.json \
            --kvrawjson='{"kkk":"xv"}'
```

### 在模板中使用环境变量

只需要在 `.tpl` 文件中使用 `.Envs.XXXX`，即可访问对应的环境变量，比如：

```tpl
// get envs: {{ .Envs.USER }}
```

生成后对应：

```go
// get envs: root
```

## 自定义桩代码中的 `app` 以及 `server` 名

示例：

```shell
trpc create -p helloworld.proto -o out --app="someappname" --server="someservername"
```
