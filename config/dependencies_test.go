package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_VersionRE(t *testing.T) {
	type arg struct {
		ver  string
		want []string
	}
	args := []arg{
		{
			ver:  "0.1.1",
			want: []string{"0", "1", "1"},
		},
		{
			ver:  "1.1.1",
			want: []string{"1", "1", "1"},
		},
		{
			ver:  "10.1.1",
			want: []string{"10", "1", "1"},
		},
		{
			ver:  "v10.1.1",
			want: []string{"10", "1", "1"},
		},
		{
			ver:  "protoc v10.1.1",
			want: []string{"10", "1", "1"},
		},
		{
			ver:  "protoc v10.1.1-beta",
			want: []string{"10", "1", "1"},
		},
	}

	for _, a := range args {
		vals := versionRE.FindStringSubmatch(a.ver)
		require.Equal(t, a.want, vals[1:])
	}
}
