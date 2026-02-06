package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScan(t *testing.T) {
	t.Run("finds basic .env files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create test structure
		writeFile(t, tmpDir, ".env", "KEY=value")
		writeFile(t, tmpDir, "app.env", "should not match")
		writeFile(t, tmpDir, ".env.example", "should not match")

		results, err := Scan(tmpDir)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 file, got %d: %v", len(results), results)
		}

		if results[0] != ".env" {
			t.Errorf("Expected .env, got %s", results[0])
		}
	})

	t.Run("finds .env variants", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create various .env files
		writeFile(t, tmpDir, ".env", "BASE=value")
		writeFile(t, tmpDir, ".env.local", "LOCAL=value")
		writeFile(t, tmpDir, ".env.production", "PROD=value")
		writeFile(t, tmpDir, ".env.development", "DEV=value")
		writeFile(t, tmpDir, ".env.test", "TEST=value")

		// Should not match
		writeFile(t, tmpDir, ".env.example", "EXAMPLE=value")
		writeFile(t, tmpDir, ".env.local.example", "LOCAL_EXAMPLE=value")

		results, err := Scan(tmpDir)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		expected := []string{".env", ".env.development", ".env.local", ".env.production", ".env.test"}
		if len(results) != len(expected) {
			t.Errorf("Expected %d files, got %d: %v", len(expected), len(results), results)
		}

		// Check that all expected files are present (order doesn't matter)
		resultMap := make(map[string]bool)
		for _, r := range results {
			resultMap[r] = true
		}

		for _, exp := range expected {
			if !resultMap[exp] {
				t.Errorf("Missing expected file: %s", exp)
			}
		}
	})

	t.Run("skips dependency directories", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create directory structure with dependencies
		mkdir(t, tmpDir, "node_modules")
		mkdir(t, tmpDir, ".git")
		mkdir(t, tmpDir, "vendor")
		mkdir(t, tmpDir, "dist")
		mkdir(t, tmpDir, "build")
		mkdir(t, tmpDir, ".next")
		mkdir(t, tmpDir, ".nuxt")
		mkdir(t, tmpDir, "__pycache__")
		mkdir(t, tmpDir, "src")

		// Add .env files in skipped directories (should not be found)
		writeFile(t, tmpDir, "node_modules/.env", "NODE_ENV=value")
		writeFile(t, tmpDir, ".git/.env", "GIT_ENV=value")
		writeFile(t, tmpDir, "vendor/.env", "VENDOR_ENV=value")
		writeFile(t, tmpDir, "dist/.env", "DIST_ENV=value")
		writeFile(t, tmpDir, "build/.env", "BUILD_ENV=value")
		writeFile(t, tmpDir, ".next/.env", "NEXT_ENV=value")
		writeFile(t, tmpDir, ".nuxt/.env", "NUXT_ENV=value")
		writeFile(t, tmpDir, "__pycache__/.env", "PYTHON_ENV=value")

		// Add .env file in regular directory (should be found)
		writeFile(t, tmpDir, "src/.env", "SRC_ENV=value")
		writeFile(t, tmpDir, ".env", "ROOT_ENV=value")

		results, err := Scan(tmpDir)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		expected := []string{".env", "src/.env"}
		if len(results) != len(expected) {
			t.Errorf("Expected %d files, got %d: %v", len(expected), len(results), results)
		}

		// Check that files from skipped directories are not present
		for _, r := range results {
			if strings.Contains(r, "node_modules") ||
				strings.Contains(r, ".git") ||
				strings.Contains(r, "vendor") ||
				strings.Contains(r, "dist") ||
				strings.Contains(r, "build") ||
				strings.Contains(r, ".next") ||
				strings.Contains(r, ".nuxt") ||
				strings.Contains(r, "__pycache__") {
				t.Errorf("Found file in skipped directory: %s", r)
			}
		}
	})

	t.Run("nested directories", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create nested structure
		mkdir(t, tmpDir, "frontend")
		mkdir(t, tmpDir, "frontend/components")
		mkdir(t, tmpDir, "backend")
		mkdir(t, tmpDir, "backend/api")
		mkdir(t, tmpDir, "backend/api/v1")

		// Add .env files at various levels
		writeFile(t, tmpDir, ".env", "ROOT=value")
		writeFile(t, tmpDir, "frontend/.env", "FRONTEND=value")
		writeFile(t, tmpDir, "frontend/components/.env", "COMPONENTS=value")
		writeFile(t, tmpDir, "backend/.env", "BACKEND=value")
		writeFile(t, tmpDir, "backend/api/.env", "API=value")
		writeFile(t, tmpDir, "backend/api/v1/.env", "API_V1=value")

		results, err := Scan(tmpDir)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		expected := []string{
			".env",
			"frontend/.env",
			"frontend/components/.env",
			"backend/.env",
			"backend/api/.env",
			"backend/api/v1/.env",
		}

		if len(results) != len(expected) {
			t.Errorf("Expected %d files, got %d: %v", len(expected), len(results), results)
		}

		resultMap := make(map[string]bool)
		for _, r := range results {
			resultMap[r] = true
		}

		for _, exp := range expected {
			if !resultMap[exp] {
				t.Errorf("Missing expected file: %s", exp)
			}
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		results, err := Scan(tmpDir)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(results) != 0 {
			t.Errorf("Expected 0 files, got %d: %v", len(results), results)
		}
	})
}

// Helper functions for test setup
func writeFile(t *testing.T, base, name, content string) {
	t.Helper()
	path := filepath.Join(base, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
}

func mkdir(t *testing.T, base, name string) {
	t.Helper()
	path := filepath.Join(base, name)
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", path, err)
	}
}
