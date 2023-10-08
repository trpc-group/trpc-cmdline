
{{ $domainName := .Domain }}
{{ $groupName := .GroupName }}
{{ $versionSuffix := .VersionSuffix }}

{{- $pkgName := .PackageName -}}
{{- $goPkgOption := "" -}}
{{- with .FileOptions.go_package -}}
  {{- $goPkgOption = . -}}
{{- end -}}
package main

import (
	_ "{{$domainName}}/{{$groupName}}/trpc-filter/debuglog{{$versionSuffix}}"
	_ "{{$domainName}}/{{$groupName}}/trpc-filter/recovery{{$versionSuffix}}"
	_ "go.uber.org/automaxprocs"

	"{{$domainName}}/{{$groupName}}/trpc-go{{$versionSuffix}}/log"

	trpc "{{$domainName}}/{{$groupName}}/trpc-go{{$versionSuffix}}"
	thttp "{{$domainName}}/{{$groupName}}/trpc-go{{$versionSuffix}}/http"
)

func main() {
	s := trpc.NewServer()

    thttp.HandleFunc("/", handle)
    thttp.RegisterDefaultService(s)

	if err := s.Serve(); err != nil {
		log.Fatal(err)
	}
}

func handle(w http.ResponseWriter, r *http.Request) error {
	return nil
}
