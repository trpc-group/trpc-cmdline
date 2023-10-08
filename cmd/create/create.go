// Package create provides create command.
package create

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"trpc.group/trpc-go/trpc-cmdline/config"
	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/parser"
	"trpc.group/trpc-go/trpc-cmdline/tpl"
	"trpc.group/trpc-go/trpc-cmdline/util/fs"
	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

// FD is an alias of descriptor.FileDescriptor.
type FD = descriptor.FileDescriptor

// Create is for the create command.
type Create struct {
	options        *params.Option
	fileDescriptor *descriptor.FileDescriptor
	preRunHook     func() error
}

// CMD returns create command.
func CMD() *cobra.Command {
	c := New(func() error { return nil })
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Quickly create projects or RPC stubs by specifying the pb file",
		Long: `Quickly create projects or RPC stubs by specifying the pb file.

'trpc create' has two modes:
- Generate a complete service project
- Generate RPC stubs for the target service, specify the '--rpconly' option
`,
		PreRunE:  c.PreRunE,
		RunE:     c.RunE,
		PostRunE: c.PostRunE,
	}
	AddCreateFlags(createCmd)
	return createCmd
}

// New initializes a create command.
func New(preRunHook func() error) *Create {
	return &Create{
		options:    &params.Option{},
		preRunHook: preRunHook,
	}
}

// PreRunE provides *cobra.Command.PreRunE.
func (c *Create) PreRunE(cmd *cobra.Command, args []string) error {
	if err := c.loadOptions(cmd.Flags()); err != nil {
		return fmt.Errorf("load create options inside pre run err: %w", err)
	}
	log.SetPrefix("[create]")
	if err := c.preRunHook(); err != nil {
		return fmt.Errorf("pre run hook err: %w", err)
	}
	// Non-pb type, pre run done.
	if c.options.OtherType != "" {
		return nil
	}
	var err error
	opts := []parser.Option{
		parser.WithAliasOn(c.options.AliasOn),
		parser.WithLanguage(c.options.Language),
		parser.WithRPCOnly(c.options.RPCOnly),
		parser.WithMultiVersion(c.options.MultiVersion),
	}
	if c.options.DescriptorSetIn != "" {
		c.fileDescriptor, err = parser.LoadDescriptorSet(
			c.options.DescriptorSetIn,
			c.options.Protofile,
			opts...,
		)
	} else {
		c.fileDescriptor, err = parser.Parse(
			c.options.Protofile,
			c.options.Protodirs,
			c.options.IDLType,
			opts...,
		)
	}
	if err != nil {
		return fmt.Errorf("parser.Parse during pre run err: %w", err)
	}
	if c.options.Verbose {
		c.fileDescriptor.Dump()
	}
	// Check and install dependencies.
	return setup([]string{c.options.Language})
}

// RunE provides *cobra.Command.RunE.
func (c *Create) RunE(cmd *cobra.Command, args []string) error {
	log.Debug("args: %v", args)
	// Create a project of non protocol type.
	if c.options.OtherType != "" {
		return c.createByNonProtocolType()
	}
	return c.createByProtocolType()

}

func (c *Create) createByProtocolType() error {
	outputDir, err := getOutputDir(c.options)
	if err != nil {
		return err
	}
	c.options.OutputDir = outputDir
	// Create by IDL protocol type.
	// Create a full project.
	// if ignore RPCOnly flag, create a full project
	if !c.options.RPCOnly || ignoreRPCOnly(c.options) {
		return c.createFullProject()
	}
	// Create only rpc stub.
	return c.createRPCOnlyStub()
}

func (c *Create) createFullProject() (err error) {
	dir := c.options.OutputDir
	// Check whether the output path is clean.
	if !isCleanDir(dir, c.options.IDLType, c.options.Language) && !c.options.Force {
		return fmt.Errorf("%s is not empty, use a clean path or provide -f to force overwrite", dir)
	}

	// Delete semi-finished files if anything wrong when
	// generating a project from scratch.
	var isNewDir bool
	if _, err := os.Lstat(dir); err != nil && os.IsNotExist(err) {
		isNewDir = true
	}
	defer func() {
		if isNewDir && err != nil {
			removeSafe(dir)
		}
	}()

	// Traverse each file in install/asset_${lang}.
	if err := tpl.GenerateFiles(c.fileDescriptor, dir, c.options); err != nil {
		return fmt.Errorf("generate files from template err: %w", err)
	}

	// Generate IDL stub code for protobuf/flatbuffers.
	if err := c.generateIDLStub(dir); err != nil {
		return fmt.Errorf("generate rpc stub from template err: %w", err)
	}
	log.Info(
		"Create tRPC project %s`%s`%s: succeed! ヾ(@^▽^@)ノ",
		log.ColorRed,
		fs.BaseNameWithoutExt(c.fileDescriptor.FilePath),
		log.ColorGreen)
	return nil
}

// isCleanDir checks if generated files exist in the directory, and returns false if existed.
// Overwrite issues should be noted carefully
func isCleanDir(dir string, idlType config.IDLType, lang string) bool {
	cfg, err := config.GetTemplate(idlType, lang)
	if err != nil {
		log.Error("config.GetTemplate failed when checking isCleanDir: %+v", err)
		return false
	}
	return noFilesInLangExt(dir, cfg)
}

func noFilesInLangExt(dir string, cfg *config.Template) bool {
	// If the directory does not exist, it means there is no risk of file overwriting.
	if _, err := os.Lstat(dir); err != nil {
		return os.IsNotExist(err)
	}

	// If the directory exists, then check whether "any" files exist in the directory.
	// If existed, there is a risk of overwriting them.
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	ext := cfg.LangFileExt
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ext) {
			return false
		}
	}

	return true
}

// removeSafe deletes the generated directory.
// Note that if the current directory is the output directory,
// it cannot be deleted.
// Returns true if deleted.
func removeSafe(path string) (bool, error) {
	dir, err := os.Getwd()
	if err != nil {
		return false, err
	}
	if path == dir {
		return false, nil
	}
	if err := os.RemoveAll(path); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Create) createByNonProtocolType() error {
	if err := c.checkCreateByNonProtocolType(); err != nil {
		return err
	}

	// Code generation.
	// Preparing the output directory.
	outputdir, err := getOutputDir(c.options)
	if err != nil {
		return err
	}

	err = os.MkdirAll(outputdir, os.ModePerm)
	if err != nil {
		return err
	}

	fd := &FD{
		PackageName: "trpc.app." + c.options.OtherType,
		Services: []*descriptor.ServiceDescriptor{
			{
				Name: c.options.OtherType,
			},
		},
		AppName: c.options.OtherType,
	}
	c.fileDescriptor = &FD{
		FilePath: c.options.OtherType,
	}

	if err = tpl.GenerateFiles(fd, outputdir, c.options); err != nil {
		return err
	}

	c.options.OutputDir = outputdir
	return nil
}

func (c *Create) checkCreateByNonProtocolType() error {
	tmplPath := filepath.Join(c.options.Assetdir)
	fileInfo, err := os.Stat(tmplPath)
	if err != nil {
		return fmt.Errorf("get template from path: %s, err: %w", tmplPath, err)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", tmplPath)
	}
	return nil
}

// getOutputDir returns OutputDir.
// Note that error will be returned if you call os.Chdir and doesn't chdir back.
func getOutputDir(option *params.Option) (string, error) {
	// 1. If `-o` is specified, use its value.
	if option.OutputDir != "" {
		if filepath.IsAbs(option.OutputDir) {
			return option.OutputDir, nil
		}
		return filepath.Abs(option.OutputDir)
	}

	// 2. If `-n` is specified, use its value, like http, kafka, etc.
	if option.OtherType != "" {
		return "trpc_" + option.OtherType + "_service", nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get wd inside get output dir err: %w", err)
	}

	// 3. If rpconly == true && not ignore rpcOnly flag, the default value is os.Getwd().
	if option.RPCOnly && !ignoreRPCOnly(option) {
		return wd, nil
	}
	// 4, If rpconly == false, the default value is equal to the basename of protofile.
	if option.GoModEx != "" && option.GoModEx == option.GoMod {
		return wd, nil
	}
	return filepath.Join(wd, fs.BaseNameWithoutExt(option.Protofile)), nil
}

func setup(languages []string) error {
	if _, err := config.Init(); err != nil {
		return err
	}
	deps, err := config.LoadDependencies(languages...)
	if err != nil {
		return err
	}
	return config.SetupDependencies(deps)
}

// should ignore rpcOnly flag
func ignoreRPCOnly(option *params.Option) bool {
	// cpp lang use bazel rule for generating stub code
	return option.Language == "cpp"
}
