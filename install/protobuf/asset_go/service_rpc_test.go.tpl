{{ $domainName := .Domain }}
{{ $groupName := .GroupName }}
{{ $versionSuffix := .VersionSuffix }}
{{- $serviceSuffix := "" -}}
{{- if not .NoServiceSuffix }}
    {{- $serviceSuffix = "Service" }}
{{- end }}

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

	trpc "{{ $domainName }}/{{ $groupName }}/trpc-go{{ $versionSuffix }}"
	_ "{{ $domainName }}/{{ $groupName }}/trpc-go{{ $versionSuffix }}/http"
    "go.uber.org/mock/gomock"
	"github.com/stretchr/testify/assert"
	pb "{{ trimright ";" $goPkgName }}"
    {{ range .ImportsX }}
    	{{.Name}} "{{.Path}}"
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

{{ $p1 := trimright ";" $goPkgName }}
{{ $p2 := splitList ";" $goPkgName | last | gopkg_simple }}

//go:generate go mod tidy
//go:generate mockgen -destination=stub/{{$p1}}/{{$fname}}_mock.go -package={{$p2}} -self_package={{$p1}} --source=stub/{{$p1}}/{{$fname}}.trpc.go
{{range $index, $method := (index .Services .ServiceIndex).RPC}}
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

{{- if and .ClientStreaming .ServerStreaming }}
func Test_{{$svrNameCamelCase}}_{{$rpcName}}(t *testing.T) {
	var {{$svrName}}Service = &{{$svrName}}Impl{}

	// Start writing mock logic.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	{{$svrNameCamelCase|untitle}}ClientProxy := pb.NewMock{{$svrNameCamelCase}}ClientProxy(ctrl)
	{{$rpcName|untitle}}Client := pb.NewMock{{$svrNameCamelCase}}_{{$rpcName}}Client(ctrl)
	{{$rpcName|untitle}}Server := pb.NewMock{{$svrNameCamelCase}}_{{$rpcName}}Server(ctrl)
	inorderClient := make([]*gomock.Call, 0)
	inorderServer := make([]*gomock.Call, 0)

	// Expected behavior.
	m := {{$svrNameCamelCase|untitle}}ClientProxy.EXPECT().{{$rpcName}}(gomock.Any(), gomock.Any()).AnyTimes()

	m.DoAndReturn(func(ctx context.Context, opts ...interface{}) (interface{}, error) {

		x := {{$rpcName|untitle}}Client.EXPECT().Send(gomock.Any()).AnyTimes()

		x.DoAndReturn(func(req interface{}) error {

			r, ok := req.(*{{$rpcReqType}})
			if !ok {
				panic("invalid request")
			}

			s := {{$rpcName|untitle}}Server.EXPECT().Recv().Return(r, nil)
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

		r, ok := rsp.(*{{$rpcRspType}})
		if !ok {
			panic("invalid response")
		}

		c := {{$rpcName|untitle}}Client.EXPECT().Recv().Return(r, nil)
		inorderClient = append(inorderClient, c)
		return nil
	})

	gomock.InOrder(inorderClient...)

	// Start writing unit test logic (for reference only, please modify as needed).
	stream, err := {{$svrNameCamelCase|untitle}}ClientProxy.{{$rpcName}}(trpc.BackgroundContext())
	require.Nil(t, err)
	require.NotNil(t, stream)

	for i := 0; i < 5; i++ {

		req := &{{$rpcReqType}}{}

		// Output each input parameter (check t.Logf output, run `go test -v`).
		t.Logf("{{$svrNameCamelCase}}_{{$rpcName}} req: %v", req)

		err := stream.Send(req)
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

	{{$svrNameCamelCase|untitle}}ClientProxy := pb.NewMock{{$svrNameCamelCase}}ClientProxy(ctrl)
	{{$rpcName|untitle}}Client := pb.NewMock{{$svrNameCamelCase}}_{{$rpcName}}Client(ctrl)
	{{$rpcName|untitle}}Server := pb.NewMock{{$svrNameCamelCase}}_{{$rpcName}}Server(ctrl)
	inorderServer := make([]*gomock.Call, 0)

	// Expected behavior.
	m := {{$svrNameCamelCase|untitle}}ClientProxy.EXPECT().{{$rpcName}}(gomock.Any(), gomock.Any()).AnyTimes()

	m.DoAndReturn(func(ctx context.Context, opts ...interface{}) (interface{}, error) {

		x := {{$rpcName|untitle}}Client.EXPECT().Send(gomock.Any()).AnyTimes()

		x.DoAndReturn(func(req interface{}) error {

			r, ok := req.(*{{$rpcReqType}})
			if !ok {
				panic("invalid request")
			}

			s := {{$rpcName|untitle}}Server.EXPECT().Recv().Return(r, nil)
			inorderServer = append(inorderServer, s)
			return nil
		})

		k := {{$rpcName|untitle}}Client.EXPECT().CloseAndRecv()

		k.DoAndReturn(func() (interface{}, error) {

			rsp := &{{$rpcRspType}}{}

			{{$rpcName|untitle}}Server.EXPECT().Recv().Return(nil, io.EOF)

			s := {{$rpcName|untitle}}Server.EXPECT().SendAndClose(gomock.Any())

			s.DoAndReturn(func(f interface{}) error {

				r, ok := f.(*{{$rpcRspType}})
				if !ok {
					panic("invalid response")
				}

				rsp = r
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

		req := &{{$rpcReqType}}{}

		// Output each input parameter (check t.Logf output, run `go test -v`).
		t.Logf("{{$svrNameCamelCase}}_{{$rpcName}} req: %v", req)

		err := stream.Send(req)
		require.Nil(t, err)
	}

	rsp, err := stream.CloseAndRecv()

	// Output the return value (check t.Logf output, run `go test -v`).
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

	{{$svrNameCamelCase|untitle}}ClientProxy := pb.NewMock{{$svrNameCamelCase}}ClientProxy(ctrl)
	{{$rpcName|untitle}}Client := pb.NewMock{{$svrNameCamelCase}}_{{$rpcName}}Client(ctrl)
	{{$rpcName|untitle}}Server := pb.NewMock{{$svrNameCamelCase}}_{{$rpcName}}Server(ctrl)
	inorderClient := make([]*gomock.Call, 0)

	// Expected behavior.
	m := {{$svrNameCamelCase|untitle}}ClientProxy.EXPECT().{{$rpcName}}(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	m.DoAndReturn(func(ctx context.Context, req interface{}, opts ...interface{}) (interface{}, error) {

		r, ok := req.(*{{$rpcReqType}})
		if !ok {
			panic("invalid request")
		}

		s := {{$rpcName|untitle}}Server.EXPECT().Send(gomock.Any()).AnyTimes()

		s.DoAndReturn(func(rsp interface{}) error {

			r, ok := rsp.(*{{$rpcRspType}})
			if !ok {
				panic("invalid response")
			}

			c := {{$rpcName|untitle}}Client.EXPECT().Recv().Return(r, nil)
			inorderClient = append(inorderClient, c)
			return nil
		})

		err := {{$svrName}}Service.{{$rpcName}}(r, {{$rpcName|untitle}}Server)
		if err != nil {
			return nil, err
		}

		{{$rpcName|untitle}}Client.EXPECT().Recv().Return(nil, io.EOF)

		return {{$rpcName|untitle}}Client, nil
	})

	gomock.InOrder(inorderClient...)

	// Start writing unit test logic (for reference only, please modify as needed).
	req := &{{$rpcReqType}}{}

	// Output the input parameters (check t.Logf output, run `go test -v`).
	t.Logf("{{$svrNameCamelCase}}_{{$rpcName}} req: %v", req)

	stream, err := {{$svrNameCamelCase|untitle}}ClientProxy.{{$rpcName}}(trpc.BackgroundContext(), req)
	require.Nil(t, err)
	require.NotNil(t, stream)

	for {
		rsp, err := stream.Recv()
		
		if err == io.EOF {
			break
		}

		// Output each return value (check t.Logf output, run `go test -v`).
		t.Logf("{{$svrNameCamelCase}}_{{$rpcName}} rsp: %v, err: %v", rsp, err)

		require.Nil(t, err)
	}
}
{{ else }}
func Test_{{$svrNameCamelCase|untitle}}Impl_{{$rpcName}}(t *testing.T) {
    {{- $svrName := $service.Name | camelcase | untitle}}
    // Start writing mock logic.
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    {{$svrName}}Service := pb.NewMock{{$svrNameCamelCase}}{{$serviceSuffix}}(ctrl)
    var inorderClient []*gomock.Call
    // Expected behavior.
    m := {{$svrName}}Service.EXPECT().{{$rpcName}}(gomock.Any(), gomock.Any()).AnyTimes()
    m.DoAndReturn(func(ctx context.Context, req *{{$rpcReqType}}) (*{{$rpcRspType}}, error) {
        s := &{{$svrName}}Impl{}
        return s.{{$rpcName}}(ctx, req)
    })
    gomock.InOrder(inorderClient...)

    // Start writing unit test logic.
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
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var rsp *{{$rpcRspType}}
            var err error
            if rsp, err = {{$svrName}}Service.{{$rpcName}}(tt.args.ctx, tt.args.req); (err != nil) != tt.wantErr {
                t.Errorf("{{$svrNameCamelCase|untitle}}Impl.{{$rpcName}}() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !reflect.DeepEqual(rsp, tt.args.rsp) {
           		t.Errorf("{{$svrNameCamelCase|untitle}}Impl.{{$rpcName}}() rsp got = %v, want %v", rsp, tt.args.rsp)
            }
        })
    }
}
{{ end }}
{{ end }}
{{- end }}
{{end}}
