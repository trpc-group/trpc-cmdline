// Package internal provides internal utilities for command to use.
package internal

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// RunAndWatch runs the command.
// It is used only in tests.
// The return string is the stdout/err of the executed command.
func RunAndWatch(cmd *cobra.Command, flags map[string]string, args []string) (string, error) {
	tmpd := filepath.Join(os.TempDir(), "trpc")
	tmpf := filepath.Join(tmpd, fmt.Sprintf("cmd_output-%d", rand.Uint64()))

	if err := os.MkdirAll(tmpd, os.ModePerm); err != nil {
		return "", fmt.Errorf("make dir %s err: %w", tmpd, err)
	}
	defer os.RemoveAll(tmpd)

	f, err := os.OpenFile(tmpf, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("open file %s err: %w", tmpf, err)
	}
	defer f.Close()

	// Redirect output and error.
	sout := os.Stdout
	serr := os.Stderr

	os.Stdout = f
	os.Stderr = f
	defer func() {
		os.Stdout = sout
		os.Stderr = serr
	}()

	for k, v := range flags {
		cmd.Flags().Set(k, v)
	}

	// PreRun.
	if cmd.PreRunE != nil {
		if err := cmd.PreRunE(cmd, args); err != nil {
			return "", err
		}
	} else if cmd.PreRun != nil {
		cmd.PreRun(cmd, args)
	}

	// Run.
	if cmd.RunE != nil {
		if err := cmd.RunE(cmd, args); err != nil {
			return "", err
		}
	} else if cmd.Run != nil {
		cmd.Run(cmd, args)
	}

	// PostRun.
	if cmd.PostRunE != nil {
		if err := cmd.PostRunE(cmd, args); err != nil {
			return "", err
		}
	} else if cmd.PostRun != nil {
		cmd.PostRun(cmd, args)
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("seek %s to start err: %w", tmpf, err)
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("read file %s err: %w", tmpf, err)
	}
	return string(b), nil
}
