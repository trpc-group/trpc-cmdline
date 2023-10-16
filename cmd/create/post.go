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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"trpc.group/trpc-go/trpc-cmdline/plugin"
	"trpc.group/trpc-go/trpc-cmdline/util/fs"
	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

// PostRunE provides *cobra.Command.PostRunE.
func (c *Create) PostRunE(cmd *cobra.Command, args []string) error {
	wd, _ := os.Getwd()
	defer os.Chdir(wd)

	if err := os.Chdir(c.options.OutputDir); err != nil {
		return err
	}

	for _, p := range append(plugin.Plugins, plugin.PluginsExt[c.options.Language]...) {
		if !p.Check(c.fileDescriptor, c.options) {
			continue
		}

		if err := p.Run(c.fileDescriptor, c.options); err != nil {
			return fmt.Errorf(
				"running plugin `%s`, err: %w",
				p.Name(), err)
		}
		if c.options.Verbose {
			log.Info(
				"running plugin %s`%s`%s, succeed",
				log.ColorRed,
				p.Name(),
				log.ColorGreen)
		}
	}

	log.Info(
		"Create tRPC project %s`%s`%s post process: succeed! (〃'▽'〃)",
		log.ColorRed,
		fs.BaseNameWithoutExt(c.fileDescriptor.FilePath),
		log.ColorGreen)
	return nil
}
