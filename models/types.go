// SPDX-FileCopyrightText: Copyright (C) 2025 Fundación Para La Transparencia del Software - STF
// SPDX-FileCopyrightText: 2025 Mariano Scasso <info@st.foundation>
//
// SPDX-License-Identifier: GPL-2.0

package models

// WFPData represents the parsed data from a WFP file
type WFPData struct {
	MD5        [16]byte
	MD5Hex     string // Hexadecimal version of MD5 for compatibility
	TotalLines int
	FilePath   string
	Hashes     []uint32
	Lines      []uint32
}

// MatchResult represents a match result (for JSON output)
type MatchResult struct {
	MatchType     string  `json:"match_type"`
	TargetLines   string  `json:"target_lines,omitempty"`
	SourceLines   string  `json:"ref_file_lines,omitempty"`
	Instances     int     `json:"instances"`
	ReferenceURL  string  `json:"reference_url"`
	ReferenceFile string  `json:"reference_file"`
	Hits          int     `json:"-"` // For internal use (not exported in JSON)
	Ranges        []Range `json:"-"` // For internal use (not exported in JSON)
}

// MatchInfo contains information about an individual match (internal use)
type MatchInfo struct {
	FileMD5Hex string
	Hits       int
	Ranges     []Range
}

// Range represents a line range in a match
type Range struct {
	From int
	To   int
	Oss  int
}

// ScanResult contains the complete results of a scan
type ScanResult struct {
	MatchType  MatchType
	MatchCount int
	Matches    []MatchInfo
	ErrorMsg   string
}

// MatchType represents the type of match found
type MatchType int

const (
	MatchNone    MatchType = 0
	MatchFile    MatchType = 1
	MatchSnippet MatchType = 2
	MatchBinary  MatchType = 3
)

// String returns a human-readable representation of the MatchType
func (m MatchType) String() string {
	switch m {
	case MatchFile:
		return "FILE"
	case MatchSnippet:
		return "SNIPPET"
	case MatchBinary:
		return "BINARY"
	case MatchNone:
		return "NONE"
	default:
		return "UNKNOWN"
	}
}
