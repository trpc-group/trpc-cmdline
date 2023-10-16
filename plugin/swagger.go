// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package plugin

import (
	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/util/apidocs/swagger"
)

// Swagger is swagger plugin.
type Swagger struct {
}

// Name return plugin's name.
func (p *Swagger) Name() string {
	return "swagger"
}

// Check run only when `--swagger=true`
func (p *Swagger) Check(fd *descriptor.FileDescriptor, opt *params.Option) bool {
	return opt.SwaggerOn
}

// Run run swagger plugin to generate swagger apidocs
func (p *Swagger) Run(fd *descriptor.FileDescriptor, opt *params.Option) error {
	return swagger.GenSwagger(fd, opt)
}
