package paths

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocateProtoFile(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "paths.test_locate_protofile")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	subdir := filepath.Join(dir, "files")
	if err := os.MkdirAll(subdir, os.ModePerm); err != nil {
		panic(err)
	}
	f, err := os.Create(filepath.Join(subdir, "trpcx.proto"))
	if err != nil {
		panic(err)
	}
	f.Close()

	// Files in the directory
	p, err := Locate("trpcx.proto", subdir)
	require.Nil(t, err)
	require.Equal(t, subdir, p)
}
