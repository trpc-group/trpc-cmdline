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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-cmdline/util/semver"
)

func TestParseNewProtocVersion(t *testing.T) {
	s := versionRE.FindStringSubmatch("22.0")
	require.Equal(t, []string{"22.0", "", "22", "0"}, s)
	version := versionNumber(strings.Join(s[1:], "."))
	require.Equal(t, ".22.0", version)
	major, minor, revision := semver.Versions(version)
	require.Equal(t, 22, major)
	require.Equal(t, 0, minor)
	require.Equal(t, 0, revision)
}
