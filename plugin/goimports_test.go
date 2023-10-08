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
