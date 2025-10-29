// SPDX-FileCopyrightText: Copyright (C) 2025 Fundación Para La Transparencia del Software - STF
// SPDX-FileCopyrightText: 2025 Mariano Scasso <info@st.foundation>
//
// SPDX-License-Identifier: GPL-2.0

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
