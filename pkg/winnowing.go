// SPDX-FileCopyrightText: Copyright (C) 2025 Fundación Para La Transparencia del Software - STF
// SPDX-FileCopyrightText: 2025 Mariano Scasso <info@st.foundation>
//
// SPDX-License-Identifier: GPL-2.0

package pkg

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
)

const GRAM = 30
const BUFFER_RATE = 4
const WINDOW = 64

var SKIP_SNIPPET_EXT = []string{
	// Executables and binaries
	".exe", ".bin", ".app", ".out", ".o", ".a", ".so", ".obj", ".dll", ".lib", ".dylib",
	// Archives
	".zip", ".tar", ".tgz", ".gz", ".7z", ".rar", ".bz2", ".xz", ".lz", ".lzma", ".Z",
	// Java
	".jar", ".war", ".ear", ".class",
	// Images
	".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico", ".tiff", ".tif", ".webp", ".svg",
	// Videos
	".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".webm", ".m4v",
	// Audio
	".mp3", ".wav", ".ogg", ".flac", ".aac", ".wma", ".m4a",
	// Documents
	".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".odt", ".ods", ".odp", ".pages", ".key", ".numbers", ".pdf",
	// Python compiled
	".pyc", ".pyo", ".pyd",
	// Fonts
	".ttf", ".otf", ".woff", ".woff2", ".eot",
	// Data/Config formats (often not useful for snippet matching)
	".json", ".xml", ".yml", ".yaml", ".toml", ".ini", ".cfg", ".conf",
	// Web
	".htm", ".html",
	// Documentation
	".md", ".txt", ".rst", ".adoc",
	// Other
	".dat", ".lst", ".mf", ".sum", ".db", ".sqlite", ".sqlite3",
}
var FILTERED_EXT = []string{".1", ".2", ".3", ".4", ".5", ".6", ".7", ".8", ".9", ".ac", ".adoc", ".am", ".asciidoc", ".bmp", ".build", ".cfg", ".chm", ".class", ".cmake", ".cnf", ".conf", ".config", ".contributors", ".copying", ".crt", ".csproj", ".css", ".csv", ".dat", ".data", ".doc", ".docx", ".dtd", ".dts", ".iws", ".c9", ".c9revisions", ".dtsi", ".dump", ".eot", ".eps", ".geojson", ".gdoc", ".gif", ".glif", ".gmo", ".gradle", ".guess", ".hex", ".htm", ".html", ".ico", ".iml", ".in", ".inc", ".info", ".ini", ".ipynb", ".jpeg", ".jpg", ".json", ".jsonld", ".lock", ".log", ".m4", ".map", ".markdown", ".md", ".md5", ".meta", ".mk", ".mxml", ".o", ".otf", ".out", ".pbtxt", ".pdf", ".pem", ".phtml", ".plist", ".png", ".po", ".ppt", ".prefs", ".properties", ".pyc", ".qdoc", ".result", ".rgb", ".rst", ".scss", ".sha", ".sha1", ".sha2", ".sha256", ".sln", ".spec", ".sql", ".sub", ".svg", ".svn-base", ".tab", ".template", ".test", ".tex", ".tiff", ".toml", ".ttf", ".txt", ".utf-8", ".vim", ".wav", ".whl", ".woff", ".xht", ".xhtml", ".xls", ".xlsx", ".xml", ".xpm", ".xsd", ".xul", ".yaml", ".yml", ".wfp", ".editorconfig", ".dotcover", ".pid", ".lcov", ".egg", ".manifest", ".cache", ".coverage", ".cover", ".gem", ".lst", ".pickle", ".pdb", ".gml", ".pot", ".plt"}

func normalize(b byte) byte {
	if b < '0' {
		return 0
	}
	if b > 'z' {
		return 0
	}
	if b <= '9' {
		return b
	}
	if b >= 'a' {
		return b
	}
	if (b >= 'A') && (b <= 'Z') {
		return b + 32
	}
	return 0

}

var basePath string
var bannedFiles map[string]bool
var onlyMD5 map[string]bool

func minHash(hashes []uint32) uint32 {

	indexMin := 0
	for r := range hashes {
		if hashes[r] <= hashes[indexMin] {
			indexMin = r
		}
	}
	return hashes[indexMin]

}
func intToByte(n uint32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, n)
	return bs

}

func LoadFilters(fileName string) {
	bannedFiles = make(map[string]bool)
	onlyMD5 = make(map[string]bool)
	for r := range SKIP_SNIPPET_EXT {
		onlyMD5[SKIP_SNIPPET_EXT[r]] = true
	}
	for r := range FILTERED_EXT {
		bannedFiles[FILTERED_EXT[r]] = true

	}

}

func SkipFile(fileName string) bool {
	return false
}
func fingerprint(filePath string) string {
	var newByte byte
	var window []byte
	//var lineArrays []int
	result := ""
	f, err := os.ReadFile(filePath)
	if err != nil {
		//	fmt.Println("No se pudo abrir el archivo")
		return ""
	}

	// Limit line length to 1KB to avoid "token too long" errors
	fileLine := fmt.Sprintf("file=%x,%d,%s\n", md5.Sum(f), len(f), filePath)
	if len(fileLine) > 1024 {
		// Truncate the file path to fit within 1KB limit
		maxPathLen := 1024 - (len(fileLine) - len(filePath)) - 1 // -1 for safety
		if maxPathLen > 0 && maxPathLen < len(filePath) {
			filePath = filePath[:maxPathLen]
		}
		fileLine = fmt.Sprintf("file=%x,%d,%s\n", md5.Sum(f), len(f), filePath)
	}
	result += fileLine
	lines := 1
	//counts := 0
	//windowPrt := 0
	crc32q := crc32.MakeTable(0x82f63b78)
	var hashes []uint32
	var last uint32
	last = 0
	//var hashesCandidates []uint32
	wfp := make(map[int][]uint32)
	for i := 0; i < len(f); i++ {
		if f[i] == '\n' {
			lines++
		}
		newByte = normalize(f[i])
		if newByte == 0 {
			continue
		}

		window = append(window, newByte)
		if len(window) >= GRAM {

			hashes = append(hashes, crc32.Checksum(window, crc32q))
			if len(hashes) >= WINDOW {
				a := minHash(hashes)
				if a != last {
					last = a
					wfp[lines] = append(wfp[lines], crc32.Checksum(intToByte(a), crc32q))
				}
				hashes = hashes[1:WINDOW]

			} else {
				//	fmt.Println("Not enoungh hashes")
			}
			window = window[1:GRAM]
		} else {
			//fmt.Println("Filling Gram", string(window))
		}
	}

	keys := make([]int, 0, len(wfp))

	for k := range wfp {
		keys = append(keys, k)
	}
	sort.Ints([]int(keys))

	for i := range keys {
		k := keys[i]
		hashLine := fmt.Sprintf("%d=", k)
		v := wfp[k]
		for w := 0; w < len(v); w++ {
			var hashStr string
			if w < len(v)-1 {
				hashStr = fmt.Sprintf("%0.8x,", v[w])
			} else {
				hashStr = fmt.Sprintf("%0.8x\n", v[w])
			}
			// Limit line length to 1KB
			if len(hashLine)+len(hashStr) > 1024 {
				// Finish current line and start a new one with same line number
				hashLine += "\n"
				result += hashLine
				hashLine = fmt.Sprintf("%d=", k)
			}
			hashLine += hashStr
		}
		result += hashLine
	}
	return result
}

var wfps []string
var progressWriter io.Writer

func walkFunc(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		// Skip hidden directories
		if len(info.Name()) > 0 && info.Name()[0] == '.' {
			return filepath.SkipDir
		}
	} else {
		// Skip hidden files (files starting with .)
		if len(info.Name()) > 0 && info.Name()[0] == '.' {
			return nil
		}

		if info.Size() > 100 {
			ext1 := filepath.Ext(path)
			baseName := filepath.Base(path)

			// Check for minimized files (.min.js, .min.css, etc.)
			// Pattern: filename.min.ext
			if len(baseName) > 4 {
				nameWithoutExt := baseName[:len(baseName)-len(ext1)]
				if len(nameWithoutExt) >= 4 && nameWithoutExt[len(nameWithoutExt)-4:] == ".min" {
					DebugLog("Skipping minimized file: %s\n", path)
					return nil
				}
			}

			_, banned := onlyMD5[ext1]
			_, skipped := bannedFiles[ext1]
			if !skipped && !banned {
				wfps = append(wfps, path)
			}
		}
	}
	return nil
}

// GenerateWFPFromFile generates WFP for a single file
func GenerateWFPFromFile(filePath string) (string, error) {
	LoadFilters("")

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("error accessing file: %v", err)
	}

	if fileInfo.IsDir() {
		return "", fmt.Errorf("path is a directory, use GenerateWFPFromDirectory instead")
	}

	if fileInfo.Size() <= 100 {
		return "", fmt.Errorf("file too small (must be > 100 bytes)")
	}

	ext := filepath.Ext(filePath)
	_, banned := onlyMD5[ext]
	_, skipped := bannedFiles[ext]

	if skipped {
		return "", fmt.Errorf("file extension %s is in skip list", ext)
	}

	if banned {
		return "", fmt.Errorf("file extension %s is banned", ext)
	}

	wfp := fingerprint(filePath)
	if wfp == "" {
		return "", fmt.Errorf("failed to generate fingerprint")
	}

	return wfp, nil
}

// GenerateWFPFromDirectory generates WFP for all files in a directory
func GenerateWFPFromDirectory(dirPath string) (string, error) {
	return GenerateWFPFromDirectoryWithProgress(dirPath, nil)
}

// GenerateWFPFromDirectoryWithProgress generates WFP for all files in a directory with progress reporting
func GenerateWFPFromDirectoryWithProgress(dirPath string, progress io.Writer) (string, error) {
	progressWriter = progress
	LoadFilters("")

	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		return "", fmt.Errorf("error accessing directory: %v", err)
	}

	if !fileInfo.IsDir() {
		return "", fmt.Errorf("path is not a directory, use GenerateWFPFromFile instead")
	}

	// Reset global variables
	wfps = []string{}
	basePath = dirPath

	// Walk the directory
	err = filepath.Walk(dirPath, walkFunc)
	if err != nil {
		return "", fmt.Errorf("error walking directory: %v", err)
	}

	if len(wfps) == 0 {
		return "", fmt.Errorf("no valid files found in directory")
	}

	// Generate WFP for all collected files
	var result string
	for i, filePath := range wfps {
		wfp := fingerprint(filePath)
		if wfp != "" {
			result += wfp
		}
		if progressWriter != nil {
			fmt.Fprintf(progressWriter, "progress:%d/%d\n", i+1, len(wfps))
		}
	}

	if result == "" {
		return "", fmt.Errorf("failed to generate any fingerprints")
	}

	return result, nil
}
