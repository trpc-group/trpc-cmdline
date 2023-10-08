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
