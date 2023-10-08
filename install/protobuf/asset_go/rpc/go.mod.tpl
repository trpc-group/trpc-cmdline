{{ $domainName := .Domain }}
{{ $groupName := .GroupName }}

{{- $pkgName := .PackageName -}}
{{- $goPkgOption := "" -}}
{{- with .FileOptions.go_package -}}
  {{- $goPkgOption = . -}}
{{- end -}}

{{- if ne $goPkgOption "" -}}
module {{trimright ";" $goPkgOption}}
{{- else -}}
module {{$pkgName}}
{{- end }}

go {{.GoVersion}}

{{ if ne $.TRPCGoVersion "" }}
require {{ $domainName }}/{{ $groupName }}/trpc-go {{.TRPCGoVersion}}
{{ end }}
