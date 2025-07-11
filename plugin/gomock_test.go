// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
)

func TestPlugin_Mock(t *testing.T) {
	if err := setup(); err != nil {
		panic(err)
	}

	u := &GoMock{}
	opt := params.Option{
		Language:  "go",
		OutputDir: filepath.Join(os.TempDir(), "trpc"),
	}
	os.MkdirAll(opt.OutputDir, os.ModePerm)
	defer os.RemoveAll(opt.OutputDir)

	t.Run("name", func(t *testing.T) {
		require.Equal(t, "mockgen", u.Name())
	})

	// Non golang are not processed.
	t.Run("lang !go", func(t *testing.T) {
		u := &GoMock{}
		opt := opt
		opt.Language = "!go"
		require.False(t, u.Check(nil, &opt))
	})

	// It is a Go language, but if the flag --mock=false is specified, it will not be processed.
	t.Run("lang go && !mockgen", func(t *testing.T) {
		u := &GoMock{}
		opt := opt
		require.False(t, u.Check(nil, &opt))
	})

	// If there is no service defined, it will not be processed even if it is a Go language.
	t.Run("lang go && mockgen && service empty", func(t *testing.T) {
		u := &GoMock{}
		opt := opt
		fd := &descriptor.FileDescriptor{
			Services: nil,
		}
		require.False(t, u.Check(fd, &opt))
	})

	t.Run("go && mockgen && !rpconly", func(t *testing.T) {
		u := &GoMock{}
		opt := opt
		opt.Mockgen = true
		opt.RPCOnly = false
		require.False(t, u.Check(nil, &opt))
	})

	t.Run("go && mockgen && !rpconly && mockgen not installed", func(t *testing.T) {
		p := gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
			return "", fmt.Errorf("not exist mockgen")
		})
		defer p.Reset()

		u := &GoMock{}
		opt := opt
		opt.Mockgen = true
		opt.RPCOnly = false
		require.False(t, u.Check(nil, &opt))
	})
}
