package pb

import (
	"errors"
	"os/exec"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/stretchr/testify/require"
)

func Test_oldVersion(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		args    args
		wantOld bool
	}{
		{"protoc-2.5.0", args{version: "2.5.0"}, true},
		{"protoc-2.6.0", args{version: "2.6.0"}, true},
		{"protoc-2.7.0", args{version: "2.7.0"}, true},
		{"protoc-3.5.0", args{version: "3.5.0"}, true},
		{"protoc-3.6.0", args{version: "3.6.0"}, false},
		{"protoc-3.6.1", args{version: "3.6.1"}, false},
		{"protoc-3.7.0", args{version: "3.7.0"}, false},
		{"protoc-3.7.1", args{version: "3.7.1"}, false},
		{"protoc-3.10.1", args{version: "3.10.1"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOld, err := oldVersion(tt.args.version)
			if err != nil {
				t.Errorf("judge protoc version error = %v", err)
			}
			if gotOld != tt.wantOld {
				t.Errorf("oldVersion() gotOld = %v, want %v", gotOld, tt.wantOld)
			}
		})
	}
}

func Test_isOldProtocVersion(t *testing.T) {
	t.Run("return version 3.15.0", func(t *testing.T) {
		p := gomonkey.ApplyFunc(protocVersion, func() (string, error) {
			return "3.15.0", nil
		})
		defer p.Reset()
		old, err := isOldProtocVersion()
		require.Nil(t, err)
		require.False(t, old)
	})

	t.Run("return version 2.5.0", func(t *testing.T) {
		p := gomonkey.ApplyFunc(protocVersion, func() (string, error) {
			return "2.5.0", nil
		})
		defer p.Reset()
		old, err := isOldProtocVersion()
		require.Nil(t, err)
		require.True(t, old)
	})

	t.Run("return version err", func(t *testing.T) {
		p := gomonkey.ApplyFunc(protocVersion, func() (string, error) {
			return "", errors.New("unexpected error")
		})
		defer p.Reset()
		_, err := isOldProtocVersion()
		require.NotNil(t, err)
	})

	t.Run("return version v2.5.0", func(t *testing.T) {
		p := gomonkey.ApplyFunc(protocVersion, func() (string, error) {
			return "v2.5.0", nil
		})
		defer p.Reset()
		old, err := isOldProtocVersion()
		require.Nil(t, err)
		require.True(t, old)
	})
}

func Test_protocVersion(t *testing.T) {
	t.Run("protoc not existed", func(t *testing.T) {
		p := gomonkey.ApplyFunc(exec.LookPath, func(p string) (string, error) {
			return "", errors.New("not found")
		})
		defer p.Reset()

		v, err := protocVersion()
		require.NotNil(t, err)
		require.Empty(t, v)
	})

	t.Run("protoc run", func(t *testing.T) {

		p := gomonkey.NewPatches()
		p.ApplyFunc(exec.LookPath, func(p string) (string, error) {
			return p, nil
		})
		defer p.Reset()

		t.Run("!success", func(t *testing.T) {
			cmd := &exec.Cmd{}
			p := gomonkey.ApplyMethod(reflect.TypeOf(cmd), "CombinedOutput", func(*exec.Cmd) ([]byte, error) {
				return nil, errors.New("permission denied")
			})
			defer p.Reset()

			v, err := protocVersion()
			require.NotNil(t, err)
			require.Empty(t, v)
		})

		t.Run("success", func(t *testing.T) {
			cmd := &exec.Cmd{}
			p := gomonkey.ApplyMethod(reflect.TypeOf(cmd), "CombinedOutput", func(*exec.Cmd) ([]byte, error) {
				return []byte("libprotoc 3.15.6"), nil
			})
			defer p.Reset()

			v, err := protocVersion()
			require.Nil(t, err)
			require.Equal(t, v, "3.15.6")
		})
	})

}

func Test_semanticVersion(t *testing.T) {
	t.Run("2.5.0 ok", func(t *testing.T) {
		a, b, c := semanticVersion("2.5.0")
		require.Equal(t, 2, a)
		require.Equal(t, 5, b)
		require.Equal(t, 0, c)
	})

	t.Run("2.5.z ok", func(t *testing.T) {
		a, b, c := semanticVersion("2.5.z")
		require.Equal(t, 2, a)
		require.Equal(t, 5, b)
		require.Equal(t, 0, c)
	})

	t.Run("2.5 ok", func(t *testing.T) {
		a, b, c := semanticVersion("2.5")
		require.Equal(t, 2, a)
		require.Equal(t, 5, b)
		require.Equal(t, 0, c)
	})

	t.Run("2 ok", func(t *testing.T) {
		a, b, c := semanticVersion("2")
		require.Equal(t, 2, a)
		require.Equal(t, 0, b)
		require.Equal(t, 0, c)
	})
}
