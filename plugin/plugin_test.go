package plugin

import (
	"os"
	"testing"

	"trpc.group/trpc-go/trpc-cmdline/config"
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func setup() error {
	if _, err := config.Init(); err != nil {
		return err
	}
	deps, err := config.LoadDependencies()
	if err != nil {
		return err
	}
	return config.SetupDependencies(deps)
}
