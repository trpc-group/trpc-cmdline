// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package create

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateCmdByFlatbuffers(t *testing.T) {
	require.Nil(t, setup(nil))

	wd, _ := os.Getwd()
	defer func() {
		os.Chdir(wd)
	}()

	pd := filepath.Dir(wd)
	pd = filepath.Dir(pd)
	testdatadir := filepath.Join(pd, "testcase/flatbuffers")

	testcases := []testcase{
		{
			name:   "1.1-without-import",
			pbdir:  "1-without-import",
			pbfile: "helloworld.fbs",
			alias:  true,
		}, {
			name:    "1.2-without-import (rpconly)",
			pbdir:   "1-without-import",
			pbfile:  "helloworld.fbs",
			rpconly: true,
		}, {
			name:          "1.3-without-import (split by method)",
			pbdir:         "1-without-import",
			pbfile:        "helloworld.fbs",
			splitByMethod: true,
		}, {
			name:   "2-multi-fb-same-namespace",
			pbdir:  "2-multi-fb-same-namespace",
			pbfile: "hello.fbs",
		}, {
			name:   "3-multi-fb-diff-namespace",
			pbdir:  "3-multi-fb-diff-namespace",
			pbfile: "helloworld.fbs",
		}, {
			name:   "4.1-multi-fb-same-namespace-diff-dir",
			pbdir:  "4.1-multi-fb-same-namespace-diff-dir",
			pbfile: "helloworld.fbs",
		}, {
			name:   "4.2-multi-fb-same-namespace-diff-dir",
			pbdir:  "4.2-multi-fb-same-namespace-diff-dir",
			pbfile: "helloworld.fbs",
		}, {
			name:   "5-multi-fb-diff-gopkg fb",
			pbdir:  "5-multi-fb-diff-gopkg",
			pbfile: "fbsread.fbs",
		}, {
			name:          "5.1-multi-fb-diff-gopkg (split by method)",
			pbdir:         "5-multi-fb-diff-gopkg",
			pbfile:        "fbsread.fbs",
			splitByMethod: true,
		},
	}

	tmp := filepath.Join(os.TempDir(), "create/generated_fb")
	os.RemoveAll(tmp)
	defer os.RemoveAll(tmp)

	for _, tt := range testcases {
		tt := tt
		t.Run("CreateCmd/fbs-"+tt.name, func(t *testing.T) {
			any := filepath.Join(testdatadir, tt.pbdir)

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

			opts := []string{}
			for _, d := range dirs {
				opts = append(opts, "--fbsdir", d)
			}
			opts = append(opts, "--fbs", tt.pbfile)

			out := filepath.Join(tmp, tt.name)
			opts = append(opts, "-o", out)

			if tt.rpconly {
				opts = append(opts, "--rpconly")
			}
			if tt.splitByMethod {
				opts = append(opts, "-s")
			}
			if tt.alias {
				opts = append(opts, "--alias")
			}
			opts = append(opts, "--mock", "false")

			resetFlags(createCmd)
			runCreateCmd(t, "fbs "+tt.name, opts, out, tt.wantErr)
		})
	}
}
