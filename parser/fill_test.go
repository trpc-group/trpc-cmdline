// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package parser

import (
	"testing"

	"trpc.group/trpc-go/trpc-cmdline/util/lang"
)

func Test_explodeImport(t *testing.T) {
	tests := []struct {
		name           string
		arg            string
		wantImportName string
		wantImportPath string
	}{
		{
			name:           "/tencent/common",
			arg:            "/tencent/common",
			wantImportName: "common",
			wantImportPath: "/tencent/common",
		},
		{
			name:           "trpc.group/tencent/common",
			arg:            "trpc.group/tencent/common",
			wantImportName: "common",
			wantImportPath: "trpc.group/tencent/common",
		},
		{
			name:           "trpc.group/tencent/common;xyz",
			arg:            "trpc.group/tencent/common;xyz",
			wantImportName: "xyz",
			wantImportPath: "trpc.group/tencent/common",
		},
		{
			name:           "common",
			arg:            "common",
			wantImportName: "common",
			wantImportPath: "common",
		},
		{
			name:           "a.b.c.d",
			arg:            "a.b.c.d",
			wantImportName: "a_b_c_d",
			wantImportPath: "a.b.c.d",
		},
		{
			name:           "trpc.group/hello/a.b.c.d",
			arg:            "trpc.group/hello/a.b.c.d",
			wantImportName: "a_b_c_d",
			wantImportPath: "trpc.group/hello/a.b.c.d",
		},
		{
			name:           "trpc.group/hello/a.b.c.d;xyz",
			arg:            "trpc.group/hello/a.b.c.d;xyz",
			wantImportName: "xyz",
			wantImportPath: "trpc.group/hello/a.b.c.d",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotImportName, gotImportPath := lang.ExplodeImport(tt.arg)
			if gotImportName != tt.wantImportName {
				t.Errorf("explodeImport() gotImportName = %v, want %v", gotImportName, tt.wantImportName)
			}
			if gotImportPath != tt.wantImportPath {
				t.Errorf("explodeImport() gotImportPath = %v, want %v", gotImportPath, tt.wantImportPath)
			}
		})
	}
}
