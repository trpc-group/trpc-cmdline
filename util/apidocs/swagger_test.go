// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package apidocs

import (
	"encoding/json"
	"fmt"
	"os"

	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/jhump/protoreflect/desc"
	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/parser"
	"trpc.group/trpc-go/trpc-cmdline/util/paths"
)

func TestNewSwagger(t *testing.T) {
	type args struct {
		fd     *descriptor.FileDescriptor
		option *params.Option
	}
	tests := []struct {
		name              string
		args              args
		want              *SwaggerJSON
		wantErr           bool
		genSwaggerInfoErr error
	}{
		{
			name: "case1-genSwaggerInfo_error",
			args: args{
				fd: &descriptor.FileDescriptor{
					FD: &descriptor.ProtoFileDescriptor{
						FD: &desc.FileDescriptor{},
					},
				},
				option: &params.Option{},
			},
			want:              nil,
			wantErr:           true,
			genSwaggerInfoErr: fmt.Errorf("error"),
		},
		{
			name: "case1-fd_nil",
			args: args{
				fd:     &descriptor.FileDescriptor{},
				option: &params.Option{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "case2-success",
			args: args{
				fd: &descriptor.FileDescriptor{
					FD: &descriptor.ProtoFileDescriptor{
						FD: &desc.FileDescriptor{},
					},
				},
				option: &params.Option{},
			},
			want: &SwaggerJSON{
				Swagger:  "2.0",
				Info:     InfoStruct{},
				Consumes: []string{"application/json"},
				Produces: []string{"application/json"},
				Paths: Paths{
					Elements: make(map[string]Methods),
					Rank:     make(map[string]int),
				},
				Definitions: make(map[string]ModelStruct),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := gomonkey.ApplyFunc(
				NewInfo,
				func(fd *descriptor.FileDescriptor) (InfoStruct, error) {
					return InfoStruct{}, tt.genSwaggerInfoErr
				},
			)

			p.ApplyFunc(
				NewDefinitions,
				func(options *params.Option, fds ...descriptor.Desc) *Definitions {
					return &Definitions{
						models: make(map[string]ModelStruct),
					}
				},
			)

			p.ApplyFunc(
				NewPaths,
				func(fd *descriptor.FileDescriptor, option *params.Option, defs *Definitions) (Paths, error) {
					return Paths{
						Elements: make(map[string]Methods),
						Rank:     make(map[string]int),
					}, nil
				},
			)

			defer p.Reset()

			got, err := NewSwagger(tt.args.fd, tt.args.option)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSwagger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSwagger() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSwagger_with_file(t *testing.T) {
	option := &params.Option{
		Protodirs: append([]string{
			".",
			"../../install",
			"../../install/protos",
		}, paths.ExpandTRPCSearch("../../install")...),
		Protofile:           "testcase/hello.proto",
		ProtofileAbs:        "testcase/hello.proto",
		SwaggerOptJSONParam: true,
		KeepOrigRPCName:     true,
	}

	fd, err := parser.ParseProtoFile(
		option.Protofile,
		option.Protodirs,
		parser.WithAliasOn(option.AliasOn),
		parser.WithLanguage(option.Language),
		parser.WithRPCOnly(option.RPCOnly),
	)
	if err != nil {
		t.Logf("cannot parse proto file, option: %+v, err: %+v", option, err)
		t.FailNow()
	}

	swagger, err := NewSwagger(fd, option)
	require.NoError(t, err)

	gotByte, err := json.MarshalIndent(swagger, "", " ")
	require.NoError(t, err)

	wantByte, err := os.ReadFile("testcase/swagger.json")
	require.NoError(t, err)

	require.Equal(t, string(wantByte), string(gotByte))
}

func TestNewSwagger_OrderByPBName_with_file(t *testing.T) {
	option := &params.Option{
		Protodirs: append([]string{
			".",
			"../../install",
			"../../install/protos",
		}, paths.ExpandTRPCSearch("../../install")...),
		Protofile:           "testcase/hello.proto",
		ProtofileAbs:        "testcase/hello.proto",
		SwaggerOptJSONParam: true,
		OrderByPBName:       true,
		KeepOrigRPCName:     true,
	}

	fd, err := parser.ParseProtoFile(
		option.Protofile,
		option.Protodirs,
		parser.WithAliasOn(option.AliasOn),
		parser.WithLanguage(option.Language),
		parser.WithRPCOnly(option.RPCOnly),
	)
	if err != nil {
		t.Logf("cannot parse proto file, option: %+v, err: %+v", option, err)
		t.FailNow()
	}

	swagger, err := NewSwagger(fd, option)
	require.NoError(t, err)

	gotByte, err := json.MarshalIndent(swagger, "", " ")
	require.NoError(t, err)

	wantByte, err := os.ReadFile("testcase/swagger_order_by_pbname.json")
	require.NoError(t, err)

	require.Equal(t, string(wantByte), string(gotByte))
}

func TestNewSwagger_Unmarshal_file(t *testing.T) {
	option := &params.Option{
		Protodirs: append([]string{
			".",
			"../../install",
			"../../install/protos",
		}, paths.ExpandTRPCSearch("../../install")...),
		Protofile:           "testcase/hello.proto",
		ProtofileAbs:        "testcase/hello.proto",
		SwaggerOptJSONParam: true,
		OrderByPBName:       true,
		KeepOrigRPCName:     true,
	}

	fd, err := parser.ParseProtoFile(
		option.Protofile,
		option.Protodirs,
		parser.WithAliasOn(option.AliasOn),
		parser.WithLanguage(option.Language),
		parser.WithRPCOnly(option.RPCOnly),
	)
	if err != nil {
		t.Logf("cannot parse proto file, option: %+v, err: %+v", option, err)
		t.FailNow()
	}

	gotSwagger, err := NewSwagger(fd, option)
	require.NoError(t, err)

	wantByte, err := os.ReadFile("testcase/swagger_order_by_pbname.json")
	require.NoError(t, err)

	var wantSwagger = &SwaggerJSON{}
	err = json.Unmarshal(wantByte, wantSwagger)
	require.NoError(t, err)

	require.Equal(t, wantSwagger, gotSwagger)
}
