{{ $domainName := .Domain }}

{{- $pkgName := .PackageName -}}
{{- $goPkgOption := "" -}}
{{- with .FileOptions.go_package -}}
  {{- $goPkgOption = . -}}
{{- end -}}
{{- $serviceSuffix := "" -}}
{{- if not .NoServiceSuffix }}
    {{- $serviceSuffix = "Service" }}
{{- end }}
package main

import (
	{{if secvtpl $.Pkg2ValidGoPkg -}}
		{{/* _ "{{$domainName}}/trpc-go/trpc-filter/validation" */}}
	{{ end -}}
	_ "{{ $domainName }}/trpc-go/trpc-filter/debuglog"
	_ "{{ $domainName }}/trpc-go/trpc-filter/recovery"
	trpc "{{ $domainName }}/trpc-go/trpc-go"
	"{{ $domainName }}/trpc-go/trpc-go/log"

    {{ if ne $goPkgOption "" -}}
   	fb "{{$goPkgOption}}"
    {{- else -}}
    fb "{{$pkgName}}"
	{{- end }}
)

{{- $appName := .AppName -}}
{{- $serverName := .ServerName }}

func main() {
	flag.Parse()
	s := trpc.NewServer()
	// If there are multiple services, it is necessary to explicitly write the service name as the first parameter; 
	// otherwise, there may be issues with streaming.
    {{range $index, $service := .Services}}
    {{- $svrNameCamelCase := $service.Name | camelcase -}}
	{{- $serviceName := $service.Name -}}
   	fb.Register{{$svrNameCamelCase}}{{$serviceSuffix}}(s.Service("{{- if and $appName $serverName -}}
        trpc.{{$appName}}.{{$serverName}}.{{$serviceName -}}
      {{- else -}}
        {{- $pkgName}}.{{$serviceName -}}
      {{- end -}}"), &{{$svrNameCamelCase|untitle}}Impl{})
	{{end -}}
	if err := s.Serve(); err != nil {
		log.Fatal(err)
	}
}
