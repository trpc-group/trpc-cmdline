// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package plugin

import (
	"os"
	"path/filepath"

	"trpc.group/trpc-go/trpc-cmdline/config"
	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/parser"
	"trpc.group/trpc-go/trpc-cmdline/util/pb"
)

// Validate is validate plugin.
type Validate struct {
}

// Name return plugin's name.
func (p *Validate) Name() string {
	return "validate"
}

// Check only run when language is supported.
func (p *Validate) Check(fd *descriptor.FileDescriptor, opt *params.Option) bool {
	if !opt.SecvEnabled {
		return false
	}
	support, ok := config.SupportValidate[opt.Language]
	if !ok || !support {
		return false
	}

	// Check if validation feature is enabled in pb and generate go files.
	return parser.CheckSECVEnabled(fd)
}

// Run runs protoc-gen-validate to generate validate.pb.go
//
// Only supports a few programming languages. See: https://trpc.group/devsec/protoc-gen-secv
func (p *Validate) Run(fd *descriptor.FileDescriptor, opt *params.Option) error {

	var (
		pbOutDir string
		err      error
	)

	outputdir := opt.OutputDir

	if !opt.RPCOnly {
		stubDir := filepath.Join(outputdir, "stub")

		pbPackage, err := parser.GetPackage(fd, opt.Language)
		if err != nil {
			return err
		}
		pbOutDir = filepath.Join(stubDir, pbPackage)
		os.MkdirAll(pbOutDir, os.ModePerm)
	}

	opts := []pb.Option{
		pb.WithPb2ImportPath(fd.Pb2ImportPath),
		pb.WithPkg2ImportPath(fd.Pkg2ImportPath),
		pb.WithSecvEnabled(true),
	}
	// Generate ${protofile}.pb.validate.go
	if !opt.RPCOnly {
		err = pb.Protoc(opt.Protodirs, opt.Protofile, opt.Language, pbOutDir, opts...)
	} else {
		err = pb.Protoc(opt.Protodirs, opt.Protofile, opt.Language, outputdir, opts...)
	}
	return err
}
