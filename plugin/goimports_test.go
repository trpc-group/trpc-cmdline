// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package plugin_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/plugin"
)

func TestGoImports(t *testing.T) {
	p := &plugin.GoImports{}
	require.Equal(t, "goimports", p.Name())
	require.False(t, p.Check(&descriptor.FileDescriptor{}, &params.Option{}))
	require.True(t, p.Check(&descriptor.FileDescriptor{}, &params.Option{Language: "go"}))
	require.Nil(t, p.Run(&descriptor.FileDescriptor{}, &params.Option{}))
}
