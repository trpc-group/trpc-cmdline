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

	var allp []plugin.Plugin
	allp = append(allp, plugin.Plugins...)
	allp = append(allp, plugin.PluginsExt[c.options.Language]...)

	err := os.Chdir(c.options.OutputDir)
	if err != nil {
		return err
	}
	for _, p := range allp {
		if !p.Check(c.fileDescriptor, c.options) {
			continue
		}

		err = p.Run(c.fileDescriptor, c.options)
		if err != nil {
			return fmt.Errorf(
				"running plugin `%s`, err: %w",
				p.Name(), err)
		}
		if c.options.Verbose {
			log.Info(
				"running plugin %s`%s`%s, err: %v",
				log.ColorRed,
				p.Name(),
				log.ColorGreen,
				err)
		}
	}

	log.Info(
		"Create tRPC project %s`%s`%s post process: succeed! (〃'▽'〃)",
		log.ColorRed,
		fs.BaseNameWithoutExt(c.fileDescriptor.FilePath),
		log.ColorGreen)
	return nil
}
