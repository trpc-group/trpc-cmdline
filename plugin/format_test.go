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
	"errors"
	"os"
	"testing"

	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/util/style"

	"github.com/agiledragon/gomonkey"
	"github.com/stretchr/testify/require"
)

func TestPlugin_Format(t *testing.T) {
	u := &Formatter{}
	opt := params.Option{
		Language: "go",
	}
	require.Equal(t, "gofmt", u.Name())

	t.Run("go", func(t *testing.T) {
		p := gomonkey.ApplyFunc(style.GoFmtDir, func(string2 string) error {
			return nil
		})
		p.ApplyFunc(os.Getwd, func() (string, error) {
			return os.TempDir(), nil
		})
		defer p.Reset()
		require.True(t, u.Check(nil, &opt))
		require.Nil(t, u.Run(nil, &opt))
	})

	t.Run("Getwd error", func(t *testing.T) {
		p := gomonkey.ApplyFunc(os.Getwd, func() (string, error) {
			return "", errors.New("getwd error")
		})
		defer p.Reset()
		require.True(t, u.Check(nil, &opt))
		require.NotNil(t, u.Run(nil, &opt))
	})
}
