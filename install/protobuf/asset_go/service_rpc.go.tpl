{{ $domainName := .Domain }}

{{- $goPkgName := .PackageName -}}
{{- with .FileOptions.go_package -}}
  {{- $goPkgName = . -}}
{{- end -}}
package main

import (
	"context"
	{{- $importio := false }}
	{{- range (index .Services .ServiceIndex).RPC }}
		{{- if or .ClientStreaming .ServerStreaming }}
			{{- $importio = true }}
		{{- end }}
	{{- end }}
	{{- if $importio }}
		"io"
	{{- end }}

	pb "{{ trimright ";" $goPkgName }}"
	{{ range $.ImportsX }}
		{{.Name}} "{{.Path}}"
	{{ end }}
)

{{/* Regardless of how the files are divided, .ServiceIndex is always valid */}}
{{ $service := (index .Services .ServiceIndex) -}}
{{- $svrName := $service.Name | camelcase | untitle -}}
{{- $svrNameCamelCase := $service.Name|camelcase -}}
{{- $unimplementedName := (printf "pb.Unimplemented%s" ($svrNameCamelCase)) -}}

{{/* Regardless of how the files are divided, generate the Service type definition when .MethodIndex <= 0 */}}
{{- if le $.MethodIndex 0 }}
type {{$svrName}}Impl struct {
	{{$unimplementedName}}
}
{{- end }}

{{/* Execute this for the current loop, regardless of whether it is divided by method or the index value passed when dividing by method */}}
{{ range $index, $method := $service.RPC }}

{{ if or (not $.PerMethod) (eq $.MethodIndex $index) }}
{{- $rpcName := $method.Name | camelcase -}}
{{- $rpcReqType := $method.RequestType -}}
{{- $rpcRspType := $method.ResponseType -}}

{{- $reqTypePkg := $method.RequestTypePkgDirective -}}
{{- with $method.RequestTypeFileOptions.go_package -}}
  {{- $reqTypePkg = . -}}
{{- end -}}

{{- $rspTypePkg := $method.ResponseTypePkgDirective -}}
{{- with $method.ResponseTypeFileOptions.go_package -}}
  {{- $rspTypePkg = . -}}
{{- end -}}

{{- if (eq $reqTypePkg $goPkgName) -}}
	{{- $rpcReqType = (printf "pb.%s" (splitList "." $rpcReqType|last|export|camelcase)) -}}
{{- else -}}
	{{- $rpcReqType = (gofulltypex $rpcReqType $.FileDescriptor) -}}
{{- end -}}

{{- if (eq $rspTypePkg $goPkgName) -}}
	{{- $rpcRspType = (printf "pb.%s" (splitList "." $rpcRspType|last|export|camelcase)) -}}
{{- else -}}
	{{- $rpcRspType = (gofulltypex $rpcRspType $.FileDescriptor) -}}
{{- end -}}

{{ with .LeadingComments }}// {{$rpcName}} {{.}}{{ end }}
{{- if and $method.ClientStreaming $method.ServerStreaming }}
func (s *{{$svrName}}Impl) {{$rpcName}}(stream {{index (splitList "." $rpcReqType) 0}}.{{$svrNameCamelCase}}_{{$rpcName}}Server) error {
	// Bidirectional streaming scenario processing logic (for reference only, please modify as needed).
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		err = stream.Send(&{{$rpcRspType}}{})
		if err != nil {
			return err
		}
	}
}
{{- else }}
{{- if $method.ClientStreaming }}
func (s *{{$svrName}}Impl) {{$rpcName}}(stream {{index (splitList "." $rpcReqType) 0}}.{{$svrNameCamelCase}}_{{$rpcName}}Server) error {
	// Client streaming scenario processing logic (for reference only, please modify as needed).
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&{{$rpcRspType}}{})
		}
		if err != nil {
			return err
		}
	}
}
{{- else }}
{{- if $method.ServerStreaming }}
func (s *{{$svrName}}Impl) {{$rpcName}}(req *{{$rpcReqType}}, stream {{index (splitList "." $rpcReqType) 0}}.{{$svrNameCamelCase}}_{{$rpcName}}Server) error {
	// Server streaming scenario processing logic (for reference only, please modify as needed).
	for i := 0; i < 5; i++ {
		err := stream.Send(&{{$rpcRspType}}{})
		if err != nil {
			return err
		}
	}
	return nil
}
{{- else }}
func (s *{{$svrName}}Impl) {{$rpcName}}(ctx context.Context, req *{{$rpcReqType}}) (*{{$rpcRspType}}, error) {
    rsp := &{{$rpcRspType}}{}
	return rsp, nil
}
{{- end }}
{{- end }}
{{- end }}
{{- end }}
{{end}}
