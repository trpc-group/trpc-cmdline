// Package fs implements cross-platform file system capabilities.
package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"trpc.group/trpc-go/trpc-cmdline/util/lang"
	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

// BaseNameWithoutExt return basename without extension of `filename`,
// in which `filename` may contains directory.
func BaseNameWithoutExt(filename string) string {
	return lang.TrimRight(".", filepath.Base(filename))
}

// LocateFile returns the absolute path of proto file.
//
// To ensure that protofile can be found,
// protodirs needs to provide the parent path of protofile and cannot be the parent path's parent path.
func LocateFile(protofile string, protodirs []string) (string, error) {
	if filepath.IsAbs(protofile) {
		_, err := os.Lstat(protofile)
		if err != nil {
			return "", fmt.Errorf("protofile stat %s err: %w", protofile, err)
		}
		return protofile, nil
	}

	// always add current directory into search dirs
	abs, err := filepath.Abs(".")
	if err != nil {
		return "", fmt.Errorf("filepath.Abs . err: %w", err)
	}

	// If we can find the protofile under the current directory,
	// directly return to prevent possible conflicts resulting from the relative path in protofile.
	// Reference: https://mk.woa.com/q/286784
	fp := filepath.Join(abs, protofile)
	if info, err := os.Lstat(fp); err == nil && !info.IsDir() {
		return fp, nil
	}

	protodirs = append(protodirs, abs)
	protodirs = UniqFilePath(protodirs)

	// Find the absolute path of protofile.
	log.Debug("protocolfile: %s", protofile)
	log.Debug("protodirs: %s", protodirs)
	fpaths, err := getPbFilePathList(protofile, protodirs)
	if err != nil {
		return "", fmt.Errorf("get pb file path listerr: %w", err)
	}

	// `-protofile=abc/d.proto`, works like `-protodir=abc -protofile=d.proto`
	return filepath.Abs(fpaths[0])
}

func getPbFilePathList(protofile string, dirs []string) ([]string, error) {
	fpaths := []string{}
	for _, dir := range dirs {
		fp := filepath.Join(dir, protofile)
		inf, err := os.Lstat(fp)
		if err == nil && !inf.IsDir() {
			fpaths = append(fpaths, fp)
		}
	}
	if len(fpaths) == 0 {
		return nil, fmt.Errorf("%s not found in dirs: %v", protofile, dirs)
	}
	if len(fpaths) > 1 {
		return nil, fmt.Errorf("%s found duplicate ones: %v", protofile, fpaths)
	}
	return fpaths, nil
}

// UniqFilePath is used to deduplicate file paths in a slice.
func UniqFilePath(dirs []string) []string {
	set := map[string]struct{}{}
	for _, p := range dirs {
		abs, _ := filepath.Abs(p)
		set[abs] = struct{}{}
	}

	uniq := []string{}
	for dir := range set {
		uniq = append(uniq, dir)
	}
	sort.Strings(uniq)

	return uniq
}

// PrepareOutputdir create outputdir if it doesn't exist,
// return error if `outputdir` existed while it is not a directory,
// return error if any other error occurs.
func PrepareOutputdir(outputdir string) (err error) {
	fin, err := os.Lstat(outputdir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return os.MkdirAll(outputdir, os.ModePerm)
	}

	if !fin.IsDir() {
		return fmt.Errorf("target %s existed, but not a directory", outputdir)
	}

	return nil
}
