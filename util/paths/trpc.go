// Package paths provides functionality related to file paths within the project.
package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
)

// Locate locates the directory of file name. It first searches in dirs and then in the template installation directory.
// name can be a regular file or directory.
func Locate(name string, search ...string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	trpcInstallUserPath := filepath.Join(home, ".trpc-cmdline-assets")
	search = append(search,
		ExpandTRPCSearch(trpcInstallUserPath)...,
	)

	for _, p := range search {
		absolutePath, e := filepath.Abs(p)
		if e != nil {
			err = multierror.Append(err, e).ErrorOrNil()
			continue
		}

		filePath := filepath.Join(absolutePath, name)
		if _, e := os.Lstat(filePath); e != nil {
			err = multierror.Append(err, e).ErrorOrNil()
			continue
		}
		return absolutePath, nil
	}
	return "", fmt.Errorf("%s not found in %s, err: %w", name, strings.Join(search, ","), err)
}

// ExpandTRPCSearch expands search path around installation path (typically ~/.trpc-cmdline-assets).
func ExpandTRPCSearch(installPath string) []string {
	return []string{
		installPath,
		filepath.Join(installPath, "submodules"),
		filepath.Join(installPath, "submodules", "trpc-protocol"),
		filepath.Join(installPath, "submodules", "trpc"),
		filepath.Join(installPath, "submodules", "protoc-gen-secv"),
		filepath.Join(installPath, "submodules", "protoc-gen-secv", "validate"),
		filepath.Join(installPath, "trpc"),
		filepath.Join(installPath, "protos"),
		filepath.Join(installPath, "protos", "trpc"),
	}
}

// ExpandSearch expands search path around trpc_options.proto path.
func ExpandSearch(protoTRPCPath string) []string { // .trpc-cmdline-assets/submodules/trpc-protocol
	parent := filepath.Dir(protoTRPCPath) // .trpc-cmdline-assets/submodules
	grandParent := filepath.Dir(parent)   // .trpc-cmdline-assets/
	return []string{
		parent,
		grandParent,
		filepath.Join(grandParent, "protos"),
		protoTRPCPath,
	}
}
