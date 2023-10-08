// Package swagger provides the ability to manipulate swagger documentation.
package swagger

import (
	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/util/apidocs"
)

// GenSwagger provides an external structure used to generate swagger JSON.
func GenSwagger(fd *descriptor.FileDescriptor, option *params.Option) error {
	// Assemble the entire Swagger JSON information.
	swaggerJSON, err := apidocs.NewSwagger(fd, option)
	if err != nil {
		return err
	}

	return apidocs.WriteJSON(option.SwaggerOut, swaggerJSON)
}
