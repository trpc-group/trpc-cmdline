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
