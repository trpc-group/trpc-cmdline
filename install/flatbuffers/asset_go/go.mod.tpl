{{- $pkgName := .PackageName -}}
{{- $svrName := (index .Services 0).Name -}}

{{ $domainName := .Domain }}

{{- $goPkgOption := "" -}}
{{- with .FileOptions.go_package -}}
  {{- $goPkgOption = . -}}
{{- end -}}

{{- if eq .GoMod "" -}}
module trpc.app.{{$svrName}}
{{- else -}}
module {{.GoMod}}
{{- end }}

go {{.GoVersion}}



{{ $rpcdir := "" -}}
{{ if ne $goPkgOption "" -}}
{{ $rpcdir = trimright ";" $goPkgOption }}
{{- else -}}
{{ $rpcdir = $pkgName }}
{{- end -}}
replace {{$rpcdir}} => ./stub/{{$rpcdir}}

{{ range $k, $v := .Pb2ImportPath -}}
{{ $v = trimright ";" $v}}
{{ if and (not (hasprefix "trpc.tech/trpc-go/trpc/v2" $v)) 
          (not (hasprefix "trpc.group/trpc-go/trpc" $v))
          (not (hasprefix "trpc.group/trpc/trpc-protocol" $v))
          (not (hasprefix "trpc.group/wineguo/trpc-protocol" $v))
          (not (hasprefix "github.com/golang/protobuf" $v)) 
          (ne $v "trpc.group/devsec/protoc-gen-secv/v2/validate")
          (ne $v "trpc.group/devsec/protoc-gen-secv/validate")
          (not (hasprefix "google/protobuf/" $k)) }}
replace {{$v}} => ./stub/{{$v}}
{{ end }}

{{ end }}


{{ if ne $.TRPCGoVersion "" }}
require trpc.group/trpc-go/trpc-go {{.TRPCGoVersion}}
{{ end }}

require github.com/google/flatbuffers v2.0.0+incompatible
