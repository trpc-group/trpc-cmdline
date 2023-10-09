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

	flatbuffers "github.com/google/flatbuffers/go" 
	fb "{{$goPkgName}}"
	"{{$domainName}}/trpc-go/trpc-go/log"

	{{ range $.Imports }}
		{{/* In the case of explicitly specifying importName, for example: import "{{ $domainName }}/..../structpb;xxx" */}}
		{{ if contains . ";" }}
			{{ $val := (splitList ";" .) }}
			{{ index $val 1}} "{{index $val 0}}"
		{{ else }}
			"{{- . -}}"
		{{ end }}
	{{ end }}
)

{{/* ".ServiceIndex" remains valid regardless of how the files are divided. */}}
{{ $service := (index .Services .ServiceIndex) -}}
{{- $svrName := $service.Name | camelcase | untitle -}}
{{- $svrNameCamelCase := $service.Name|camelcase -}}

{{/* Generate the Service type definition when .MethodIndex is less than or equal to 0, regardless of how the files are divided. */}}
{{- if le $.MethodIndex 0 }}
type {{$svrName}}Impl struct {}
{{- end }}

{{/* During the current loop, the execution is necessary if the files are not divided by method, or if the index value passed when dividing the files by method needs to be executed. */}}
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
	{{- $rpcReqType = (printf "fb.%s" (splitList "." $rpcReqType|last|export|camelcase)) -}}
{{- else -}}
	{{- $rpcReqType = (gofulltype $rpcReqType $.FileDescriptor) -}}
{{- end -}}

{{- if (eq $rspTypePkg $goPkgName) -}}
	{{- $rpcRspType = (printf "fb.%s" (splitList "." $rpcRspType|last|export|camelcase)) -}}
{{- else -}}
	{{- $rpcRspType = (gofulltype $rpcRspType $.FileDescriptor) -}}
{{- end -}}

{{ with .LeadingComments }}// {{$rpcName}} {{.}}{{ end }}
{{- if and $method.ClientStreaming $method.ServerStreaming }}
func (s *{{$svrName}}Impl) {{$rpcName}}(
	stream {{index (splitList "." $rpcReqType) 0}}.{{$svrNameCamelCase}}_{{$rpcName}}Server,
) error {
	// Bidirectional streaming scenario processing logic (for reference only, please modify as needed).
	for {
		req, err := stream.Recv()
		log.Debugf("Bidi server receive %v", req)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		b := flatbuffers.NewBuilder(0)
		for _, greeting := range [...]string{"Hello", "Hi   ", "Hola "} {
			log.Debugf("Bidi server is about to send %v", greeting)
			// Example of Adding a Field.
			// Replace the "String" in CreateString with the field type you want to work with.
			// Replace "Message" in AddMessage with the field name you want to work with.
			// idx := b.CreateString(fmt.Sprintf("%v %v", greeting, string(req.Message())))
			{{$rpcRspType}}Start(b)
			// {{$rpcRspType}}AddMessage(b, idx)
			b.Finish({{$rpcRspType}}End(b))
			if err := stream.Send(b); err != nil {
				return err
			}
		}
	}
}
{{- else if $method.ClientStreaming }}
func (s *{{$svrName}}Impl) {{$rpcName}}(
	stream {{index (splitList "." $rpcReqType) 0}}.{{$svrNameCamelCase}}_{{$rpcName}}Server,
) error {
	// Client streaming scenario processing logic (for reference only, please modify as needed).
	// all := []string{}
	for {
		req, err := stream.Recv()
		log.Debugf("StreamClient server receive %v", req)
		if err == io.EOF {
			b := flatbuffers.NewBuilder(0)
			// Example of Adding a Field.
			// Replace the "String" in CreateString with the field type you want to work with.
			// Replace "Message" in AddMessage with the field name you want to work with.
			// idx := b.CreateString(strings.Join(all, ", "))
			{{$rpcRspType}}Start(b)
			// {{$rpcRspType}}AddMessage(b, idx)
			b.Finish({{$rpcRspType}}End(b))
			return stream.SendAndClose(b)
		}
		if err != nil {
			return err
		}
		// Replace 'Message' with the field name you want to operate on.
		// all = append(all, string(req.Message()))
	}
}
{{- else if $method.ServerStreaming }}
func (s *{{$svrName}}Impl) {{$rpcName}}(
	req *{{$rpcReqType}}, 
	stream {{index (splitList "." $rpcReqType) 0}}.{{$svrNameCamelCase}}_{{$rpcName}}Server,
) error {
	// Server streaming scenario processing logic (for reference only, please modify as needed).
	log.Debugf("StreamClient server receive %v", req)
	for i := 0; i < 5; i++ {
		b := flatbuffers.NewBuilder(0)
		// Example of Adding a Field.
		// Replace the "String" in CreateString with the field type you want to work with.
		// Replace "Message" in AddMessage with the field name you want to work with.
		// idx := b.CreateString(fmt.Sprintf("Hello %v %v", string(req.Message()), i))
		{{$rpcRspType}}Start(b)
		// {{$rpcRspType}}AddMessage(b, idx)
		b.Finish({{$rpcRspType}}End(b))
		if err := stream.Send(b); err != nil {
			return err
		}
	}
	return nil
}
{{- else }}
func (s *{{$svrName}}Impl) {{$rpcName}}(
	ctx context.Context, 
	req *{{$rpcReqType}},
) (*flatbuffers.Builder, error) {
	// Unary call: flatbuffers processing logic (for reference only, please modify as needed).
	log.Debugf("Simple server receive %v", req)
	// Replace 'Message' with the field name you want to operate on.
	// v := req.Message() // Get Message field of request.
	// var m string
	// if v == nil {
	// 	m = "Unknown"
	// } else {
	// 	m = string(v)
	// }
	// Example of Adding a Field.
	// Replace the "String" in CreateString with the field type you want to work with.
	// Replace "Message" in AddMessage with the field name you want to work with.
	// idx := b.CreateString("welcome " + m) // Create a string in flatbuffers.
	b := &flatbuffers.Builder{}
	{{$rpcRspType}}Start(b)
	// {{$rpcRspType}}AddMessage(b, idx)
	b.Finish({{$rpcRspType}}End(b))
	return b, nil
}
{{- end }}
{{- end }}
{{- end }}
