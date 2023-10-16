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
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"trpc.group/trpc-go/trpc-cmdline/util/semver"
)

// IDLType is IDL type.
type IDLType uint

// Related constants of IDL type.
const (
	IDLTypeProtobuf    IDLType = iota // IDL is protobuf
	IDLTypeFlatBuffers                // IDL is flatbuffers
	IDLTypeInvalid
)

var idlTypeDesc = map[IDLType]string{
	IDLTypeProtobuf:    "protobuf",
	IDLTypeFlatBuffers: "flatbuffers",
}

// Valid Check if the IDL type is valid.
func (t IDLType) Valid() bool {
	return t < IDLTypeInvalid
}

// String Return the description of the IDL type.
func (t IDLType) String() string {
	if !t.Valid() {
		return fmt.Sprintf("unknown type: %d", t)
	}
	return idlTypeDesc[t]
}

// Template is language templates.
type Template struct {
	Language    string `yaml:"language"`      // Programming language
	LangFileExt string `yaml:"lang_file_ext"` // Source file extension for the programming language
	// The output filename for the stub corresponds to: service name + separator + original file name
	Separator string `yaml:"separator"`
	// Keep the original file name as a suffix for the generated files
	KeepOrigName bool `yaml:"keep_orig_name"`
	// Follow the camelcase style for generated file names
	CamelCaseName     bool     `yaml:"camelcase_name"`
	AssetDir          string   `yaml:"asset_dir"`            // Code template directory
	RPCServerStub     string   `yaml:"rpc_server_stub"`      // Server stub
	RPCServerTestStub string   `yaml:"rpc_server_test_stub"` // Server test stub
	RPCClientStub     []string `yaml:"rpc_client_stub"`      // Client stub
	// Whether the client stub includes all service definitions
	RPCClientStubPerService bool `yaml:"rpc_client_stub_per_service"`
}

// OpSys is the system of operation (运营体系).
type OpSys struct {
	Name    string   `yaml:"name"`
	Imports []string `yaml:"imports"`
}

// Config is the global config.
type Config struct {
	Domain     string                          `yaml:"domain"` // host in importPath
	TplFileExt string                          `yaml:"tpl_file_ext"`
	IDL        map[string]*Dependency          `yaml:"idl"`       // IDL name -> IDL tool
	Tools      map[string][]*Dependency        `yaml:"tools"`     // Programming language -> Dependency tools
	Plugins    map[string][]string             `yaml:"plugins"`   // Programming language -> Dependency plugins
	Templates  map[string]map[string]*Template `yaml:"templates"` // idltype -> Code templates for each language
}

// Dependency is the description of dependencies.
type Dependency struct {
	Executable  string `yaml:"executable"`   // Executable file name.
	VersionMin  string `yaml:"version_min"`  // Min version.
	VersionCmd  string `yaml:"version_cmd"`  // Max version.
	ArtifactURL string `yaml:"artifact_url"` // Artifact download URL.
	Repository  string `yaml:"repository"`   // Repository URL.
	MD5         string `yaml:"md5"`          // md5 sum up.
	Fallback    string `yaml:"fallback"`     // Failure prompt.
}

// TryInstallTo tries to install the dependency to the given path.
func (d *Dependency) TryInstallTo(path string) error {
	rsp, err := http.Get(d.ArtifactURL)
	if err != nil {
		return fmt.Errorf("install %s from %s err: %w", d.Executable, d.ArtifactURL, err)
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(rsp.Body)
		return fmt.Errorf("download from %s, rsp body: %q, body read err: %w", d.ArtifactURL, body, err)
	}
	dst := filepath.Join(path, d.Executable)
	bs, err := io.ReadAll(rsp.Body)
	if err != nil {
		return fmt.Errorf("read from %s err: %w", d.ArtifactURL, err)
	}
	if err := os.WriteFile(dst, bs, os.ModePerm); err != nil {
		return fmt.Errorf("write to %s err: %w", dst, err)
	}
	return nil
}

// Installed checks installed or not.
func (d *Dependency) Installed() bool {
	if _, err := exec.LookPath(d.Executable); err != nil {
		return false
	}
	return true
}

// version returns installed version.
func (d *Dependency) version() (string, error) {
	if d.VersionCmd == "" {
		return "", nil
	}

	cmd := exec.Command("sh", "-c", d.VersionCmd)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", d.VersionCmd)
	}

	buf, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("read version failed: %w, command: %s", err, d.VersionCmd)
	}

	s := string(buf)
	strs := versionRE.FindStringSubmatch(s)
	if len(strs) < 2 {
		return "", fmt.Errorf("version %s does not match pattern %s", s, versionPattern)
	}
	version := strings.Join(strs[1:], ".")
	return version, nil
}

// CheckVersion check if installed version meet the version requirement.
func (d *Dependency) CheckVersion() (passed bool, error error) {
	// skip checking if cmd/version not specified
	if d.VersionMin == "" || d.VersionCmd == "" {
		return true, nil
	}

	// load version
	v, err := d.version()
	if err != nil {
		return false, err
	}

	// compare version and required version
	version := v
	required := d.VersionMin

	version = versionNumber(version)
	required = versionNumber(required)
	return semver.NewerThan(version, required), nil
}

func versionNumber(v string) string {
	if len(v) != 0 && (v[0] == 'v' || v[0] == 'V') {
		return v[1:]
	}
	return v
}
