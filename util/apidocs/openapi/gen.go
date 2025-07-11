// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package openapi provides the ability to manipulate OpenAPI documents.
package openapi

import (
	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/util/apidocs"
)

// GenOpenAPI provides external structure used to generate openapi json.
func GenOpenAPI(fd *descriptor.FileDescriptor, option *params.Option) error {
	// Assemble the entire JSON information.
	openapi, err := apidocs.NewOpenAPIJSON(fd, option)
	if err != nil {
		return err
	}

	return apidocs.WriteJSON(option.OpenAPIOut, openapi)
}
