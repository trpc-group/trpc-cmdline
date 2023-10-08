package plugin

import (
	"fmt"
	"os/exec"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

// GoImports is goimports plugin.
type GoImports struct{}

// Name return plugin's name.
func (p *GoImports) Name() string {
	return "goimports"
}

// Check only run when `--lang=go && --goimports=true`
func (p *GoImports) Check(_ *descriptor.FileDescriptor, opt *params.Option) bool {
	if opt.Language == "go" {
		return true
	}
	return false
}

// Run runs goimports action.
func (p *GoImports) Run(_ *descriptor.FileDescriptor, _ *params.Option) error {
	goimports, err := exec.LookPath("goimports")
	if err != nil {
		return fmt.Errorf("goimports not found, install it first")
	}

	// Under some rare circumstances, we need run goimports multiple times to
	// prevent duplicate imports.
	const maxGoImports = 5
	for i := 0; i < maxGoImports; i++ {
		buf, err := exec.Command(goimports, "-w", ".").CombinedOutput()
		if err != nil {
			log.Error("run goimports -w . error: %+v,\n%s", err, string(buf))
			return err
		}
		buf, err = exec.Command(goimports, "-d", ".").CombinedOutput()
		if err != nil {
			log.Error("run goimports -d . error: %+v,\n%s", err, string(buf))
			return err
		}
		if len(buf) == 0 {
			break
		}
	}

	return nil
}
