// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package fs

import (
	"os"
	"path/filepath"
	"syscall"
)

const lstat = "lstat"

// Move move `src` to `dest`
//
// the behavior of fs.Move is consistent with bash shell `mv` command:
//
// when move a file, actions are following:
// ------------------------------------------------------------------------------------------------
// | No. | src existed | src type | dst existed | dst type | behavior                             |
// ------------------------------------------------------------------------------------------------
// | 1   | False       | -        | -           | -        | error: No such file or directory     |
// ------------------------------------------------------------------------------------------------
// | 2   | True        | File     | False       | -        | if dir(dst) existed:                 |
// |     |             |          |             |          | - Yes, is dir, mv `src` to dir(dst)  |
// |     |             |          |             |          | - Yes, not dir, err: Not a directory |
// |     |             |          |             |          | - No, err: No such file or directory |
// ------------------------------------------------------------------------------------------------
// | 3   | True        | File     | True        | Folder   | if dst/basename(src) existed:        |
// |     |             |          |             |          | - Yes, mv `src` to dst/basename(src) |
// |     |             |          |             |          | - No, mv `src` to dst/basename(src)  |
// ------------------------------------------------------------------------------------------------
// | 4   | True        | File     | True        | File     | mv `src` to dst                      |
// ------------------------------------------------------------------------------------------------
//
// when move a directory, actions are following:
// ------------------------------------------------------------------------------------------------
// | 5   | True        | Folder   | False       | -        | if dir(dst) existed:                 |
// |     |             |          |             |          | - Yes, is dir, mv `src` to dir(dst)  |
// |     |             |          |             |          | - Yes, not dir, err: File Exists     |
// |     |             |          |             |          | - No, err: No such file or directory |
// ------------------------------------------------------------------------------------------------
// | 6   | True        | Folder   | True        | File     | error: File Already Existed          |
// ------------------------------------------------------------------------------------------------
// | 7   | True        | Folder   | True        | Folder   | t = dst/basename(src), if t existed: |
// |     |             |          |             |          | - Yes, t empty, mv src to t          |
// |     |             |          |             |          | -      t notempty, err: t Not empty  |
// |     |             |          |             |          | - No, mv src to t                    |
// ------------------------------------------------------------------------------------------------
//
// Why keep the behavior consistent? It makes the usage much more friendly when it behaves as users expected.
func Move(src, dst string) error {
	var (
		inf os.FileInfo
		err error
	)

	// check whether `src` is valid or not
	if inf, err = os.Lstat(src); err != nil {
		return err
	}

	// move directory
	if inf.IsDir() {
		return moveDirectory(src, dst)
	}

	// move file
	return moveFile(src, dst)
}

// moveFile move a file `src` to `dst`
//
// `src` is a normal file, dst can be a file or directory.
// 1. if `dst` not existed
//   - if dir(dst) existed and is a directory, then move `src` under dir(dst),
//   - if dir(dst) existed and not a directory, return err: &PathError(Op: "lstat", Path: dir(dstErr:), syscall.EEXIST}
//   - if dir(dst) not existed, return err: &PathError(Op: "lstat", Path: dir(dstErr:), os.ENOENT}
//
// 2. if `dst` existed
// - if dst is a normal file, rename src to dst
// - if dst is a folder, rename src to dst/basename(src)
func moveFile(src, dst string) error {
	dstInf, err := os.Lstat(dst)

	// if dst existed
	if err == nil {
		if !dstInf.IsDir() {
			return Rename(src, dst)
		}

		p := filepath.Join(dst, filepath.Base(src))
		return Rename(src, p)
	}

	if !os.IsNotExist(err) {
		return err
	}

	// if dst not existed
	p := filepath.Dir(dst)
	if inf, err := os.Lstat(p); err != nil {
		return err
	} else {
		// p is a symlink to valid directory
		if inf.Mode()&os.ModeSymlink != 0 {
			yes, err := isSymLinkToDir(p)
			if err != nil {
				return err
			}
			if !yes {
				return &os.PathError{Op: "lstat", Path: p, Err: syscall.EEXIST}
			}
			return Rename(src, dst)
		}
		// p isn't directory, neither
		if !inf.IsDir() {
			return &os.PathError{Op: "lstat", Path: p, Err: syscall.EEXIST}
		}
		return Rename(src, dst)
	}
}

// isSymLinkToDir check if symlink `p` points to a valid directory
func isSymLinkToDir(symlink string) (yes bool, err error) {
	linkTo, err := filepath.EvalSymlinks(symlink)
	if err != nil {
		return
	}

	fin, err := os.Lstat(linkTo)
	if err != nil {
		return
	}

	if !fin.IsDir() {
		return false, nil
	}
	return true, nil
}

// moveDirectory move a directory `src` to `dst`
//
// `src` is a directory, dst should always be a directory.
// 1. if `dst` existed
//   - if `dst` is not a directory, return error &PathError{Op: "lstat", Path: dst, Err: os.EEXIST}
//   - if `dst` is a directory
//   - if dst/basename(src) is empty, then rename src to dst/basename(src)
//   - if dst/basename(src) not empty, return error &PathError{Op: "lstat", Path: dst, Err: syscall.ENOTEMPTY}
//
// 2. if `dst` not existed
//   - if dir(dst) existed, rename src to dst
//   - if dir(dst) not existed, return error &PathError{Op: "lstat", Path: dst, Err: syscall.ENOENT}
func moveDirectory(src, dst string) error {
	dstInf, err := os.Lstat(dst)

	// if dst existed
	if err == nil {

		if !dstInf.IsDir() {
			return &os.PathError{Op: "lstat", Path: dst, Err: syscall.EEXIST}
		}

		target := filepath.Join(dst, filepath.Base(src))
		inf, err := os.Lstat(target)
		if err != nil {
			if os.IsNotExist(err) {
				return Rename(src, target)
			}
			return err
		}

		if !inf.IsDir() {
			return &os.PathError{Op: "lstat", Path: target, Err: syscall.EEXIST}
		}

		files, err := os.ReadDir(target)
		if err != nil {
			return err
		}
		if len(files) != 0 {
			return &os.PathError{Op: "lstat", Path: target, Err: syscall.ENOTEMPTY}
		}

		if err = os.RemoveAll(target); err != nil {
			return err
		}
		return Rename(src, target)
	}

	// if dst not existed
	inf, err := os.Lstat(filepath.Dir(dst))
	if err != nil {
		return err
	}
	if !inf.IsDir() {
		return &os.PathError{Op: "lstat", Path: filepath.Base(dst), Err: syscall.EEXIST}
	}
	return Rename(src, dst)
}

func coverNoFileTarget(src string, target string) error {
	files, err := os.ReadDir(target)
	if err != nil {
		return err
	}
	if len(files) != 0 {
		return &os.PathError{Op: lstat, Path: target, Err: syscall.ENOTEMPTY}
	}

	if err = os.RemoveAll(target); err != nil {
		return err
	}
	if err = Copy(src, target); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

// Rename fs.Rename is just a wrapper of syscall rename,
// it may fail if renamimg across different devices, so we must provide
// a solution like: copy(src,dst)+rm(dst).
//
// see: https://groups.google.com/g/golang-dev/c/5w7Jmg_iCJQ
func Rename(src, dst string) error {
	var err error
	if src, err = filepath.Abs(src); err != nil {
		return err
	}
	if dst, err = filepath.Abs(dst); err != nil {
		return err
	}
	if src == dst {
		return nil
	}
	if err := Copy(src, dst); err != nil {
		return err
	}
	return os.RemoveAll(src)
}
