// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFile(t *testing.T) {
	wd, _ := os.Getwd()

	// The suffix of the go_package field may vary between the proto files and their imported proto files.
	// The imported proto files may have the same package but different suffix of go_package.
	t.Run("1-suffix of importpath not equal", func(t *testing.T) {
		proto := filepath.Join(wd, "testcase/importpath/case1/hello.proto")
		dir, file := filepath.Split(proto)
		fd, err := ParseProtoFile(file, []string{dir})
		if err != nil {
			t.Fatalf("parse proto failed: %v", err)
		}
		//package: hello
		//import:  trpc.group/dep/dep1
		//import:  trpc.group/dep/dep2
		require.Len(t, fd.ImportsX, 1)
		require.Equal(t, "dep1", fd.ImportsX[0].Name)
		require.Equal(t, "trpc.group/dep/dep1", fd.ImportsX[0].Path)
	})

	// The package name and the suffix of go_package are different from imported proto files.
	// For imported proto files, the package name is the same but the suffix of go_package is different.
	t.Run("2-suffix of importpath not equal", func(t *testing.T) {
		proto := filepath.Join(wd, "testcase/importpath/case2/hello.proto")
		dir, file := filepath.Split(proto)
		fd, err := ParseProtoFile(file, []string{dir})
		if err != nil {
			t.Fatalf("parse proto failed: %v", err)
		}
		for _, v := range fd.FD.GetDependencies() {
			fmt.Println("import: ", v.GetFileOptions().GetGoPackage())
		}
		//package: hello
		//import:  trpc.group/dep1/proto
		//import:  trpc.group/dep2/proto
		require.Len(t, fd.ImportsX, 2)
		require.Equal(t, "proto1", fd.ImportsX[0].Name)
		require.Equal(t, "trpc.group/dep1/proto", fd.ImportsX[0].Path)
		require.Equal(t, "proto2", fd.ImportsX[1].Name)
		require.Equal(t, "trpc.group/dep2/proto", fd.ImportsX[1].Path)
	})
}
