// SPDX-FileCopyrightText: Copyright (C) 2025 Fundación Para La Transparencia del Software - STF
// SPDX-FileCopyrightText: 2025 Mariano Scasso <info@st.foundation>
//
// SPDX-License-Identifier:GPL-2.0-only

package main

// #cgo CFLAGS: -I.. -I../../inc -I../../external/inc
// #cgo LDFLAGS: ${SRCDIR}/../libsnippets_wrapper.a -lldb -lssl -lcrypto -lz -lm -lpthread
// #include "snippets_wrapper.h"
// #include <stdlib.h>
import "C"
import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Software-Transparency-Foundation/stf-plagicheck/deps"
	"github.com/Software-Transparency-Foundation/stf-plagicheck/models"
)

func main() {
	// Parse command line flags
	var debugMode bool
	ossDbName := flag.String("oss-db-name", "oss", "OSS database name")
	flag.BoolVar(&debugMode, "q", false, "Enable debug output")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: wfp_scanner [options] <wfp_file_path>")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExample WFP file format:")
		fmt.Println("file=e27b911d391391f94a862ebbe40ddcc0,1652,path/to/file.c")
		fmt.Println("1=63e9a57f")
		fmt.Println("3=e6f64278")
		fmt.Println("6=aa323afd,31466ee5,87dece99")
		os.Exit(0)
	}

	wfpFilePath := flag.Arg(0)

	// Initialize wrapper with database name and debug mode
	if !deps.SnippetWrapperInit(*ossDbName, debugMode) {
		log.Fatalf("Failed to initialize Snippet Wrapper with DB: %s", *ossDbName)
		os.Exit(1)
	}
	// Parse WFP file
	fmt.Printf("Parsing WFP file: %s\n", wfpFilePath)
	wfpData, err := deps.ParseWFPFile(wfpFilePath)
	if err != nil {
		log.Fatalf("Failed to parse WFP file: %v", err)
	}

	fmt.Printf("File: %s\n", wfpData.FilePath)
	fmt.Printf("MD5: %x\n", wfpData.MD5)
	fmt.Printf("Total lines: %d\n", wfpData.TotalLines)
	fmt.Printf("Number of hashes: %d\n", len(wfpData.Hashes))

	// Scan the WFP data
	fmt.Println("\nScanning snippets...")
	result, err := deps.ScanWFP(wfpData, debugMode)
	if err != nil {
		log.Fatalf("Failed to scan: %v", err)
	}

	// Print results
	fmt.Println("\n=== Scan Results ===")
	fmt.Printf("✓ Match Type: %s\n", result.MatchType)

	switch result.MatchType {
	case models.MatchFile:
		fmt.Println("  Complete file match found!")
	case models.MatchSnippet:
		fmt.Println("  Code snippet match found!")
	case models.MatchBinary:
		fmt.Println("  Binary file match found!")
	case models.MatchNone:
		fmt.Println("  No match found")
	}

	if result.ErrorMsg != "" {
		fmt.Printf("\nError: %s\n", result.ErrorMsg)
	}

	// Print matching MD5s
	if result.MatchCount > 0 {
		fmt.Printf("\n=== Matching Files (%d) ===\n", result.MatchCount)
		for _, match := range result.Matches {
			fmt.Printf("  %s (hits: %d)", match.FileMD5Hex, match.Hits)

			// Print ranges if available
			if len(match.Ranges) > 0 {
				fmt.Printf(" - ranges: ")
				for i, r := range match.Ranges {
					if i > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%d-%d", r.From, r.To)
				}
			}
			fmt.Println()
		}
	}
}
