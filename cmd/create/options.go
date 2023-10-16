// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package create

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	"trpc.group/trpc-go/trpc-cmdline/config"
	"trpc.group/trpc-go/trpc-cmdline/plugin"
	"trpc.group/trpc-go/trpc-cmdline/util/fs"
	"trpc.group/trpc-go/trpc-cmdline/util/lang"
	"trpc.group/trpc-go/trpc-cmdline/util/paths"
	"trpc.group/trpc-go/trpc-cmdline/util/pb"
)

// loadOptions loads options from flags.
func (c *Create) loadOptions(flags *pflag.FlagSet) error {
	if err := c.parse(flags); err != nil {
		return fmt.Errorf("load options err: %w", err)
	}
	if err := c.fixOptions(); err != nil {
		return fmt.Errorf("fix create options inside prerun err: %w", err)
	}
	return nil
}

// parse verifies and parses the flags.
func (c *Create) parse(flags *pflag.FlagSet) error {
	if err := c.parseIDLOptions(flags); err != nil {
		return fmt.Errorf("parse idl options err: %w", err)
	}

	if err := c.parseGeneralOptions(flags); err != nil {
		return fmt.Errorf("parse general options err: %w", err)
	}

	if err := c.parseSwaggerOptions(flags); err != nil {
		return fmt.Errorf("parse swagger options err: %w", err)
	}

	if err := c.parseProtocolOptions(flags); err != nil {
		return fmt.Errorf("parse protocol options err: %w", err)
	}

	if err := c.parseAuxiliaryOptions(flags); err != nil {
		return fmt.Errorf("parse auxiliary options err: %w", err)
	}

	if err := c.parseSyncGitOptions(flags); err != nil {
		return fmt.Errorf("parse sync git options err: %w", err)
	}
	return nil
}

// fixOptions fixes the options.
func (c *Create) fixOptions() error {
	c.fixGoMod()
	if err := c.fixIDL(); err != nil {
		return fmt.Errorf("fix idl err: %w", err)
	}
	return nil
}

// fixGoMod fixes the module name.
func (c *Create) fixGoMod() {
	//  1. Use module name specified by -mod.
	//  2. Use local go.mod if -mod is not specified (for backward compatibility).
	//  3. Use package name defined in pb (implemented by template).
	if c.options.GoMod != "" {
		return
	}
	mod, err := lang.LoadGoMod()
	if err != nil {
		return
	}
	if mod == "" {
		return
	}
	c.options.GoModEx = mod
	c.options.GoMod = mod
	return
}

// fixIDL fixes the IDL.
func (c *Create) fixIDL() error {
	if c.options.OtherType != "" {
		return c.fixOtherType() // Non-IDL type, such as kafka, HTTP.
	}
	if c.options.Protofile == "" {
		return errors.New("protobuf/flatbuffers file both empty")
	}
	if err := c.fixProtoDirs(); err != nil {
		return fmt.Errorf("fix proto dirs err: %w", err)
	}
	return c.fixProtocolType()
}

// fixOtherType updates the options related to "OtherType".
func (c *Create) fixOtherType() error {
	installPath, err := config.CurrentTemplatePath()
	if err != nil {
		return fmt.Errorf("failed to get current template path for other type err: %w", err)
	}

	if c.options.Assetdir == "" { // Do not override options provided by the user.
		c.options.Assetdir = filepath.Join(installPath, "without_idl", c.options.Language, c.options.OtherType)
	}

	c.options.Protocol = c.options.OtherType
	// May consider using c.options.OutputDir here.

	c.options.Protofile = c.options.OtherType

	// Plugins for code generation.
	plugin.Plugins = []plugin.Plugin{
		&plugin.GoImports{}, // goimports, runs before mockgen, to eliminate `package import but unused` errors
		&plugin.Formatter{}, // gofmt
	}

	plugin.PluginsExt[c.options.Language] = nil

	return nil
}

// fixProtoDirs updates the options related to proto directories.
// If DescriptorSetIn is provided, it locates the file and updates the file path.
// Otherwise, it locates the proto file and updates the options accordingly.
func (c *Create) fixProtoDirs() error {
	if c.options.DescriptorSetIn != "" {
		var err error
		filePath, err := fs.LocateFile(c.options.DescriptorSetIn, nil)
		if err != nil {
			return fmt.Errorf("fs locate file %s err: %w", c.options.DescriptorSetIn, err)
		}
		c.options.DescriptorSetIn = filePath
		return nil
	}

	p, err := paths.Locate(pb.ProtoTRPC)
	if err != nil {
		return fmt.Errorf("paths locate %s failed err: %w", pb.ProtoTRPC, err)
	}

	c.options.Protodirs = fs.UniqFilePath(append(append(c.options.Protodirs, p),
		paths.ExpandSearch(p)...,
	))

	target, err := fs.LocateFile(c.options.Protofile, c.options.Protodirs)
	if err != nil {
		return fmt.Errorf("locate file in proto dirs failed err: %w", err)
	}

	if c.options.UseBaseName {
		c.options.Protofile = filepath.Base(target)
	}

	c.options.ProtofileAbs = target
	c.options.Protodirs = append(c.options.Protodirs, filepath.Dir(target))

	return nil
}

// fixProtocolType updates the options related to the protocol type.
// It loads configurations from trpc.yaml and updates the options accordingly.
func (c *Create) fixProtocolType() error {
	// Load configurations from trpc.yaml.
	cfg, err := config.GetTemplate(c.options.IDLType, c.options.Language)
	if err != nil {
		return fmt.Errorf("config get template failed err: %w", err)
	}
	if c.options.Assetdir == "" {
		c.options.Assetdir = cfg.AssetDir
	}
	if c.options.Domain == "" {
		c.options.Domain = config.GlobalConfig().Domain
	}
	if c.options.VersionSuffix != "" {
		c.options.VersionSuffix = "/" + c.options.VersionSuffix
	}
	return nil
}

// parseIDLOptions parses the IDL-related options from the command line flags.
// It parses the "usebasename" flag and delegates to other functions to parse protobuf/flatbuffers options.
func (c *Create) parseIDLOptions(flags *pflag.FlagSet) error {
	var err error
	c.options.UseBaseName, err = flags.GetBool("usebasename")
	if err != nil {
		return fmt.Errorf("flags parse usebasename %w", err)
	}
	// Parse protobuf/flatbuffers options.
	if err := c.parsePBIDLOptions(flags); err != nil {
		return fmt.Errorf("flags parse pb idl options err: %w", err)
	}
	// If protofile field is empty, try parse flatbuffers related flags.
	if c.options.Protofile == "" {
		if err := c.parseFBIDLOptions(flags); err != nil {
			return fmt.Errorf("flags parse fb idl options, err: %w", err)
		}
	}
	return nil
}

// parseGeneralOptions parses the general options from the command line flags.
// It parses the "verbose" flag and delegates to other functions to parse input/output options.
func (c *Create) parseGeneralOptions(flags *pflag.FlagSet) error {
	var err error
	c.options.Verbose, err = flags.GetBool("verbose")
	if err != nil {
		return fmt.Errorf("flags parse verbose string err: %w", err)
	}
	if err := c.parseInputOptions(flags); err != nil {
		return err
	}
	if err := c.parseOutputOptions(flags); err != nil {
		return err
	}
	return nil
}

// parseAuxiliaryOptions parses the auxiliary options from the command line flags.
// It parses various boolean and string flags related to auxiliary options.
func (c *Create) parseAuxiliaryOptions(flags *pflag.FlagSet) error {
	var err error
	c.options.MultiVersion, err = flags.GetBool("multi-version")
	if err != nil {
		return fmt.Errorf("flags parse multi-version bool err: %w", err)
	}
	c.options.NoServiceSuffix, err = flags.GetBool("noservicesuffix")
	if err != nil {
		return fmt.Errorf("flags parse noservicesuffix bool err: %w", err)
	}
	return nil
}

// parseInputOptions parses the input options from the command line flags.
// It parses various string and boolean flags related to input options.
func (c *Create) parseInputOptions(flags *pflag.FlagSet) error {
	var err error
	c.options.Assetdir, err = flags.GetString("assetdir")
	if err != nil {
		return fmt.Errorf("flags parse assetdir string err: %w", err)
	}
	c.options.Language, err = flags.GetString("lang")
	if err != nil {
		return fmt.Errorf("flags parse lang string err: %w", err)
	}
	c.options.AliasOn, err = flags.GetBool("alias")
	if err != nil {
		return fmt.Errorf("flags parse alias bool err: %w", err)
	}
	c.options.AliasAsClientRPCName, err = flags.GetBool("alias-as-client-rpcname")
	if err != nil {
		return fmt.Errorf("flags parse alias-as-client-rpcname bool err: %w", err)
	}
	c.options.GoMod, err = flags.GetString("mod")
	if err != nil {
		return fmt.Errorf("flags parse mod string err: %w", err)
	}
	c.options.GoVersion, err = flags.GetString("goversion")
	if err != nil {
		return fmt.Errorf("flags parse goversion string err: %w", err)
	}
	c.options.TRPCGoVersion, err = flags.GetString("trpcgoversion")
	if err != nil {
		return fmt.Errorf("flags parse trpcgoversion string err: %w", err)
	}
	c.options.CustomAPPName, err = flags.GetString("app")
	if err != nil {
		return fmt.Errorf("flags parse app string err: %w", err)
	}
	c.options.CustomServerName, err = flags.GetString("server")
	if err != nil {
		return fmt.Errorf("flags parse server string err: %w", err)
	}
	c.options.DescriptorSetIn, err = flags.GetString("descriptor_set_in")
	if err != nil {
		return fmt.Errorf("flags parse descriptor_set_in string err: %w", err)
	}
	return nil
}

// parseOutputOptions parses the output options from the command line flags.
// It parses various string and boolean flags related to output options.
func (c *Create) parseOutputOptions(flags *pflag.FlagSet) error {
	var err error
	c.options.OutputDir, err = flags.GetString("output")
	if err != nil {
		return fmt.Errorf("flags parse output string err: %w", err)
	}
	c.options.RPCOnly, err = flags.GetBool("rpconly")
	if err != nil {
		return fmt.Errorf("flags parse rpconly bool err: %w", err)
	}
	c.options.DependencyStub, err = flags.GetBool("dependencystub")
	if err != nil {
		return fmt.Errorf("flags parse dependencystub %w", err)
	}
	c.options.NoGoMod, err = flags.GetBool("nogomod")
	if err != nil {
		return fmt.Errorf("flags parse nogomod bool err: %w", err)
	}
	c.options.KeepOrigRPCName = true // Always true.
	c.options.SecvEnabled, err = flags.GetBool("secvenabled")
	if err != nil {
		return fmt.Errorf("flags parse secvenabled bool err: %w", err)
	}
	kvFile, err := flags.GetString("kvfile")
	if err != nil {
		return fmt.Errorf("flags parse kvfile string err: %w", err)
	}
	if kvFile != "" {
		bs, err := os.ReadFile(kvFile)
		if err != nil {
			return fmt.Errorf("read kv file %s err: %w", kvFile, err)
		}
		if err := json.Unmarshal(bs, &c.options.KVs); err != nil {
			return fmt.Errorf("json unmarshal kv file %s into %T err: %w", kvFile, c.options.KVs, err)
		}
	}
	kvRawJSON, err := flags.GetString("kvrawjson")
	if err != nil {
		return fmt.Errorf("flags parse kvrawjson string err: %w", err)
	}
	if kvRawJSON != "" {
		if err := json.Unmarshal([]byte(kvRawJSON), &c.options.KVs); err != nil {
			return fmt.Errorf("json unmarshal kv raw json %s into %T err: %w", kvRawJSON, c.options.KVs, err)
		}
	}
	c.options.Force, err = flags.GetBool("force")
	if err != nil {
		return fmt.Errorf("flags parse force bool err: %w", err)
	}
	c.options.Mockgen, err = flags.GetBool("mock")
	if err != nil {
		return fmt.Errorf("flags parse mock bool err: %w", err)
	}
	c.options.PerMethod, err = flags.GetBool("split-by-method")
	if err != nil {
		return fmt.Errorf("flags parse split-by-method bool err: %w", err)
	}
	c.options.Domain, err = flags.GetString("domain")
	if err != nil {
		return fmt.Errorf("flags parse domain string err: %w", err)
	}
	c.options.GroupName, err = flags.GetString("groupname")
	if err != nil {
		return fmt.Errorf("flags parse groupname string err: %w", err)
	}
	c.options.VersionSuffix, err = flags.GetString("versionsuffix")
	if err != nil {
		return fmt.Errorf("flags parse versionsuffix string err: %w", err)
	}
	return nil
}

// parseSwaggerOptions parses the swagger options from the command line flags.
// It parses various string and boolean flags related to swagger options.
func (c *Create) parseSwaggerOptions(flags *pflag.FlagSet) error {
	var err error
	c.options.SwaggerOn, err = flags.GetBool("swagger")
	if err != nil {
		return fmt.Errorf("flags parse swagger bool err: %w", err)
	}
	c.options.SwaggerOut, err = flags.GetString("swagger-out")
	if err != nil {
		return fmt.Errorf("flags parse swagger-out string err: %w", err)
	}
	c.options.SwaggerOptJSONParam, err = flags.GetBool("swagger-json-param")
	if err != nil {
		return fmt.Errorf("flags parse swagger-json-param bool err: %w", err)
	}
	return nil
}

// parseProtocolOptions parses the protocol options from the command line flags.
// It parses various string flags related to protocol options.
func (c *Create) parseProtocolOptions(flags *pflag.FlagSet) error {
	var err error
	c.options.Protocol, err = flags.GetString("protocol")
	if err != nil {
		return fmt.Errorf("flags parse protocol string err: %w", err)
	}
	c.options.OtherType, err = flags.GetString("non-protocol-type")
	if err != nil {
		return fmt.Errorf("flags parse non-protocol-type string err: %w", err)
	}
	return nil
}

// parsePBIDLOptions parses the protobuf IDL options from the command line flags.
// It parses various string flags and an array of strings related to protobuf options.
func (c *Create) parsePBIDLOptions(flags *pflag.FlagSet) error {
	dirs, err := flags.GetStringArray("protodir")
	if err != nil {
		return fmt.Errorf("flags get protodir string array failed err: %w", err)
	}
	// Always append the current working directory.
	c.options.Protodirs = fs.UniqFilePath(append(dirs, "."))
	c.options.Protofile, err = flags.GetString("protofile")
	if err != nil {
		return fmt.Errorf("flags get protofile string failed err: %w", err)
	}
	c.options.Gotag, err = flags.GetBool("gotag")
	if err != nil {
		return fmt.Errorf("flags get gotag bool failed err: %w", err)
	}
	c.options.IDLType = config.IDLTypeProtobuf
	return nil
}

// parseFBIDLOptions parses the FlatBuffers IDL options from the command line flags.
// It parses various string flags and an array of strings related to FlatBuffers options.
func (c *Create) parseFBIDLOptions(flags *pflag.FlagSet) error {
	dirs, err := flags.GetStringArray("fbsdir")
	if err != nil {
		return fmt.Errorf("flags get fbsdir string array failed err: %w", err)
	}
	// Always append the current working directory.
	c.options.Protodirs = fs.UniqFilePath(append(dirs, "."))
	c.options.Protofile, err = flags.GetString("fbs")
	if err != nil {
		return fmt.Errorf("flags get fbs string failed err: %w", err)
	}
	c.options.IDLType = config.IDLTypeFlatBuffers
	return nil
}

// parseSyncGitOptions parses the synchronization and git options from the command line flags.
// It parses various string and boolean flags related to git synchronization options.
func (c *Create) parseSyncGitOptions(flags *pflag.FlagSet) error {
	sync, err := flags.GetBool("sync")
	if err != nil {
		return fmt.Errorf("flags get git sync bool failed err: %w", err)
	}
	c.options.Sync = sync
	remote, err := flags.GetString("remote")
	if err != nil {
		return fmt.Errorf("flags get git remote address url failed err: %w", err)
	}
	c.options.Remote = remote
	newTag, err := flags.GetBool("newtag")
	if err != nil {
		return fmt.Errorf("flags get git new tag bool failed err: %w", err)
	}
	c.options.NewTag = newTag
	tag, err := flags.GetString("tag")
	if err != nil {
		return fmt.Errorf("flags get git tag failed err: %w", err)
	}
	c.options.Tag = tag
	return nil
}
