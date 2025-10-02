package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"scanoss.com/openkb-engine/pkg"
)

var (
	kbName  string = "osskbopen"
	version string = "dev"
	commit  string = "unknown"
)

func main() {
	generateMode := flag.Bool("fp", false, "Generate WFP from file or directory (output only, no scan)")
	outputFile := flag.String("output", "", "Output file for generated WFP (optional, default: stdout)")
	minHits := flag.Int("min-hits", 3, "Minimum number of hits required for valid snippet match (default: 3)")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("plagicheck version %s (commit: %s)\n", version, commit)
		os.Exit(0)
	}

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [-fp] [--output <file>] [--min-hits <N>] <file|directory|file.wfp>\n", os.Args[0])
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
			wfp, err = pkg.GenerateWFPFromDirectory(path)
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
	results, err := pkg.ScanWFPFile(kbName, wfpFile, *minHits)
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
