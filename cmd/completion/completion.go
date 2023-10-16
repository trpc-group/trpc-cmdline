// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package completion provides completion command.
package completion

import (
	"os"

	"github.com/spf13/cobra"

	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

// CMD returns completion command.
// Users can run `trpc completion your_shell_name` to get script for auto-complete.
// To append to command to your shell environment, run
//
//	`trpc completion your_shell_name >> your_shell_init_file`
func CMD() *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell autocompletion scripts",
		Long: `Generate shell autocompletion scripts. Usage example:

Bash:

$ source <(yourprogram completion bash)

# To load completions for each session, execute once:
Linux:
	$ yourprogram completion bash > /etc/bash_completion.d/yourprogram
MacOS:
	$ yourprogram completion bash > /usr/local/etc/bash_completion.d/yourprogram

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ yourprogram completion zsh > "${fpath[1]}/_yourprogram"

# You will need to start a new shell for this Setup to take effect.

Fish:

$ yourprogram completion fish | source

# To load completions for each session, execute once:
$ yourprogram completion fish > ~/.config/fish/completions/yourprogram.fish
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletion(os.Stdout)
			default:
				log.Error("%s is not supported", args[0])
			}
		},
	}
	return completionCmd
}
