// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package config provides configuration-related capabilities for this project.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"trpc.group/trpc-go/trpc-cmdline/bindata/compress"
	"trpc.group/trpc-go/trpc-cmdline/gobin"
	"trpc.group/trpc-go/trpc-cmdline/util/fs"
	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

// LanguageTemplates is the templates for each programming language.
type LanguageTemplates map[string]*Template

var once sync.Once
var globalConfig Config

func init() {
	// Auto initialize.
	if _, err := Init(); err != nil {
		log.Error("init error: %+v", err)
	}
}

type options struct {
	force bool
}

// Option is the option provided for config.Init.
type Option func(*options)

// WithForce sets whether to force initialize the asset files, i.e. extract the files
// inside the binary to overwrite the existing files located at $HOME/.trpc-cmdline-assets.
// Default is false.
func WithForce(force bool) Option {
	return func(o *options) {
		o.force = force
	}
}

// Init initializes the configuration.
// If the configuration file trpc.yaml is missing, or the code template is missing,
// or the configuration version does not match, the configuration is automatically installed.
// After the initialization is successful, returns the path of config file's directory.
func Init(opts ...Option) (string, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	// List of candidate template paths.
	paths, err := candidatedInstallPath()
	if err != nil {
		return "", fmt.Errorf("get template search path error: %w", err)
	}

	// Check if the candidate template path exists.
	var reinstall bool
	installPath, err := templateInstallPath(paths)
	if err != nil {
		if err != errTemplateNotFound {
			return "", fmt.Errorf("get template instal path from %+v error: %w", paths, err)
		}
		reinstall = true
		installPath = paths[0]
	} else {
		// Version mismatch, needs to be reinstalled.
		dat, err := os.ReadFile(filepath.Join(installPath, "VERSION"))
		if err != nil || string(dat) != TRPCCliVersion {
			reinstall = true
		}
	}
	// Install the template as needed.
	if o.force || reinstall {
		log.Debug("reinstall template to %s", installPath)
		if err := installTemplate(installPath); err != nil {
			return "", fmt.Errorf("install template to %s error: %w", installPath, err)
		}
	}
	return installPath, nil
}

// CurrentTemplatePath gets the installation path of the trpc configuration file,
// which is installed to $HOME/.trpc-cmdline-assets.
func CurrentTemplatePath() (dir string, err error) {
	candicates, err := candidatedInstallPath()
	if err != nil {
		return "", err
	}
	return templateInstallPath(candicates)
}

var errTemplateNotFound = errors.New("Template directory not found")

func candidatedInstallPath() ([]string, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	candidates := []string{filepath.Join(u.HomeDir, ".trpc-cmdline-assets")}
	return candidates, nil
}

func templateInstallPath(dirs []string) (dir string, err error) {
	for _, d := range dirs {
		if fin, err := os.Lstat(d); err == nil && fin.IsDir() {
			return d, nil
		}
	}
	return "", errTemplateNotFound
}

func installTemplate(installTo string) error {
	tmp := filepath.Join(os.TempDir(), "trpc"+fmt.Sprintf("tmp+%d", rand.Uint64()))
	if err := os.RemoveAll(installTo); err != nil {
		return fmt.Errorf("remove %s err: %w", installTo, err)
	}
	if err := os.RemoveAll(tmp); err != nil {
		return fmt.Errorf("remove %s err: %w", tmp, err)
	}
	if err := os.MkdirAll(tmp, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir err: %w", err)
	}
	if err := compress.Untar(tmp, bytes.NewBuffer(gobin.AssetsGo)); err != nil {
		return fmt.Errorf("untar to %s err: %w", tmp, err)
	}
	if err := fs.Move(tmp, installTo); err != nil {
		return fmt.Errorf("move %s -> %s err: %w", tmp, installTo, err)
	}
	if err := os.RemoveAll(tmp); err != nil {
		return fmt.Errorf("remove %s after move err: %w", tmp, err)
	}
	return nil
}

// GlobalConfig returns global config.
func GlobalConfig() Config {
	once.Do(func() {
		cfg, err := LoadConfig()
		if err != nil {
			panic(err)
		}
		globalConfig = *cfg
	})
	return globalConfig
}

// GetTemplate returns the corresponding configuration information
// based on the given index of the serialization protocol and the language type.
func GetTemplate(idl IDLType, lang string) (*Template, error) {
	if !idl.Valid() {
		return nil, fmt.Errorf("invalid idltype: %s", idl)
	}

	tpl, ok := GlobalConfig().Templates[idl.String()][lang]
	if !ok {
		return nil, fmt.Errorf("language: %s not supported", lang)
	}
	return tpl, nil
}

func repair(cfg *Config) error {
	installedPath, err := CurrentTemplatePath()
	if err != nil {
		return err
	}

	for _, tpls := range cfg.Templates {
		for lang, tpl := range tpls {
			tpl.AssetDir = filepath.Join(installedPath, tpl.AssetDir)
			if tpl.Language == "" {
				tpl.Language = lang
			}
			if tpl.LangFileExt == "" {
				tpl.LangFileExt = lang
			}
		}
	}
	return nil
}

// LoadConfig loads trpc config.
func LoadConfig() (cfg *Config, err error) {
	defer func() {
		if err == nil && cfg != nil {
			repair(cfg)
		}
	}()
	// Try to load configuration from the file specified by --config flag.
	fp := viper.ConfigFileUsed()
	if fp != "" {
		return loadConfigFile(fp)
	}

	// Try to load configuration from the template installation path.
	d, err := CurrentTemplatePath()
	if err != nil {
		return nil, err
	}

	fp = filepath.Join(d, "trpc.yaml")
	return loadConfigFile(fp)
}

func loadConfigFile(fp string) (*Config, error) {
	b, err := os.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	cfg := Config{}
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// UninstallTemplate cleans up installed templates.
func UninstallTemplate() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	os.RemoveAll(filepath.Join(user.HomeDir, ".trpc-cmdline-assets"))
}
