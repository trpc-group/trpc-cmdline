{{ $domainName := .Domain }}
{{ $groupName := .GroupName }}
{{ $versionSuffix := .VersionSuffix }}
{{- $serviceSuffix := "" -}}
{{- if not .NoServiceSuffix }}
    {{- $serviceSuffix = "Service" }}
{{- end }}

{{- $pkgName := .PackageName -}}
{{- $goPkgOption := "" -}}
{{- with .FileOptions.go_package -}}
  {{- $goPkgOption = . -}}
{{- end -}}
package main

import (
	{{- if secvtpl $.Pkg2ValidGoPkg }}
		{{/* _ "{{$domainName}}/{{$groupName}}/trpc-filter/validation{{$versionSuffix}}" */}}
	{{- end }}
	_ "{{$domainName}}/{{$groupName}}/trpc-filter/debuglog{{$versionSuffix}}"
	_ "{{$domainName}}/{{$groupName}}/trpc-filter/recovery{{$versionSuffix}}"
	{{- if (or .ValidateEnabled .SecvEnabled)  }}
	_ "{{ $domainName }}/{{ $groupName }}/trpc-filter/validation{{ $versionSuffix }}"
	{{- end }}
	trpc "{{$domainName}}/{{$groupName}}/trpc-go{{$versionSuffix}}"
	"{{$domainName}}/{{$groupName}}/trpc-go{{$versionSuffix}}/log"
    {{ if ne $goPkgOption "" -}}
   	pb "{{ trimright ";" $goPkgOption }}"
    {{- else -}}
    pb "{{$pkgName}}"
	{{- end }}
)

{{- $appName := .AppName -}}
{{- $serverName := .ServerName }}

func main() {
	s := trpc.NewServer()
    {{range $index, $service := .Services}}
    {{- $svrNameCamelCase := $service.Name | camelcase -}}
	{{- $serviceName := $service.Name -}}
   	pb.Register{{$svrNameCamelCase}}{{$serviceSuffix}}(s.Service("{{- if and $appName $serverName -}}
        trpc.{{$appName}}.{{$serverName}}.{{$serviceName -}}
      {{- else -}}
        {{- $pkgName}}.{{$serviceName -}}
      {{- end -}}"), &{{$svrNameCamelCase|untitle}}Impl{})
	{{end -}}
	if err := s.Serve(); err != nil {
		log.Fatal(err)
	}
}
