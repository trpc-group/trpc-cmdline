# Usage

## 目录文件说明

- `requirements.txt`: 安装依赖所需的包依赖
- `test_service_rpc.py`: 测试服务的客户端代码
- `trpc_main.py`: 服务启动代码
- `trpc_python.yaml`: 服务启动依赖代码
- `.gitignore`: git忽略文件
- `build.sh`: 检查环境中的python版本,创建对应的虚拟环境,安装依赖
- `clean.sh`: 清理运行生成的非必要文件
- `coverage.sh`: 覆盖率生成文件
- `run_server.sh`: 启动服务端脚本
- `run_client.sh`: 启动客户端测试
- `stub`: 生成的桩代码目录,不允许修改
- `requirements`: 包含框架运行的包和测试包依赖;可以分多个文件存储,方便管理

## 代码运行

- 执行脚本运行服务: `sh +x run_server.sh` 
- 执行脚本执行客户端访问: `sh +x run_client.sh` 


## 其他说明

工具会拷贝业务的proto文件到stub的子目录中, 工具根据proto文件中的service字段对应的名称构建服务代码文件,用户如果需要修改业务逻辑在其中修改即可;由于生成的代码没有具体实现接口逻辑;要求用户去实现,不然测试会抛异常
