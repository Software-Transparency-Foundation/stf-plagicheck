// SPDX-FileCopyrightText: Copyright (C) 2025 Fundación Para La Transparencia del Software - STF
// SPDX-FileCopyrightText: 2025 Mariano Scasso <info@st.foundation>
//
// SPDX-License-Identifier: GPL-2.0

package pkg

import (
	"testing"

	"scanoss.com/openkb-engine/models"
)

func TestMergeRanges(t *testing.T) {
	tests := []struct {
		name      string
		ranges    []models.Range
		tolerance int
		expected  int // expected number of merged ranges
	}{
		{
			name:      "empty ranges",
			ranges:    []models.Range{},
			tolerance: 3,
			expected:  0,
		},
		{
			name: "overlapping ranges",
			ranges: []models.Range{
				{From: 10, To: 20, Oss: 5},
				{From: 15, To: 25, Oss: 10},
			},
			tolerance: 3,
			expected:  1,
		},
		{
			name: "ranges within tolerance",
			ranges: []models.Range{
				{From: 10, To: 20, Oss: 5},
				{From: 22, To: 30, Oss: 15},
			},
			tolerance: 3,
			expected:  1,
		},
		{
			name: "ranges beyond tolerance",
			ranges: []models.Range{
				{From: 10, To: 20, Oss: 5},
				{From: 30, To: 40, Oss: 25},
			},
			tolerance: 3,
			expected:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeRanges(tt.ranges, tt.tolerance)
			if len(result) != tt.expected {
				t.Errorf("expected %d ranges, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestFilterValidRanges(t *testing.T) {
	tests := []struct {
		name     string
		ranges   []models.Range
		expected int // expected number of valid ranges
	}{
		{
			name:     "empty ranges",
			ranges:   []models.Range{},
			expected: 0,
		},
		{
			name: "all single-line ranges",
			ranges: []models.Range{
				{From: 10, To: 10, Oss: 5},
				{From: 20, To: 20, Oss: 15},
			},
			expected: 0,
		},
		{
			name: "all multi-line ranges",
			ranges: []models.Range{
				{From: 10, To: 20, Oss: 5},
				{From: 30, To: 40, Oss: 25},
			},
			expected: 2,
		},
		{
			name: "mixed ranges",
			ranges: []models.Range{
				{From: 10, To: 10, Oss: 5},
				{From: 20, To: 30, Oss: 15},
				{From: 40, To: 40, Oss: 35},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterValidRanges(tt.ranges)
			if len(result) != tt.expected {
				t.Errorf("expected %d valid ranges, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestFormatRanges(t *testing.T) {
	ranges := []models.Range{
		{From: 10, To: 20, Oss: 5},
		{From: 30, To: 40, Oss: 25},
	}

	target, source := FormatRanges(ranges)

	expectedTarget := "10-20,30-40"
	expectedSource := "5-15,25-35"

	if target != expectedTarget {
		t.Errorf("expected target '%s', got '%s'", expectedTarget, target)
	}

	if source != expectedSource {
		t.Errorf("expected source '%s', got '%s'", expectedSource, source)
	}
}

func TestReadWFPFile(t *testing.T) {
	// Test with file_match.wfp
	entries, err := ReadWFPFile("../test/file_match.wfp")
	if err != nil {
		t.Fatalf("failed to read WFP file: %v", err)
	}

	if len(entries) == 0 {
		t.Error("expected at least one entry")
	}

	// Verify first entry has required fields
	if entries[0].MD5Hex == "" {
		t.Error("expected MD5Hex to be set")
	}

	if entries[0].FilePath == "" {
		t.Error("expected FilePath to be set")
	}
}
