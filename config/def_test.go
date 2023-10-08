package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-cmdline/util/semver"
)

func TestParseNewProtocVersion(t *testing.T) {
	s := versionRE.FindStringSubmatch("22.0")
	require.Equal(t, []string{"22.0", "", "22", "0"}, s)
	version := versionNumber(strings.Join(s[1:], "."))
	require.Equal(t, ".22.0", version)
	major, minor, revision := semver.Versions(version)
	require.Equal(t, 22, major)
	require.Equal(t, 0, minor)
	require.Equal(t, 0, revision)
}
