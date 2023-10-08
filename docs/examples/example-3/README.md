## example-3 文档相关

准确地说是 api 文档，在后台同学与 Web 同学或者相互之间沟通服务接口的时候，api 文档非常有帮助。

目前支持 swagger 以及 openapi 3.0 规范的 api 文档的生成。

为了保持功能的纯粹性，我们将 api 文档生成逻辑移动到了`trpc apidocs`子命令：

- 生成 swagger 文档，`trpc apidocs -p <protofile> --swagger`，默认输出到 apidocs.swagger.json
- 生成 openapi 文档，`trpc apidocs -p <protofile> --openapi`，默认输出到 apidocs.openapi.json

您也可以指定对应的输出选项来控制输出到的文件，或者控制请求参数的编码方式。

关于 swagger、openapi 相关 option 的使用，请参考 example-2 中的说明。

当已经生成了 swagger 文档之后，您可以安装 swagger 命令，执行`swagger server apidocs.swagger.json`来查看 api 文档。
您复制、粘贴到在线的 swagger editor 来查看 api 文档信息。

openapi 的使用方式大致与 swagger 类似，只是结构上有点差异。


