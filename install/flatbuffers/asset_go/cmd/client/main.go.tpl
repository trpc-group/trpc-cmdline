package main 

{{ $domainName := .Domain }}
{{- $sevriceProtocol := .Protocol -}}
{{- $goPkgName := .PackageName -}}
{{- with .FileOptions.go_package -}}
  {{- $goPkgName = . -}}
{{- end -}}
import (
	trpc "{{$domainName}}/trpc-go/trpc-go"
	"{{$domainName}}/trpc-go/trpc-go/client"
	"{{$domainName}}/trpc-go/trpc-go/log"
	fb "{{$goPkgName}}"
	flatbuffers "github.com/google/flatbuffers/go" 
	{{ range $.Imports }}
		{{/* Specify the importName explicitly, for example: import "{{ $domainName }}/..../structpb;xxx" */}}
		{{ if contains . ";" }}
			{{ $val := (splitList ";" .) }}
			{{ index $val 1}} "{{index $val 0}}"
		{{ else }}
			"{{- . -}}"
		{{ end }}
	{{ end }}
)
{{- range $index, $service := .Services -}}
{{- $svrNameCamelCase := $service.Name | camelcase -}}
{{- range $mindex, $method := $service.RPC -}}
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
{{- end }}

func call{{$svrNameCamelCase}}{{$rpcName}}() {
	proxy := fb.New{{$svrNameCamelCase}}ClientProxy(
		client.WithTarget("ip://127.0.0.1:{{add 8000 $index}}"),
		client.WithProtocol("{{$sevriceProtocol}}"),
	)
	ctx := trpc.BackgroundContext()
{{- if and $method.ClientStreaming $method.ServerStreaming}}
	// Example of using a bidirectional streaming client.
	stream, err := proxy.{{$rpcName}}(ctx)
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	for i := 0; i < 5; i++ {
		b := flatbuffers.NewBuilder(clientFBBuilderInitialSize)
		// Example of Adding a Field.
		// Replace the "String" in CreateString with the field type you want to work with.
		// Replace "Message" in AddMessage with the field name you want to work with.
		// idx := b.CreateString(fmt.Sprintf("{{$svrNameCamelCase}}{{$rpcName}} %v", i))
		{{$rpcReqType}}Start(b)
		// {{$rpcReqType}}AddMessage(b, idx)
		b.Finish({{$rpcReqType}}End(b))
		if err := stream.Send(b); err != nil {
			log.Fatalf("err: %v", err)
		}
	}
	if err := stream.CloseSend(); err != nil {
		log.Fatalf("err: %v", err)
	}
	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			break 
		}
		if err != nil {
			log.Fatalf("err: %v", err)
		}
		// Replace "Message" with the field name you need to access.
		// log.Debugf(" bidi  stream receive: %q", rsp.Message())
		log.Debugf(" bidi  stream receive: %v", rsp)
	}
{{- else if $method.ClientStreaming}}
	// Example usage of client-side streaming.
	stream, err := proxy.{{$rpcName}}(ctx)
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	for i := 0; i < 5; i++ {
		b := flatbuffers.NewBuilder(clientFBBuilderInitialSize)
		// Example of Adding a Field.
		// Replace the "String" in CreateString with the field type you want to work with.
		// Replace "Message" in AddMessage with the field name you want to work with.
		// idx := b.CreateString(fmt.Sprintf("{{$svrNameCamelCase}}{{$rpcName}} %v", i))
		{{$rpcReqType}}Start(b)
		// {{$rpcReqType}}AddMessage(b, idx)
		b.Finish({{$rpcReqType}}End(b))
		if err := stream.Send(b); err != nil {
			log.Fatalf("err: %v", err)
		}
	}
	rsp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	// Replace "Message" with the field name you need to access.
	// log.Debugf("client stream receive: %q", rsp.Message())
	log.Debugf("client stream receive: %v", rsp)
{{- else if $method.ServerStreaming}}
	// Example usage of server-side streaming.
	b := flatbuffers.NewBuilder(clientFBBuilderInitialSize)
	// Example of Adding a Field.
	// Replace the "String" in CreateString with the field type you want to work with.
	// Replace "Message" in AddMessage with the field name you want to work with.
	// i := b.CreateString("{{$svrNameCamelCase}}{{$rpcName}}")
	{{$rpcReqType}}Start(b)
	// {{$rpcReqType}}AddMessage(b, i)
	b.Finish({{$rpcReqType}}End(b))
	stream, err := proxy.{{$rpcName}}(ctx, b)
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("err: %v", err)
		}
		// Replace "Message" with the field name you need to access.
		// log.Debugf("server stream receive: %q", reply.Message())
		log.Debugf("server stream receive: %v", reply)
	}
{{- else}}
	// Example usage of unary client.
	b := flatbuffers.NewBuilder(clientFBBuilderInitialSize)
	// Example of Adding a Field.
	// Replace the "String" in CreateString with the field type you want to work with.
	// Replace "Message" in AddMessage with the field name you want to work with.
	// i := b.CreateString("{{$svrNameCamelCase}}{{$rpcName}}")
	{{$rpcReqType}}Start(b)
	// {{$rpcReqType}}AddMessage(b, i)
	b.Finish({{$rpcReqType}}End(b))
	reply, err := proxy.{{$rpcName}}(ctx, b)
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	// Replace "Message" with the field name you need to access.
	// log.Debugf("simple  rpc   receive: %q", reply.Message())
	log.Debugf("simple  rpc   receive: %v", reply)
{{- end}}
}
{{- end}}
{{- end}}

// clientFBBuilderInitialSize sets the initial size for initializing flatbuffers.NewBuilder on the client side.
var clientFBBuilderInitialSize int

func init() {
	flag.IntVar(&clientFBBuilderInitialSize, "n", 1024, "set client flatbuffers builder's initial size")
}

func main() {
	flag.Parse()
{{- range $index, $service := .Services -}}
{{- $svrNameCamelCase := $service.Name | camelcase -}}
{{- range $mindex, $method := $service.RPC -}}
{{- $rpcName := $method.Name | camelcase }}
	call{{$svrNameCamelCase}}{{$rpcName}}() 
{{- end}}
{{- end}}
}