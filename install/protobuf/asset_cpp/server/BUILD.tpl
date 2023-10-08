cc_binary(
    name = "server_bin",
    deps = [
        ":server",
    ],
)

cc_library(
    name = "server",
    srcs = ["server.cc"],
    hdrs = ["server.h"],
    deps = [
        ":service",
        "@com_github_fmtlib_fmt//:fmtlib",
        "@trpc_cpp//trpc/common:trpc_app",
        "@trpc_cpp//trpc/log:trpc_log",
    ],
)

cc_library(
    name = "service",
    srcs = ["service.cc"],
    hdrs = ["service.h"],
    deps = [
        {{- if eq (dir .RelatvieFilePath) "." }}
        "@proto//:{{ basenamewithoutext .FilePath | lower }}_proto",
        {{- else }}
        "@proto//:{{ (join (splitList "/" (dir .RelatvieFilePath )) "_") }}_{{ basenamewithoutext .FilePath | lower }}_proto",
        {{- end }}
        "@trpc_cpp//trpc/log:trpc_log",
    ],
)
