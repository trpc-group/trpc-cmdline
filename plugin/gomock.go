package plugin

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/parser"
	"trpc.group/trpc-go/trpc-cmdline/util/fs"
	"trpc.group/trpc-go/trpc-cmdline/util/lang"
	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

// GoMock is gomock plugin.
type GoMock struct {
}

// Name return plugin's name.
func (p *GoMock) Name() string {
	return "mockgen"
}

// Check only run when `--lang=go || --mockgen=true`
func (p *GoMock) Check(fd *descriptor.FileDescriptor, opt *params.Option) bool {
	if opt.Language != "go" || !opt.Mockgen || fd == nil || len(fd.Services) == 0 {
		return false
	}

	// If not installed, only prompt to install, do not fail.
	_, err := exec.LookPath("mockgen")
	if err != nil {
		log.Error("mockgen not found: %v", err)
		return false
	}
	return true
}

// Run runs mockgen action.
func (p *GoMock) Run(fd *descriptor.FileDescriptor, opt *params.Option) error {
	if !opt.RPCOnly && opt.Mockgen {
		return p.runGoGenerateAllAround(opt)
	}

	wd, _ := os.Getwd()
	defer os.Chdir(wd)

	os.Chdir(opt.OutputDir)

	pkgName, err := parser.GetPbPackage(fd, "go_package")
	if err != nil {
		return err
	}

	if !opt.NoGoMod {
		if err := p.ensureGoMod(pkgName); err != nil {
			return err
		}
	}

	pkg := lang.PBValidGoPackage(pkgName)
	fname := fs.BaseNameWithoutExt(fd.FilePath)
	dest := fmt.Sprintf("-destination=%s_mock.go", strcase.ToSnake(fname))
	pkgv := fmt.Sprintf("-package=%s", pkg)
	var selfpkgv string
	if !opt.NoGoMod {
		selfpkgv = fmt.Sprintf("-self_package=%s", pkgName)
	}
	source := fmt.Sprintf("--source=%s.trpc.go", fname)

	if err := runCmd(fmt.Sprintf("mockgen %s %s %s %s", dest, pkgv, source, selfpkgv)); err != nil {
		return fmt.Errorf("go mock mockgen err: %w, "+
			"if the error is caused by go mod tidy, "+
			"you may try adding --nogomod flag to use the outer go.mod of your project, "+
			"or you can use --mock=false to disable go mod tidy and mockgen completely", err)
	}
	return nil
}

// ensureGoMod ensure go mod is valid
func (p GoMock) ensureGoMod(pkgName string) error {
	if err := p.checkGoMod(pkgName); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := p.initGoMod(pkgName); err != nil {
			return err
		}
		return nil
	}
	if err := runCmd("go mod tidy"); err != nil {
		return fmt.Errorf("go mock ensure go mod err: %w", err)
	}
	return nil
}

func (p *GoMock) runGoGenerateAllAround(opt *params.Option) error {
	return filepath.Walk(opt.OutputDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		matches, err := filepath.Glob(opt.OutputDir + "/*.go")
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			return nil
		}
		// Do not run go generate under stub dir.
		if strings.Contains(path, "stub") {
			return nil
		}

		// Run "go generate".
		// If wd is an actual path and path is a symbolic link, there may be problems with go generate failure,
		// so path is not specified here to execute.
		wd, _ := os.Getwd()
		defer os.Chdir(wd)
		os.Chdir(path)
		log.Debug("switch to path %s from working dir %s before go generate", path, wd)
		// run `go mod tidy` before `mockgen` which is specified by //go:generate
		if err := runCmd("go mod tidy"); err != nil {
			return fmt.Errorf("run go mod tidy inside go mock, err: %w", err)
		}
		if err := runCmd("go generate"); err != nil {
			return fmt.Errorf("run go generate inside go mock, err: %w", err)
		}
		return nil
	})
}

func (p *GoMock) initGoMod(pkg string) error {
	mod := lang.TrimRight(";", pkg)
	if err := runCmd("go mod init " + mod); err != nil {
		return fmt.Errorf("go mock: go mod init err: %w", err)
	}

	if err := runCmd("go mod tidy"); err != nil {
		return fmt.Errorf("go mock: go mod tidy err: %w", err)
	}
	return nil
}

func runCmd(cmd string) error {
	log.Debug("run cmd: %s", cmd)
	args := strings.Split(cmd, " ")
	c := exec.Command(args[0], args[1:]...)
	b, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cmd exec err: %v, msg:%s", err, string(b))
	}
	return nil
}

// checkGoMod check the mod is valid
func (p *GoMock) checkGoMod(mod string) error {
	f, err := os.Open("go.mod")
	if err != nil {
		return err
	}
	defer f.Close()

	const module = "module"

	br := bufio.NewScanner(f)
	for br.Scan() {
		s := strings.TrimSpace(br.Text())
		if !strings.HasPrefix(s, module+" ") {
			continue
		}

		name := s[len(module)+1:]
		if name != mod {
			return fmt.Errorf("当前目录已经包含go.mod (%s != %s)，请通过-o指定其他输出目录", name, mod)
		}
		return nil
	}
	return errors.New("invalid go.mod")
}
