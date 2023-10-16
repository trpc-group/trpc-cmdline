// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package sync

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

// FileManager is an interface for managing operating system files.
type FileManager interface {
	RemoveAll(path string) error
	WalkDir(root string, fn fs.WalkDirFunc) error
	MkdirAll(path string, perm os.FileMode) error
	UserHomeDir() (string, error)

	Open(name string) (*os.File, error)
	Create(name string) (*os.File, error)
	Close(*os.File)
	Copy(dst io.Writer, src io.Reader) (written int64, err error)
}

type defaultFileManager struct{}

// DefaultFileManager is a default file manager constructor which creates a FileManager.
var DefaultFileManager = &defaultFileManager{}

// RemoveAll removes path and any children it contains.
func (d *defaultFileManager) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// WalkDir walks the file tree rooted at root, calling fn for each file or
// directory in the tree, including root.
func (d *defaultFileManager) WalkDir(root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
}

// MkdirAll creates a directory named path
func (d *defaultFileManager) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// UserHomeDir returns the current user's home directory.
func (d *defaultFileManager) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

// Open opens the named file for reading.
func (d *defaultFileManager) Open(name string) (*os.File, error) {
	return os.Open(name)
}

// Create creates or truncates the named file.
func (d *defaultFileManager) Create(name string) (*os.File, error) {
	return os.Create(name)
}

// Close closes the File, rendering it unusable for I/O.
func (d *defaultFileManager) Close(file *os.File) {
	if err := file.Close(); err != nil {
		log.Error("default file manager close error%v, file:%s", err, file.Name())
	}
	return
}

// Copy copies from src to dst until either EOF is reached
// on src or an error occurs. It returns the number of bytes
// copied and the first error encountered while copying, if any.
func (d *defaultFileManager) Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	return io.Copy(dst, src)
}
