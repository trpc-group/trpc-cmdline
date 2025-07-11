// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package plugin

import (
	"os"
	"testing"

	"trpc.group/trpc-go/trpc-cmdline/config"
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func setup() error {
	if _, err := config.Init(); err != nil {
		return err
	}
	deps, err := config.LoadDependencies()
	if err != nil {
		return err
	}
	return config.SetupDependencies(deps)
}
