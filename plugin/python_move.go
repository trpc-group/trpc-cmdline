// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/parser"
	"trpc.group/trpc-go/trpc-cmdline/util/fs"
	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

const pythonTRPCPBPackage = "trpc_pb"

// PythonMove is the plugin for adjusting the directory structure of python stubs.
type PythonMove struct {
}

// Name returns the plugin name.
func (p *PythonMove) Name() string {
	return "move"
}

// Check only runs when `--lang=python`.
func (p *PythonMove) Check(fd *descriptor.FileDescriptor, opt *params.Option) bool {
	return opt.Language == "python"
}

// Run moves the stub files into stub/<package> and fixes the import paths of *_pb2.py files.
func (p *PythonMove) Run(fd *descriptor.FileDescriptor, opt *params.Option) error {
	log.Debug("execute plugin for %s in %s", opt.Language, opt.OutputDir)

	pbPackage, err := parser.GetPackage(fd, opt.Language)
	if err != nil {
		return err
	}

	stubDir, pbOutDir, err := prepareStubPackageOutputDir(pbPackage, opt.OutputDir)
	if err != nil {
		return err
	}

	// Move stub/<package>/setup.py to stub/setup.py.
	if err = moveSetupFile(pbOutDir, stubDir); err != nil {
		return err
	}

	log.Debug("Fix the import sentence in *_pb2 files in %s for %s", pbOutDir, pbPackage)
	return fixPBImport(pbOutDir, pbPackage)
}

func prepareStubPackageOutputDir(pbPackage string, outputDir string) (string, string, error) {
	stubDir := filepath.Join(outputDir, "stub")
	pbOutDir := filepath.Join(stubDir, pbPackage)

	if _, err := os.Stat(pbOutDir); err == nil {
		return stubDir, pbOutDir, nil
	}

	files, err := os.ReadDir(outputDir)
	if err != nil {
		return "", "", err
	}

	log.Debug("create stub package dir: %s", pbOutDir)
	if err := os.MkdirAll(pbOutDir, os.ModePerm); err != nil {
		return "", "", err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		log.Debug("move %s from %s to %s", file.Name(), outputDir, pbOutDir)
		if err := fs.Rename(filepath.Join(outputDir, file.Name()), filepath.Join(pbOutDir, file.Name())); err != nil {
			return "", "", err
		}
	}

	return stubDir, pbOutDir, nil
}

func moveSetupFile(pbOutDir string, stubDir string) error {
	setupFilePath := filepath.Join(pbOutDir, "setup.py")

	if fin, err := os.Stat(setupFilePath); err == nil && !fin.IsDir() {
		log.Debug("move setup.py file from %s to %s", setupFilePath, stubDir)
		if err = fs.Move(setupFilePath, stubDir); err != nil {
			return err
		}
	}

	return nil
}

func fixPBImport(pbDir, pbPackage string) error {
	moduleImports, err := copyFilesAndBuildModuleImports(pbDir)
	if err != nil {
		return err
	}
	return fixImportPath(pbDir, pbPackage, moduleImports)
}

// copyFilesAndBuildModuleImports copies __init__.py into every sub directory of pbDir,
// and collects the import sentences of all *_pb2.py modules.
func copyFilesAndBuildModuleImports(pbDir string) (map[string]string, error) {
	moduleImports := make(map[string]string)

	err := filepath.Walk(pbDir, func(fp string, info os.FileInfo, err error) error {
		// Continue to process other files if err encountered.
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if info.Name() != filepath.Base(pbDir) {
				_ = fs.Copy(filepath.Join(pbDir, "__init__.py"), filepath.Join(fp, "__init__.py"))
			}
			return nil
		}

		baseName := filepath.Base(fp)
		if strings.HasSuffix(baseName, "_pb2.py") {
			moduleName := fs.BaseNameWithoutExt(baseName)
			modulePath, _ := filepath.Split(strings.Replace(fp, pbDir+"/", "", -1))
			modulePath = fs.BaseNameWithoutExt(strings.ReplaceAll(modulePath, "/", "."))

			if len(modulePath) == 0 {
				moduleImports[moduleName] = fmt.Sprintf("import %s", moduleName)
			} else {
				moduleImports[moduleName] = fmt.Sprintf("from %s import %s", modulePath, moduleName)
			}
			log.Debug("Found module %s using %s for %s", moduleName, moduleImports[moduleName], fp)
		}
		return nil
	})

	return moduleImports, err
}

func fixImportPath(pbDir, pbPackage string, moduleImports map[string]string) error {
	return filepath.Walk(pbDir, func(filePath string, info os.FileInfo, err error) error {
		// Continue to process other files if err encountered.
		if err != nil {
			return nil
		}

		basename := filepath.Base(filePath)
		if !strings.HasSuffix(basename, "_pb2.py") && basename != "pb.py" {
			return nil
		}

		log.Debug("Rewrite %s", filePath)
		in, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		out := string(in)

		for moduleName, moduleImport := range moduleImports {
			targetContent := fmt.Sprintf("from . %s", moduleImport)
			if !strings.HasPrefix(moduleImport, "import ") {
				targetContent = fmt.Sprintf("from %s.%s", pbPackage, strings.Replace(moduleImport, "from ", "", -1))
			}
			log.Debug("%s: %s -> %s", moduleName, moduleImport, targetContent)
			out = strings.ReplaceAll(out, moduleImport, targetContent)
		}

		out = fixTRPCPBImportPath(out)
		return os.WriteFile(filePath, []byte(out), os.ModePerm)
	})
}

// fixTRPCPBImportPath redirects imports of trpc protocol modules to the trpc_pb package,
// to avoid conflicting with the trpc framework package.
func fixTRPCPBImportPath(content string) string {
	fromTRPC := "from trpc."
	importTRPC := "import trpc."
	fromTRPCPB := fmt.Sprintf("from %s.trpc.", pythonTRPCPBPackage)
	importTRPCPB := fmt.Sprintf("import %s.trpc.", pythonTRPCPBPackage)

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, fromTRPC):
			lines[i] = strings.Replace(line, fromTRPC, fromTRPCPB, 1)
		case strings.HasPrefix(line, importTRPC):
			lines[i] = strings.Replace(line, importTRPC, importTRPCPB, 1)
		}
	}
	return strings.Join(lines, "\n")
}
