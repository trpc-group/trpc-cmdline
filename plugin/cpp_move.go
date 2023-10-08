package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/util/fs"
	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

// CppMove is the plugin for moving proto files to proto directory in generated project.
type CppMove struct {
}

// Name returns the plugin name.
func (p *CppMove) Name() string {
	return "cpp_move"
}

// Check only run when `--lang=cpp`
func (p *CppMove) Check(fd *descriptor.FileDescriptor, opt *params.Option) bool {
	return opt.Language == "cpp" && !opt.RPCOnly
}

// Run runs moving directories action.
func (p *CppMove) Run(fd *descriptor.FileDescriptor, opt *params.Option) error {
	log.Debug("execute plugin for %s in %s", opt.Language, opt.OutputDir)
	pbOutDir := filepath.Join(opt.OutputDir, "proto")
	// moving proto files to proto directory in generated project.
	for pbFile := range fd.Pb2DepsPbs {
		pbAbsPath, err := fs.LocateFile(pbFile, opt.Protodirs)
		if err != nil {
			fmt.Println("fs.LocateFile err: %w", err)
			return nil
		}
		err = fs.Copy(pbAbsPath, filepath.Join(pbOutDir, pbFile))
		if err != nil {
			fmt.Println("file copy err: %w", err)
			return err
		}
	}
	// add executing mode for script
	scriptLs := []string{"build.sh", "clean.sh", "run_client.sh", "run_server.sh"}
	for _, script := range scriptLs {
		err := os.Chmod(filepath.Join(opt.OutputDir, script), 0755)
		if err != nil {
			fmt.Println("chmod failed err: %w", err)
			return err
		}
	}
	return nil
}
