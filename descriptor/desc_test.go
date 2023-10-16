// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package descriptor

import (
	"log"
	"testing"

	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/fbs"
)

func TestFbsDesc(t *testing.T) {
	emptySchema := &fbs.SchemaDesc{Name: "empty.fbs"}
	reqDesc := &fbs.TableDesc{Name: "HelloRequest"}
	rspDesc := &fbs.TableDesc{Name: "HelloResponse"}
	methodDesc := &fbs.MethodDesc{Name: "hello", InputTypeDesc: reqDesc, OutputTypeDesc: rspDesc}
	rpcDesc := &fbs.RPCDesc{Name: "Service1", Methods: []*fbs.MethodDesc{methodDesc}}
	schema := &fbs.SchemaDesc{
		Name:         "test1.fbs",
		Namespaces:   []string{".", "trpc.testapp.testserver"},
		Root:         "HelloRequest",
		FileExt:      "fext",
		FileIdent:    "fident",
		Attrs:        []string{"go_package=trpc.group/testapp/testserver"},
		Includes:     []string{"test2.fbs"},
		Dependencies: []*fbs.SchemaDesc{emptySchema},
		Tables:       []*fbs.TableDesc{reqDesc},
		RPCs:         []*fbs.RPCDesc{rpcDesc},
	}
	reqDesc.Schema = schema
	fbsDesc := &FbsFileDescriptor{FD: schema}
	require.Equal(t, "test1.fbs", fbsDesc.GetName())
	require.Equal(t, "test1.fbs", fbsDesc.GetFullyQualifiedName())
	require.Equal(t, "trpc.testapp.testserver", fbsDesc.GetPackage())
	require.Equal(t, NewFbsAttrs(fbsDesc.FD.Attrs), fbsDesc.GetFileOptions())
	require.Equal(t, []Desc{&FbsFileDescriptor{FD: emptySchema}}, fbsDesc.GetDependencies())
	require.Equal(t, []ServiceDesc{&FbsServiceDescriptor{SD: rpcDesc}}, fbsDesc.GetServices())
	require.Equal(t, []MessageDesc{&FbsMessageDescriptor{MD: reqDesc}}, fbsDesc.GetMessageTypes())
	require.Equal(t, "trpc.group/testapp/testserver", fbsDesc.GetFileOptions().GetGoPackage())
	fbsDesc.FD.Attrs = []string{}
	require.Equal(t, "", fbsDesc.GetFileOptions().GetGoPackage())
	fbsServiceDesc := &FbsServiceDescriptor{SD: rpcDesc}
	require.Equal(t, "Service1", fbsServiceDesc.GetName())
	require.Equal(t, []MethodDesc{&FbsMethodDescriptor{MD: methodDesc}}, fbsServiceDesc.GetMethods())
	fbsMethodDesc := &FbsMethodDescriptor{MD: methodDesc}
	require.Equal(t, "hello", fbsMethodDesc.GetName())
	require.Equal(t, &FbsMessageDescriptor{MD: reqDesc}, fbsMethodDesc.GetInputType())
	require.Equal(t, &FbsMessageDescriptor{MD: rspDesc}, fbsMethodDesc.GetOutputType())
	require.False(t, fbsMethodDesc.IsClientStreaming())
	require.False(t, fbsMethodDesc.IsServerStreaming())
	require.Equal(t, &FbsSourceInfo{}, fbsMethodDesc.GetSourceInfo())
	fbsMessageDesc := &FbsMessageDescriptor{MD: reqDesc}
	require.Equal(t, &FbsFileDescriptor{FD: schema}, fbsMessageDesc.GetFile())
	require.Equal(t, ".HelloRequest", fbsMessageDesc.GetFullyQualifiedName())
	fbsMessageDesc.MD.Namespace = "trpc.testapp.testserver"
	require.Equal(t, ".trpc.testapp.testserver.HelloRequest", fbsMessageDesc.GetFullyQualifiedName())
	fbsSourceInfo := &FbsSourceInfo{}
	require.Equal(t, "", fbsSourceInfo.GetLeadingComments())
	require.Equal(t, "", fbsSourceInfo.GetTrailingComments())
}

func TestProtoDesc(t *testing.T) {
	testpath := "./testcase/"
	p := protoparse.Parser{
		ImportPaths: []string{testpath},
	}
	fds, err := p.ParseFiles("hello.proto")
	require.Nil(t, err)
	require.NotNil(t, fds)
	fd := fds[0]
	require.NotNil(t, fd)
	protoFileDesc := &ProtoFileDescriptor{FD: fd}
	log.Print(fd)
	require.Equal(t, "hello.proto", protoFileDesc.GetName())
	require.Equal(t, "hello.proto", protoFileDesc.GetFullyQualifiedName())
	require.Equal(t, "hello", protoFileDesc.GetPackage())
	require.Equal(t, "trpc.group/examples/hello", protoFileDesc.GetFileOptions().GetGoPackage())
	require.Equal(t, 1, len(protoFileDesc.GetDependencies()))
	require.Equal(t, 1, len(protoFileDesc.GetServices()))
	require.Equal(t, 1, len(protoFileDesc.GetMessageTypes()))
	service := protoFileDesc.GetServices()[0]
	require.Equal(t, "hello_svr", service.GetName())
	require.Equal(t, 1, len(service.GetMethods()))
	method := service.GetMethods()[0]
	require.Equal(t, "Hello", method.GetName())
	require.Equal(t, "message.proto", method.GetInputType().GetFile().GetName())
	require.Equal(t, "message.proto", method.GetOutputType().GetFile().GetName())
	require.False(t, method.IsClientStreaming())
	require.False(t, method.IsServerStreaming())
	require.Equal(t, "", method.GetSourceInfo().GetLeadingComments())
	require.Equal(t, "hello.HelloReq", method.GetInputType().GetFullyQualifiedName())
}
