## 使用 trpc 工具生成 flatbuffers 桩代码

### 环境配置

用 `trpc` 工具创建 trpc-go flatbuffers 工程需要用到 `flatc` 工具，即 flatbuffers 官方提供的编译器

当前依赖的 flatbuffers 为 `v2.0.0`，官方 release 页面提供了编译好的二进制下载，但是在机器上可能会由于动态链接库的缺失而无法使用，这时我们需要从源码编译出 `flatc` 工具

首先得到相应版本的仓库：

```shell
$ git clone -b v2.0.0 --depth=1 https://github.com/google/flatbuffers.git
```

然后进行编译

```shell
$ cd flatbuffers 
$ # 如果没有 cmake 的话可以通过 yum install cmake -y 来安装
$ cmake . 
$ make -j 16 # 设置为 cpu 的核数来加快编译速度
$ make install # 头文件以及编译好的二进制文件就会被安装到 /usr/local 的相关目录下
```

__注：__ 假如在 make 步骤时因为 `-Werror=shadow` 而报错，可以将 `CMakeLists.txt` 中的这部分去掉，示例操作如下：

```shell
$ sed -i "s/-Werror=shadow//g" CMakeLists.txt
$ cmake . && make -j 16 && make install # 然后再运行 cmake 和 make 等
```

可以查看 `flatc` 自带的命令行选项说明：

```shell
$ flatc --help 
```

### 快速上手

类似使用 protobuf 的方式，先写好你的 flatbuffers 对应的 `.fbs` 文件，然后运行 `trpc` 工具来生成相应的桩代码

`testcase/flatbuffers/` 目录下面提供了一系列的例子，里面既包含给定的 `.fbs` 文件，还包含了其相应的 `.proto` 文件，这样可以方便大家进行对照比较。除此之外还提供了使用这些文件生成桩代码的脚本 `testcase/flatbuffers/run*.sh`，可以查看脚本内容进行学习。注意这些脚本需要在 `trpc-cmdline` 文件夹下执行


首先看 `greeter.fbs`，这是一个简单但又相当全面的测试文件，里面包含了两个 service 的定义，每个 service 又有四种类型：1. 一发一收 2. 客户端流式 3. 服务端流式 4. 双向流式

类似于 protobuf 中的 `option go_package`，flatbuffers 中通过添加一个 attribute 来支持这一功能，如：

```
attribute "go_package=trpc.group/trpcprotocol/testapp/greeter";
```

在 `trpc-cmdline` 目录下执行 `./testcase/flatbuffers/rungreeter.sh` 即可生成该文件的桩代码，或者可以直接把脚本中的命令拿出来跑：

```shell
$ trpc create --fbs testcase/flatbuffers/greeter.fbs -o out-greeter --mod trpc.group/testapp/testgreeter 
```

其中 `--fbs` 指定了 flatbuffers 的文件名（带相对路径），`-o` 指定了输出路径，`--mod` 指定了生成文件 `go.mod` 中 `package` 的内容，假如没有 `--mod` 的话，它会寻找当前目录下的 `go.mod` 文件，以该文件中的 `package` 内容作为 `--mod` 的内容

生成的代码目录结构如下：

```shell
├── cmd/client/main.go # 客户端代码
├── go.mod
├── go.sum
├── greeter_2.go       # 第二个 service 的服务端实现
├── greeter_2_test.go  # 第二个 service 的服务端测试
├── greeter.go         # 第一个 service 的服务端实现
├── greeter_test.go    # 第一个 service 的服务端测试
├── main.go            # 服务启动代码
├── stub/trpc.group/trpcprotocol/testapp/greeter # 桩代码文件
└── trpc_go.yaml       # 配置文件
```

在一个终端内，编译并运行服务端：

```shell
$ go build      # 编译
$ ./testgreeter # 运行
```

在另一个终端内，运行客户端：

```shell
$ go run cmd/client/main.go 
```

然后可以在两个终端的 log 中查看相互发送的消息

如果想要在 service 中实现一些自己的业务逻辑，可以查看相应 service 中 method 的实现，查看注释中提供的示例来熟悉 flatbuffers 消息的构建方式，比如查看 `cmd/client/main.go` 来了解客户端如何构建一个 flatbuffers 的消息来发送给服务端：

```go
func callGreeterSayHello() {
	proxy := fb.NewGreeterClientProxy(
		client.WithTarget("ip://127.0.0.1:8000"),
		client.WithProtocol("trpc"),
	)
	ctx := trpc.BackgroundContext()
	// 一发一收 client 用法示例
	b := flatbuffers.NewBuilder(0)
	// 添加字段示例
	// 将 CreateString 中的 String 替换为你想要操作的字段类型
	// 将 AddMessage 中的 Message 替换为你想要操作的字段名
	// i := b.CreateString("GreeterSayHello")
	fb.HelloRequestStart(b)
	// fb.HelloRequestAddMessage(b, i)
	b.Finish(fb.HelloRequestEnd(b))
	reply, err := proxy.SayHello(ctx, b)
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	// 将 Message 替换为你需要访问的字段名
	// log.Debugf("simple  rpc   receive: %q", reply.Message())
	log.Debugf("simple  rpc   receive: %v", reply)
}
```

一般流程如下：

```go
// 导入桩代码的 package
import fb "trpc.group/trpcprotocol/testapp/greeter"
// 首先创建一个 *flatbuffers.Builder
// 其参数为底层 buffer 的初始大小
// 一般默认为 1024 
b := flatbuffers.NewBuilder(1024) 
// 想要为结构体填充字段的话
// 首先创建一个该字段类型的对象
// 比如想要填充的字段类型为 String
// 就可以调用 b.CreateString("a string") 来创建这个字符串
// 该方法返回的是在 flatbuffer 中的 index
i := b.CreateString("GreeterSayHello")
// 想要构造一个 HelloRequest 结构体
// 需要调用桩代码中提供的 XXXXStart 方法
// 表示该结构体构造的开始
// 其相对应的结束为 fb.HelloRequestEnd 
fb.HelloRequestStart(b)
// 该填充字段的名字为 message
// 就可以调用 fb.HelloRequestAddMessage(b, i)
// 通过传入 builder 以及之前构造的字符串的 index 来构造这个 message 字段
// 其他字段可以通过这种方式不断进行构造
fb.HelloRequestAddMessage(b, i)
// 当结构体构造结束时调用 XXXEnd 方法
// 该方法会返回这个结构体在 flatbuffer 中的 index
// 然后调用 b.Finish 可以结束这个 flatbuffer 的构造
b.Finish(fb.HelloRequestEnd(b))
```

flatbuffers 在发消息的时候是 `flatbuffers.Builder`，序列化时调用 `Marshal` 是直接将其中的 `[]byte` 拿了出来，反序列化时调用 `Unmarshal` 则是从这个 `[]byte` 中构建出原本的结构体，所以服务端在收到消息时得到的直接是该结构体，比如可以查看 `greeter.go` 文件里面 `SayHello` 的实现：

```go
func (s *greeterServiceImpl) SayHello(ctx context.Context, req *fb.HelloRequest, b *flatbuffers.Builder) error {
	// 单发单收 flatbuffers 处理逻辑（仅供参考，请根据需要修改）
	log.Debugf("Simple server receive %v", req)
	// 将 Message 替换为你想要操作的字段名
	v := req.Message() // Get Message field of request.
	var m string
	if v == nil {
	 	m = "Unknown"
	} else {
	 	m = string(v)
	}
	// 添加字段示例
	// 将 CreateString 中的 String 替换为你想要操作的字段类型
	// 将 AddMessage 中的 Message 替换为你想要操作的字段名
	idx := b.CreateString("welcome " + m) // 创建一个 flatbuffers 中的字符串
	fb.HelloReplyStart(b)
	fb.HelloReplyAddMessage(b, idx)
	b.Finish(fb.HelloReplyEnd(b))
	return nil
}
```

想要访问收到消息中的字段时，直接如下访问即可：

```go
req.Message() // 访问 req 中的 message 字段
```

更多用法可以查看生成文件中的示例以及 `testcase/flatbuffers/` 下面各种其他的测试用例，各测试用例简要特点概括如下：

* 1-without-import: RPC 方法需要的结构体定义都放在一个文件中，没有 include 其他文件，这是最简单的情况

* 2-multi-fb-same-namespace: 在同一目录下有多个 `.fbs` 文件，每个 `.fbs` 文件的 `namespace` 都是一样的（flatbuffers 中的 `namespace` 等同于 protobuf 中的 `package` 语句），然后其中一个主文件 include 了其他 `.fbs` 文件

* 3-multi-fb-diff-namespace: 在同一个目录下有多个 `.fbs` 文件，每个 `.fbs` 文件的 `namespace` 不一样，比如定义 RPC 的主文件中引用了不同 `namespace` 中的类型

* 4.1-multi-fb-same-namespace-diff-dir: 多个 `.fbs` 文件的 `namespace` 相同，但是在不同的目录下，主文件 `helloworld.fbs` 中在 include 其他文件时使用相对路径，可以看下 `run4.1.sh`，其中并不需要用 `--fbsdir` 来指定搜索路径

* 4.2-multi-fb-same-namespace-diff-dir: 除了 `helloworld.fbs` 文件中 include 语句里面只使用文件名以外，其余和 4.1 完全相同，这个例子想要正确运行，需要添加 `--fbsdir` 来指定搜索路径，见 `run4.2.sh`：
```shell
trpc create --fbsdir testcase/flatbuffers/4.2-multi-fb-same-namespace-diff-dir/request \
			--fbsdir testcase/flatbuffers/4.2-multi-fb-same-namespace-diff-dir/response \
			--fbs testcase/flatbuffers/4.2-multi-fb-same-namespace-diff-dir/helloworld.fbs \
			-o out-4-2 \
			--mod trpc.group/testapp/testserver42
```
所以为了尽可能简化命令行参数，建议在 include 语句时写上文件的相对路径（如果不在一个文件夹中的话）

* 5-multi-fb-diff-gopkg: 多个 `.fbs` 文件，多文件之间有 include 关系，他们的 `go_package` 不相同。注意：由于 `flatc` 的限制，目前不支持两个文件在 `namespace` 相同的情况下 `go_package` 却不同，并要求一个文件中的 `namespace` 和 `go_package` 的最后一段必须相同（比如 `trpc.testapp.testserver` 和 `trpc.group/testapp/testserver` 最后一段 `testserver` 是相同的）

### 实现细节

1. 添加了 `--fbs` 标志，用于指定 flatbuffers 的文件路径
2. 添加了 `--fbsdir` 标志，用于指定 include 文件的搜索路径
3. 添加对 `.fbs` 文件的解析，解析操作由 `https://trpc.group/trpc-go/fbs` 完成，见 `parser/parser.go`
4. `descriptor.FileDescriptor` 中的 `FD` 字段提为 `interface`，分成了 protobuf 以及 flatbuffers 两种不同的实现，分别放在 `descriptor/protodesc.go` 以及 `descriptor/fbsdesc.go` 中
5. 为 flatbuffers 添加自己的模板文件，放在了 `install/flatbuffers` 目录下，protobuf 相关的模板文件移动到了 `install/protobuf` 目录下，相应修改了配置文件 `install/trpc.yaml` 以及 `config.go` 中的定义
6. 添加对 `flatc` 调用的封装，见 `util/fb/`，相关使用见 `create.go`
7. 添加了各种 testcase 以及相关运行脚本，见 `testcase/flatbuffers/`，测试代码见 `create_test.go` 中的 `TestCreateCmdByFlatbuffers`
