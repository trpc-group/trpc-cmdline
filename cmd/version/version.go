// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package version provides version command.
package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"trpc.group/trpc-go/trpc-cmdline/config"
)

// CMD returns the version command.
func CMD() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version of trpc command (commit hash)",
		Long:  "Show the version of trpc command (commit hash).",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("trpc-group/trpc-cmdline version:", config.TRPCCliVersion)
		},
	}
	return versionCmd
}
