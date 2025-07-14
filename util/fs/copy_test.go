// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package fs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	type args struct {
		src  string
		dest string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"1-copy/fa", args{"copy/fa", "target/fa"}, false},
		{"2-copy/fa", args{"copy/fa", "target/copy/fa"}, false},
		{"3-copy/d", args{"copy/d", "target/d"}, false},
		{"4-copy/d", args{"copy/d", "target/e/d"}, false},
	}
	tmpdir := filepath.Join(wd, "testcase/target")
	err := os.MkdirAll(tmpdir, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := filepath.Join(wd, "testcase", tt.args.src)
			dest := filepath.Join(wd, "testcase", tt.args.dest)

			if err := Copy(src, dest); (err != nil) != tt.wantErr {
				t.Errorf("Copy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_lcopy(t *testing.T) {
	t.Run("case fail", func(t *testing.T) {
		p := gomonkey.NewPatches()
		p.ApplyFunc(os.Readlink, func(name string) (string, error) {
			return "", errors.New("fake error")
		})
		p.ApplyFunc(os.Symlink, func(oldname, newname string) error {
			return nil
		})
		defer p.Reset()

		info, _ := os.Lstat("copy/a")
		err := lcopy("src", "dst", info)
		require.NotNil(t, err)
	})

	t.Run("case succ", func(t *testing.T) {
		p := gomonkey.NewPatches()
		p.ApplyFunc(os.Readlink, func(name string) (string, error) {
			return "", nil
		})
		p.ApplyFunc(os.Symlink, func(oldname, newname string) error {
			return nil
		})
		defer p.Reset()

		info, _ := os.Lstat("copy/a")
		err := lcopy("src", "dst", info)
		require.Nil(t, err)
	})
}
