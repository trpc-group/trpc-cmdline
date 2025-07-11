// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package create

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"trpc.group/trpc-go/trpc-cmdline/parser"
	"trpc.group/trpc-go/trpc-cmdline/tpl"
	"trpc.group/trpc-go/trpc-cmdline/util/fs"
)

func (c *Create) createRPCOnlyStub() error {
	fd, options, outputDir := c.fileDescriptor, c.options, c.options.OutputDir
	// In case where the current output directory is not empty,
	// generate all the files inside a temporary directory.
	// Then move them into the expected output directory.
	tmpDir := filepath.Join(outputDir, fmt.Sprintf("tmp-%d", time.Now().UnixNano()))
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		return fmt.Errorf("os.MkdirAll inside create rpc only stub %w", err)
	}
	defer os.RemoveAll(tmpDir)
	// The format of pkg is like: "trpc.group/testapp/testserver".
	pkg, err := parser.GetPackage(fd, options.Language)
	if err != nil {
		return fmt.Errorf("parser get package inside create rpc stub err: %w", err)
	}
	// Traverse each file in install/asset_${lang}.
	if len(fd.Services) != 0 {
		if err := tpl.GenerateFiles(fd, tmpDir, options); err != nil {
			return fmt.Errorf("generate files from template inside create rpc stub err: %w", err)
		}
	}

	// Generate IDL stub code for protobuf/flatbuffers.
	if err := c.generateIDLStub(tmpDir); err != nil {
		return fmt.Errorf("generate rpc stub from template inside create rpc stub err: %w", err)
	}
	savedir := filepath.Join(tmpDir, "stub", pkg)
	saveFiles, err := os.ReadDir(savedir)
	if err != nil {
		return err
	}
	for _, f := range saveFiles {
		if strings.HasSuffix(f.Name(), "proto") {
			continue
		}
		if options.NoGoMod && f.Name() == "go.mod" {
			continue
		}
		fs.Move(filepath.Join(savedir, f.Name()), outputDir)
	}
	return nil
}
