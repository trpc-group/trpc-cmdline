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
