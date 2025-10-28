// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * go-wrapper/snippets_wrapper.h
 *
 * Go wrapper header for SCANOSS snippet scanning
 *
 * Copyright (C) 2018-2021 SCANOSS.COM
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package deps

// #cgo CFLAGS: -I.
// #cgo LDFLAGS: ${SRCDIR}/libsnippets_wrapper.a -lldb -lssl -lcrypto -lz -lm -lpthread
// #include "snippets_wrapper.h"
// #include <stdlib.h>
import "C"
import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unsafe"

	"scanoss.com/openkb-engine/models"
)

func ParseWFPFile(filepath string) (*models.WFPData, error) {
	return ParseWFPFileForMD5(filepath, "")
}

// ParseWFPFileForMD5 parses a WFP file and extracts only the data from the file with the specified MD5.
// If targetMD5 is empty, it parses the first file found (legacy behavior).
func ParseWFPFileForMD5(filepath string, targetMD5 string) (*models.WFPData, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open WFP file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	wfpData := &models.WFPData{
		Hashes: make([]uint32, 0),
		Lines:  make([]uint32, 0),
	}

	var processingTarget bool = false
	var foundTarget bool = false

	for scanner.Scan() {
		line := scanner.Text()

		// Parse file header line
		if strings.HasPrefix(line, "file=") {
			parts := strings.Split(strings.TrimPrefix(line, "file="), ",")
			if len(parts) < 3 {
				continue
			}

			currentMD5 := parts[0]

			// If we already found and processed the target file, stop
			if foundTarget && targetMD5 != "" {
				break
			}

			// Determine if this is the file we want to process
			if targetMD5 == "" || currentMD5 == targetMD5 {
				processingTarget = true
				foundTarget = true

				// Parse MD5
				md5Bytes, err := hex.DecodeString(parts[0])
				if err != nil {
					return nil, fmt.Errorf("failed to decode MD5: %v", err)
				}
				copy(wfpData.MD5[:], md5Bytes)

				// Parse total lines
				wfpData.TotalLines, err = strconv.Atoi(parts[1])
				if err != nil {
					return nil, fmt.Errorf("failed to parse total lines: %v", err)
				}

				// Parse file path
				wfpData.FilePath = parts[2]
			} else {
				// This is not the target file, stop processing until the next file=
				processingTarget = false
			}

		} else if strings.Contains(line, "=") && processingTarget {
			// Only parse hashes if we are processing the target file
			// Parse hash lines (format: line_number=hash1,hash2,...)
			parts := strings.Split(line, "=")
			if len(parts) != 2 {
				continue
			}

			lineNum, err := strconv.Atoi(parts[0])
			if err != nil {
				continue
			}

			// Parse hashes for this line
			hashStrings := strings.Split(parts[1], ",")
			for _, hashStr := range hashStrings {
				hashValue, err := strconv.ParseUint(hashStr, 16, 32)
				if err != nil {
					continue
				}
				wfpData.Hashes = append(wfpData.Hashes, uint32(hashValue))
				wfpData.Lines = append(wfpData.Lines, uint32(lineNum))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	if targetMD5 != "" && !foundTarget {
		return nil, fmt.Errorf("file with MD5 %s not found in WFP", targetMD5)
	}

	return wfpData, nil
}

func SnippetWrapperInit(ossDbName string, debugMode bool) bool {
	cDbName := C.CString(ossDbName)
	defer C.free(unsafe.Pointer(cDbName))
	return bool(C.snippets_wrapper_init(cDbName, C.bool(debugMode)))
}

func SnippetWrapperCleanup() {
	C.snippets_wrapper_cleanup()
}

func ScanWFP(wfpData *models.WFPData, debugMode bool) (*models.ScanResult, error) {
	if len(wfpData.Hashes) == 0 {
		return nil, fmt.Errorf("no hashes found in WFP data")
	}

	// Prepare C struct
	cInput := C.wrapper_scan_input_t{}

	// Copy MD5
	for i := 0; i < 16; i++ {
		cInput.md5[i] = C.uint8_t(wfpData.MD5[i])
	}

	// Set file path
	cFilePath := C.CString(wfpData.FilePath)
	defer C.free(unsafe.Pointer(cFilePath))
	cInput.file_path = cFilePath

	// Allocate C memory for hashes and lines
	hashCount := len(wfpData.Hashes)
	cHashes := (*C.uint32_t)(C.malloc(C.size_t(hashCount * 4)))
	cLines := (*C.uint32_t)(C.malloc(C.size_t(hashCount * 4)))
	defer C.free(unsafe.Pointer(cHashes))
	defer C.free(unsafe.Pointer(cLines))

	// Copy data to C memory
	hashesSlice := (*[1 << 30]C.uint32_t)(unsafe.Pointer(cHashes))[:hashCount:hashCount]
	linesSlice := (*[1 << 30]C.uint32_t)(unsafe.Pointer(cLines))[:hashCount:hashCount]

	for i := 0; i < hashCount; i++ {
		hashesSlice[i] = C.uint32_t(wfpData.Hashes[i])
		linesSlice[i] = C.uint32_t(wfpData.Lines[i])
	}

	// Set pointers and counts
	cInput.hashes = cHashes
	cInput.lines = cLines
	cInput.hash_count = C.uint32_t(hashCount)
	cInput.total_lines = C.int(wfpData.TotalLines)

	if debugMode {
		fmt.Fprintf(os.Stderr, "[GO DEBUG] About to call C.snippets_wrapper_scan\n")
		fmt.Fprintf(os.Stderr, "[GO DEBUG] hash_count=%d, total_lines=%d\n", hashCount, wfpData.TotalLines)
		fmt.Fprintf(os.Stderr, "[GO DEBUG] cInput.hash_count=%d, cInput.total_lines=%d\n", cInput.hash_count, cInput.total_lines)
		fmt.Fprintf(os.Stderr, "[GO DEBUG] cHashes=%p, cLines=%p\n", cHashes, cLines)
	}

	// Call the scan function
	if debugMode {
		fmt.Fprintf(os.Stderr, "[GO DEBUG] Calling C.snippets_wrapper_scan now...\n")
	}
	cResult := C.snippets_wrapper_scan(&cInput)
	if debugMode {
		fmt.Fprintf(os.Stderr, "[GO DEBUG] Returned from C.snippets_wrapper_scan\n")
	}
	if cResult == nil {
		return nil, fmt.Errorf("scan failed: result is nil")
	}
	defer C.snippets_wrapper_free_result(cResult)

	// Convert C result to Go model
	result := &models.ScanResult{
		MatchType:  models.MatchType(cResult.match_type),
		MatchCount: int(cResult.match_count),
		Matches:    make([]models.MatchInfo, 0),
	}

	if cResult.error_msg != nil {
		result.ErrorMsg = C.GoString(cResult.error_msg)
	}

	// Convert matches
	if cResult.match_count > 0 {
		matchesSlice := (*[1 << 30]C.wrapper_match_info_t)(unsafe.Pointer(cResult.matches))[:cResult.match_count:cResult.match_count]
		for i := 0; i < int(cResult.match_count); i++ {
			matchInfo := models.MatchInfo{
				FileMD5Hex: C.GoString(&matchesSlice[i].file_md5_hex[0]),
				Hits:       int(matchesSlice[i].hits),
				Ranges:     make([]models.Range, 0),
			}

			// Convert ranges if available
			rangeCount := int(matchesSlice[i].range_count)
			if rangeCount > 0 && matchesSlice[i].range_from != nil {
				rangeFromSlice := (*[1 << 30]C.int)(unsafe.Pointer(matchesSlice[i].range_from))[:rangeCount:rangeCount]
				rangeToSlice := (*[1 << 30]C.int)(unsafe.Pointer(matchesSlice[i].range_to))[:rangeCount:rangeCount]
				rangeOSSSlice := (*[1 << 30]C.int)(unsafe.Pointer(matchesSlice[i].oss_line))[:rangeCount:rangeCount]

				for r := 0; r < rangeCount; r++ {
					matchInfo.Ranges = append(matchInfo.Ranges, models.Range{
						From: int(rangeFromSlice[r]),
						To:   int(rangeToSlice[r]),
						Oss:  int(rangeOSSSlice[r]),
					})
				}
			}

			result.Matches = append(result.Matches, matchInfo)
		}
	}

	return result, nil
}
