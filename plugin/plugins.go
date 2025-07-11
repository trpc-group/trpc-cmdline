// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

//go:build !experimental
// +build !experimental

package plugin

import "trpc.group/trpc-go/trpc-cmdline/plugin/sync"

// Plugins is the chain of public plugins.
var Plugins = []Plugin{
	&Swagger{},  // swagger apidoc
	&OpenAPI{},  // openapi apidoc
	&Validate{}, // protoc-gen-secv
	sync.NewGit(sync.DefaultFileManager, sync.DefaultGitManager,
		sync.AuthSupplier), // sync stub to git repository
}

// PluginsExt is the language-specific plugin chain.
var PluginsExt = map[string][]Plugin{
	"go": {
		&GoImports{}, // goimports,  runs before mockgen, to eliminate `package import but unused` errors
		&Formatter{}, // gofmt
		&GoMock{},    // gomock
		&GoTag{},     // custom go tag by proto field options
	},
	"cpp": {
		&CppMove{}, // Move the PB files to proto directory.
	},
}
