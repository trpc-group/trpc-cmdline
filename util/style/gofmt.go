// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package style provides formatting functions.
package style

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"

	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

// Format the code in place.
func Format(fpath string, lang string) error {
	switch lang {
	case "go":
		return GoFmt(fpath)
	default:
		return nil
	}
}

// GoFmt formats go code in place.
func GoFmt(fpath string) error {
	fin, err := os.ReadFile(fpath)
	if err != nil {
		return err
	}

	buf, err := format.Source(fin)
	if err != nil {
		return fmt.Errorf("format error: %v", err)
	}
	return os.WriteFile(fpath, buf, 0644)
}

// GoFmtDir formats Go code in-place in a directory.
func GoFmtDir(dir string) error {
	err := filepath.Walk(dir, func(fpath string, info os.FileInfo, err error) error {
		if strings.HasSuffix(fpath, ".go") && !info.IsDir() {
			err := GoFmt(fpath)
			if err != nil {
				log.Error("Warn: style file:%s error:%v", fpath, err)
			}
		}
		if info.IsDir() && dir != fpath {
			return GoFmtDir(fpath)
		}
		return nil
	})
	return err
}
