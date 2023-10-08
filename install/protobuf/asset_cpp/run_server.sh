#!/bin/bash

bazel build //server/...
bazel-bin/server/server_bin --config=server/conf/trpc_cpp_fiber.yaml
# bazel-bin/server/server_bin --config=server/conf/trpc_cpp.yaml
