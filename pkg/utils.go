package pkg

import (
	"fmt"
	"os"
)

var debugMode bool = false

// SetDebugMode enables or disables debug messages
func SetDebugMode(enabled bool) {
	debugMode = enabled
}

// DebugLog prints debug messages to stderr if debug mode is enabled
func DebugLog(format string, args ...interface{}) {
	if debugMode {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format, args...)
	}
}
