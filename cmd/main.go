package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"scanoss.com/openkb-engine/deps"
	"scanoss.com/openkb-engine/models"
)

var kbName string = "osskbopen"

// Tolerancia de líneas para unificar rangos (rangos separados por menos de esta cantidad se unifican)
const rangeMergeTolerance = 3

func getFirstUrlrecords(hash string) ([]string, error) {
	// Ejecutar la consulta ldb
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo select from %s/file-url key %s csv hex 8 | ldb | head -n 1", kbName, hash))
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Procesar la salida
	queryResult := strings.TrimSpace(string(output))
	if queryResult != "" {
		// Tomar solo la primera línea del resultado
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

// mergeRanges unifica rangos que se superponen o están separados por menos de 'tolerance' líneas
func mergeRanges(ranges []models.Range, tolerance int) []models.Range {
	if len(ranges) == 0 {
		return ranges
	}

	// Ordenar rangos por inicio
	sorted := make([]models.Range, len(ranges))
	copy(sorted, ranges)

	// Bubble sort simple (suficiente para cantidad pequeña de rangos)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].From < sorted[i].From {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Unificar rangos
	merged := []models.Range{sorted[0]}

	for i := 1; i < len(sorted); i++ {
		last := &merged[len(merged)-1]
		current := sorted[i]

		// Si el rango actual se superpone o está dentro de la tolerancia
		if current.From <= last.To+tolerance+1 {
			// Extender el último rango si es necesario
			if current.To > last.To {
				last.To = current.To
			}
		} else {
			// No hay superposición, agregar como nuevo rango
			merged = append(merged, current)
		}
	}

	return merged
}

// formatRanges convierte un slice de ranges a string formato "15-45,120-135"
func formatRanges(ranges []models.Range) string {
	var parts []string
	for _, r := range ranges {
		parts = append(parts, fmt.Sprintf("%d-%d", r.From, r.To))
	}
	return strings.Join(parts, ",")
}

// ReadWFPFile lee archivos WFP y extrae los datos de cada archivo
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

// processWFPEntry procesa una entrada WFP y devuelve los resultados de coincidencia
// Primero intenta búsqueda por MD5 completo, luego por snippets si no hay coincidencia
func processWFPEntry(entry *models.WFPData, wfpFilePath string) (*models.MatchResult, error) {
	// Paso 1: Intentar búsqueda por MD5 completo
	records, err := getFirstUrlrecords(entry.MD5Hex)
	if err == nil && len(records) >= 3 {
		// Coincidencia completa encontrada
		var instances int
		if i, err := strconv.Atoi(records[2]); err == nil {
			instances = i
		}

		result := &models.MatchResult{
			MatchType:     "full_file",
			Instances:     instances,
			ReferenceURL:  records[1], // URL está en el índice 1
			ReferenceFile: records[0], // Archivo está en el índice 0
		}
		return result, nil
	}

	// Paso 2: No hay coincidencia completa, intentar búsqueda de snippets
	// Parsear el archivo WFP completo para obtener hashes y líneas
	wfpData, err := deps.ParseWFPFile(wfpFilePath)
	if err != nil {
		return nil, fmt.Errorf("error parsing WFP file: %v", err)
	}

	// Ejecutar scan de snippets
	deps.SnippetWrapperInit(kbName, false)
	scanResult, err := deps.ScanWFP(wfpData, false)
	if err != nil {
		return nil, fmt.Errorf("error scanning snippets: %v", err)
	}

	// Si no hay coincidencias de snippets
	if scanResult.MatchCount == 0 || len(scanResult.Matches) == 0 {
		return nil, fmt.Errorf("no matches found")
	}

	// Paso 3: Seleccionar el candidato con mayor cantidad de hits
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

	// Paso 4: Obtener detalles del archivo candidato usando su MD5
	records, err = getFirstUrlrecords(bestMatch.FileMD5Hex)
	if err != nil {
		return nil, fmt.Errorf("error getting URL records for best match: %v", err)
	}

	var instances int
	if len(records) >= 3 {
		if i, err := strconv.Atoi(records[2]); err == nil {
			instances = i
		}
	}

	// Paso 5: Unificar rangos con tolerancia y generar resultado con formato code_snippet
	mergedRanges := mergeRanges(bestMatch.Ranges, rangeMergeTolerance)
	targetLines := formatRanges(mergedRanges)

	result := &models.MatchResult{
		MatchType:     "code_snippet",
		TargetLines:   targetLines,
		SourceLines:   targetLines, // Por ahora copiamos target_lines
		Instances:     instances,
		ReferenceURL:  records[1],
		ReferenceFile: records[0],
		Hits:          bestMatch.Hits,
		Ranges:        mergedRanges,
	}

	return result, nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Uso: %s <archivo.wfp>\n", os.Args[0])
		os.Exit(1)
	}

	wfpFile := os.Args[1]

	// Leer el archivo WFP
	entries, err := ReadWFPFile(wfpFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error leyendo archivo WFP: %v\n", err)
		os.Exit(1)
	}

	// Mapa para almacenar los resultados por archivo
	results := make(map[string][]*models.MatchResult)

	// Procesar cada entrada
	for _, entry := range entries {
		match, err := processWFPEntry(entry, wfpFile)
		if err != nil {
			// Si hay error, agregar un resultado no_match
			results[entry.FilePath] = []*models.MatchResult{{
				MatchType:     "no_match",
				Instances:     0,
				ReferenceURL:  "",
				ReferenceFile: "",
			}}
			continue
		}

		results[entry.FilePath] = []*models.MatchResult{match}
	}

	// Convertir a JSON y mostrar
	jsonOutput, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generando JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
}
