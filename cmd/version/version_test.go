package version

import (
	"strings"
	"testing"

	"trpc.group/trpc-go/trpc-cmdline/cmd/internal"
	"trpc.group/trpc-go/trpc-cmdline/config"
)

func TestCmd_Version(t *testing.T) {
	if _, err := config.Init(); err != nil {
		t.Errorf("config init error: %v", err)
	}

	versionCmd := CMD()
	output, err := internal.RunAndWatch(versionCmd, nil, nil)
	if err != nil {
		t.Errorf("versionCmd run and watch error: %v", err)
	}

	if !strings.Contains(output, config.TRPCCliVersion) {
		t.Errorf("versionCmd.Run() output version mismatch")
	}
}
