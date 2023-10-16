// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package style_test

import (
	"errors"
	"go/format"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-cmdline/util/style"
)

var sourceRaw = `package main

func main() {
fmt.Println("hello world")
	fmt.Println("Hello world")
fmt.Println("hello world")
return
}
`

var sourceFormatted = `package main

func main() {
	fmt.Println("hello world")
	fmt.Println("Hello world")
	fmt.Println("hello world")
	return
}
`

func TestFormat(t *testing.T) {

	t.Run("format go", func(t *testing.T) {
		var err error
		// setup
		dir := filepath.Join(os.TempDir(), "gofmt")
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			panic(err)
		}
		defer os.RemoveAll(dir)

		fp := filepath.Join(dir, "sourceRaw.go")
		if err = os.WriteFile(fp, []byte(sourceRaw), 0666); err != nil {
			panic(err)
		}

		// format
		if err = style.Format(fp, "go"); err != nil {
			t.Fatal(err)
		}

		// validate
		var buf []byte
		if buf, err = os.ReadFile(fp); err != nil {
			panic(err)
		}
		require.Equal(t, string(buf), sourceFormatted)
	})

	t.Run("format !go", func(t *testing.T) {
		err := style.Format("fpath", "other")
		require.Nil(t, err)
	})
}

func TestGoFmt_Exception(t *testing.T) {

	t.Run("GoFmt (ReadFile error)", func(t *testing.T) {
		p := gomonkey.ApplyFunc(os.ReadFile, func(string) ([]byte, error) {
			return nil, errors.New("read file error")
		})
		defer p.Reset()
		err := style.GoFmt("fpath")
		require.NotNil(t, err)
	})

	t.Run("GoFmt (Source error)", func(t *testing.T) {
		p := gomonkey.ApplyFunc(os.ReadFile, func(string) ([]byte, error) {
			return nil, nil
		})
		p.ApplyFunc(format.Source, func(src []byte) ([]byte, error) {
			return nil, errors.New("source error")
		})
		defer p.Reset()
		err := style.GoFmt("fpath")
		require.NotNil(t, err)
	})

	t.Run("GoFmt (WriteFile error)", func(t *testing.T) {
		p := gomonkey.NewPatches()
		p.ApplyFunc(os.ReadFile, func(filename string) ([]byte, error) {
			return nil, nil
		})
		p.ApplyFunc(format.Source, func(src []byte) ([]byte, error) {
			return nil, nil
		})
		p.ApplyFunc(os.WriteFile, func(filename string, data []byte, perm os.FileMode) error {
			return errors.New("fake error")
		})
		defer p.Reset()
		err := style.GoFmt("fpath")
		require.NotNil(t, err)
	})
}

func TestGoFmtDir(t *testing.T) {
	p := gomonkey.ApplyFunc(style.GoFmt, func(string) error {
		return nil
	})
	defer p.Reset()
	err := style.GoFmtDir("./")
	require.Nil(t, err)
}
