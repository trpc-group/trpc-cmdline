package create

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"trpc.group/trpc-go/trpc-cmdline/config"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/parser"
	"trpc.group/trpc-go/trpc-cmdline/util/fb"
	"trpc.group/trpc-go/trpc-cmdline/util/fs"
	"trpc.group/trpc-go/trpc-cmdline/util/lang"
	"trpc.group/trpc-go/trpc-cmdline/util/log"
	"trpc.group/trpc-go/trpc-cmdline/util/pb"
)

// generateIDLStub generates *.pb.go under outputdir/rpc/.
func (c *Create) generateIDLStub(dir string) error {
	// Cpp IDL stub code will be generated using bazel rules, do nothing here.
	if c.options.Language == "cpp" {
		return nil
	}

	fd, options := c.fileDescriptor, c.options
	stubdir, err := prepareOutputStub(dir)
	if err != nil {
		return fmt.Errorf("prepare output stub: %w", err)
	}

	pkg, err := parser.GetPackage(fd, options.Language)
	if err != nil {
		return fmt.Errorf("parser get package: %w", err)
	}

	if err := generatePBFB(fd, options, pkg, stubdir); err != nil {
		return fmt.Errorf("generate pb fb: %w", err)
	}

	if !options.RPCOnly || options.DependencyStub {
		if err := handleDependencies(fd, options, pkg, stubdir); err != nil {
			return fmt.Errorf("handle dependencies: %w", err)
		}
	}

	// Move dir/rpc into dir/$gopkgdir/.
	src := filepath.Join(dir, "rpc")
	dest := filepath.Join(stubdir, pkg)
	defer os.RemoveAll(src)

	return filepath.Walk(src, func(fpath string, _ os.FileInfo, _ error) (e error) {
		if fpath == src {
			return nil
		}
		if fname := filepath.Base(fpath); fname == "trpc.go" {
			return fs.Move(fpath, filepath.Join(dest, fs.BaseNameWithoutExt(fd.FilePath)+".trpc.go"))
		}
		return fs.Move(fpath, filepath.Join(dest, filepath.Base(fpath)))
	})
}

func prepareOutputStub(outputdir string) (string, error) {
	stubDir := filepath.Join(outputdir, "stub")

	if _, err := os.Lstat(stubDir); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		if err := os.Mkdir(stubDir, os.ModePerm); err != nil {
			return "", err
		}
		return stubDir, nil
	}
	return stubDir, nil
}

// generatePBFB generates stub code based on option.IDLType by calling runProtoc or runFlatc.
func generatePBFB(fd *FD, option *params.Option, packageName, stubDir string) error {
	out := filepath.Join(stubDir, packageName)
	if err := os.MkdirAll(out, os.ModePerm); err != nil {
		return err
	}
	if option.IDLType == config.IDLTypeProtobuf {
		log.Debug("generate code for file %s from %v into %s", option.Protofile, option.Protodirs, out)
		// Invoke protoc and copy the .proto file to the generated folder.
		return protocAndCopy(fd, option, out)
	}
	log.Debug("generate code for file %s into %s", option.Protofile, out)
	// Invoke flatc and copy the .fbs file to the generated folder.
	return flatcAndCopy(fd, option, out)
}

// protocAndCopy invokes runProtoc and copies the .proto file to the generated folder.
func protocAndCopy(fd *FD, option *params.Option, pbOutDir string) error {
	if _, err := runProtoc(fd, pbOutDir, option); err != nil {
		return fmt.Errorf("run protoc err: %w", err)
	}
	if option.DescriptorSetIn != "" {
		// When passing in the descriptor_set_in parameter, skip copying the .proto file as it does not exist.
		return nil
	}
	// copy *.proto to outpoutdir/rpc/
	basename := filepath.Base(fd.FilePath)
	return fs.Copy(fd.FilePath, filepath.Join(pbOutDir, basename))
}

// runProtoc sets the required options and invokes util/pb.Protoc for processing.
func runProtoc(fd *FD, pbOutDir string, option *params.Option) ([]string, error) {
	opts := []pb.Option{
		pb.WithPb2ImportPath(fd.Pb2ImportPath),
		pb.WithPkg2ImportPath(fd.Pkg2ImportPath),
		pb.WithDescriptorSetIn(option.DescriptorSetIn),
	}

	var files []string

	// run protoc --$lang_out
	if err := pb.Protoc(option.Protodirs, option.Protofile, option.Language, pbOutDir, opts...); err != nil {
		return nil, fmt.Errorf("run protoc --$lang_out err: %v", err)
	}

	// run protoc-gen-secv
	opts = append(opts, pb.WithSecvEnabled(true))
	if err := pb.Protoc(option.Protodirs, option.Protofile, option.Language, pbOutDir, opts...); err != nil {
		return nil, fmt.Errorf("run protoc-gen-secv err: %v", err)
	}
	log.Debug("pbOutDir = %s", pbOutDir)
	// collect the generated files
	matches, err := filepath.Glob(pbOutDir)
	if err != nil {
		return nil, fmt.Errorf("filepath glob pb out dir: %s, err: %w", pbOutDir, err)
	}
	for _, v := range matches {
		if v == pbOutDir {
			continue
		}
		inf, err := os.Lstat(v)
		if err != nil {
			continue
		}
		if inf.IsDir() {
			continue
		}
		files = append(files, v)
	}

	return files, nil
}

// runFlatc sets the required options and invokes util.fb.Flatc.
func runFlatc(fd *FD, fbsOutDir string, option *params.Option) ([]string, error) {
	opts := []fb.Option{
		fb.WithFbsDirs(option.Protodirs),
		fb.WithFbsfile(option.Protofile),
		fb.WithLanguage(option.Language),
		fb.WithPackagePath(fd.BaseGoPackageName),
		fb.WithOutputdir(fbsOutDir),
		fb.WithFb2ImportPath(fd.Pb2ImportPath),
		fb.WithPkg2ImportPath(fd.Pkg2ImportPath),
	}
	// FIXME, return generate filenames
	return nil, fb.NewFbs(opts...).Flatc()
}

// flatcAndCopy constructs the parameter list and invokes fb.Flatc to generate stub code for each type in the .fbs file.
// Then, the .fbs file is copied to the generated folder.
func flatcAndCopy(fd *FD, option *params.Option, outdir string) error {
	if _, err := runFlatc(fd, outdir, option); err != nil {
		return err
	}

	// The basename is in the form of "file1.fbs".
	basename := filepath.Base(fd.FilePath)

	// Copy the *.fbs file to the directory where the generated files are located.
	return fs.Copy(fd.FilePath, filepath.Join(outdir, basename))
}

// handleDependencies processes other pb files imported by the pb files specified in the "-protofile" option.
// It also processes protoc and copies pb files.
//
// Preparing to generate *.pb.go files corresponding to the PB files using protoc.
// Note that to avoid generating code with circular dependencies.
//
// Parse the result using jhump/protoreflect.
// If the pkgname is the same as the one specified in "-protofile", the importpath will be "".
//
// runProtoc --go_out=M$pb=$pkgname, we need to do compatibility processing:
//  1. Avoid passing $pkgname as empty, otherwise protoc will generate code like this.：
//     ```go
//     package $pkgname
//     import (
//     "."
//     )
//     ```
//  2. Avoid passing the same pkgname as -protofile, otherwise it will cause circular dependencies.
//     ```go
//     package $pkgname
//     import (
//     "$pkgname"
//     )
//     ```
func handleDependencies(fd *FD, option *params.Option, pbpkg string, outputDir string) error {
	outputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("filepath abs output dir: %s, err: %w", outputDir, err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("os get working directory err: %w", err)
	}
	defer os.Chdir(wd)

	return doHandleDependencies(fd, pbpkg, outputDir, wd, option)
}

func doHandleDependencies(fd *FD, pbpkg, outputDir, wd string, option *params.Option) error {
	includeDirs := genIncludeDirs(fd)
	for fname, importPath := range fd.Pb2ImportPath {
		if skipThisProtofile(fd, fname) {
			continue
		}

		param := &genDependencyRPCStubParam{
			fd:          fd,
			option:      option,
			pbpkg:       pbpkg,
			outputDir:   outputDir,
			fname:       fname,
			importPath:  importPath,
			wd:          wd,
			includeDirs: includeDirs,
		}
		param.importPath = lang.TrimRight(";", param.importPath)
		pbOutDir, err := param.genDependencyRPCStub()
		if err != nil {
			return fmt.Errorf("generate dependency rpc stub err: %w", err)
		}
		importPath = lang.TrimRight(";", importPath)
		if err := moduleInit(option, pbpkg, fname, importPath, pbOutDir); err != nil {
			return fmt.Errorf("module init err: %w", err)
		}
	}
	return nil
}

func genIncludeDirs(fd *FD) []string {
	includeDirs := []string{}
	for fname := range fd.Pb2ImportPath {
		dir, _ := filepath.Split(fname)
		includeDirs = append(includeDirs, dir)
	}
	return includeDirs
}

func skipThisProtofile(fd *FD, fname string) bool {
	// If it is ${protofile}, skip and do not process it.
	if filepath.Base(fd.FilePath) == fname {
		return true
	}

	// Skip the pb files, trpc extension files, and swagger extension files provided by Google.
	return pb.IsInternalProto(fname)
}

type genDependencyRPCStubParam struct {
	fd          *FD
	option      *params.Option
	pbpkg       string
	outputDir   string
	fname       string
	importPath  string
	wd          string
	includeDirs []string
}

func (g *genDependencyRPCStubParam) genDependencyRPCStub() (string, error) {
	var err error

	g.outputDir, err = prepareOutputDir(g.outputDir, g.importPath, g.option.Language, g.pbpkg)
	if err != nil {
		return "", fmt.Errorf("prepare output dir, err: %w", err)
	}

	switch g.option.IDLType {
	case config.IDLTypeProtobuf:
		err = g.genDependencyRPCStubPB()
	case config.IDLTypeFlatBuffers:
		err = g.genDependencyRPCStubFB()
	default:
		return "", errors.New("invalid IDL type")
	}

	return g.outputDir, err
}

func prepareOutputDir(outputDir, importPath, lang, pbPackage string) (string, error) {
	var pbOutDir string
	if lang == "go" {
		pbOutDir = filepath.Join(outputDir, importPath)
	} else {
		pbOutDir = filepath.Join(outputDir, pbPackage)
	}
	if err := os.MkdirAll(pbOutDir, os.ModePerm); err != nil {
		return "", err
	}
	return pbOutDir, nil
}

func (g *genDependencyRPCStubParam) genDependencyRPCStubPB() error {
	// Inherit the directory from the parent level to avoid directory not found issues.
	searchPath, err := genProtocProtoPath(g.option, g.wd, g.includeDirs)
	if err != nil {
		return fmt.Errorf("generate protoc proto path err: %w", err)
	}
	log.Debug("Generate code for proto file %s from %v into %s", g.fname, searchPath, g.outputDir)

	// run protoc-gen-go
	opts := []pb.Option{
		pb.WithPb2ImportPath(g.fd.Pb2ImportPath),
		pb.WithPkg2ImportPath(g.fd.Pkg2ImportPath),
		pb.WithDescriptorSetIn(g.option.DescriptorSetIn),
	}
	if err = pb.Protoc(searchPath, g.fname, g.option.Language, g.outputDir, opts...); err != nil {
		return fmt.Errorf("GenerateFiles: %v", err)
	}

	// run protoc-gen-secv
	opts = append(opts, pb.WithSecvEnabled(true))
	if err = pb.Protoc(searchPath, g.fname, g.option.Language, g.outputDir, opts...); err != nil {
		return fmt.Errorf("GenerateFiles: %v", err)
	}
	if g.option.DescriptorSetIn != "" {
		return nil // skip copy if descriptor_set_in is passed.
	}

	// Copy pb file.
	if err := copyProtofile(g.fname, g.outputDir, g.option); err != nil {
		return fmt.Errorf("copy proto file err: %w", err)
	}
	return nil
}

func genProtocProtoPath(option *params.Option, wd string, includeDirs []string) ([]string, error) {
	searchPath := option.Protodirs
	parentDirs := []string{wd}
	parentDirs = append(parentDirs, option.Protodirs...)
	for _, pDir := range parentDirs {
		newSearchPath, err := genProtocProtoPathByParentDir(includeDirs, pDir)
		if err != nil {
			return nil, err
		}
		searchPath = append(searchPath, newSearchPath...)
	}
	return fs.UniqFilePath(searchPath), nil
}

func genProtocProtoPathByParentDir(includeDirs []string, pDir string) ([]string, error) {
	var searchPath []string
	for _, incDir := range includeDirs {

		includeDir := filepath.Join(pDir, incDir)
		includeDir = filepath.Clean(includeDir)

		if fin, err := os.Lstat(includeDir); err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("os lstat err err: %w", err)
			}
		} else {
			if !fin.IsDir() {
				return nil, fmt.Errorf("import path: %s, not directory", includeDir)
			}
			searchPath = append(searchPath, includeDir)
		}
	}
	return searchPath, nil
}

func copyProtofile(fname, pbOutDir string, option *params.Option) error {
	p, err := fs.LocateFile(fname, option.Protodirs)
	if err != nil {
		return fmt.Errorf("fs locate file err: %w", err)
	}

	_, baseName := filepath.Split(fname)
	src := p
	dst := filepath.Join(pbOutDir, baseName)

	log.Debug("Copy file %s to %s", src, dst)
	if err := fs.Copy(src, dst); err != nil {
		return err
	}
	return nil
}

func (g *genDependencyRPCStubParam) genDependencyRPCStubFB() error {
	strs := strings.Split(g.importPath, "/")
	baseGoPackageName := strs[len(strs)-1]
	filename, err := fs.LocateFile(g.fname, g.option.Protodirs)
	if err != nil {
		return fmt.Errorf("cannot locate file %v: %v", g.fname, err)
	}
	opts := []fb.Option{
		fb.WithFbsDirs(g.option.Protodirs),
		fb.WithFbsfile(filename),
		fb.WithLanguage(g.option.Language),
		fb.WithPackagePath(baseGoPackageName),
		fb.WithOutputdir(g.outputDir),
		fb.WithFb2ImportPath(g.fd.Pb2ImportPath),
		fb.WithPkg2ImportPath(g.fd.Pkg2ImportPath),
	}
	f := fb.NewFbs(opts...)
	if err := f.Flatc(); err != nil {
		return fmt.Errorf("flatc: %v", err)
	}
	// Copy fbs file.
	_, baseName := filepath.Split(filename)
	src := filename
	dst := filepath.Join(g.outputDir, baseName)
	log.Debug("Copy file %s to %s", src, dst)
	if err := fs.Copy(src, dst); err != nil {
		return err
	}
	return nil
}

func moduleInit(option *params.Option, pbpkg string, fname string, importPath string, pbOutDir string) error {
	// Fixme: move to createCmd.PostRun
	// Run "go mod init". If it is the same as pbPackage, no initialization is required.
	if option.Language != "go" {
		return nil
	}

	return genGoModInit(importPath, pbpkg, pbOutDir, fname)
}

func genGoModInit(importPath, pbPackage, pbOutDir, fname string) error {
	// Initialize go.mod to avoid duplicating initialization of go.mod.
	fp := filepath.Join(pbOutDir, "go.mod")
	fin, err := os.Stat(fp)
	if err == nil && !fin.IsDir() {
		return nil
	}

	// Run "go mod init".
	if !canExecGoModInit(importPath, pbPackage) {
		return nil
	}
	_ = os.Chdir(pbOutDir)

	cmd := exec.Command("go", "mod", "init", importPath)
	if buf, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("process %s, init go.mod in stub/%s error: %v", fname, importPath, string(buf))
	}
	log.Debug("process %s, init go.mod success in stub/%s: go mod init %s", fname, importPath, importPath)
	return nil
}

func canExecGoModInit(importPath string, pbPackage string) bool {
	return len(importPath) != 0 && importPath != pbPackage
}
