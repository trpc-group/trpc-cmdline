package main

import (
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/stretchr/testify/require"
)

func Test_readFromInputSource(t *testing.T) {
	t.Run("case invalid input", func(t *testing.T) {
		p := gomonkey.NewPatches()
		p.ApplyFunc(os.Lstat, func(name string) (os.FileInfo, error) {
			return nil, errors.New("fake error")
		})
		defer p.Reset()
		data, err := readFromInputSource("")
		require.NotNil(t, err)
		require.Nil(t, data)
	})

	t.Run("case success", func(t *testing.T) {
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		_, err = readFromInputSource(dir)
		require.Nil(t, err)
	})
}
