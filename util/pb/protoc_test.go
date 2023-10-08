package pb

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey"

	"trpc.group/trpc-go/trpc-cmdline/config"
)

var wd string

func TestMain(m *testing.M) {
	if _, err := config.Init(); err != nil {
		panic(err)
	}
	if err := setup(); err != nil {
		panic(err)
	}

	d, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	wd = filepath.Join(d, "testcase")

	os.Exit(m.Run())
}

func TestProtoc(t *testing.T) {
	languages := []string{"go"}
	outputdir := filepath.Join(wd, "generated")
	os.Mkdir(outputdir, os.ModePerm)
	defer os.RemoveAll(outputdir)

	type args struct {
		protodirs []string
		protofile string
		pb2impt   map[string]string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			"case1",
			args{
				[]string{wd},
				"helloworld.proto",
				nil,
			},
		},
		{
			"case2",
			args{
				[]string{wd},
				"helloworld.proto",
				map[string]string{"helloworld.proto": "trpc.group/examples/helloworld"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, lang := range languages {
				err := Protoc(tt.args.protodirs, tt.args.protofile, lang, outputdir, WithPb2ImportPath(tt.args.pb2impt))
				if err != nil {
					t.Errorf("Protoc() error = %v", err)
				}
			}
		})
	}

	// clean
	os.RemoveAll(outputdir)
}

func Test_makeProtocOutByLanguage(t *testing.T) {
	type args struct {
		language  string
		pbpkg     string
		outputdir string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case_other_language",
			args: args{
				language:  "other",
				pbpkg:     "pbpkg",
				outputdir: "outputdir",
			},
			want: "--other_out=pbpkg:outputdir",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := gomonkey.ApplyFunc(os.MkdirAll, func(path string, perm os.FileMode) error {
				return nil
			})
			defer p.Reset()

			if got := makeProtocOutByLanguage(tt.args.language, tt.args.pbpkg, tt.args.outputdir); got != tt.want {
				t.Errorf("makeProtocOutByLanguage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_genRelPathFromWd(t *testing.T) {
	type args struct {
		wd        string
		protofile string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
		absRsp  map[string]string
		absErr  map[string]error
	}{
		{
			name: "get wd abs path error",
			args: args{
				wd:        "wd",
				protofile: "protofile",
			},
			want:    "",
			wantErr: true,
			absRsp:  map[string]string{},
			absErr: map[string]error{
				"wd": fmt.Errorf("error"),
			},
		},
		{
			name: "get protofile abs path error",
			args: args{
				wd:        "wd",
				protofile: "protofile",
			},
			want:    "",
			wantErr: true,
			absRsp:  map[string]string{},
			absErr: map[string]error{
				"protofile": fmt.Errorf("error"),
			},
		},
		{
			name: "gen ref path error",
			args: args{
				wd:        "wd",
				protofile: "protofile",
			},
			want:    "",
			wantErr: true,
			absRsp: map[string]string{
				"wd":        "/wd",
				"protofile": "a/protofile",
			},
			absErr: map[string]error{},
		},
		{
			name: "gen ref path succ",
			args: args{
				wd:        "wd",
				protofile: "protofile",
			},
			want:    "a/protofile",
			wantErr: false,
			absRsp: map[string]string{
				"wd":        "/wd",
				"protofile": "/wd/a/protofile",
			},
			absErr: map[string]error{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := gomonkey.ApplyFunc(filepath.Abs, func(path string) (string, error) {
				return tt.absRsp[path], tt.absErr[path]
			})
			defer p.Reset()

			got, err := genRelPathFromWd(tt.args.protofile, tt.args.wd)
			if (err != nil) != tt.wantErr {
				t.Errorf("genRelPathFromWd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("genRelPathFromWd() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockFileInfo struct {
	err   error
	isDir bool
}

func (s *mockFileInfo) Name() string       { return "" }
func (s *mockFileInfo) Size() int64        { return 0 }
func (s *mockFileInfo) Mode() os.FileMode  { return 0 }
func (s *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (s *mockFileInfo) IsDir() bool        { return s.isDir }
func (s *mockFileInfo) Sys() interface{}   { return nil }

func Test_genRelPathFromWdWithDirs(t *testing.T) {
	type args struct {
		protodirs []string
		protofile string
		wd        string
	}
	tests := []struct {
		name                string
		args                args
		want                string
		wantErr             bool
		lstatErr            error
		isDir               bool
		genRelPathFromWdRsp string
		genRelPathFromWdErr error
	}{
		{
			name: "os.Lstat error",
			args: args{
				protodirs: []string{"dir1"},
				protofile: "file",
				wd:        "path",
			},
			want:     "",
			wantErr:  true,
			lstatErr: fmt.Errorf("error"),
		},
		{
			name: "genRelPathFromWd error",
			args: args{
				protodirs: []string{"dir1"},
				protofile: "file",
				wd:        "path",
			},
			want:                "",
			wantErr:             true,
			isDir:               false,
			genRelPathFromWdRsp: "path",
			genRelPathFromWdErr: fmt.Errorf("error"),
		},
		{
			name: "genRelPathFromWd succ",
			args: args{
				protodirs: []string{"dir1"},
				protofile: "file",
				wd:        "path",
			},
			want:                "path",
			wantErr:             false,
			isDir:               false,
			genRelPathFromWdRsp: "path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := gomonkey.ApplyFunc(os.Lstat, func(name string) (os.FileInfo, error) {
				return &mockFileInfo{isDir: tt.isDir}, tt.lstatErr
			})
			defer p.Reset()
			p.ApplyFunc(genRelPathFromWd, func(protofile, wd string) (string, error) {
				return tt.genRelPathFromWdRsp, tt.genRelPathFromWdErr
			})

			got, err := genRelPathFromWdWithDirs(tt.args.protodirs, tt.args.protofile, tt.args.wd)
			if (err != nil) != tt.wantErr {
				t.Errorf("genRelPathFromWdWithDirs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("genRelPathFromWdWithDirs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func setup() error {
	if _, err := config.Init(); err != nil {
		return err
	}
	deps, err := config.LoadDependencies()
	if err != nil {
		return err
	}
	return config.SetupDependencies(deps)
}
