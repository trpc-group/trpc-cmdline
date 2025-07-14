// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package create

import (
	"errors"
	"fmt"

	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"unsafe"

	"github.com/agiledragon/gomonkey"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-cmdline/config"
	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/parser"
	"trpc.group/trpc-go/trpc-cmdline/plugin"
	"trpc.group/trpc-go/trpc-cmdline/tpl"
	"trpc.group/trpc-go/trpc-cmdline/util/pb"
)

func TestMain(m *testing.M) {
	if _, err := config.Init(); err != nil {
		panic(err)
	}
	if err := setup(nil); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

// prepare testcases
type testcase struct {
	name          string
	pbdir         string
	pbfile        string
	rpconly       bool
	splitByMethod bool
	alias         bool
	lang          string
	wantErr       bool
	opts          []string
}

func Test_CreateCmd(t *testing.T) {
	wd, _ := os.Getwd()
	defer os.Chdir(wd)

	pd := filepath.Dir(wd)
	pd = filepath.Dir(pd)
	testdir := filepath.Join(pd, "testcase/create")
	testcases := []testcase{
		{
			name:   "1.1-without-import",
			pbdir:  "1-without-import",
			pbfile: "helloworld.proto",
		}, {
			name:   "1.2-without-import (alias)",
			pbdir:  "1-without-import",
			pbfile: "helloworld.proto",
			alias:  true,
		}, {
			name:    "1.3-without-import (rpconly)",
			pbdir:   "1-without-import",
			pbfile:  "helloworld.proto",
			rpconly: true,
		}, {
			name:          "1.5-without-import (split by method)",
			pbdir:         "1-without-import",
			pbfile:        "helloworld.proto",
			splitByMethod: true,
		}, {
			name:   "2-multi-pb-same-package",
			pbdir:  "2-multi-pb-same-package",
			pbfile: "hello.proto",
		}, {
			name:   "3-multi-pb-diff-package",
			pbdir:  "3-multi-pb-diff-package",
			pbfile: "helloworld.proto",
		}, {
			name:   "4.1-multi-pb-same-package-diff-protodir",
			pbdir:  "4.1-multi-pb-same-package-diff-protodir",
			pbfile: "helloworld.proto",
		}, {
			name:   "4.2-multi-pb-same-package-diff-protodir",
			pbdir:  "4.2-multi-pb-same-package-diff-protodir",
			pbfile: "helloworld.proto",
		}, {
			name:   "5-multi-pb-same-pkgdirective-diff-gopkgoption",
			pbdir:  "5-multi-pb-same-pkgdirective-diff-gopkgoption",
			pbfile: "helloworld.proto",
		}, {
			name:          "5.1-multi-pb-same-pkgdirective-diff-gopkgoption（split by method）",
			pbdir:         "5-multi-pb-same-pkgdirective-diff-gopkgoption",
			pbfile:        "helloworld.proto",
			splitByMethod: true,
		}, {
			name:   "6.1-other-scene google/protobuf/any",
			pbdir:  "6-other-scene/google",
			pbfile: "google.proto",
		}, {
			name:   "6.2-other-scene hello_service",
			pbdir:  "6-other-scene/hello_service",
			pbfile: "hello.proto",
		}, {
			name:    "8-service-not-existed",
			pbdir:   "8-service-not-existed",
			pbfile:  "helloworld.proto",
			rpconly: true,
		}, {
			name:   "9.1-restful (service)",
			pbdir:  "9-restful",
			pbfile: "helloworld.proto",
		}, {
			name:    "9.2-restful (rpconly)",
			pbdir:   "9-restful",
			pbfile:  "helloworld.proto",
			rpconly: true,
		}, {
			name:   "10-validate-pgv",
			pbdir:  "10-validate-pgv",
			pbfile: "helloworld.proto",
			opts:   []string{"--validate"},
		},
	}

	tmp := filepath.Join(os.TempDir(), "create/generated")
	os.RemoveAll(tmp)
	// Reset plugin configuration.
	// First, run create_idl_non_test.go, then execute setNonProtocolTypeOption,
	// which modifies the global variables plugin.Plugins and plugin.PluginsExt.
	// Finally, when running create_test.go test, it does not reset, causing the mockgan plugin not to run.
	resetPlugin()

	// run createCmd
	for _, tt := range testcases {
		tt := tt
		t.Run("CreateCmd/"+tt.name, func(t *testing.T) {
			any := filepath.Join(testdir, tt.pbdir)
			if err := os.Chdir(any); err != nil {
				panic(err)
			}

			dirs := []string{}
			err := filepath.Walk(any, func(path string, info os.FileInfo, _ error) error {
				if info.IsDir() {
					dirs = append(dirs, path)
				}
				return nil
			})
			if err != nil {
				panic("walk testcase error")
			}

			opts := tt.opts
			for _, d := range dirs {
				opts = append(opts, "--protodir", d)
			}
			out := filepath.Join(tmp, tt.name)
			opts = append(opts, "--protofile", tt.pbfile)
			opts = append(opts, "-o", out)
			opts = append(opts, "--check-update", "true")
			opts = append(opts, "-v")

			if tt.rpconly {
				opts = append(opts, "--rpconly")
			}
			if tt.splitByMethod {
				opts = append(opts, "-s")
			}
			if tt.alias {
				opts = append(opts, "--alias")
			}
			if tt.lang != "" {
				opts = append(opts, "--lang", tt.lang)
			}
			resetFlags(createCmd)
			runCreateCmd(t, tt.name, opts, out, tt.wantErr)
		})
	}
}

var createCmd *cobra.Command

func init() {
	c := New(func() error {
		return nil
	})
	createCmd = &cobra.Command{
		Use:   "create",
		Short: "指定 pb 文件快速创建工程或 rpcstub",
		Long: `指定 pb 文件快速创建工程或 rpcstub, 

'trpc create' 有两种模式：
- 生成一个完整的服务工程
- 生成被调服务的 rpcstub, 需指定'--rpconly'选项
`,
		PreRunE:  c.PreRunE,
		RunE:     c.RunE,
		PostRunE: c.PostRunE,
	}
	var (
		cfgFile           string
		defaultConfigFile string
		verboseFlag       bool
		checkUpdateFlag   bool
	)
	createCmd.PersistentFlags().StringVar(&cfgFile, "config", defaultConfigFile, "配置文件路径 (自动计算)")
	createCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "显示详细日志信息")
	createCmd.PersistentFlags().BoolVar(&checkUpdateFlag, "check-update", false, "是否检查版本更新")
	AddCreateFlags(createCmd)
}

func resetPlugin() {
	// Public plugin chain.
	plugin.Plugins = []plugin.Plugin{
		&plugin.Swagger{},  // swagger apidoc
		&plugin.OpenAPI{},  // openapi apidoc
		&plugin.Validate{}, // protoc-gen-secv
	}
	// Language-specific plugin chain.
	plugin.PluginsExt = map[string][]plugin.Plugin{
		"go": {
			&plugin.GoImports{}, // goimports,  runs before mockgen, to eliminate `package import but unused` errors
			&plugin.Formatter{}, // gofmt
			&plugin.GoMock{},    // gomock
			&plugin.GoTag{},     // custom go tag by proto field options
		},
	}
}

func runCreateCmd(t *testing.T, name string, opts []string, out string, wantErr bool) {
	if err := createCmd.ParseFlags(opts); err != nil {
		t.Fatalf("TEST CreateCmd/%s ParseFlags error = %v", name, err)
	}
	if err := createCmd.PreRunE(createCmd, opts); (err != nil) != wantErr {
		t.Fatalf("TEST CreateCmd/%s PreRunE() error = %v, wantErr %v", name, err, wantErr)
	}
	if err := createCmd.RunE(createCmd, opts); (err != nil) != wantErr {
		t.Fatalf("TEST CreateCmd/%s RunE() error = %v, wantErr %v", name, err, wantErr)
	}

	if _, err := os.Lstat(out); err != nil {
		t.Fatalf("TEST CreateCmd/%s RunE() didn't generate output", name)
	}
	if err := createCmd.PostRunE(createCmd, opts); (err != nil) != wantErr {
		t.Fatalf("TEST CreateCmd/%s PostRunE() error = %v, wantErr %v", name, err, wantErr)
	}
	if err := os.Chdir(out); (err != nil) != wantErr {
		t.Fatalf("TEST CreateCmd/%s Chdir() error = %v, wantErr %v", name, err, wantErr)
	}
}

func resetFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Value.Type() == "stringSlice" {
			// XXX: unfortunately, flag.Value.Set() appends to original
			// slice, not resets it, so we retrieve pointer to the slice here
			// and set it to new empty slice manually
			value := reflect.ValueOf(flag.Value).Elem().FieldByName("value")
			ptr := (*[]string)(unsafe.Pointer(value.Pointer()))
			*ptr = make([]string, 0)
		}
		if flag.Value.Type() == "stringArray" {
			// XXX: unfortunately, flag.Value.Set() appends to original
			// slice, not resets it, so we retrieve pointer to the slice here
			// and set it to new empty slice manually
			value := reflect.ValueOf(flag.Value).Elem().FieldByName("value")
			ptr := (*[]string)(unsafe.Pointer(value.Pointer()))
			*ptr = make([]string, 0)
		}
		_ = flag.Value.Set(flag.DefValue)
	})
	for _, cmd := range cmd.Commands() {
		resetFlags(cmd)
	}
}

func TestCmd_Create_Exception(t *testing.T) {
	// prepare workdir
	pwd, err := os.Getwd()
	require.Nil(t, err)

	dir := filepath.Dir(filepath.Dir(pwd))
	dir = filepath.Join(dir, "testcase/create/1-without-import")
	if err := os.Chdir(dir); err != nil {
		panic(err)
	}
	defer os.Chdir(pwd)

	// case1: invalid loadCreateOption (invalid --protofile)
	out := "helloworld"
	flags := map[string]string{
		"protodir":  dir,
		"protofile": "",
		"mock":      "false",
		"output":    out,
	}
	defer os.RemoveAll(filepath.Join(dir, out))

	createCmd.ResetFlags()
	_, err = runAndWatch(createCmd, flags, nil)
	require.NotNil(t, err)

	// case2: proto parse error
	p := gomonkey.ApplyFunc(parser.ParseProtoFile, func(string, []string, ...parser.Option) (*descriptor.FileDescriptor, error) {
		return nil, errors.New("parse error")
	})

	flags = map[string]string{
		"protodir":  dir,
		"protofile": "helloworld.proto",
		"mock":      "false",
	}

	_, err = runAndWatch(createCmd, flags, nil)
	p.Reset()

	require.NotNil(t, err)
}

func TestCreate_Exception(t *testing.T) {
	t.Run("outputdir error", func(t *testing.T) {
		p := gomonkey.ApplyFunc(getOutputDir, func(option *params.Option) (string, error) {
			return "", errors.New("get outputdir error")
		})
		require.NotNil(t, (&Create{}).createByProtocolType())
		p.Reset()
	})

	t.Run("isCleanDir error", func(t *testing.T) {
		p := gomonkey.ApplyFunc(getOutputDir, func(option *params.Option) (string, error) {
			return "xxxxx", nil
		})
		p.ApplyFunc(isCleanDir, func(a string, idlType config.IDLType, c string) bool {
			return false
		})
		defer p.Reset()
		require.NotNil(t, (&Create{options: &params.Option{Language: "go"}}).createFullProject())
	})

	t.Run("generate files error", func(t *testing.T) {
		p := gomonkey.NewPatches()
		p.ApplyFunc(getOutputDir, func(option *params.Option) (string, error) {
			return "xxxxx", nil
		})
		p.ApplyFunc(isCleanDir, func(a string, b config.IDLType, c string) bool {
			return true
		})
		p.ApplyFunc(tpl.GenerateFiles, func(*descriptor.FileDescriptor, string, *params.Option) error {
			return errors.New("generate files error")
		})
		defer p.Reset()
		require.NotNil(t, (&Create{options: &params.Option{Language: "go"}}).createFullProject())
	})

	t.Run("prepare outputdir error", func(t *testing.T) {
		p := gomonkey.NewPatches()
		p.ApplyFunc(getOutputDir, func(option *params.Option) (string, error) {
			return "xxxxx", nil
		})
		p.ApplyFunc(isCleanDir, func(a string, b config.IDLType, c string) bool {
			return true
		})
		p.ApplyFunc(tpl.GenerateFiles, func(*descriptor.FileDescriptor, string, *params.Option) error {
			return nil
		})
		p.ApplyFunc(prepareOutputStub, func(string) (string, error) {
			return "", errors.New("prepare outputdir error")
		})
		defer p.Reset()
		require.NotNil(t, (&Create{options: &params.Option{Language: "go"}}).createFullProject())
	})

	t.Run("get package error", func(t *testing.T) {
		p := gomonkey.NewPatches()
		p.ApplyFunc(getOutputDir, func(option *params.Option) (string, error) {
			return "xxxxx", nil
		})
		p.ApplyFunc(isCleanDir, func(a string, b config.IDLType, c string) bool {
			return true
		})
		p.ApplyFunc(tpl.GenerateFiles, func(*descriptor.FileDescriptor, string, *params.Option) error {
			return nil
		})
		p.ApplyFunc(prepareOutputStub, func(string) (string, error) {
			return "xxxx", nil
		})
		p.ApplyFunc(parser.GetPackage, func(*descriptor.FileDescriptor, string) (string, error) {
			return "", errors.New("get package error")
		})
		defer p.Reset()
		require.NotNil(t, (&Create{options: &params.Option{Language: "go"}}).createFullProject())
	})

	t.Run("mkdir all error", func(t *testing.T) {
		p := gomonkey.NewPatches()
		p.ApplyFunc(getOutputDir, func(option *params.Option) (string, error) {
			return "xxxxx", nil
		})
		p.ApplyFunc(isCleanDir, func(a string, b config.IDLType, c string) bool {
			return true
		})
		p.ApplyFunc(tpl.GenerateFiles, func(*descriptor.FileDescriptor, string, *params.Option) error {
			return nil
		})
		p.ApplyFunc(prepareOutputStub, func(string) (string, error) {
			return "xxxx", nil
		})
		p.ApplyFunc(parser.GetPackage, func(*descriptor.FileDescriptor, string) (string, error) {
			return "xxxx", nil
		})
		p.ApplyFunc(os.MkdirAll, func(string, os.FileMode) error {
			return errors.New("mkdirall error")
		})
		defer p.Reset()
		require.NotNil(t, (&Create{options: &params.Option{Language: "go"}}).createFullProject())
	})

	t.Run("run protoc error", func(t *testing.T) {
		p := gomonkey.NewPatches()
		p.ApplyFunc(getOutputDir, func(option *params.Option) (string, error) {
			return "xxxxx", nil
		})
		p.ApplyFunc(isCleanDir, func(a string, b config.IDLType, c string) bool {
			return true
		})
		p.ApplyFunc(tpl.GenerateFiles, func(*descriptor.FileDescriptor, string, *params.Option) error {
			return nil
		})
		p.ApplyFunc(prepareOutputStub, func(string) (string, error) {
			return "xxxx", nil
		})
		p.ApplyFunc(parser.GetPackage, func(*descriptor.FileDescriptor, string) (string, error) {
			return "xxxx", nil
		})
		p.ApplyFunc(os.MkdirAll, func(string, os.FileMode) error {
			return nil
		})
		p.ApplyFunc(pb.Protoc, func(protodirs []string, protofile, lang, outputdir string, opts ...pb.Option) error {
			return errors.New("run protoc error")
		})
		defer p.Reset()
		require.NotNil(t, (&Create{
			fileDescriptor: &descriptor.FileDescriptor{},
			options:        &params.Option{Language: "go"},
		}).createFullProject())
	})
}

func Test_isCleanDir(t *testing.T) {
	t.Run("dirty directory", func(t *testing.T) {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		v := isCleanDir(wd, config.IDLTypeProtobuf, "go")
		require.False(t, v)
	})

	t.Run("clean directory", func(t *testing.T) {
		wd := os.TempDir()
		tmp := filepath.Join(wd, "trpc")
		os.RemoveAll(tmp)
		os.MkdirAll(tmp, os.ModePerm)
		defer os.RemoveAll(tmp)
		v := isCleanDir(tmp, config.IDLTypeProtobuf, "go")
		require.True(t, v)
	})
}

func Test_removeSafe(t *testing.T) {
	tmp := filepath.Join(os.TempDir(), "trpc")

	p := gomonkey.NewPatches()
	p.ApplyFunc(os.Getwd, func() (dir string, err error) {
		return tmp, nil
	})

	rm, err := removeSafe(tmp)
	require.Nil(t, err)
	require.False(t, rm)

	p.Reset()
	rm, err = removeSafe(tmp)
	require.Nil(t, err)
	require.True(t, rm)
}

func Test_prepareOutputDir(t *testing.T) {
	a := os.TempDir()
	b := "b"
	c := "hello"

	p := gomonkey.ApplyFunc(os.MkdirAll, func(string, os.FileMode) error {
		return nil
	})
	defer p.Reset()
	t.Run("prepare go outputdir", func(t *testing.T) {
		d, err := prepareOutputDir(a, b, "go", c)
		require.Nil(t, err)
		require.Equal(t, filepath.Join(a, b), d)
	})
}

func Test_prepareOutputStubDir(t *testing.T) {
	tmp := os.TempDir()
	stub := filepath.Join(tmp, "stub")

	t.Run("stub notExist, create", func(t *testing.T) {
		p := gomonkey.ApplyFunc(os.MkdirAll, func(string, os.FileMode) error {
			return nil
		})
		defer p.Reset()

		dir, err := prepareOutputStub(tmp)
		require.Nil(t, err)
		require.Equal(t, stub, dir)
	})

	t.Run("stub exist, return", func(t *testing.T) {
		os.MkdirAll(stub, os.ModePerm)
		defer os.RemoveAll(stub)

		dir, err := prepareOutputStub(tmp)
		require.Nil(t, err)
		require.Equal(t, stub, dir)
	})

	t.Run("stub lstat error", func(t *testing.T) {
		p := gomonkey.ApplyFunc(os.Lstat, func(string) (os.FileInfo, error) {
			return nil, errors.New("lstat error")
		})
		_, err := prepareOutputStub(tmp)
		require.NotNil(t, err)
		p.Reset()
	})
}

func TestOutputDir(t *testing.T) {

	t.Run("case os.Getwd error", func(t *testing.T) {
		p := gomonkey.ApplyFunc(os.Getwd, func() (dir string, err error) {
			return "", errors.New("fake error")
		})
		defer p.Reset()
		opts := &params.Option{}
		dir, err := getOutputDir(opts)
		require.NotNil(t, err)
		require.Empty(t, dir)
	})

	t.Run("case rpconly abs", func(t *testing.T) {
		p := gomonkey.ApplyFunc(os.Getwd, func() (dir string, err error) {
			return "wd", nil
		})
		defer p.Reset()
		opts := &params.Option{
			RPCOnly:   true,
			OutputDir: "/is_abs",
		}
		dir, err := getOutputDir(opts)
		require.Nil(t, err)
		require.Equal(t, "/is_abs", dir)
	})

	t.Run("case rpconly not abs", func(t *testing.T) {
		p := gomonkey.ApplyFunc(os.Getwd, func() (dir string, err error) {
			return "wd", nil
		})
		defer p.Reset()
		opts := &params.Option{
			RPCOnly:   true,
			OutputDir: "is_not_abs",
		}
		dir, err := getOutputDir(opts)
		require.Nil(t, err)
		require.Equal(t, "wd/is_not_abs", dir)
	})

	t.Run("case not rpconly with -o", func(t *testing.T) {
		p := gomonkey.ApplyFunc(os.Getwd, func() (dir string, err error) {
			return "wd", nil
		})
		defer p.Reset()
		opts := &params.Option{
			RPCOnly:   false,
			OutputDir: "outputdir",
		}
		dir, err := getOutputDir(opts)
		require.Nil(t, err)
		require.Equal(t, "wd/outputdir", dir)
	})

	t.Run("case not rpconly with mod", func(t *testing.T) {
		p := gomonkey.ApplyFunc(os.Getwd, func() (dir string, err error) {
			return "wd", nil
		})
		defer p.Reset()
		opts := &params.Option{
			RPCOnly:   false,
			GoModEx:   "mod",
			GoMod:     "mod",
			OutputDir: "",
		}
		dir, err := getOutputDir(opts)
		require.Nil(t, err)
		require.Equal(t, "wd", dir)
	})

	t.Run("case not rpconly with mod", func(t *testing.T) {
		p := gomonkey.ApplyFunc(os.Getwd, func() (dir string, err error) {
			return "wd", nil
		})
		defer p.Reset()
		opts := &params.Option{
			RPCOnly:   false,
			GoModEx:   "",
			GoMod:     "",
			OutputDir: "",
			Protofile: "filename.proto",
			IDLType:   config.IDLTypeProtobuf,
		}
		dir, err := getOutputDir(opts)
		require.Nil(t, err)
		require.Equal(t, "wd/filename", dir)
	})
}

func runAndWatch(cmd *cobra.Command, flags map[string]string, args []string) (string, error) {
	n := rand.Int() % 65535
	tmp := os.TempDir()
	tmpd := filepath.Join(tmp, "trpc")
	tmpf := filepath.Join(tmpd, fmt.Sprintf("cmd_output-%d", n))

	os.MkdirAll(tmpd, os.ModePerm)
	defer os.RemoveAll(tmpd)

	f, err := os.OpenFile(tmpf, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpf)

	sout := os.Stdout
	serr := os.Stderr

	os.Stdout = f
	os.Stderr = f
	defer func() {
		os.Stdout = sout
		os.Stderr = serr
	}()

	for k, v := range flags {
		cmd.Flags().Set(k, v)
	}

	// PreRun
	if cmd.PreRunE != nil {
		err = cmd.PreRunE(cmd, args)
		if err != nil {
			return "", err
		}
	} else if cmd.PreRun != nil {
		cmd.PreRun(cmd, args)
	}

	// Run
	if cmd.RunE != nil {
		err = cmd.RunE(cmd, args)
		if err != nil {
			return "", err
		}
	} else if cmd.Run != nil {
		cmd.Run(cmd, args)
	}

	// PostRun
	if cmd.PostRunE != nil {
		err = cmd.PostRunE(cmd, args)
		if err != nil {
			return "", err
		}
	} else if cmd.PostRun != nil {
		cmd.PostRun(cmd, args)
	}

	f.Close()

	b, err := os.ReadFile(tmpf)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
