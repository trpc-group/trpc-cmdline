// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package cmd provides commands to help developer generating project, testing, etc.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"trpc.group/trpc-go/trpc-cmdline/cmd/apidocs"
	"trpc.group/trpc-go/trpc-cmdline/cmd/completion"
	"trpc.group/trpc-go/trpc-cmdline/cmd/create"
	"trpc.group/trpc-go/trpc-cmdline/cmd/setup"
	"trpc.group/trpc-go/trpc-cmdline/cmd/version"
	"trpc.group/trpc-go/trpc-cmdline/config"
	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf(`Execution err:
	%+v
Please run "trpc -h" or "trpc create -h" (or "trpc {some-other-subcommand} -h") for help messages.
`, err)
		os.Exit(1) // Exist with non-zero errcode to indicate failure.
	}
}

var (
	cfgFile     string
	verboseFlag bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "trpc",
	Short: "trpc is an efficiency tool for convenient trpc service development",
	Long: `trpc is an efficiency tool for convenient trpc service development.

For example:
- Generate a complete project or corresponding RPC stub by specifying the pb file
- Send RPC test requests to the target service

Try using trpc framework + trpc tool to write your next trpc service!
`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

const defaultConfigFile = "path/to/trpc.yaml"

func init() {
	cobra.OnInitialize(func() {
		if err := initConfig(); err != nil {
			panic(err)
		}
	})

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", defaultConfigFile,
		"Path to the configuration file (automatically calculated)")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Display detailed log information")

	rootCmd.AddCommand(create.CMD())
	rootCmd.AddCommand(setup.CMD())
	rootCmd.AddCommand(completion.CMD())
	rootCmd.AddCommand(apidocs.CMD())
	rootCmd.AddCommand(version.CMD())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	log.SetVerbose(verboseFlag)

	d, err := config.Init()
	if err != nil {
		return fmt.Errorf("config init err: %w", err)
	}

	if cfgFile == defaultConfigFile {
		cfgFile = filepath.Join(d, "trpc.yaml")
	}
	if cfgFile != "" {
		// Use config file from the flag.
		log.Debug("using config from %s", cfgFile)
		viper.SetConfigFile(cfgFile)
	}

	viper.AutomaticEnv() // Read in environment variables that matched.

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("viper read in config: %w", err)
	}

	return nil
}
