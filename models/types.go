// SPDX-FileCopyrightText: Copyright (C) 2025 Fundación Para La Transparencia del Software - STF
// SPDX-FileCopyrightText: 2025 Mariano Scasso <info@st.foundation>
//
// SPDX-License-Identifier: GPL-2.0

package models

// WFPData representa los datos parseados de un archivo WFP
type WFPData struct {
	MD5        [16]byte
	MD5Hex     string // Versión hexadecimal del MD5 para compatibilidad
	TotalLines int
	FilePath   string
	Hashes     []uint32
	Lines      []uint32
}

// MatchResult representa un resultado de coincidencia (para salida JSON)
type MatchResult struct {
	MatchType     string  `json:"match_type"`
	TargetLines   string  `json:"target_lines,omitempty"`
	SourceLines   string  `json:"ref_file_lines,omitempty"`
	Instances     int     `json:"instances"`
	ReferenceURL  string  `json:"reference_url"`
	ReferenceFile string  `json:"reference_file"`
	Hits          int     `json:"-"` // Para uso interno (no se exporta en JSON)
	Ranges        []Range `json:"-"` // Para uso interno (no se exporta en JSON)
}

// MatchInfo contiene información sobre una coincidencia individual (uso interno)
type MatchInfo struct {
	FileMD5Hex string
	Hits       int
	Ranges     []Range
}

// Range representa un rango de líneas en una coincidencia
type Range struct {
	From int
	To   int
	Oss  int
}

// ScanResult contiene los resultados completos de un scan
type ScanResult struct {
	MatchType  MatchType
	MatchCount int
	Matches    []MatchInfo
	ErrorMsg   string
}

// MatchType representa el tipo de coincidencia encontrada
type MatchType int

const (
	MatchNone    MatchType = 0
	MatchFile    MatchType = 1
	MatchSnippet MatchType = 2
	MatchBinary  MatchType = 3
)

// String devuelve una representación legible del MatchType
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
