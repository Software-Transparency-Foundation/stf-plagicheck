package pkg

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"sort"
)

const GRAM = 30
const BUFFER_RATE = 4
const WINDOW = 64

var SKIP_SNIPPET_EXT = []string{".exe", ".zip", ".tar", ".tgz", ".gz", ".7z", ".rar", ".jar", ".war", ".ear", ".class", ".pyc", ".o", ".a", ".so", ".obj", ".dll", ".lib", ".out", ".app", ".bin", ".lst", ".dat", ".json", ".htm", ".html", ".xml", ".md", ".txt", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".odt", ".ods", ".odp", ".pages", ".key", ".numbers", ".pdf", ".min.js", ".mf", ".sum"}
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
var jobs chan string

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

	result += fmt.Sprintf("file=%x,%d,%s\n", md5.Sum(f), len(f), filePath)
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
		result += fmt.Sprintf("%d=", k)
		v := wfp[k]
		for w := 0; w < len(v); w++ {
			if w < len(v)-1 {
				result += fmt.Sprintf("%0.8x,", v[w])
			} else {
				result += fmt.Sprintf("%0.8x\n", v[w])
			}

		}

	}
	return result
}

var wfps []string

func walkFunc(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
	} else {
		if path[0] != '.' && info.Size() > 100 {
			ext1 := filepath.Ext(path)
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
	for _, filePath := range wfps {
		wfp := fingerprint(filePath)
		if wfp != "" {
			result += wfp
		}
	}

	if result == "" {
		return "", fmt.Errorf("failed to generate any fingerprints")
	}

	return result, nil
}
