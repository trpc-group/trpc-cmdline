package parser

import (
	annotations "trpc.group/trpc/trpc-protocol/pb/go/trpc/api"
)

// HttpRule provide interface for http rule.
type HttpRule[HR any] interface {
	*annotations.HttpRule
	GetAdditionalBindings() []HR
	GetBody() string
	GetResponseBody() string
}
