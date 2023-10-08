// Package log encapsulates logging functionalities of the project.
package log

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	logVerbose bool
	logPrefix  string
)

// SetVerbose set logging level
func SetVerbose(verbose bool) {
	logVerbose = verbose
}

// SetPrefix set logging prefix
func SetPrefix(prefix string) {
	logPrefix = prefix
}

// Info print logging info at level INFO, if flag verbose true, filename and lineno will be logged.
func Info(format string, vals ...interface{}) {
	fileno, _ := callerAddress(3)
	if logVerbose {
		fmt.Printf("%s%s[Info][%s] %s%s\n", ColorGreen, logPrefix, fileno, fmt.Sprintf(format, vals...), ColorReset)
	} else {
		fmt.Printf("%s%s %s%s\n", ColorGreen, logPrefix, fmt.Sprintf(format, vals...), ColorReset)
	}
}

// Debug print logging info at level DEBUG, if flag verbose true, filename and lineno will be logged.
func Debug(format string, vals ...interface{}) {
	fileno, _ := callerAddress(3)
	if logVerbose {
		fmt.Printf("%s%s[Debug][%s] %s%s\n", ColorPink, logPrefix, fileno, fmt.Sprintf(format, vals...), ColorReset)
	}
}

// Error print logging info at level ERROR, if flag verbose true, filename and lineno will be logged.
func Error(format string, vals ...interface{}) {
	fileno, _ := callerAddress(3)
	if logVerbose {
		fmt.Printf("%s%s[Error][%s] %s%s\n", ColorRed, logPrefix, fileno, fmt.Sprintf(format, vals...), ColorReset)
	} else {
		fmt.Printf("%s%s %s%s\n", ColorRed, logPrefix, fmt.Sprintf(format, vals...), ColorReset)
	}
}

// callerAddress skip N level to get the caller's filename and lineno, if no caller return error.
func callerAddress(skip int) (string, error) {
	fpcs := make([]uintptr, 1)
	// Skip N levels to get the caller
	n := runtime.Callers(skip, fpcs)
	if n == 0 {
		return "", fmt.Errorf("MSG: NO CALLER")
	}

	caller := runtime.FuncForPC(fpcs[0] - 1)
	if caller == nil {
		return "", fmt.Errorf("MSG: CALLER IS NIL")
	}

	// Print the file name and line number
	fileName, lineNo := caller.FileLine(fpcs[0] - 1)
	baseName := fileName[strings.LastIndex(fileName, "/")+1:]

	return fmt.Sprintf("%s:%d", baseName, lineNo), nil
}
