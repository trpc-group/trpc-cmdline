{{- $pkgName := .PackageName -}}
{{- $svrName := (index .Services 0).Name -}}

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
