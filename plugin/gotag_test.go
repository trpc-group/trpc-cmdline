package plugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/parser"
	"trpc.group/trpc-go/trpc-cmdline/util/paths"
	"trpc.group/trpc-go/trpc-cmdline/util/pb"
)

func TestPlugin_GoTag(t *testing.T) {
	require.Nil(t, setup())
	u := &GoTag{}
	require.Equal(t, "gotag", u.Name())

	opt := params.Option{
		Language:  "go",
		Gotag:     true,
		OutputDir: t.TempDir(),
	}
	os.MkdirAll(opt.OutputDir, os.ModePerm)
	defer os.RemoveAll(opt.OutputDir)

	// Don't run if not golang.
	t.Run("lang !go", func(t *testing.T) {
		opt := opt
		opt.Language = "!go"
		require.False(t, u.Check(nil, &opt))
	})

	// Golang, but no gotag.
	t.Run("go && !gotag", func(t *testing.T) {
		opt := opt
		opt.Gotag = false
		require.False(t, u.Check(nil, &opt))
	})

	// Run normally.
	t.Run("go && gotag", func(t *testing.T) {
		require.True(t, u.Check(nil, &opt))

		// setup
		opt := opt
		pbf, fd, err := parseGoTagSampleProtofile()
		if err != nil {
			panic(err)
		}

		// No corresponding pb.go file is generated for the pb file.
		t.Run("lstat error", func(t *testing.T) {
			require.NotNil(t, u.Run(fd, &opt))
		})

		// Parse tags area.
		t.Run("tags area", func(t *testing.T) {
			opt := opt
			opt.RPCOnly = true
			pbd := filepath.Dir(pbf)

			wd, _ := os.Getwd()
			defer os.Chdir(wd)

			os.Chdir(pbd)

			root := filepath.Clean(filepath.Join(wd, ".."))
			protodirs := append([]string{
				root,
				filepath.Join(root, "install"),
				filepath.Join(root, "install/submodules"),
				filepath.Join(root, "install/submodules/trpc"),
				filepath.Join(root, "install/protos"),
				filepath.Join(root, "testcase/plugins/gotag"),
			}, paths.ExpandTRPCSearch(filepath.Join(root, "install"))...)
			err = pb.Protoc(protodirs, pbf, "go", opt.OutputDir)
			if err != nil {
				panic(err)
			}

			require.Nil(t, u.Run(fd, &opt))
		})
	})
}

func parseGoTagSampleProtofile() (string, *descriptor.FileDescriptor, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}
	// parse protofile
	pbd := filepath.Clean(filepath.Join(wd, "../testcase/plugins/gotag"))
	pbf := filepath.Join(pbd, "gotag.proto")

	dir1, err := paths.Locate(pb.ProtoTRPC)
	if err != nil {
		panic(err)
	}

	fd, err := parser.ParseProtoFile(
		"gotag.proto",
		append(paths.ExpandSearch(dir1), pbd, dir1),
		parser.WithLanguage("go"),
	)
	if err != nil {
		return "", nil, err
	}
	return pbf, fd, err
}
