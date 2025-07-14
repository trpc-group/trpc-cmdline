// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package config

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-cmdline/util/fs"
)

// Backup the current configuration information before running the tests,
// and restore the original configuration information after the tests are complete.
func TestMain(m *testing.M) {
	mustBackupTemplate()
	ret := m.Run()
	mustRestoreTemplate()
	os.Exit(ret)
}

func Test_templateCandidatedPath(t *testing.T) {
	t.Run("user: root install for self", func(t *testing.T) {
		p := gomonkey.NewPatches()
		p.ApplyFunc(os.Executable, func() (string, error) {
			return "/root/go/bin/trpc", nil
		})
		p.ApplyFunc(user.Current, func() (*user.User, error) {
			return &user.User{
				Username: "root",
				HomeDir:  "/root/",
			}, nil
		})
		defer p.Reset()

		paths, err := candidatedInstallPath()
		require.Nil(t, err)
		require.Equal(t, []string{"/root/.trpc-cmdline-assets"}, paths)
	})
}

// Load the trpc.yaml configuration file and return the parsed configuration.
func Test_LoadConfig(t *testing.T) {
	t.Run("load config by ~/.trpc-cmdline-assets/trpc.yaml", func(t *testing.T) {
		_, err := Init()
		require.Nil(t, err)

		c, err := LoadConfig()
		require.Nil(t, err)
		require.NotZero(t, c)
	})

	t.Run("load config by --config=", func(t *testing.T) {
		globalConfig = Config{}

		orig := viper.ConfigFileUsed()

		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		src := filepath.Join(filepath.Dir(wd), "install/trpc.yaml")

		defer func() {
			viper.Reset()
			viper.SetConfigFile(orig)
		}()

		viper.SetConfigFile(src)
		UninstallTemplate()

		c, err := LoadConfig()
		require.Nil(t, err)
		require.NotZero(t, c)
	})
}

// Load the trpc.yaml configuration file and return the parsed configuration.
func Test_CurrentTemplatePath(t *testing.T) {
	t.Run("template not installed", func(t *testing.T) {
		// Remove all candidate installation paths.
		// At this point, the check for the template installation path should return templateNotFound.
		UninstallTemplate()

		d, err := CurrentTemplatePath()
		require.Equal(t, errTemplateNotFound, err)
		require.Zero(t, d)
	})
}

func Test_Init_Exception(t *testing.T) {
	t.Run("templateInstallPath lstat error", func(t *testing.T) {
		UninstallTemplate()
		p := gomonkey.ApplyFunc(templateInstallPath, func([]string) (string, error) {
			return "", errors.New("error != templateNotFound")
		})
		_, err := Init()
		require.NotNil(t, err)
		require.NotEqual(t, err, errTemplateNotFound)
		p.Reset()
	})

	t.Run("installTemplate, mkdir error", func(t *testing.T) {
		UninstallTemplate()
		p := gomonkey.ApplyFunc(os.MkdirAll, func(string, os.FileMode) error {
			return errors.New("prevent mkdirall")
		})
		_, err := Init()
		require.NotNil(t, err)
		p.Reset()
	})

	t.Run("initConfig, by ~/.trpc-cmdline-assets/trpc.yaml", func(t *testing.T) {
		UninstallTemplate()
		p := gomonkey.ApplyFunc(viper.ConfigFileUsed, func() string {
			return ""
		})
		_, err := Init()
		require.Nil(t, err)
		p.Reset()
	})
}

func Test_GetTemplate(t *testing.T) {
	_, err := Init()
	require.Nil(t, err)

	t.Run("protobuf + go", func(t *testing.T) {
		cfg, err := GetTemplate(IDLTypeProtobuf, "go")
		require.Nil(t, err)
		require.NotNil(t, cfg)
	})

	t.Run("protobuf + aaaaaaa", func(t *testing.T) {
		cfg, err := GetTemplate(IDLTypeProtobuf, "aaaaaaa")
		require.NotNil(t, err)
		require.Nil(t, cfg)
	})

	t.Run("flatbuffers + go", func(t *testing.T) {
		fbsCfg, err := GetTemplate(IDLTypeFlatBuffers, "go")
		require.Nil(t, err)
		// The following content is asserted based on install/trpc.yaml.
		require.Contains(t, fbsCfg.AssetDir, "flatbuffers/asset_go")
		require.Equal(t, ".tpl", globalConfig.TplFileExt)
		require.Equal(t, "go", fbsCfg.LangFileExt)
		require.Equal(t, "service_rpc.go.tpl", fbsCfg.RPCServerStub)
		require.Equal(t, "service_rpc_test.go.tpl", fbsCfg.RPCServerTestStub)
		require.Equal(t, "trpc.group", GlobalConfig().Domain)
	})
}

func Test_InstallTemplate(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "__install_trpc_test")

	os.MkdirAll(dir, os.ModePerm)
	defer os.RemoveAll(dir)

	err := installTemplate(dir)
	if err != nil {
		t.Errorf("install Template error: %v", err)
	}
}

func fileExisted(fp string) bool {
	_, err := os.Lstat(fp)
	if err != nil {
		return false
	}
	return true
}

func mustBackupTemplate() {
	u, _ := user.Current()
	src := filepath.Join(u.HomeDir, ".trpc-cmdline-assets")
	dst := filepath.Join(os.TempDir(), "trpc.bak")

	_ = os.RemoveAll(dst)
	fs.Copy(src, dst)
}

func mustRestoreTemplate() {
	u, _ := user.Current()
	src := filepath.Join(u.HomeDir, ".trpc-cmdline-assets")
	dst := filepath.Join(os.TempDir(), "trpc.bak")

	_ = os.RemoveAll(src)
	fs.Copy(dst, src)
}
