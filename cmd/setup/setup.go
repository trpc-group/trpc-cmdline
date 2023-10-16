// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package setup provides setup command.
package setup

import (
	"fmt"

	"github.com/spf13/cobra"

	"trpc.group/trpc-go/trpc-cmdline/config"
	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

// CMD returns setup command.
func CMD() *cobra.Command {
	setup := func(languages []string) error {
		// Load dependencies according to languages specified.
		deps, err := config.LoadDependencies(languages...)
		if err != nil {
			return err
		}
		// Setup dependencies.
		return config.SetupDependencies(deps)
	}
	setupCmd := &cobra.Command{
		Use:           "setup",
		Short:         "Initialize setup && Install dependency tools",
		Long:          "Initialize setup && Install dependency tools.",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			log.SetPrefix("[setup]")
			log.SetVerbose(true)
			log.Info("Initializing setup && Installing dependency tools")
			lang, err := cmd.Flags().GetStringArray("lang")
			if err != nil {
				return fmt.Errorf("get lang flag err: %w", err)
			}
			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return fmt.Errorf("get force flag err: %w", err)
			}
			if _, err := config.Init(config.WithForce(force)); err != nil {
				return fmt.Errorf("init config with force=%v, err: %w", force, err)
			}
			// Do setup according to language.
			if err := setup(lang); err != nil {
				return fmt.Errorf("setup failed: %v", err)
			}
			log.Info("Setup completed")
			return nil
		},
	}
	setupCmd.Flags().StringArrayP("lang", "l", nil, "setup tools for languages")
	setupCmd.Flags().BoolP("force", "f", false, "force extracting assets to overwrite the existing asset files")
	return setupCmd
}
