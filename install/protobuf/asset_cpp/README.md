# Usage

- You can directly compile and run via scripts `run_server.sh` and `run_client.sh`.
- You can only do compile via script `build.sh` and check the binary generated at bazel-bin. The raw command to run server and client is like below, you are free to check the yaml file too:
```bash
./bazel-bin/server/server_bin --config=server/conf/trpc_cpp_fiber.yaml
./bazel-bin/client/fiber_client --client_config=client/conf/trpc_cpp_fiber.yaml
```
- You can clean all the resources build out via script `clean.sh`.
