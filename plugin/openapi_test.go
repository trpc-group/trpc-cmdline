// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package plugin

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/parser"
	"trpc.group/trpc-go/trpc-cmdline/util/apidocs/openapi"
	"trpc.group/trpc-go/trpc-cmdline/util/fs"
)

func TestPlugin_OpenAPI(t *testing.T) {
	u := &OpenAPI{}
	require.Equal(t, "openapi", u.Name())

	opt := params.Option{
		OpenAPIOn: true,
	}

	// swaggeron=false, do not process.
	t.Run("!openapion", func(t *testing.T) {
		opt := opt
		opt.OpenAPIOn = false
		require.False(t, u.Check(nil, &opt))
	})

	// openapi=true, process normally.
	t.Run("openapion", func(t *testing.T) {
		require.True(t, u.Check(nil, &opt))

		pbf, fd, err := parseSampleProtofile()
		if err != nil {
			panic(err)
		}
		wd, _ := os.Getwd()
		defer os.Chdir(wd)
		os.Chdir(filepath.Dir(pbf))

		opt := opt
		opt.Protodirs = []string{filepath.Dir(pbf)}
		opt.Protofile = pbf
		opt.OpenAPIOut = filepath.Join(os.TempDir(), "xxxxxx.json")
		defer os.RemoveAll(opt.OpenAPIOut)

		require.Nil(t, u.Run(fd, &opt))
	})

	t.Run("openapion fail", func(t *testing.T) {
		require.True(t, u.Check(nil, &opt))

		p := gomonkey.ApplyFunc(openapi.GenOpenAPI, func(*descriptor.FileDescriptor, *params.Option) error {
			return errors.New("gen error")
		})
		defer p.Reset()
		opt := opt
		require.NotNil(t, u.Run(nil, &opt))
	})
}

func parseSampleProtofile() (string, *descriptor.FileDescriptor, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", nil, fmt.Errorf("getwd error: %v", err)
	}
	// parse protofile
	pbd := filepath.Clean(filepath.Join(wd, "../testcase/plugins/format"))
	pbf := filepath.Join(pbd, "helloworld.proto")
	outputdir := filepath.Join(os.TempDir(), "trpc")
	if err := os.MkdirAll(outputdir, os.ModePerm); err != nil {
		return "", nil, fmt.Errorf("mkdirall error: %v", err)
	}

	fd, err := parser.ParseProtoFile(
		"helloworld.proto",
		[]string{pbd},
	)
	if err != nil {
		return "", fd, fmt.Errorf("parse proto error: %v", err)
	}

	dir := filepath.Join(filepath.Dir(wd), "testcase/plugins/format")
	err = fs.Copy(filepath.Join(dir, "helloworld.pb.go.txt"), filepath.Join(outputdir, "helloworld.pb.go"))
	if err != nil {
		return "", nil, err
	}
	err = fs.Copy(filepath.Join(dir, "helloworld.trpc.go.txt"), filepath.Join(outputdir, "helloworld.trpc.go"))
	if err != nil {
		return "", nil, err
	}

	return pbf, fd, err
}
