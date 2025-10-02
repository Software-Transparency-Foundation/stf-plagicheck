package pkg

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateWFPFromFile(t *testing.T) {
	// Create a temporary test file
	content := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
	fmt.Println("This is a test")
	fmt.Println("With multiple lines")
	fmt.Println("To generate WFP")
}
`
	tmpFile, err := os.CreateTemp("", "test-*.go")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Test WFP generation
	wfp, err := GenerateWFPFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to generate WFP: %v", err)
	}

	// Verify WFP format
	if !strings.HasPrefix(wfp, "file=") {
		t.Error("WFP should start with 'file='")
	}

	if !strings.Contains(wfp, tmpFile.Name()) {
		t.Error("WFP should contain file path")
	}

	// Should have fingerprint lines
	lines := strings.Split(wfp, "\n")
	if len(lines) < 2 {
		t.Error("WFP should have multiple lines")
	}
}

func TestGenerateWFPFromFile_SmallFile(t *testing.T) {
	// Create a very small file (should fail)
	tmpFile, err := os.CreateTemp("", "test-small-*.go")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString("small"); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	_, err = GenerateWFPFromFile(tmpFile.Name())
	if err == nil {
		t.Error("expected error for small file, got nil")
	}

	if !strings.Contains(err.Error(), "too small") {
		t.Errorf("expected 'too small' error, got: %v", err)
	}
}

func TestGenerateWFPFromDirectory(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir, err := os.MkdirTemp("", "test-dir-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test Go file
	content := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
	fmt.Println("This is a test")
	fmt.Println("With multiple lines")
	fmt.Println("To generate WFP")
	fmt.Println("More lines for testing")
}
`
	testFile := tmpDir + "/test.go"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Test WFP generation from directory
	wfp, err := GenerateWFPFromDirectory(tmpDir)
	if err != nil {
		t.Fatalf("failed to generate WFP from directory: %v", err)
	}

	// Should have at least one file entry
	if !strings.Contains(wfp, "file=") {
		t.Error("WFP should contain at least one file entry")
	}

	// Count file entries
	fileCount := strings.Count(wfp, "file=")
	if fileCount == 0 {
		t.Error("expected at least one file in WFP")
	}
}

func TestLoadFilters(t *testing.T) {
	LoadFilters("")

	// Check that banned files map is populated
	if len(bannedFiles) == 0 {
		t.Error("bannedFiles should be populated")
	}

	// Check that onlyMD5 map is populated
	if len(onlyMD5) == 0 {
		t.Error("onlyMD5 should be populated")
	}

	// Test specific extensions
	if !bannedFiles[".json"] {
		t.Error(".json should be in bannedFiles")
	}

	if !onlyMD5[".zip"] {
		t.Error(".zip should be in onlyMD5")
	}
}
