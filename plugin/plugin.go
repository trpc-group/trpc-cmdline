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
