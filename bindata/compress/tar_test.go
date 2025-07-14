// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package compress

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTar(t *testing.T) {
	// setup
	tmp := os.TempDir()
	dir := filepath.Join(tmp, "bindata")
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		panic(err)
	}
	fp := filepath.Join(dir, "test.txt")
	txt := []byte("helloworld")
	err = os.WriteFile(fp, txt, 0666)
	if err != nil {
		panic(err)
	}

	// tar
	buf := bytes.Buffer{}
	err = Tar(fp, &buf)
	require.Nil(t, err)
	require.NotZero(t, buf.Len())

	// untar
	reader := bytes.NewReader(buf.Bytes())
	dst := filepath.Join(dir, "test2.txt")
	err = Untar(dst, reader)
	require.Nil(t, err)

	dat, err := os.ReadFile(dst)
	require.Nil(t, err)
	require.Equal(t, txt, dat)

	os.RemoveAll(dir)
}

func TestTar_Install(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	src := wd

	// tar
	buf := bytes.Buffer{}
	err = Tar(src, &buf)
	require.Nil(t, err)
	require.NotZero(t, buf.Len())

	// untar
	tmp := os.TempDir()
	dst := filepath.Join(tmp, "bindata")
	defer os.RemoveAll(dst)

	err = Untar(dst, &buf)
	require.Nil(t, err)

	// compare file list
	srcFileSet := make(map[string]struct{})
	filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		path = strings.TrimPrefix(path, src)
		srcFileSet[path] = struct{}{}
		return nil
	})
	dstFileSet := make(map[string]struct{})
	filepath.Walk(dst, func(path string, info os.FileInfo, err error) error {
		path = strings.TrimPrefix(path, dst)
		dstFileSet[path] = struct{}{}
		return nil
	})
	require.Equal(t, srcFileSet, dstFileSet)
}
