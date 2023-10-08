
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
    "{{$domainName}}/{{$groupName}}/trpc-database/kafka{{$versionSuffix}}"
)

func main() {
	s := trpc.NewServer()

    kafka.RegisterHandlerService(s, handle)

	if err := s.Serve(); err != nil {
		log.Fatal(err)
	}
}

func handle(ctx context.Context, key, value []byte, topic string, partition int32, offset int64) error {
	return nil
}
