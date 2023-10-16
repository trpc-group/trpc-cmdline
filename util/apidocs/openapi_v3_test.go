// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package apidocs

import (
	"encoding/json"
	"os"

	"testing"

	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/parser"
	"trpc.group/trpc-go/trpc-cmdline/util/paths"
)

func TestNewOpenAPIJSON(t *testing.T) {
	option := &params.Option{
		Protodirs: append([]string{
			".",
			"../../install",
			"../../install/submodules",
			"../../install/submodules/trpc-protocol",
			"../../install/protos",
		}, paths.ExpandTRPCSearch("../../install")...),
		Protofile:       "testcase/hello.proto",
		ProtofileAbs:    "testcase/hello.proto",
		KeepOrigRPCName: true,
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

	openapi, err := NewOpenAPIJSON(fd, option)
	require.NoError(t, err)

	gotByte, err := json.MarshalIndent(openapi, "", " ")
	require.NoError(t, err)

	wantByte, err := os.ReadFile("testcase/openapi.json")
	require.NoError(t, err)
	require.Equal(t, string(wantByte), string(gotByte))
}

func TestNewOpenAPIJSON_OptJSONParam(t *testing.T) {
	option := &params.Option{
		Protodirs: append([]string{
			".",
			"../../install",
			"../../install/submodules",
			"../../install/submodules/trpc-protocol",
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

	openapi, err := NewOpenAPIJSON(fd, option)
	require.NoError(t, err)

	gotByte, err := json.MarshalIndent(openapi, "", " ")
	require.NoError(t, err)

	wantByte, err := os.ReadFile("testcase/openapi_json_param.json")
	require.NoError(t, err)
	require.Equal(t, string(wantByte), string(gotByte))
}

func TestNewOpenAPIJSON_OrderByPBName(t *testing.T) {
	option := &params.Option{
		Protodirs: append([]string{
			".",
			"../../install",
			"../../install/submodules",
			"../../install/submodules/trpc-protocol",
			"../../install/protos",
		}, paths.ExpandTRPCSearch("../../install")...),
		Protofile:           "testcase/hello.proto",
		ProtofileAbs:        "testcase/hello.proto",
		OrderByPBName:       true,
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

	openapi, err := NewOpenAPIJSON(fd, option)
	require.NoError(t, err)

	gotByte, err := json.MarshalIndent(openapi, "", " ")
	require.NoError(t, err)

	wantByte, err := os.ReadFile("testcase/openapi_order_by_pbname.json")

	require.NoError(t, err)
	require.Equal(t, string(wantByte), string(gotByte))
}

func TestNewOpenAPIJSON_Unmarshal(t *testing.T) {
	option := &params.Option{
		Protodirs: append([]string{
			".",
			"../../install",
			"../../install/submodules",
			"../../install/submodules/trpc-protocol",
			"../../install/protos",
		}, paths.ExpandTRPCSearch("../../install")...),
		Protofile:           "testcase/hello.proto",
		ProtofileAbs:        "testcase/hello.proto",
		OrderByPBName:       true,
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

	gotOpenapi, err := NewOpenAPIJSON(fd, option)
	require.NoError(t, err)

	wantByte, err := os.ReadFile("testcase/openapi_order_by_pbname.json")
	var wantOpenapi = &OpenAPIJSON{}
	err = json.Unmarshal(wantByte, wantOpenapi)
	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, wantOpenapi, gotOpenapi)
}
