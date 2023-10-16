[English](README.md) | 中文

# trpc-cmdline

[![Go Reference](https://pkg.go.dev/badge/github.com/trpc.group/trpc-cmdline.svg)](https://pkg.go.dev/github.com/trpc.group/trpc-cmdline)
[![Go Report Card](https://goreportcard.com/badge/github.com/trpc.group/trpc-go/trpc-cmdline)](https://goreportcard.com/report/github.com/trpc.group/trpc-go/trpc-cmdline)
[![LICENSE](https://img.shields.io/github/license/trpc.group/trpc-cmdline.svg?style=flat-square)](https://github.com/trpc.group/trpc-cmdline/blob/main/LICENSE)
[![Releases](https://img.shields.io/github/release/trpc.group/trpc-cmdline.svg?style=flat-square)](https://github.com/trpc.group/trpc-cmdline/releases)
[![Docs](https://img.shields.io/badge/docs-latest-green)](http://test.trpc.group.woa.com/docs/)
[![Tests](https://github.com/trpc.group/trpc-cmdline/actions/workflows/prc.yaml/badge.svg)](https://github.com/trpc.group/trpc-cmdline/actions/workflows/prc.yaml)
[![Coverage](https://codecov.io/gh/trpc.group/trpc-cmdline/branch/main/graph/badge.svg)](https://app.codecov.io/gh/trpc.group/trpc-cmdline/tree/main)

trpc-cmdline 是 [trpc-cpp](https://github.com/trpc-group/trpc-cpp) 和 [trpc-go](https://github.com/trpc-group/trpc-go) 的命令行工具。

本项目支持 [Go](https://go.dev/doc/devel/release) 最新发布的三个版本。

## 安装

### 安装 trpc-cmdline

#### 使用 go 命令进行安装

首先将以下内容添加到你的 `~/.gitconfig` 中:

```bash
[url "ssh://git@github.com/"]
    insteadOf = https://github.com/
```

然后执行以下命令以安装 `trpc-cmdline`:

```bash
go install trpc.group/trpc-go/trpc-cmdline/trpc@latest
```

<!-- #### Install from release

<details><summary>Click to show the bash script</summary><br><pre>
$ TAG="v0.0.1" # Choose tag.
$ OS=linux # Choose from "linux", "darwin" or "windows".
$ wget -O trpc https://github.com/trpc-group/trpc-cmdline/releases/download/${TAG}/trpc_${OS}
$ mkdir -p ~/go/bin && chmod +x trpc && mv trpc ~/go/bin
$ export PATH=~/go/bin:$PATH # Add this to your `~/.bashrc`.
</pre></details> -->

### 安装依赖

 <!-- by using one of the following methods.

#### Using trpc setup

After installation of trpc-cmdline, simply running `trpc setup` will automatically install all the dependencies. 

#### Install separately -->

<details><summary>Install protoc </summary><br><pre>
$ # Reference: https://grpc.io/docs/protoc-installation/
$ PB_REL="https://github.com/protocolbuffers/protobuf/releases"
$ curl -LO $PB_REL/download/v3.15.8/protoc-3.15.8-linux-x86_64.zip
$ unzip -o protoc-3.15.8-linux-x86_64.zip -d $HOME/.local
$ export PATH=~/.local/bin:$PATH # Add this to your `~/.bashrc`.
$ protoc --version
libprotoc 3.15.8
</pre></details>

<details><summary>Install flatc </summary><br><pre>
$ # Reference: https://github.com/google/flatbuffers/releases
$ wget https://github.com/google/flatbuffers/releases/download/v23.5.26/Linux.flatc.binary.g++-10.zip
$ unzip -o Linux.flatc.binary.g++-10.zip -d $HOME/.bin
$ export PATH=~/.bin:$PATH # Add this to your `~/.bashrc`.
$ flatc --version
flatc version 23.5.26
</pre></details>

<details><summary>Install protoc-gen-go</summary><br><pre>
$ # Reference: https://grpc.io/docs/languages/go/quickstart/
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
</pre></details>

<details><summary>Install goimports</summary><br><pre>
$ go install golang.org/x/tools/cmd/goimports@latest
</pre></details>

<details><summary>Install mockgen</summary><br><pre>
$ # Reference: https://github.com/uber-go/mock
$ go install go.uber.org/mock/mockgen@latest
</pre></details>


## 快速上手

### 生成完整项目

* 将以下内容复制到 `helloworld.proto`, 原始文件为 [./docs/helloworld/helloworld.proto](./docs/helloworld/helloworld.proto):

```protobuf
syntax = "proto3";
package helloworld;

option go_package = "github.com/some-repo/examples/helloworld";

// HelloRequest is hello request.
message HelloRequest {
  string msg = 1;
}

// HelloResponse is hello response.
message HelloResponse {
  string msg = 1;
}

// HelloWorldService handles hello request and echo message.
service HelloWorldService {
  // Hello says hello.
  rpc Hello(HelloRequest) returns(HelloResponse);
}
```

* 使用 trpc-cmdline 来生成完整项目:
```go
$ trpc create -p helloworld.proto -o out
```

注意: `-p` 用于指定 proto 文件, `-o` 用于指定输出目录, 
更多 flag 信息可以运行 `trpc -h` 以及 `trpc create -h` 来进行查看。

* 进入输出目录，运行服务端:
```bash
$ cd out
$ go run .
...
... trpc service:helloworld.HelloWorldService launch success, tcp:127.0.0.1:8000, serving ...
...
```

* 在另一个终端中进入输出目录，运行客户端:
```bash
$ go run cmd/client/main.go 
... simple  rpc   receive: 
```

注意: 由于生成的代码默认都是空操作，因此日志中显示的收到的数据内容也为空。

* 现在你可以尝试修改 `hello_world_service.go` 中的服务端代码以及 `cmd/client/main.go` 中的客户端代码来创建一个 echo 服务器。你可以参考 [https://github.com/trpc-group/trpc-go/tree/main/examples/helloworld](https://github.com/trpc-group/trpc-go/tree/main/examples/helloworld) 以获取灵感

* 生成文件的详细解释如下:

```bash
$ tree
.
|-- cmd
|   `-- client
|       `-- main.go  # Generated client code.
|-- go.mod
|-- go.sum
|-- hello_world_service.go  # Generated server service implementation.
|-- hello_world_service_test.go
|-- main.go  # Server entrypoint.
|-- stub  # Stub code.
|   `-- github.com
|       `-- some-repo
|           `-- examples
|               `-- helloworld
|                   |-- go.mod
|                   |-- helloworld.pb.go
|                   |-- helloworld.proto
|                   |-- helloworld.trpc.go
|                   `-- helloworld_mock.go
`-- trpc_go.yaml  # Configuration file for trpc-go.
```

### 仅生成桩代码

* 只需要添加 `--rpconly` 选项就可以只生成桩代码:
```go
$ trpc create -p helloworld.proto -o out --rpconly
$ tree out
out
|-- go.mod
|-- go.sum
|-- helloworld.pb.go
|-- helloworld.trpc.go
`-- helloworld_mock.go
```

### 常用的指令

下面列举了一些常用的命令行选项：

* `-f`: 用于强制覆盖输出目录中的内容
* `-d some-dir`: 添加 proto 文件的查找路径（包括依赖的 proto 文件），可以指定多次
* `--mock=false`: 禁止生成 mock 代码
* `--nogomod=true`: 在生成桩代码时不生成 `go.mod` 文件，只在 `--rpconly=true` 的时候生效, 默认为 `false`

更多命令行选项可以执行 `trpc -h` 以及 `trpc [subcmd] -h` 来进行查看。

### 更多功能

请查看 [文档](./docs/)

## 贡献

本开源项目欢迎任何贡献，请阅读 [贡献指南](CONTRIBUTING.md) 以获取更多信息。
