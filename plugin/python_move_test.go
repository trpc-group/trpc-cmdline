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
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/parser"
)

func TestPlugin_PythonMove(t *testing.T) {
	u := &PythonMove{}
	opt := params.Option{
		Language:  "python",
		OutputDir: t.TempDir(),
	}
	t.Run("name", func(t *testing.T) {
		require.Equal(t, "move", u.Name())
	})
	t.Run("lang !python", func(t *testing.T) {
		opt := opt
		opt.Language = "go"
		require.False(t, u.Check(nil, &opt))
	})
	t.Run("lang python", func(t *testing.T) {
		require.True(t, u.Check(nil, &opt))
		t.Run("getpackage error", func(t *testing.T) {
			p := gomonkey.ApplyFunc(parser.GetPackage, func(*descriptor.FileDescriptor, string) (string, error) {
				return "", errors.New("getpackage error")
			})
			defer p.Reset()
			require.NotNil(t, u.Run(nil, &opt))
		})
		t.Run("move setup.py error", func(t *testing.T) {
			p := gomonkey.NewPatches()
			p.ApplyFunc(parser.GetPackage, func(*descriptor.FileDescriptor, string) (string, error) {
				return "xxxx", nil
			})
			p.ApplyFunc(moveSetupFile, func(string, string) error {
				return errors.New("error")
			})
			p.ApplyFunc(fixPBImport, func(string, string) error {
				return nil
			})
			defer p.Reset()
			require.NotNil(t, u.Run(nil, &opt))
		})
		t.Run("move setup.py ok, fix pb import error", func(t *testing.T) {
			p := gomonkey.NewPatches()
			pbPackage := "trpc_soho_team"
			p.ApplyFunc(parser.GetPackage, func(*descriptor.FileDescriptor, string) (string, error) {
				return pbPackage, nil
			})
			p.ApplyFunc(fixPBImport, func(string, string) error {
				return errors.New("error")
			})
			defer p.Reset()
			stubDir := filepath.Join(opt.OutputDir, "stub")
			pbOutDir := filepath.Join(stubDir, pbPackage)
			require.Nil(t, os.MkdirAll(pbOutDir, os.ModePerm))
			require.Nil(t, os.WriteFile(filepath.Join(pbOutDir, "setup.py"), []byte("..."), 0666))
			require.NotNil(t, u.Run(nil, &opt))
			require.FileExists(t, filepath.Join(stubDir, "setup.py"))
		})
		t.Run("fix pb import", func(t *testing.T) {
			p := gomonkey.NewPatches()
			pbPackage := "trpc_soho_team"
			p.ApplyFunc(parser.GetPackage, func(*descriptor.FileDescriptor, string) (string, error) {
				return pbPackage, nil
			})
			p.ApplyFunc(moveSetupFile, func(string, string) error {
				return nil
			})
			defer p.Reset()
			// stub/trpc_soho_team/
			pbOutDir := filepath.Join(opt.OutputDir, "stub", pbPackage)
			require.Nil(t, os.MkdirAll(pbOutDir, os.ModePerm))
			require.Nil(t, os.WriteFile(filepath.Join(pbOutDir, "__init__.py"), []byte("# init"), 0666))
			serviceFile := filepath.Join(pbOutDir, "app_pb2.py")
			require.Nil(t, os.WriteFile(serviceFile, []byte("from mpc.common import mpc_comm_pb2"), 0666))
			// stub/trpc_soho_team/mpc/common
			commonDir := filepath.Join(pbOutDir, "mpc/common")
			require.Nil(t, os.MkdirAll(commonDir, os.ModePerm))
			dependencyFile := filepath.Join(commonDir, "mpc_comm_pb2.py")
			require.Nil(t, os.WriteFile(dependencyFile,
				[]byte("from google.protobuf import message as _message"), 0666))
			pbFile := filepath.Join(pbOutDir, "pb.py")
			require.Nil(t, os.WriteFile(pbFile, []byte("from trpc.common.validate_pb2 import *"), 0666))

			require.Nil(t, u.Run(nil, &opt))

			content, err := os.ReadFile(pbFile)
			require.Nil(t, err)
			require.Equal(t, "from trpc_pb.trpc.common.validate_pb2 import *", string(content))
			content, err = os.ReadFile(serviceFile)
			require.Nil(t, err)
			require.Equal(t, "from trpc_soho_team.mpc.common import mpc_comm_pb2", string(content))
			require.FileExists(t, filepath.Join(pbOutDir, "mpc", "__init__.py"))
			require.FileExists(t, filepath.Join(commonDir, "__init__.py"))
		})
		t.Run("prepareStubPackageOutputDir error", func(t *testing.T) {
			p := gomonkey.ApplyFunc(parser.GetPackage, func(*descriptor.FileDescriptor, string) (string, error) {
				return "xxxx", nil
			})
			p.ApplyFunc(prepareStubPackageOutputDir, func(string, string) (string, string, error) {
				return "", "", errors.New("prepareStubPackageOutputDir error")
			})
			defer p.Reset()
			require.NotNil(t, u.Run(nil, &opt))
		})
		t.Run("prepareStubPackageOutputDir read dir error", func(t *testing.T) {
			p := gomonkey.ApplyFunc(parser.GetPackage, func(*descriptor.FileDescriptor, string) (string, error) {
				return "xxxx", nil
			})
			p.ApplyFunc(os.Stat, func(string) (os.FileInfo, error) {
				return nil, errors.New("dir not exist")
			})
			p.ApplyFunc(os.ReadDir, func(string) ([]os.DirEntry, error) {
				return nil, errors.New("fail to read dir")
			})
			defer p.Reset()
			require.NotNil(t, u.Run(nil, &opt))
		})
		t.Run("prepareStubPackageOutputDir mkdir error", func(t *testing.T) {
			p := gomonkey.ApplyFunc(parser.GetPackage, func(*descriptor.FileDescriptor, string) (string, error) {
				return "xxxx", nil
			})
			p.ApplyFunc(os.Stat, func(string) (os.FileInfo, error) {
				return nil, errors.New("dir not exist")
			})
			p.ApplyFunc(os.MkdirAll, func(string, os.FileMode) error {
				return errors.New("fail to mkdir")
			})
			defer p.Reset()
			require.NotNil(t, u.Run(nil, &opt))
		})
		t.Run("prepareStubPackageOutputDir success", func(t *testing.T) {
			pbPackage := "trpc_demo_hello"
			p := gomonkey.NewPatches()
			p.ApplyFunc(parser.GetPackage, func(*descriptor.FileDescriptor, string) (string, error) {
				return pbPackage, nil
			})
			p.ApplyFunc(moveSetupFile, func(string, string) error {
				return nil
			})
			p.ApplyFunc(fixPBImport, func(string, string) error {
				return nil
			})
			defer p.Reset()
			require.Nil(t, os.MkdirAll(filepath.Join(opt.OutputDir, "test"), os.ModePerm))
			initFile := filepath.Join(opt.OutputDir, "__init__.py")
			require.Nil(t, os.WriteFile(initFile, []byte("# -*- coding: utf-8 -*-"), 0666))

			require.Nil(t, u.Run(nil, &opt))
			require.NoFileExists(t, initFile)
			require.FileExists(t, filepath.Join(opt.OutputDir, "stub", pbPackage, "__init__.py"))
			require.DirExists(t, filepath.Join(opt.OutputDir, "test"))
		})
	})
}

func TestFixTRPCPBImportPath(t *testing.T) {
	content := `from trpc.common import validate_pb2 as trpc_dot_common_dot_validate__pb2
from trpc.mcp import mcp_pb2 as trpc_dot_mcp_dot_mcp__pb2
import trpc.api.annotations_pb2 as trpc_dot_api_dot_annotations__pb2
from google.protobuf import descriptor as _descriptor
from trpc_yuanbao import service_pb2`

	expected := `from trpc_pb.trpc.common import validate_pb2 as trpc_dot_common_dot_validate__pb2
from trpc_pb.trpc.mcp import mcp_pb2 as trpc_dot_mcp_dot_mcp__pb2
import trpc_pb.trpc.api.annotations_pb2 as trpc_dot_api_dot_annotations__pb2
from google.protobuf import descriptor as _descriptor
from trpc_yuanbao import service_pb2`

	require.Equal(t, expected, fixTRPCPBImportPath(content))
}
