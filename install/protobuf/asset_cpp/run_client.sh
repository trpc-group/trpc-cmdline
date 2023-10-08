#!/bin/bash

bazel build //client/...
bazel-bin/client/fiber_client --client_config=client/conf/trpc_cpp_fiber.yaml
bazel-bin/client/future_client --client_config=client/conf/trpc_cpp_future.yaml
