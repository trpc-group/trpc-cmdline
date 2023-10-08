{{ $domainName := .Domain }}

{{- $svrNameCamelCase := (index .Services .ServiceIndex).Name | camelcase -}}
{{- $goPkgName := .PackageName -}}
{{- with .FileOptions.go_package -}}
{{- $goPkgName = . -}}
{{- end -}}

{{ $fname := (basenamewithoutext .FilePath) -}}
{{- $serviceIndex := .ServiceIndex -}}
{{ $service := (index .Services .ServiceIndex) -}}

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
    "testing"

	trpc "{{ $domainName }}/trpc-go/trpc-go"
	_ "{{ $domainName }}/trpc-go/trpc-go/http"

    "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	flatbuffers "github.com/google/flatbuffers/go" 
	fb "{{$goPkgName}}"

    {{ range .Imports }}
   		{{ if contains . ";" }}
   			{{ $val := (splitList ";" .) }}
   			{{ index $val 1}} "{{index $val 0}}"
   		{{ else }}
			"{{- . -}}"
		{{ end }}
    {{- end }}
)

{{$svrName := $service.Name | camelcase | untitle}}
{{ $commaSeparated := (print ($svrName|title) "ClientProxy")}}

{{- range (index .Services .ServiceIndex).RPC }}
{{- if or .ClientStreaming .ServerStreaming }}
{{- $rpcName := .Name | camelcase -}}

{{- $commaSeparated = (print $commaSeparated "," $svrNameCamelCase "_" $rpcName "Client") }}
{{- $commaSeparated = (print $commaSeparated "," $svrNameCamelCase "_" $rpcName "Server") }}
{{- end }}
{{- end }}

//go:generate go mod tidy
{{ $p1 := trimright ";" $goPkgName }}
{{ $p2 := splitList ";" $goPkgName | last | gopkg_simple }}
//go:generate mockgen -destination=stub/{{$p1}}/{{$fname}}_mock.go -package={{$p2}} -self_package={{$p1}} --source=stub/{{$p1}}/{{$fname}}.trpc.go

{{range $index, $method := (index .Services .ServiceIndex).RPC}}
{{- $rpcName := $method.Name | camelcase -}}
{{-  $rpcReqType  := $method.RequestType -}}
{{-  $rpcRspType  := $method.ResponseType -}}

{{- $reqTypePkg := $method.RequestTypePkgDirective -}}
{{- with $method.RequestTypeFileOptions.go_package -}}
  {{- $reqTypePkg = . -}}
{{- end -}}

{{- $rspTypePkg := $method.ResponseTypePkgDirective -}}
{{- with $method.ResponseTypeFileOptions.go_package -}}
  {{- $rspTypePkg = . -}}
{{- end -}}

{{- if (eq $reqTypePkg $goPkgName) -}}
	{{-  $rpcReqType  = (printf "fb.%s" (splitList "."  $rpcReqType |last|export|camelcase)) -}}
{{- else -}}
	{{-  $rpcReqType  = (gofulltype  $rpcReqType  $.FileDescriptor) -}}
{{- end -}}

{{- if (eq $rspTypePkg $goPkgName) -}}
	{{-  $rpcRspType  = (printf "fb.%s" (splitList "."  $rpcRspType |last|export|camelcase)) -}}
{{- else -}}
	{{-  $rpcRspType  = (gofulltype  $rpcRspType  $.FileDescriptor) -}}
{{- end -}}

{{- /* $reqType and $rspType are the underscored versions of $rpcReqType and $rpcRspType, used for naming xxFromBuilder functions. */ -}}
{{- $reqType := (printf "%s_%s_%s" $svrNameCamelCase $rpcName ($rpcReqType | gopkg) ) -}}
{{- $rspType := (printf "%s_%s_%s" $svrNameCamelCase $rpcName ($rpcRspType | gopkg) ) -}}
// {{$reqType}}FromBuilder retrieves the corresponding structure of the Request from the *flatbuffers.Builder type.
func {{$reqType}}FromBuilder(b *flatbuffers.Builder) *{{$rpcReqType}} {
	// By calling b.FinishedBytes, you can obtain the byte stream corresponding to Marshal.
	reqbytes := b.FinishedBytes()
	req := &{{$rpcReqType}}{}
	// Calling Init allows you to construct a Request from a byte stream, which is equivalent to Unmarshal.
	req.Init(reqbytes, flatbuffers.GetUOffsetT(reqbytes))
	return req
}

// {{$rspType}}FromBuilder retrieves the corresponding structure of the Reply from the *flatbuffers.Builder type.
func {{$rspType}}FromBuilder(b *flatbuffers.Builder) *{{$rpcRspType}} {
	// By calling b.FinishedBytes, you can obtain the byte stream corresponding to Marshal.
	rspbytes := b.FinishedBytes()
	rsp := &{{$rpcRspType}}{}
	// Calling Init allows you to construct a Reply from a byte stream, which is equivalent to Unmarshal.
	rsp.Init(rspbytes, flatbuffers.GetUOffsetT(rspbytes))
	return rsp
}

{{- if and .ClientStreaming .ServerStreaming }}
func Test_{{$svrNameCamelCase}}_{{$rpcName}}(t *testing.T) {
	var {{$svrName}}Service = &{{$svrName}}Impl{}
	// Start writing mock logic.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	{{$svrNameCamelCase|untitle}}ClientProxy := fb.NewMock{{$svrNameCamelCase}}ClientProxy(ctrl)
	{{$rpcName|untitle}}Client := fb.NewMock{{$svrNameCamelCase}}_{{$rpcName}}Client(ctrl)
	{{$rpcName|untitle}}Server := fb.NewMock{{$svrNameCamelCase}}_{{$rpcName}}Server(ctrl)
	inorderClient := make([]*gomock.Call, 0)
	inorderServer := make([]*gomock.Call, 0)
	// Expected behavior.
	m := {{$svrNameCamelCase|untitle}}ClientProxy.EXPECT().{{$rpcName}}(gomock.Any(), gomock.Any()).AnyTimes()
	m.DoAndReturn(func(ctx context.Context, opts ...interface{}) (interface{}, error) {
		x := {{$rpcName|untitle}}Client.EXPECT().Send(gomock.Any()).AnyTimes()
		x.DoAndReturn(func(req interface{}) error {
			b, ok := req.(*flatbuffers.Builder)
			if !ok {
				panic("invalid request")
			}
			s := {{$rpcName|untitle}}Server.EXPECT().Recv().Return({{$reqType}}FromBuilder(b), nil)
			inorderServer = append(inorderServer, s)
			return nil
		})
		k := {{$rpcName|untitle}}Client.EXPECT().CloseSend()
		k.DoAndReturn(func() error {
			{{$rpcName|untitle}}Server.EXPECT().Recv().Return(nil, io.EOF)
			err := {{$svrName}}Service.{{$rpcName}}({{$rpcName|untitle}}Server)
			if err != nil {
				return err
			}
			{{$rpcName|untitle}}Client.EXPECT().Recv().Return(nil, io.EOF)
			return nil
		})
		return {{$rpcName|untitle}}Client, nil
	})
	gomock.InOrder(inorderServer...)
	s := {{$rpcName|untitle}}Server.EXPECT().Send(gomock.Any()).AnyTimes()
	s.DoAndReturn(func(rsp interface{}) error {
		b, ok := rsp.(*flatbuffers.Builder)
		if !ok {
			panic("invalid response")
		}
		c := {{$rpcName|untitle}}Client.EXPECT().Recv().Return({{$rspType}}FromBuilder(b), nil)
		inorderClient = append(inorderClient, c)
		return nil
	})
	gomock.InOrder(inorderClient...)
	// Start writing unit test logic (for reference only, please modify as needed).
	stream, err := {{$svrNameCamelCase|untitle}}ClientProxy.{{$rpcName}}(trpc.BackgroundContext())
	require.Nil(t, err)
	require.NotNil(t, stream)
	for i := 0; i < 5; i++ {
		b := flatbuffers.NewBuilder(0)
		// Example of Adding a Field.
		// Replace the "String" in CreateString with the field type you want to work with.
		// Replace "Message" in AddMessage with the field name you want to work with.
		// idx := b.CreateString(fmt.Sprintf("D %v", i))
		{{$rpcReqType}}Start(b)
		// {{$rpcReqType}}AddMessage(b, idx)
		b.Finish({{$rpcReqType}}End(b))
		// Output each input parameter (check t.Logf output, run `go test -v`).
		// The Message field can be modified as needed.
		// t.Logf("{{$svrNameCamelCase}}_{{$rpcName}} req: %q", {{$reqType}}FromBuilder(b).Message())
		t.Logf("{{$svrNameCamelCase}}_{{$rpcName}} req: %v", {{$reqType}}FromBuilder(b))
		err := stream.Send(b)
		require.Nil(t, err)
	}
	err = stream.CloseSend()
	require.Nil(t, err)
	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		// Output each return value (check t.Logf output, run `go test -v`).
		// The Message field can be modified as needed.
		// t.Logf("{{$svrNameCamelCase}}_{{$rpcName}} rsp: %q, err: %v", rsp.Message(), err)
		t.Logf("{{$svrNameCamelCase}}_{{$rpcName}} rsp: %v, err: %v", rsp, err)
		require.Nil(t, err)
	}
}
{{ else }}
{{ if .ClientStreaming }}
func Test_{{$svrNameCamelCase}}_{{$rpcName}}(t *testing.T) {
	var {{$svrName}}Service = &{{$svrName}}Impl{}
	// Start writing mock logic.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	{{$svrNameCamelCase|untitle}}ClientProxy := fb.NewMock{{$svrNameCamelCase}}ClientProxy(ctrl)
	{{$rpcName|untitle}}Client := fb.NewMock{{$svrNameCamelCase}}_{{$rpcName}}Client(ctrl)
	{{$rpcName|untitle}}Server := fb.NewMock{{$svrNameCamelCase}}_{{$rpcName}}Server(ctrl)
	inorderServer := make([]*gomock.Call, 0)
	// Expected behavior.
	m := {{$svrNameCamelCase|untitle}}ClientProxy.EXPECT().{{$rpcName}}(gomock.Any(), gomock.Any()).AnyTimes()
	m.DoAndReturn(func(ctx context.Context, opts ...interface{}) (interface{}, error) {
		x := {{$rpcName|untitle}}Client.EXPECT().Send(gomock.Any()).AnyTimes()
		x.DoAndReturn(func(req interface{}) error {
			b, ok := req.(*flatbuffers.Builder)
			if !ok {
				panic("invalid request")
			}
			s := {{$rpcName|untitle}}Server.EXPECT().Recv().Return({{$reqType}}FromBuilder(b), nil)
			inorderServer = append(inorderServer, s)
			return nil
		})
		k := {{$rpcName|untitle}}Client.EXPECT().CloseAndRecv()
		k.DoAndReturn(func() (interface{}, error) {
			rsp := &{{$rpcRspType}}{}
			{{$rpcName|untitle}}Server.EXPECT().Recv().Return(nil, io.EOF)
			s := {{$rpcName|untitle}}Server.EXPECT().SendAndClose(gomock.Any())
			s.DoAndReturn(func(f interface{}) error {
				b, ok := f.(*flatbuffers.Builder)
				if !ok {
					panic("invalid response")
				}
				rsp = {{$rspType}}FromBuilder(b)
				return nil
			})
			err := {{$svrName}}Service.{{$rpcName}}({{$rpcName|untitle}}Server)
			if err != nil {
				return nil, err
			}
			return rsp, nil
		})
		return {{$rpcName|untitle}}Client, nil
	})
	gomock.InOrder(inorderServer...)
	// Start writing unit test logic (for reference only, please modify as needed).
	stream, err := {{$svrNameCamelCase|untitle}}ClientProxy.{{$rpcName}}(trpc.BackgroundContext())
	require.Nil(t, err)
	require.NotNil(t, stream)
	for i := 0; i < 5; i++ {
		b := flatbuffers.NewBuilder(0)
		// Example of Adding a Field.
		// Replace the "String" in CreateString with the field type you want to work with.
		// Replace "Message" in AddMessage with the field name you want to work with.
		// idx := b.CreateString(fmt.Sprintf("B %v", i))
		{{$rpcReqType}}Start(b)
		// {{$rpcReqType}}AddMessage(b, idx)
		b.Finish({{$rpcReqType}}End(b))
		// Output each input parameter (check t.Logf output, run `go test -v`).
		// The Message field can be modified as needed.
		// t.Logf("{{$svrNameCamelCase}}_{{$rpcName}} req: %q", {{$reqType}}FromBuilder(b).Message())
		t.Logf("{{$svrNameCamelCase}}_{{$rpcName}} req: %q", {{$reqType}}FromBuilder(b))
		err := stream.Send(b)
		require.Nil(t, err)
	}
	rsp, err := stream.CloseAndRecv()
	// The Message field can be modified as needed.
	// t.Logf("{{$svrNameCamelCase}}_{{$rpcName}} rsp: %q, err: %v", rsp.Message(), err)
	t.Logf("{{$svrNameCamelCase}}_{{$rpcName}} rsp: %v, err: %v", rsp, err)
	require.Nil(t, err)
}
{{ else }}
{{ if .ServerStreaming }}
func Test_{{$svrNameCamelCase}}_{{$rpcName}}(t *testing.T) {
	var {{$svrName}}Service = &{{$svrName}}Impl{}
	// Start writing mock logic.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	{{$svrNameCamelCase|untitle}}ClientProxy := fb.NewMock{{$svrNameCamelCase}}ClientProxy(ctrl)
	{{$rpcName|untitle}}Client := fb.NewMock{{$svrNameCamelCase}}_{{$rpcName}}Client(ctrl)
	{{$rpcName|untitle}}Server := fb.NewMock{{$svrNameCamelCase}}_{{$rpcName}}Server(ctrl)
	inorderClient := make([]*gomock.Call, 0)
	// Expected behavior.
	m := {{$svrNameCamelCase|untitle}}ClientProxy.EXPECT().{{$rpcName}}(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	m.DoAndReturn(func(ctx context.Context, req interface{}, opts ...interface{}) (interface{}, error) {
		b, ok := req.(*flatbuffers.Builder)
		if !ok {
			panic("invalid request")
		}
		s := {{$rpcName|untitle}}Server.EXPECT().Send(gomock.Any()).AnyTimes()
		s.DoAndReturn(func(rsp interface{}) error {
			b, ok := rsp.(*flatbuffers.Builder)
			if !ok {
				panic("invalid response")
			}
			c := {{$rpcName|untitle}}Client.EXPECT().Recv().Return({{$rspType}}FromBuilder(b), nil)
			inorderClient = append(inorderClient, c)
			return nil
		})
		err := {{$svrName}}Service.{{$rpcName}}({{$reqType}}FromBuilder(b), {{$rpcName|untitle}}Server)
		if err != nil {
			return nil, err
		}
		{{$rpcName|untitle}}Client.EXPECT().Recv().Return(nil, io.EOF)
		return {{$rpcName|untitle}}Client, nil
	})
	gomock.InOrder(inorderClient...)
	// Start writing unit test logic (for reference only, please modify as needed).
	b := flatbuffers.NewBuilder(0)
	// Example of Adding a Field.
	// Replace the "String" in CreateString with the field type you want to work with.
	// Replace "Message" in AddMessage with the field name you want to work with.
	// i := b.CreateString("C")
	{{$rpcReqType}}Start(b)
	// {{$rpcReqType}}AddMessage(b, i)
	b.Finish({{$rpcReqType}}End(b))
	// The Message field can be modified as needed.
	// t.Logf("{{$svrNameCamelCase}}_{{$rpcName}} req: %q", {{$reqType}}FromBuilder(b).Message())
	t.Logf("{{$svrNameCamelCase}}_{{$rpcName}} req: %v", {{$reqType}}FromBuilder(b))
	stream, err := {{$svrNameCamelCase|untitle}}ClientProxy.{{$rpcName}}(trpc.BackgroundContext(), b)
	require.Nil(t, err)
	require.NotNil(t, stream)
	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		// The Message field can be modified as needed.
		// t.Logf("{{$svrNameCamelCase}}_{{$rpcName}} rsp: %q, err: %v", rsp.Message(), err)
		t.Logf("{{$svrNameCamelCase}}_{{$rpcName}} rsp: %v, err: %v", rsp, err)
		require.Nil(t, err)
	}
}
{{ else }}
// As an example, create{{$reqType}} creates a string in flatbuffers 
// and uses that string as the Message field in the Request.
// During development, replace the Message field with the desired field 
// to be constructed, and replace String with the corresponding field type.
func create{{$reqType}}(s string) *{{$rpcReqType}} {
	b := flatbuffers.NewBuilder(0)
	// Example of Adding a Field.
	// Replace the "String" in CreateString with the field type you want to work with.
	// Replace "Message" in AddMessage with the field name you want to work with.
	// i := b.CreateString(s)
	{{$rpcReqType}}Start(b)
	// {{$rpcReqType}}AddMessage(b, i)
	b.Finish({{$rpcReqType}}End(b))
	return {{$reqType}}FromBuilder(b)
}
func create{{$rspType}}(s string) *{{$rpcRspType}} {
	b := flatbuffers.NewBuilder(0)
	// Example of Adding a Field.
	// Replace the "String" in CreateString with the field type you want to work with.
	// Replace "Message" in AddMessage with the field name you want to work with.
	// i := b.CreateString(s)
	{{$rpcRspType}}Start(b)
	// {{$rpcRspType}}AddMessage(b, i)
	b.Finish({{$rpcRspType}}End(b))
	return {{$rspType}}FromBuilder(b)
}
func Test_{{$svrNameCamelCase|untitle}}Impl_{{$rpcName}}(t *testing.T) {
    type args struct {
        ctx context.Context
        req *{{$rpcReqType}}
        rsp *{{$rpcRspType}}
    }
    tests := []struct {
        name    string
        args    args
        wantErr bool
    }{
        // TODO: Add test cases.
		// Here's an example that you can modify according to your needs.
		{
			name: "basic test",
			args: args{
				ctx: trpc.BackgroundContext(),
				req: create{{$reqType}}("A"),
				rsp: create{{$rspType}}("welcome A"),
			},
			wantErr: false, 
		},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            {{- $svrName := $service.Name | camelcase | untitle}}
            s := &{{$svrName}}Impl{}
            b := &flatbuffers.Builder{}
            var err error
            if b, err = s.{{$rpcName}}(tt.args.ctx, tt.args.req); (err != nil) != tt.wantErr {
            	t.Errorf("{{$svrNameCamelCase|untitle}}Impl.{{$rpcName}}() error = %v, wantErr %v", err, tt.wantErr)
            }
			rsp := {{$rspType}}FromBuilder(b)
            if !reflect.DeepEqual(rsp, tt.args.rsp) {
				// The Message field can be modified as needed.
           		// t.Errorf("{{$svrNameCamelCase|untitle}}Impl.{{$rpcName}}() rsp got = %q, want %q", rsp.Message(), tt.args.rsp.Message())
           		t.Errorf("{{$svrNameCamelCase|untitle}}Impl.{{$rpcName}}() rsp got = %v, want %v", rsp, tt.args.rsp)
            }
        })
    }
}
{{ end }}
{{ end }}
{{- end }}
{{end}}
