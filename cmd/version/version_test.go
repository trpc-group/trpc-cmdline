// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package version

import (
	"strings"
	"testing"

	"trpc.group/trpc-go/trpc-cmdline/cmd/internal"
	"trpc.group/trpc-go/trpc-cmdline/config"
)

func TestCmd_Version(t *testing.T) {
	if _, err := config.Init(); err != nil {
		t.Errorf("config init error: %v", err)
	}

	versionCmd := CMD()
	output, err := internal.RunAndWatch(versionCmd, nil, nil)
	if err != nil {
		t.Errorf("versionCmd run and watch error: %v", err)
	}

	if !strings.Contains(output, config.TRPCCliVersion) {
		t.Errorf("versionCmd.Run() output version mismatch")
	}
}
