{{- /* $pkgName .PackageName example: trpc.testapp.testserver */}}
{{- $pkgName := .PackageName -}}
{{- /* $goPkgOption example: "trpc.group/trpcprotocol/testapp/testserver/greeter" */}}
{{- $goPkgOption := "" -}}
{{- with .FileOptions.go_package -}}
  {{- $goPkgOption = . -}}
{{- end -}}

{{- if ne $goPkgOption "" -}}
module {{$goPkgOption}}
{{- else -}}
module {{$pkgName}}
{{- end }}

go {{.GoVersion}}



{{ if ne $.TRPCGoVersion "" }}
require trpc.group/trpc-go/trpc-go {{.TRPCGoVersion}}
{{ end }}

require github.com/google/flatbuffers v2.0.0+incompatible
