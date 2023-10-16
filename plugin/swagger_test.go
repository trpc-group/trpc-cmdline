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
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/util/apidocs/swagger"
)

func TestPlugin_Swagger(t *testing.T) {
	u := &Swagger{}
	require.Equal(t, "swagger", u.Name())

	opt := params.Option{
		SwaggerOn: true,
	}

	// swaggeron=false, do not process.
	t.Run("!swagger", func(t *testing.T) {
		opt := opt
		opt.SwaggerOn = false
		require.False(t, u.Check(nil, &opt))
	})

	// swaggeron=true, process normally.
	t.Run("swagger", func(t *testing.T) {
		//os.chdir
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
		opt.SwaggerOut = filepath.Join(os.TempDir(), "xxxxxx.json")
		defer os.RemoveAll(opt.SwaggerOut)

		require.True(t, u.Check(nil, &opt))
		require.Nil(t, u.Run(fd, &opt))
	})

	t.Run("swagger fail", func(t *testing.T) {
		p := gomonkey.ApplyFunc(swagger.GenSwagger, func(*descriptor.FileDescriptor, *params.Option) error {
			return errors.New("gen error")
		})
		defer p.Reset()

		opt := opt
		require.True(t, u.Check(nil, &opt))
		require.NotNil(t, u.Run(nil, &opt))
	})
}
