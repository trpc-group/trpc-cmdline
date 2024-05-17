// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package config

import (
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

var (
	versionPattern = `[v]?(?:(0|[1-9]\d*)\.)?(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-.*)?`
	versionRE      = regexp.MustCompile(versionPattern)
)

// LoadDependencies loads the configuration on demand based on the given language type.
func LoadDependencies(languages ...string) ([]*Dependency, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	var deps []*Dependency
	depsSet := make(map[string]*Dependency)

	deps = append(deps, cfg.IDL[IDLTypeProtobuf.String()])
	deps = append(deps, cfg.IDL[IDLTypeFlatBuffers.String()])
	if len(languages) == 0 {
		// Default configures the tools required by all languages.
		deps = append(deps, cfg.Tools["go"]...)
	} else {
		// Configure according to the specified language.
		for _, l := range languages {
			v, ok := cfg.Tools[l]
			if !ok {
				continue
			}
			deps = append(deps, v...)
		}
	}
	for _, v := range deps {
		if v == nil {
			continue
		}
		depsSet[v.Executable] = v
	}

	deps = []*Dependency{}
	for _, v := range depsSet {
		if runtime.GOOS == "windows" {
			v.ArtifactURL += ".exe"
			v.Executable += ".exe"
		}
		deps = append(deps, v)
	}
	return deps, nil
}

// SetupDependencies configures dependency installation.
func SetupDependencies(deps []*Dependency) error {
	path, err := getCandidate()
	if err != nil {
		return err
	}
	for _, dep := range deps {
		// Check whether installed, if no, try to install it.
		if dep.Installed() {
			ok, err := dep.CheckVersion()
			if err != nil {
				return fmt.Errorf("%s installed, check version, err: %w", dep.Executable, err)
			}
			if ok {
				log.Debug("%s check passed", dep.Executable)
				continue
			}
			log.Debug("%s installed, check version, too old, need reinstall", dep.Executable)
		}
		log.Debug("%s not installed, need install", dep.Executable)
		// Try to install if needed.
		if err := dep.TryInstallTo(path); err != nil {
			return fmt.Errorf("install %s to %s failed: %w", dep.Executable, path, err)
		}
		if !dep.Installed() {
			log.Error("%s is installed to %s, but it cannot be found, please ensure it is added to your $PATH variable",
				dep.Executable, path)
		} else {
			log.Info("%s is installed to %s", dep.Executable, path)
		}
		ok, err := dep.CheckVersion()
		if err != nil {
			return fmt.Errorf("%s is installed, check version err: %w", dep.Executable, err)
		}
		if !ok {
			return fmt.Errorf("%s is still too old, check $PATH, maybe there're several existed", dep.Executable)
		}
	}
	return nil
}

func getCandidate() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("install dependency error: %w", err)
	}
	cmd := exec.Command("go", "env", "GOPATH")
	output, _ := cmd.Output()
	gopath := strings.TrimSpace(string(output))
	candidates := []string{
		filepath.Join(gopath, "bin"),
		filepath.Join(build.Default.GOPATH, "bin"),
		filepath.Join(home, "go", "bin"),
		filepath.Join(home, "bin"),
	}
	dest := candidates[0]
	path := os.ExpandEnv(os.Getenv("PATH"))
	var found bool
	for _, p := range candidates {
		// Must be a directory to avoid interference from files with the same name.
		fin, err := os.Lstat(p)
		if err != nil {
			continue
		}
		if !fin.IsDir() {
			continue
		}

		// Should be searchable in `PATH`.
		if !strings.Contains(path, p) {
			continue
		}
		dest = p
		found = true
		break
	}
	if !found {
		return "", fmt.Errorf("get install path error: %s does not exist or cannot be found under $PATH, "+
			"consider creating it manually and adding to $PATH", dest)
	}
	return dest, nil
}
