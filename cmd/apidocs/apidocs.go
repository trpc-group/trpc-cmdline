// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package apidocs provides apidocs command.
package apidocs

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"trpc.group/trpc-go/trpc-cmdline/config"
	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/parser"
	"trpc.group/trpc-go/trpc-cmdline/util/apidocs/openapi"
	"trpc.group/trpc-go/trpc-cmdline/util/apidocs/swagger"
	"trpc.group/trpc-go/trpc-cmdline/util/fs"
	"trpc.group/trpc-go/trpc-cmdline/util/log"
	"trpc.group/trpc-go/trpc-cmdline/util/paths"
	"trpc.group/trpc-go/trpc-cmdline/util/pb"
)

// CMD returns apidocs command.
func CMD() *cobra.Command {
	apidocsCmd := &cobra.Command{
		Use:   "apidocs",
		Short: "Generate apidocs",
		Long: `Generate apidocs, supporting swagger and openapi.
When generating swagger documentation, the summary information of rpc methods
includes the leading comments of the rpc method in the pb file.
The field information of input and output also includes the leading and trailing comments of the field.
	`,
		RunE: runAPIDocs,
	}

	// Flags related to swagger.
	apidocsCmd.Flags().Bool("swagger", true, "Generate swagger apidocs")
	apidocsCmd.Flags().Bool("swagger-json-param", false, "Generate swagger apidocs using json body")
	apidocsCmd.Flags().String("swagger-out", "apidocs.swagger.json", "Output path for swagger apidocs")

	// Flags related to openapi.
	apidocsCmd.Flags().Bool("openapi", false, "Generate openapi apidocs")
	apidocsCmd.Flags().String("openapi-out", "apidocs.openapi.json", "Output path for openapi apidocs")

	// Proto files and search paths.
	apidocsCmd.Flags().StringP("protofile", "p", "", "Specify the pb file for the service")
	apidocsCmd.Flags().StringArrayP("protodir", "d",
		[]string{"."}, "Search paths for pb files (including dependency pb files), can be specified multiple times")

	// Flag for alias.
	apidocsCmd.Flags().Bool("alias", false, "Use alias mode for rpcname")
	// Preserve original rpcname.
	apidocsCmd.Flags().BoolP("keep-orig-rpcname", "k", true,
		"Preserve the original rpcname (if --alias=true), set it to false if you only want alias names.")

	// The rules of documents order.
	apidocsCmd.Flags().Bool("order-by-pbname", false,
		"Use the order defined in the PB for api documentation, defaults to alphabetical order")
	return apidocsCmd
}

// runAPIDocs generates API documents.
func runAPIDocs(cmd *cobra.Command, _ []string) error {
	if _, err := config.Init(); err != nil {
		return fmt.Errorf("init config err: %w", err)
	}
	// Check the command line arguments.
	option, err := loadAPIDocsOptions(cmd.Flags())
	if err != nil {
		return fmt.Errorf("error checking command options: %w", err)
	}
	// Parse pb.
	fileDescriptor, err := parser.Parse(
		option.Protofile,
		option.Protodirs,
		option.IDLType,
		parser.WithAliasOn(option.AliasOn),
		parser.WithLanguage(option.Language),
		parser.WithRPCOnly(option.RPCOnly),
	)
	if err != nil {
		return fmt.Errorf("error parsing pb file %s: %w", option.Protofile, err)
	}
	// Dump fd for debugging.
	fileDescriptor.Dump()
	return genAPIDocs(fileDescriptor, option)
}

func genAPIDocs(fileDescriptor *descriptor.FileDescriptor, option *params.Option) error {
	if option.SwaggerOn {
		if err := swagger.GenSwagger(fileDescriptor, option); err != nil {
			return fmt.Errorf("create swagger apidocs error: %w", err)
		}
		log.Info("Generate the swagger apidocs of ```%s``` success", option.Protofile)
	}
	if option.OpenAPIOn {
		if err := openapi.GenOpenAPI(fileDescriptor, option); err != nil {
			return fmt.Errorf("create openapi apidocs error: %w", err)
		}
		log.Info("Generate the openapi apidocs of ```%s``` success", option.Protofile)
	}
	return nil
}

func loadAPIDocsOptions(flagSet *pflag.FlagSet) (*params.Option, error) {
	option := &params.Option{}

	// Flags related to swagger.
	option.SwaggerOn, _ = flagSet.GetBool("swagger")
	option.SwaggerOptJSONParam, _ = flagSet.GetBool("swagger-json-param")
	option.SwaggerOut, _ = flagSet.GetString("swagger-out")

	// Flags related to openapi.
	option.OpenAPIOn, _ = flagSet.GetBool("openapi")
	option.OpenAPIOut, _ = flagSet.GetString("openapi-out")
	option.OrderByPBName, _ = flagSet.GetBool("order-by-pbname")

	// Proto files and search paths.
	var err error
	option.Protofile, err = flagSet.GetString("protofile")
	if err != nil {
		return nil, err
	}
	option.Protodirs, _ = flagSet.GetStringArray("protodir")
	// Always append the current working directory.
	option.Protodirs = append(option.Protodirs, ".")
	option.AliasOn, _ = flagSet.GetBool("alias")
	option.KeepOrigRPCName, _ = flagSet.GetBool("keep-orig-rpcname")

	// Check if the pb file is valid.
	if len(option.Protofile) == 0 {
		return nil, errors.New("invalid protofile")
	}

	// Locate the pb file.
	target, err := fs.LocateFile(option.Protofile, option.Protodirs)
	if err != nil {
		return nil, err
	}
	option.Protofile = filepath.Base(target)
	option.ProtofileAbs = target
	option.Protodirs = append(option.Protodirs, filepath.Dir(target))
	option.IDLType = config.IDLTypeProtobuf

	// Adjust the search path for pb files.
	if err := fixProtodirs(option); err != nil {
		return nil, err
	}
	return option, nil
}

// fixProtodirs fixes the search path for pb files.
func fixProtodirs(option *params.Option) error {
	p, err := paths.Locate(pb.ProtoTRPC)
	if err != nil {
		return err
	}
	option.Protodirs = fs.UniqFilePath(append(append(option.Protodirs, p),
		paths.ExpandSearch(p)...))
	return nil
}
