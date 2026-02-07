package parser_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jellydn/dotenv-tui/internal/parser"
)

// TestParseRealEnvFiles tests the parser against real example files
func TestParseRealEnvFiles(t *testing.T) {
	// Get the testdata directory relative to the test file
	_, filename, _, _ := runtime.Caller(0)
	testdataDir := filepath.Join(filepath.Dir(filename), "..", "..", "testdata")

	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Skipf("testdata directory not found: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, ".env") {
			continue
		}

		t.Run(name, func(t *testing.T) {
			path := filepath.Join(testdataDir, name)
			file, err := os.Open(path)
			if err != nil {
				t.Fatalf("failed to open %s: %v", path, err)
			}
			defer func() {
				if err := file.Close(); err != nil {
					t.Logf("failed to close %s: %v", path, err)
				}
			}()

			entries, err := parser.Parse(file)
			if err != nil {
				t.Fatalf("failed to parse %s: %v", path, err)
			}

			// Basic sanity checks
			if len(entries) == 0 {
				t.Logf("warning: %s has no entries", name)
			}

			// Count entry types
			var kvCount, commentCount, blankCount int
			for _, e := range entries {
				switch e.(type) {
				case parser.KeyValue:
					kvCount++
				case parser.Comment:
					commentCount++
				case parser.BlankLine:
					blankCount++
				}
			}

			t.Logf("%s: %d key-values, %d comments, %d blank lines",
				name, kvCount, commentCount, blankCount)
		})
	}
}
