// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package create

import "github.com/spf13/cobra"

// AddCreateFlags adds flags to create sub command.
func AddCreateFlags(createCmd *cobra.Command) {
	// There is stub code generated by idl (protobuf, flatbuffers).
	// - pb
	createCmd.Flags().StringArrayP("protodir", "d", []string{"."},
		"Search paths for pb files (including dependent pb files), can be specified multiple times")
	createCmd.Flags().StringP("protofile", "p", "", "Specify the pb file corresponding to the service")
	// - fb
	createCmd.Flags().StringArray("fbsdir", []string{"."}, "Search paths for flatbuffers include files")
	createCmd.Flags().String("fbs", "", "Specify the flatbuffers file corresponding to the service")

	// Whether to pass protoc/flatc by the basename of "--protofile/--fbs" provided above.
	createCmd.Flags().Bool("usebasename", false, "Whether to pass the basename of --protofile/--fbs to protoc/flatc")

	// Generate stub code without IDL.
	createCmd.Flags().StringP("non-protocol-type", "n", "",
		"Generate project types without pb protocol support, supported types: kafka, http, hippo")

	// Select code template.
	createCmd.Flags().String("assetdir", "", "Specify the custom template path, e.g., ~/.trpc-cmdline-assets/protobuf/asset_go")
	createCmd.Flags().StringP("lang", "l", "go",
		"Specify the programming language to use, supported languages: go, cpp")
	createCmd.Flags().Bool("rpconly", false,
		"Only generate stub code (recommended to execute under the stub directory), can be used with -o")
	createCmd.Flags().Bool("dependencystub", false,
		"Whether to generate stub code for dependencies, only effective when --rpconly=true, defaults to false")
	createCmd.Flags().Bool("nogomod", false,
		"Do not generate go.mod file in the stub code, only effective when --rpconly=true, defaults to false")
	createCmd.Flags().Bool("secvenabled", false,
		"Enable generation of validate.go file using protoc-gen-secv, defaults to false")
	createCmd.Flags().Bool("validate", false,
		"Enable generation of validate.pb.go file using protoc-gen-validate, defaults to false")
	createCmd.Flags().String("kvfile", "",
		"Provide a json file path to unmarshal into key-value pairs (KVs) for usage in template files")
	createCmd.Flags().String("kvrawjson", "",
		"Provide raw json content to unmarshal into key-value pairs (KVs) for usage in template files")
	createCmd.Flags().String("app", "",
		"Provide custom app name to use in stub code")
	createCmd.Flags().String("server", "",
		"Provide custom server name to use in stub code")

	// Add functionality similar to "protoc --go_out=. testdesc.proto --descriptor_set_in=testdesc.pb".
	createCmd.Flags().StringP("descriptor_set_in", "", "",
		"Similar to the same flag in protoc, can pass in the parsed descriptor_set to generate the project")

	// Parameters passed to template.
	createCmd.Flags().String("protocol", "trpc",
		"Specify the protocol type to use, supported types: trpc, http, etc")
	createCmd.Flags().String("domain", "",
		"Specify the code address domain to generate, defaults to the value specified in ~/.trpc-cmdline-assets/trpc.yaml")
	createCmd.Flags().String("groupname", "trpc-go",
		"Specify the code address group name to generate, defaults to trpc-go")
	createCmd.Flags().String("versionsuffix", "",
		"Specify the version suffix in the code address, "+
			"defaults to empty, can be set to v2, v3, etc. (effective for both the main library and dependencies)")
	createCmd.Flags().Bool("multi-version", false,
		"Multi-version protocol support, true: supported; false: not supported, "+
			"defaults to not support importing multiple version protocols")
	createCmd.Flags().Bool("noservicesuffix", false,
		"Whether the Service Descriptor naming in the generated Go stub code includes the Service suffix, "+
			"defaults to false")
	addOutputFlags(createCmd)
}

func addOutputFlags(createCmd *cobra.Command) {
	// Controls related to output.
	// Common options.
	createCmd.Flags().StringP("output", "o", "",
		"Specify the output directory (default: output to the directory with the same name as the pb file, "+
			"rpconly defaults to the current directory)")
	createCmd.Flags().BoolP("force", "f", false, "Force overwrite existing code")
	createCmd.Flags().StringP("mod", "m", "", "Specify the go module, default: trpc.app.${pb.package}")
	createCmd.Flags().String("goversion", "1.18", "Specify the Go version in the generated go.mod file, default: 1.18")
	createCmd.Flags().String("trpcgoversion", "",
		"Specify the trpc-go version in the generated go.mod file")
	createCmd.Flags().Bool("mock", true,
		"Generate mock stub code (can be updated by running `go generate` in the project)")

	// Enable rpcname aliases.
	createCmd.Flags().Bool("alias", false, "Use rpcname aliases")
	createCmd.Flags().Bool("alias-as-client-rpcname", true, "Use alias name as client rpcname in stub code")

	// Customize pb message struct tag.
	createCmd.Flags().Bool("gotag", true, "Generate custom pb struct tag")

	// By default, files are split by service.
	// If the following option is enabled, files will be split by RPC.
	createCmd.Flags().BoolP("split-by-method", "s", false, "Whether to split files by method")

	// Whether to output Swagger documentation.
	createCmd.Flags().Bool("swagger", false, "Generate Swagger API documentation")
	createCmd.Flags().String("swagger-out", "apidocs.swagger.json", "Output file path for Swagger API documentation")
	createCmd.Flags().Bool("swagger-json-param", false,
		"Generate Swagger API documentation request parameters using JSON body")

	// Parameters related to git address.
	createCmd.Flags().Bool("sync", false,
		"Whether to sync git repository, default address: specified by go_package, can be specified by --remote")
	createCmd.Flags().String("remote", "", "If sync git repository, specify the git repository address")
	createCmd.Flags().Bool("newtag", false, "Whether to tag the uploaded repository")
	createCmd.Flags().String("tag", "",
		"If tagging the repository, specify the tag name. If not specified, "+
			"tag will be created in the format of v1.1.1, with rule of base 100")
}
