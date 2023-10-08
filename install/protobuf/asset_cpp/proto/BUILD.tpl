package(default_visibility = ["//visibility:public"])

load("@trpc_cpp//trpc:trpc.bzl", "trpc_proto_library")
{{ range $pb, $deps := .Pb2DepsPbs -}}
{{- $depsCount := (len $deps) }}
trpc_proto_library(
    {{ if eq (dir $pb ) "." -}}
    name = "{{ basenamewithoutext $pb | lower }}_proto",
    {{ else -}}
    name = "{{ (join (splitList "/" (dir $pb )) "_") }}_{{ basenamewithoutext $pb | lower }}_proto",
    {{ end -}}
    srcs = ["{{ $pb }}"],
    rootpath = "@trpc_cpp",
    use_trpc_plugin = True,
    {{- if ne $depsCount 0 }}
    deps = [
      {{- range $dep := $deps }}
      {{ if eq (dir $dep ) "." -}}
      ":{{ basenamewithoutext $dep | lower }}_proto",
      {{- else -}}
      ":{{ (join (splitList "/" (dir $dep )) "_") }}_{{ basenamewithoutext $dep | lower }}_proto",
      {{- end }}
      {{- end }}
    ],
    {{- end }}
)
{{ end }}