package plugin

import (
	"fmt"

	"trpc.group/trpc-go/trpc-cmdline/util/apidocs/openapi"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
)

// OpenAPI is swagger plugin.
type OpenAPI struct {
}

// Name return plugin's name.
func (p *OpenAPI) Name() string {
	return "openapi"
}

// Check only run when `--openapi=true`
func (p *OpenAPI) Check(fd *descriptor.FileDescriptor, opt *params.Option) bool {
	if opt.OpenAPIOn {
		return true
	}
	return false
}

// Run runs openapi action to generate openapi apidocs
func (p *OpenAPI) Run(fd *descriptor.FileDescriptor, opt *params.Option) error {
	if err := openapi.GenOpenAPI(fd, opt); err != nil {
		return fmt.Errorf("create open api document error: %v", err)
	}
	return nil
}
