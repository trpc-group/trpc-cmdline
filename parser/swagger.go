// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package parser

import (
	"trpc.group/trpc/trpc-protocol/pb/go/trpc/swagger"
)

type SwaggerRule[SP SwaggerParam] interface {
	*swagger.SwaggerRule
	GetTitle() string
	GetDescription() string
	GetMethod() string
	GetParams() []SP
}

type SwaggerParam interface {
	*swagger.SwaggerParam
	GetName() string
	GetRequired() bool
	GetDefault() string
}
