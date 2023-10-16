English | [中文](README.zh_CN.md)

# trpc-cmdline

[![Go Reference](https://pkg.go.dev/badge/github.com/trpc.group/trpc-cmdline.svg)](https://pkg.go.dev/github.com/trpc.group/trpc-cmdline)
[![Go Report Card](https://goreportcard.com/badge/github.com/trpc.group/trpc-go/trpc-cmdline)](https://goreportcard.com/report/github.com/trpc.group/trpc-go/trpc-cmdline)
[![LICENSE](https://img.shields.io/github/license/trpc.group/trpc-cmdline.svg?style=flat-square)](https://github.com/trpc.group/trpc-cmdline/blob/main/LICENSE)
[![Releases](https://img.shields.io/github/release/trpc.group/trpc-cmdline.svg?style=flat-square)](https://github.com/trpc.group/trpc-cmdline/releases)
[![Docs](https://img.shields.io/badge/docs-latest-green)](http://test.trpc.group.woa.com/docs/)
[![Tests](https://github.com/trpc-group/trpc-cmdline/actions/workflows/prc.yml/badge.svg)](https://github.com/trpc-group/trpc-cmdline/actions/workflows/prc.yml)
[![Coverage](https://codecov.io/gh/trpc.group/trpc-cmdline/branch/main/graph/badge.svg)](https://app.codecov.io/gh/trpc.group/trpc-cmdline/tree/main)

trpc-cmdline is the command line tool for [trpc-cpp](https://github.com/trpc-group/trpc-cpp) and [trpc-go](https://github.com/trpc-group/trpc-go).

It supports the latest three major releases of [Go](https://go.dev/doc/devel/release).

## Installation

### Install trpc-cmdline

#### Install using go command

First, add the following into your `~/.gitconfig`:

```bash
[url "ssh://git@github.com/"]
    insteadOf = https://github.com/
```

Then run the following to install `trpc-cmdline`:

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

### Dependencies

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


## Quick Start

### Generation of Full Project

* Copy and paste the following to `helloworld.proto`, you can get it from [./docs/helloworld/helloworld.proto](./docs/helloworld/helloworld.proto):

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

* Using trpc-cmdline to generate a full project:
```go
$ trpc create -p helloworld.proto -o out
```

Note: `-p` specifies proto file, `-o` specifies the output directory, 
for more information please run `trpc -h` and `trpc create -h`

* Enter the output directory and start the server:
```bash
$ cd out
$ go run .
...
... trpc service:helloworld.HelloWorldService launch success, tcp:127.0.0.1:8000, serving ...
...
```

* Open the output directory in another terminal and start the client:
```bash
$ go run cmd/client/main.go 
... simple  rpc   receive: 
```

Note: Since the implementation of server service is an empty operation and the client sends empty data, therefore the log shows that the simple rpc receives an empty string.

* Now you may try to modify the service implementation located in `hello_world_service.go` and the client implementation located in `cmd/client/main.go` to create an echo server. You can refer to [https://github.com/trpc-group/trpc-go/tree/main/examples/helloworld](https://github.com/trpc-group/trpc-go/tree/main/examples/helloworld) for inspiration.

* The generated files are explained below:

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

### Generation of RPC Stub

* Simply add `--rpconly` flag to generate rpc stub instead of a full project:
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

### Frequently Used Flags

The following lists some frequently used flags.

* `-f`: Force overwrite the existing code.
* `-d some-dir`: Search paths for pb files (including dependent pb files), can be specified multiple times.
* `--mock=false`: Disable generation of mock stub code.
* `--nogomod=true`: Do not generate go.mod file in the stub code, only effective when --rpconly=true, defaults to false.

For additional flags please run `trpc -h` and `trpc [subcmd] -h`.

### Functionalities

Please check [Documentation](./docs/)

## Contributing

This project is open-source and accepts contributions. See the [contribution guide](CONTRIBUTING.md) for more information.
