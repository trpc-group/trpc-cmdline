// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package tpl encapsulates go's template operations and
// supports generating stub codes and configurations based on template files.
package tpl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"trpc.group/trpc-go/trpc-cmdline/config"
	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/util/fs"
	"trpc.group/trpc-go/trpc-cmdline/util/log"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
)

const (
	// ServiceIndexDefault represents the default value of the service index (by default,
	// services are separated into different files, a valid value needs to be passed)
	ServiceIndexDefault = 0
	// MethodIndexDefault represents the default value for service index (services are separated into different files,
	// effective value is required)
	MethodIndexDefault = -1
)

// FD is the a type alias of file descriptor.
type FD = descriptor.FileDescriptor

// GenerateFiles processes the go template files and outputs them to the outputdir directory.
func GenerateFiles(fd *FD, outputdir string, option *params.Option) error {
	// Preparing output directory.
	if err := fs.PrepareOutputdir(outputdir); err != nil {
		return fmt.Errorf("create outputdir: %v", err)
	}

	// Template file extension name.
	var cfg *config.Template
	if option.OtherType == "" {
		c, err := config.GetTemplate(option.IDLType, option.Language)
		if err != nil {
			return err
		}
		cfg = c
	}

	// Traverse through the template files for processing.
	f := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == option.Assetdir {
			return nil
		}
		mixed := MixedOptions{
			OutputDir: outputdir,
			Cfg:       cfg,
		}
		return ProcessTemplateFile(fd, path, info, option, &mixed)
	}
	return filepath.Walk(option.Assetdir, f)
}

// GenerateOptions is extension options.
type GenerateOptions struct {
	// Index position of the proto service, starting from 0, -1 means it doesn't exist.
	serviceIndex int
	// The index position of the service in proto, starting from 0,
	// and -1 indicates that it does not exist.
	methodIndex int
}

// ServiceIndex returns the index of the service to be generated.
func (o *GenerateOptions) ServiceIndex() int {
	if o != nil && o.serviceIndex >= 0 {
		return o.serviceIndex
	}
	return ServiceIndexDefault
}

// MethodIndex returns the index of the method to be generated.
func (o *GenerateOptions) MethodIndex() int {
	if o != nil && o.methodIndex >= 0 {
		return o.methodIndex
	}
	return MethodIndexDefault
}

// GenerateFile generates file for fd.
func GenerateFile(fd *FD, infile, outfile string, opt *params.Option, extOpt *GenerateOptions) error {
	if !filepath.IsAbs(opt.Assetdir) {
		return errors.New("assetdir must be absolute path")
	}

	// stat template
	if _, err := os.Lstat(infile); err != nil {
		return fmt.Errorf("lstat file err: %v", err)
	}

	// create output file
	fout, err := os.Create(outfile)
	if err != nil {
		return fmt.Errorf("create file err: %v", err)
	}
	defer fout.Close()

	// template execute and populate the output file
	var tplInstance *template.Template
	var baseName = filepath.Base(infile)

	if funcMap == nil {
		tplInstance, err = template.New(baseName).ParseFiles(infile)
	} else {
		tplInstance, err = template.New(baseName).Funcs(funcMap).ParseFiles(infile)
	}
	if err != nil {
		return fmt.Errorf("template initialize err: %v", err)
	}

	// Pass in descriptor information, command-line control parameter information,
	// and other serviceIndex information required by other split files.
	err = tplInstance.Execute(fout, struct {
		*descriptor.FileDescriptor
		*params.Option
		ServiceIndex       int
		MethodIndex        int
		TRPCCmdlineVersion string
	}{
		fd,
		opt,
		extOpt.ServiceIndex(),
		extOpt.MethodIndex(),
		config.TRPCCliVersion,
	})
	log.Debug("outfile:%s, genExtOption:%+v", outfile, extOpt)

	if err != nil {
		return fmt.Errorf("template execute err: %v", err)
	}
	return nil
}

// ProcessTemplateFile executes the Go template processing to generate a file
// based on the content of the entry template.
func ProcessTemplateFile(fd *FD, entry string, info os.FileInfo, option *params.Option, opts *MixedOptions) error {
	log.Debug("file entry srcPath:%s", entry)

	// keep same files/folders hierarchy in the outputdir/assetdir
	relPath := strings.TrimPrefix(entry, filepath.Clean(option.Assetdir)+string(filepath.Separator))
	if len(relPath) == 0 {
		return nil
	}
	tplFileExt := config.GlobalConfig().TplFileExt
	outPath := strings.TrimSuffix(filepath.Join(opts.OutputDir, relPath), tplFileExt)
	log.Debug("file entry destPath: %s", outPath)

	// if `entry` is directory, create the same entry in `outputdir`
	if info.IsDir() {
		return os.MkdirAll(outPath, os.ModePerm)
	}
	// if `entry` is client/server stub
	outdir := filepath.Dir(filepath.Join(opts.OutputDir, relPath))
	if opts.Cfg != nil {
		if isServerStubFile(relPath, opts.Cfg.RPCServerStub) {
			return generateServerStub(fd, entry, outdir, opts.Cfg, option)
		}
		if isServerTestStubFile(relPath, opts.Cfg.RPCServerTestStub) {
			return generateServerTestStub(fd, entry, outdir, opts.Cfg.LangFileExt, option)
		}
		if isRPCClientStubFile(relPath, opts.Cfg.RPCClientStub) {
			return generateClientStub(fd, entry, outdir, opts.Cfg, option)
		}
	}
	// if `entry` is normal go template file
	return GenerateFile(fd, entry, outPath, option, nil)
}

func isServerStubFile(fp, serverStub string) bool {
	stub := strings.ReplaceAll(serverStub, "/", string(filepath.Separator))
	return fp == stub
}

func isServerTestStubFile(fp, serverTestStub string) bool {
	stub := strings.ReplaceAll(serverTestStub, "/", string(filepath.Separator))
	return fp == stub
}

func isRPCClientStubFile(fp string, rpcClientStub []string) bool {
	for _, f := range rpcClientStub {
		stub := strings.ReplaceAll(f, "/", string(filepath.Separator))
		if fp == stub {
			return true
		}
	}
	return false
}

// generateServerStub generates server-side code corresponding to the service in the IDL.
func generateServerStub(fd *FD, infile, outdir string, cfg *config.Template, opt *params.Option) error {
	if opt.PerMethod {
		return generatePerMethod(fd, infile, outdir, cfg.LangFileExt, opt)
	}
	var camelcase bool
	return generatePerService(fd, infile, outdir, cfg.LangFileExt, camelcase, opt)
}

// generatePerService splits the generated code into separate files per service.
func generatePerService(fd *FD, infile, outdir, langFileExt string, camelcase bool, opt *params.Option) error {
	for sIdx, sd := range fd.Services {
		base := strcase.ToSnake(sd.Name) + "." + langFileExt
		if camelcase {
			base = strcase.ToCamel(sd.Name) + "." + langFileExt
		}
		outfile := filepath.Join(outdir, base)
		if err := GenerateFile(fd, infile, outfile, opt, &GenerateOptions{sIdx, -1}); err != nil {
			return err
		}
	}
	return nil
}

// generatePerMethod splits the generated code into separate files per method.
func generatePerMethod(fd *FD, inFile, outdir, langFileExt string, option *params.Option) error {
	for sIdx, sd := range fd.Services {
		for mIdx, method := range sd.RPC {
			base := strcase.ToSnake(sd.Name) + "_" + strcase.ToSnake(method.Name) + "." + langFileExt
			outfile := filepath.Join(outdir, base)
			if err := GenerateFile(fd, inFile, outfile, option, &GenerateOptions{sIdx, mIdx}); err != nil {
				return err
			}
		}
	}
	return nil
}

// generateServerTestStub generates server-side test code for the service in the IDL.
func generateServerTestStub(fd *FD, entry, outdir, langFileExt string, option *params.Option) error {
	for idx, sd := range fd.Services {
		base := strcase.ToSnake(sd.Name) + "_test." + langFileExt
		outfile := filepath.Join(outdir, base)
		if err := GenerateFile(fd, entry, outfile, option, &GenerateOptions{serviceIndex: idx}); err != nil {
			return err
		}
		log.Debug("entry destPath: %s", outfile)
		continue
	}
	return nil
}

// generateClientStub generates the client code corresponding to the service in the IDL.
func generateClientStub(fd *FD, infile, outdir string, cfg *config.Template, opt *params.Option) error {
	// generate only one rpc stub file, which contains all services
	if !cfg.RPCClientStubPerService {
		base := strings.TrimSuffix(filepath.Base(infile), config.GlobalConfig().TplFileExt)
		outPath := filepath.Join(outdir, base)
		return GenerateFile(fd, infile, outPath, opt, nil)
	}

	// generate rpc stub file per service
	keepOrigName := cfg.KeepOrigName
	camelcaseName := cfg.CamelCaseName
	langFileExt := cfg.LangFileExt

	for idx, sd := range fd.Services {
		var base string
		if !keepOrigName {
			base = strcase.ToSnake(sd.Name) + "." + langFileExt
			if camelcaseName {
				base = strcase.ToCamel(sd.Name) + "." + langFileExt
			}
		} else {
			base = strcase.ToSnake(sd.Name) + cfg.Separator + fs.BaseNameWithoutExt(infile)
			if camelcaseName {
				base = strcase.ToCamel(sd.Name) + cfg.Separator + fs.BaseNameWithoutExt(infile)
			}
		}
		outfile := filepath.Join(outdir, base)
		if err := GenerateFile(fd, infile, outfile, opt, &GenerateOptions{idx, -1}); err != nil {
			return err
		}
	}
	return nil
}

// GenerateFilePerService outputs clientStub by service splitting files.
func GenerateFilePerService(fd *FD, infile, outfile, stubName string, opt *params.Option) error {
	cfg, err := config.GetTemplate(opt.IDLType, opt.Language)
	if err != nil {
		return err
	}

	dir := filepath.Dir(outfile)

	for i, sd := range fd.Services {
		var base string
		if cfg.KeepOrigName {
			base = fmt.Sprintf("%s%s%s.%s", strcase.ToCamel(sd.Name), cfg.Separator, stubName, cfg.Language)
		} else {
			base = fmt.Sprintf("%s%s.%s", strcase.ToSnake(sd.Name), cfg.Separator, cfg.Language)
		}

		outfile = filepath.Join(dir, base)
		extopt := &GenerateOptions{serviceIndex: i}
		if err := GenerateFile(fd, infile, outfile, opt, extopt); err != nil {
			return err
		}
	}
	return nil
}

// MixedOptions aggregates many options to simplify method signatures.
type MixedOptions struct {
	OutputDir string
	Cfg       *config.Template
}
