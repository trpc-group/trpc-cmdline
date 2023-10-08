cc_binary(
    name = "future_client",
    srcs = ["future_client.cc"],
    deps = [
        {{- if eq (dir .RelatvieFilePath) "." }}
        "@proto//:{{ basenamewithoutext .FilePath | lower }}_proto",
        {{- else }}
        "@proto//:{{ (join (splitList "/" (dir .RelatvieFilePath )) "_") }}_{{ basenamewithoutext .FilePath | lower }}_proto",
        {{- end }}
        "@trpc_cpp//trpc/client:make_client_context",
        "@trpc_cpp//trpc/client:trpc_client",
        "@trpc_cpp//trpc/common:runtime_manager",
        "@trpc_cpp//trpc/log:trpc_log",
        "@trpc_cpp//trpc/util/thread:latch",
        "@com_github_gflags_gflags//:gflags",
    ],
)

cc_binary(
    name = "fiber_client",
    srcs = ["fiber_client.cc"],
    deps = [
        {{- if eq (dir .RelatvieFilePath) "." }}
        "@proto//:{{ basenamewithoutext .FilePath | lower }}_proto",
        {{- else }}
        "@proto//:{{ (join (splitList "/" (dir .RelatvieFilePath )) "_") }}_{{ basenamewithoutext .FilePath | lower }}_proto",
        {{- end }}
        "@trpc_cpp//trpc/client:make_client_context",
        "@trpc_cpp//trpc/client:trpc_client",
        "@trpc_cpp//trpc/common/config:trpc_config",
        "@trpc_cpp//trpc/common:runtime_manager",
        "@trpc_cpp//trpc/log:trpc_log",
        "@com_github_gflags_gflags//:gflags",
    ],
)
