//go:build !windows
// +build !windows

package log

// Color directives.
const (
	ColorReset = "\033[0m"
	ColorGreen = "\033[1;32m"
	ColorRed   = "\033[1;31m"
	ColorPink  = "\033[1;35m"
)
