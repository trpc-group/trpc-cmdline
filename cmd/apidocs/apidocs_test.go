// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package apidocs

import (
	"os"
	"path/filepath"
	"testing"

	"trpc.group/trpc-go/trpc-cmdline/cmd/internal"
)

func TestCmd_ApiDocs(t *testing.T) {
	pwd, _ := os.Getwd()
	defer func() {
		os.Chdir(pwd)
	}()

	wd := filepath.Dir(filepath.Dir(pwd))
	pbdir := filepath.Join(wd, "testcase/apidocs")

	if err := os.Chdir(pbdir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)

	type testCase struct {
		pb        string
		generated string
		flags     map[string]string
		wantErr   bool
	}

	cases := []testCase{
		{
			pb:        "helloworld.proto",
			generated: "helloworld.swagger.json",
			flags: map[string]string{
				"protodir":     ".",
				"protofile":    "helloworld.proto",
				"swagger-out":  "helloworld.swagger.json",
				"check-update": "true",
			},
			wantErr: false,
		},
		{
			pb:        "helloworld.proto",
			generated: "helloworld.openapi.json",
			flags: map[string]string{
				"protodir":     ".",
				"protofile":    "helloworld.proto",
				"swagger":      "false",
				"openapi":      "true",
				"openapi-out":  "helloworld.openapi.json",
				"check-update": "true",
			},
			wantErr: false,
		},
		{
			pb:        "helloworld_restful.proto",
			generated: "helloworld_restful.swagger.json",
			flags: map[string]string{
				"swagger":      "true",
				"protofile":    "helloworld_restful.proto",
				"swagger-out":  "helloworld_restful.swagger.json",
				"check-update": "true",
			},
			wantErr: false,
		},
	}
	apidocsCmd := CMD()
	for _, arg := range cases {
		generated := filepath.Join(pbdir, arg.generated)
		defer os.Remove(generated)
		if _, err := internal.RunAndWatch(apidocsCmd, arg.flags, nil); (err != nil) != arg.wantErr {
			t.Errorf("apidocs cmd, wantErr = %v, got = %v", arg.wantErr, err)
		}
	}
}
