// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package plugin provides the ability to implement extension functionality using plugins.
package plugin

import (
	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
)

// Plugin represents some customized operation.
type Plugin interface {
	// Name return plugin's name.
	Name() string

	// Check return whether this plugin should be run.
	Check(fd *descriptor.FileDescriptor, opt *params.Option) bool

	// Run runs plugin.
	Run(fd *descriptor.FileDescriptor, opt *params.Option) error
}
