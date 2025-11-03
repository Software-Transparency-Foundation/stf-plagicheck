// SPDX-FileCopyrightText: Copyright (C) 2025 Fundación Para La Transparencia del Software - STF
// SPDX-FileCopyrightText: 2025 Mariano Scasso <info@st.foundation>
//
// SPDX-License-Identifier: GPL-2.0

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Software-Transparency-Foundation/stf-plagicheck/pkg"
	"github.com/schollz/progressbar/v3"
)

var (
	kbName  string = "osskb-core"
	version string = "dev"
	commit  string = "unknown"
)

// progressWriter captures progress messages and updates a progress bar
type progressWriter struct {
	bar *progressbar.ProgressBar
}

func (pw *progressWriter) Write(p []byte) (n int, err error) {
	// Parse "progress:N/M" messages
	msg := strings.TrimSpace(string(p))
	if strings.HasPrefix(msg, "progress:") {
		parts := strings.Split(strings.TrimPrefix(msg, "progress:"), "/")
		if len(parts) == 2 {
			current, err1 := strconv.Atoi(parts[0])
			total, err2 := strconv.Atoi(parts[1])
			if err1 == nil && err2 == nil {
				if pw.bar == nil {
					pw.bar = progressbar.Default(int64(total))
				}
				pw.bar.Set(current)
			}
		}
	}
	return len(p), nil
}

func main() {
	generateMode := flag.Bool("fp", false, "Generate WFP from file or directory (output only, no scan)")
	outputFile := flag.String("output", "", "Output file for generated WFP (optional, default: stdout)")
	minHits := flag.Int("min-hits", 3, "Minimum number of hits required for valid snippet match (default: 3)")
	numThreads := flag.Int("T", 3, "Number of parallel threads for processing files (default: 3)")
	debugMode := flag.Bool("d", false, "Enable debug mode (show detailed processing information)")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Set debug mode
	pkg.SetDebugMode(*debugMode)

	if *showVersion {
		fmt.Printf("plagicheck version %s (commit: %s)\n", version, commit)
		os.Exit(0)
	}

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [-fp] [--output <file>] [--min-hits <N>] [-T <threads>] [-d] <file|directory|file.wfp>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s --version\n", os.Args[0])
		os.Exit(1)
	}

	path := flag.Arg(0)

	// Determine input type
	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	isWFPFile := !fileInfo.IsDir() && strings.HasSuffix(strings.ToLower(path), ".wfp")

	// Generate-only mode (with -fp flag)
	if *generateMode {
		var wfp string
		if fileInfo.IsDir() {
			wfp, err = pkg.GenerateWFPFromDirectory(path)
		} else {
			wfp, err = pkg.GenerateWFPFromFile(path)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating WFP: %v\n", err)
			os.Exit(1)
		}

		// Write output
		if *outputFile != "" {
			err = os.WriteFile(*outputFile, []byte(wfp), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "WFP successfully generated at: %s\n", *outputFile)
		} else {
			fmt.Print(wfp)
		}
		return
	}

	// Scan mode
	var wfpFile string
	var tempFile *os.File

	if isWFPFile {
		// It's a .wfp file, use it directly
		wfpFile = path
	} else {
		// Not a .wfp file, generate temporary WFP
		var wfp string
		if fileInfo.IsDir() {
			fmt.Fprintf(os.Stderr, "Generating WFP...\n")
			progress := &progressWriter{}
			wfp, err = pkg.GenerateWFPFromDirectoryWithProgress(path, progress)
			if progress.bar != nil {
				progress.bar.Finish()
				fmt.Fprintln(os.Stderr)
			}
		} else {
			wfp, err = pkg.GenerateWFPFromFile(path)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating WFP: %v\n", err)
			os.Exit(1)
		}

		// Create temporary file
		tempFile, err = os.CreateTemp("", "scan-*.wfp")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating temporary file: %v\n", err)
			os.Exit(1)
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		_, err = tempFile.WriteString(wfp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing temporary WFP: %v\n", err)
			os.Exit(1)
		}
		tempFile.Close()

		wfpFile = tempFile.Name()
	}

	// Scan WFP file
	fmt.Fprintf(os.Stderr, "Scanning files with %d threads...\n", *numThreads)
	progress := &progressWriter{}
	results, err := pkg.ScanWFPFile(kbName, wfpFile, *minHits, progress, *numThreads)
	if progress.bar != nil {
		progress.bar.Finish()
		fmt.Fprintln(os.Stderr)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning WFP: %v\n", err)
		os.Exit(1)
	}

	// Convert to JSON and display
	jsonOutput, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
}
