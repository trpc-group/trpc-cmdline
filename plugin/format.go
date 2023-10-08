package plugin

import (
	"os"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/util/style"
)

// Formatter is a plugin used to format generated code.
type Formatter struct {
}

// Name return plugin's name.
func (p *Formatter) Name() string {
	return "gofmt"
}

// Check don't run if `--lang != go`
func (p *Formatter) Check(fd *descriptor.FileDescriptor, opt *params.Option) bool {
	if opt.Language != "go" {
		return false
	}
	return true
}

// Run runs gofmt action.
func (p *Formatter) Run(fd *descriptor.FileDescriptor, opt *params.Option) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	return style.GoFmtDir(dir)
}
