// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_VersionRE(t *testing.T) {
	type arg struct {
		ver  string
		want []string
	}
	args := []arg{
		{
			ver:  "0.1.1",
			want: []string{"0", "1", "1"},
		},
		{
			ver:  "1.1.1",
			want: []string{"1", "1", "1"},
		},
		{
			ver:  "10.1.1",
			want: []string{"10", "1", "1"},
		},
		{
			ver:  "v10.1.1",
			want: []string{"10", "1", "1"},
		},
		{
			ver:  "protoc v10.1.1",
			want: []string{"10", "1", "1"},
		},
		{
			ver:  "protoc v10.1.1-beta",
			want: []string{"10", "1", "1"},
		},
	}

	for _, a := range args {
		vals := versionRE.FindStringSubmatch(a.ver)
		require.Equal(t, a.want, vals[1:])
	}
}
