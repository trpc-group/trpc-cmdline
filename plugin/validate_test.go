package plugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/parser"
	"trpc.group/trpc-go/trpc-cmdline/util/paths"
	"trpc.group/trpc-go/trpc-cmdline/util/pb"
)

func TestPlugin_Validate(t *testing.T) {
	u := &Validate{}
	require.Equal(t, "validate", u.Name())

	t.Run("!go", func(t *testing.T) {
		require.False(t, u.Check(nil, &params.Option{}))
	})

	// rpconly test
	t.Run("validate && rpconly", func(t *testing.T) {
		// setup, refer to pb related to validate under testcase directory.
		pbf, fd, err := parseValidateSampleProtofile()
		require.Nil(t, err)
		output := t.TempDir()

		wd, _ := os.Getwd()
		defer os.Chdir(wd)

		// setup, generate a pb.go file for the above pb files.
		opt := &params.Option{
			Protofile: "helloworld.proto",
			Protodirs: []string{filepath.Clean(filepath.Join(wd, "../install")), filepath.Dir(pbf)},
			Language:  "go",
			RPCOnly:   true,
			OutputDir: output,
		}
		require.True(t, u.Check(fd, opt))

		dir, err := paths.Locate(pb.ProtoTRPC)
		require.Nil(t, err)
		opt.Protodirs = append(append(opt.Protodirs, dir),
			paths.ExpandSearch(dir)...)

		os.Chdir(filepath.Dir(pbf))
		require.Nil(t, u.Run(fd, opt))
	})

	// !rpconly test
	t.Run("validate && !rpconly", func(t *testing.T) {
		pbf, fd, err := parseValidateSampleProtofile()
		if err != nil {
			panic(err)
		}
		output := t.TempDir()

		wd, _ := os.Getwd()
		defer os.Chdir(wd)

		opt := &params.Option{
			Protofile: "helloworld.proto",
			Protodirs: []string{filepath.Clean(filepath.Join(wd, "../install")), filepath.Dir(pbf)},
			Language:  "go",
			RPCOnly:   false,
			OutputDir: output,
		}

		p := gomonkey.ApplyFunc(pb.Protoc, func([]string, string, string, string, ...pb.Option) error {
			return nil
		})
		defer p.Reset()

		require.True(t, u.Check(fd, opt))
		require.Nil(t, u.Run(fd, opt))
	})
}

func parseValidateSampleProtofile() (string, *descriptor.FileDescriptor, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}
	// parse protofile
	pbd := filepath.Clean(filepath.Join(wd, "../testcase/plugins/validate"))
	pbf := filepath.Join(pbd, "helloworld.proto")
	ins := filepath.Clean(filepath.Join(wd, "../install"))
	fd, err := parser.ParseProtoFile("helloworld.proto", append([]string{pbd, ins}, paths.ExpandTRPCSearch(ins)...))
	if err != nil {
		return "", fd, err
	}
	return pbf, fd, err
}
