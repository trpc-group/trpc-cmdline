package log

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInfo(t *testing.T) {
	SetVerbose(true)
	require.Equal(t, logVerbose, true)
	Info("log content is: %s", "message")
	Debug("log content is: %s", "message")
	Error("log content is: %s", "message")

	SetVerbose(false)
	require.Equal(t, logVerbose, false)
	Info("log content is: %s", "message")
	Debug("log content is: %s", "message")
	Error("log content is: %s", "message")
}
