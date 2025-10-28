package pkg

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"scanoss.com/openkb-engine/deps"
	"scanoss.com/openkb-engine/models"
)

// Line tolerance for merging ranges (ranges separated by less than this amount will be merged)
const RangeMergeTolerance = 3

var wfpAvailable bool = false // Indicates if WFP scanning is available

// GetFirstURLRecords retrieves the first URL record for a given file hash from the KB
func GetFirstURLRecords(kbName, hash string) ([]string, error) {
	// Execute ldb query
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo select from %s/file-url key %s csv hex 8 | ldb | head -n 1", kbName, hash))
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Process output
	queryResult := strings.TrimSpace(string(output))
	if queryResult != "" {
		// Take only the first line of the result
		resultLines := strings.Split(queryResult, "\n")
		if len(resultLines) == 0 {
			return nil, nil
		}

		resultFields := strings.Split(resultLines[0], ",")
		if len(resultFields) == 0 {
			return nil, fmt.Errorf("empty result")
		}

		return resultFields[1:], nil
	}

	return nil, fmt.Errorf("empty result")
}

// MergeRanges merges ranges that overlap or are separated by less than 'tolerance' lines
// Iteratively increases tolerance to ensure a maximum of 10 ranges
func MergeRanges(ranges []models.Range, tolerance int) []models.Range {
	if len(ranges) == 0 {
		return ranges
	}

	const maxRanges = 10
	currentTolerance := tolerance
	var merged []models.Range

	// Iteratively merge with increasing tolerance until we have <= maxRanges
	for {
		merged = mergeRangesWithTolerance(ranges, currentTolerance)

		DebugLog("MergeRanges: tolerance=%d, resulted in %d ranges\n", currentTolerance, len(merged))

		// If we have acceptable number of ranges, or we only have 1 range, stop
		if len(merged) <= maxRanges || len(merged) == 1 {
			break
		}

		// Double the tolerance and try again
		currentTolerance *= 2
	}

	return merged
}

// mergeRangesWithTolerance performs the actual merging with a given tolerance
func mergeRangesWithTolerance(ranges []models.Range, tolerance int) []models.Range {
	if len(ranges) == 0 {
		return ranges
	}

	// Sort ranges by start position
	sorted := make([]models.Range, len(ranges))
	copy(sorted, ranges)

	// Simple bubble sort (sufficient for small number of ranges)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].From < sorted[i].From {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Merge ranges
	merged := []models.Range{sorted[0]}

	for i := 1; i < len(sorted); i++ {
		last := &merged[len(merged)-1]
		current := sorted[i]

		// If current range overlaps or is within tolerance
		if current.From <= last.To+tolerance+1 {
			// Extend last range if necessary
			if current.To > last.To {
				last.To = current.To
			}
		} else {
			// No overlap, add as new range
			merged = append(merged, current)
		}
	}

	return merged
}

// FormatRanges converts a slice of ranges to string format "15-45,120-135"
func FormatRanges(ranges []models.Range) (string, string) {
	var parts []string
	var oss []string
	for _, r := range ranges {
		parts = append(parts, fmt.Sprintf("%d-%d", r.From, r.To))
		oss = append(oss, fmt.Sprintf("%d-%d", r.Oss, r.Oss+r.To-r.From))
	}
	return strings.Join(parts, ","), strings.Join(oss, ",")
}

// ReadWFPFile reads WFP files and extracts data for each file
func ReadWFPFile(filename string) ([]*models.WFPData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []*models.WFPData
	filePattern := regexp.MustCompile(`^file=([a-f0-9]{32}),([0-9]+),(.+)$`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if matches := filePattern.FindStringSubmatch(line); matches != nil {
			md5Bytes, err := hex.DecodeString(matches[1])
			if err != nil {
				continue
			}

			totalLines, err := strconv.Atoi(matches[2])
			if err != nil {
				continue
			}

			entry := &models.WFPData{
				MD5Hex:     matches[1],
				TotalLines: totalLines,
				FilePath:   matches[3],
			}
			copy(entry.MD5[:], md5Bytes)

			entries = append(entries, entry)
		}
	}

	return entries, scanner.Err()
}

// FilterValidRanges filters out ranges that only span a single line
func FilterValidRanges(ranges []models.Range) []models.Range {
	var valid []models.Range
	for _, r := range ranges {
		// Range is valid only if it spans more than one line
		if r.To > r.From {
			valid = append(valid, r)
		}
	}
	return valid
}

// ProcessWFPEntry processes a WFP entry and returns match results
// First tries full MD5 match, then snippet matching if no full match is found
// minHits: minimum number of hits required for a valid snippet match (default: 3)
func ProcessWFPEntry(kbName string, entry *models.WFPData, wfpFilePath string, minHits int) (*models.MatchResult, error) {
	// Step 1: Try full MD5 match
	DebugLog("Step 1: Checking full MD5 match...\n")
	records, err := GetFirstURLRecords(kbName, entry.MD5Hex)
	if err == nil && len(records) >= 3 {
		// Full match found
		var instances int
		if i, err := strconv.Atoi(records[2]); err == nil {
			instances = i
		}

		result := &models.MatchResult{
			MatchType:     "full_file",
			Instances:     instances,
			ReferenceURL:  records[1], // URL is at index 1
			ReferenceFile: records[0], // File is at index 0
		}
		return result, nil
	}

	// Step 2: No full match, try snippet matching
	DebugLog("Step 2: No full match, parsing WFP for snippet matching...\n")
	// Parse only the specific file from WFP using its MD5
	wfpData, err := deps.ParseWFPFileForMD5(wfpFilePath, entry.MD5Hex)
	if err != nil {
		return nil, fmt.Errorf("error parsing WFP file: %v", err)
	}

	// Execute snippet scan
	if wfpAvailable {
		DebugLog("Step 2b: Scanning snippets (this may take a while)...\n")
		scanResult, err := deps.ScanWFP(wfpData, false)
		DebugLog("Step 2c: Snippet scan completed.\n")
		if err != nil {
			return nil, fmt.Errorf("error scanning snippets: %v", err)
		}

		// If no snippet matches
		if scanResult.MatchCount == 0 || len(scanResult.Matches) == 0 {
			return nil, fmt.Errorf("no matches found")
		}

		// Step 3: Select candidate with highest number of hits
		var bestMatch *models.MatchInfo
		maxHits := 0
		for i := range scanResult.Matches {
			if scanResult.Matches[i].Hits > maxHits {
				maxHits = scanResult.Matches[i].Hits
				bestMatch = &scanResult.Matches[i]
			}
		}

		if bestMatch == nil {
			return nil, fmt.Errorf("no valid match found")
		}

		// Validate minimum hits requirement
		if bestMatch.Hits < minHits {
			return nil, fmt.Errorf("insufficient hits: %d (minimum required: %d)", bestMatch.Hits, minHits)
		}

		// Filter ranges to keep only those spanning more than one line
		validRanges := FilterValidRanges(bestMatch.Ranges)
		if len(validRanges) == 0 {
			return nil, fmt.Errorf("no valid ranges found (all ranges span single line)")
		}

		// Step 4: Get candidate file details using its MD5
		records, err = GetFirstURLRecords(kbName, bestMatch.FileMD5Hex)
		if err != nil {
			return nil, fmt.Errorf("error getting URL records for best match: %v", err)
		}

		var instances int
		if len(records) >= 3 {
			if i, err := strconv.Atoi(records[2]); err == nil {
				instances = i
			}
		}

		// Step 5: Merge ranges with tolerance and generate result in code_snippet format
		mergedRanges := MergeRanges(validRanges, RangeMergeTolerance)
		targetLines, ossLines := FormatRanges(mergedRanges)
		result := &models.MatchResult{
			MatchType:     "code_snippet",
			TargetLines:   targetLines,
			SourceLines:   ossLines,
			Instances:     instances,
			ReferenceURL:  records[1],
			ReferenceFile: records[0],
			Hits:          bestMatch.Hits,
			Ranges:        mergedRanges,
		}

		return result, nil
	}
	return nil, nil
}

// ScanWFPFile scans a WFP file with progress reporting and parallel processing
func ScanWFPFile(kbName, wfpFilePath string, minHits int, progress io.Writer, numThreads int) (map[string][]*models.MatchResult, error) {
	// Read WFP file
	entries, err := ReadWFPFile(wfpFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading WFP file: %v", err)
	}

	// Initialize snippet scanner once for all files
	wfpAvailable = deps.SnippetWrapperInit(kbName, false)
	defer deps.SnippetWrapperCleanup()

	// Ensure at least 1 thread
	if numThreads < 1 {
		numThreads = 1
	}

	DebugLog("Processing %d files with %d threads\n", len(entries), numThreads)

	// Map to store results per file (protected by mutex)
	results := make(map[string][]*models.MatchResult)
	var resultsMutex sync.Mutex

	// Progress tracking
	var processedCount int
	var progressMutex sync.Mutex

	// Create work channel and wait group
	type workItem struct {
		index int
		entry *models.WFPData
	}
	workChan := make(chan workItem, len(entries))
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for item := range workChan {
				// Use unique key: if multiple files with same name exist,
				// add MD5 to distinguish them
				key := item.entry.FilePath

				// Debug logging
				DebugLog("\n[Worker %d] Processing file %d/%d: %s (MD5: %s)\n",
					workerID, item.index+1, len(entries), item.entry.FilePath, item.entry.MD5Hex)

				match, err := ProcessWFPEntry(kbName, item.entry, wfpFilePath, minHits)

				// Store result
				resultsMutex.Lock()
				// Check if this key already exists in results
				if _, exists := results[key]; exists {
					// A file with this name already exists, use FilePath+MD5 as key
					key = fmt.Sprintf("%s [%s]", item.entry.FilePath, item.entry.MD5Hex)
				}

				if err != nil {
					// If error, add a no_match result
					results[key] = []*models.MatchResult{{
						MatchType:     "no_match",
						Instances:     0,
						ReferenceURL:  "",
						ReferenceFile: "",
					}}
				} else {
					results[key] = []*models.MatchResult{match}
				}
				resultsMutex.Unlock()

				// Update progress
				progressMutex.Lock()
				processedCount++
				if progress != nil {
					fmt.Fprintf(progress, "progress:%d/%d\n", processedCount, len(entries))
				}
				progressMutex.Unlock()
			}
		}(i)
	}

	// Send work to workers
	for i, entry := range entries {
		workChan <- workItem{index: i, entry: entry}
	}
	close(workChan)

	// Wait for all workers to complete
	wg.Wait()

	return results, nil
}
