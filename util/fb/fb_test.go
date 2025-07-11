// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package fb

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlatcNormal(t *testing.T) {
	dir := "./testcase/normal"
	file := path.Join(dir, "hello.fbs")
	fileAbs, err := filepath.Abs(file)
	require.Nil(t, err)
	out := filepath.Join(os.TempDir(), "flatc_generated")
	os.RemoveAll(out)
	fb2ImportPath := map[string]string{
		"hello.fbs": "trpc.group/examples/hello",
	}
	pkg2ImportPath := map[string]string{
		"hello": "trpc.group/examples/hello",
	}
	opts := []Option{
		WithFbsDirs([]string{dir}),
		WithFbsfile(fileAbs),
		WithLanguage("go"),
		WithPackagePath("hello"),
		WithOutputdir(out),
		WithFb2ImportPath(fb2ImportPath),
		WithPkg2ImportPath(pkg2ImportPath),
	}
	f := NewFbs(opts...)
	err = f.Flatc()
	require.Nil(t, err, fmt.Sprintf("err: %+v", err))
}

func TestFlatcError(t *testing.T) {
	dir := "./testcase/error"
	file := path.Join(dir, "hello.fbs")
	fileAbs, err := filepath.Abs(file)
	require.Nil(t, err)
	out := filepath.Join(os.TempDir(), "flatc_generated")
	os.RemoveAll(out)
	fb2ImportPath := map[string]string{
		"hello.fbs": "trpc.group/examples/hello",
	}
	pkg2ImportPath := map[string]string{
		"hello": "trpc.group/examples/hello",
	}
	opts := []Option{
		WithFbsDirs([]string{dir}),
		WithFbsfile(fileAbs),
		WithLanguage("go"),
		WithPackagePath("hello"),
		WithOutputdir(out),
		WithFb2ImportPath(fb2ImportPath),
		WithPkg2ImportPath(pkg2ImportPath),
	}
	f := NewFbs(opts...)
	err = f.Flatc()
	require.NotNil(t, err)
}

func TestFlatcMultiFBDiffGopkg(t *testing.T) {
	dir := "./testcase/multi-fb-diff-gopkg"
	file := path.Join(dir, "fbsread.fbs")
	fileAbs, err := filepath.Abs(file)
	require.Nil(t, err)
	out := filepath.Join(os.TempDir(), "flatc_generated")
	os.RemoveAll(out)
	fb2ImportPath := map[string]string{
		"fbsread.fbs": "trpc.group/trpcprotocol/circlesearch/common_feedcloud_fbsread",
		"circlesearch/common/feedcloud/fbsmeta.fbs": "trpc.group/trpcprotocol/circlesearch/common_feedcloud_fbsmeta",
		"circlesearch/common/feedcloud/common.fbs":  "trpc.group/trpcprotocol/circlesearch/common_feedcloud_common",
	}
	pkg2ImportPath := map[string]string{
		"trpc.circlesearch.common_feedcloud_fbsread": "trpc.group/trpcprotocol/circlesearch/common_feedcloud_fbsread",
		"trpc.circlesearch.common_feedcloud_fbsmeta": "trpc.group/trpcprotocol/circlesearch/common_feedcloud_fbsmeta",
		"trpc.circlesearch.common_feedcloud_common":  "trpc.group/trpcprotocol/circlesearch/common_feedcloud_common",
	}
	opts := []Option{
		WithFbsDirs([]string{dir}),
		WithFbsfile(fileAbs),
		WithLanguage("go"),
		WithPackagePath("hello"),
		WithOutputdir(out),
		WithFb2ImportPath(fb2ImportPath),
		WithPkg2ImportPath(pkg2ImportPath),
	}
	f := NewFbs(opts...)
	err = f.Flatc()
	require.Nil(t, err, fmt.Sprintf("err: %+v", err))
}

func Test_replacePkgNameInFiles(t *testing.T) {
	type args struct {
		files        []string
		originImport string
		importPath   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "file not exists",
			args: args{
				files:        []string{"a_file_that_doesnot_exist.fbs"},
				originImport: "nana",
				importPath:   "lala",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := replacePkgNameInFiles(tt.args.files, tt.args.originImport, tt.args.importPath); (err != nil) != tt.wantErr {
				t.Errorf("replacePkgNameInFiles got error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
