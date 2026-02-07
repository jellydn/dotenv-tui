// Package scanner recursively finds .env files in project directories.
package scanner

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

var skipDirs = map[string]bool{
	"node_modules": true,
	".git":         true,
	"vendor":       true,
	"dist":         true,
	"build":        true,
	".next":        true,
	".nuxt":        true,
	"__pycache__":  true,
}

func scanFiles(root string, match func(fileName string) bool) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		pathParts := strings.Split(relPath, string(filepath.Separator))
		for _, part := range pathParts {
			if skipDirs[part] {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
		}

		if d.IsDir() {
			return nil
		}

		fileName := d.Name()
		if match(fileName) {
			files = append(files, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning directory: %w", err)
	}

	return files, nil
}

// Scan recursively finds .env files in a project tree, skipping dependency directories.
func Scan(root string) ([]string, error) {
	return scanFiles(root, isEnvFile)
}

// ScanExamples finds .env.example files in a project tree, skipping dependency directories.
func ScanExamples(root string) ([]string, error) {
	return scanFiles(root, isExampleFile)
}

func isEnvFile(fileName string) bool {
	if strings.HasSuffix(fileName, ".example") {
		return false
	}

	return strings.HasPrefix(fileName, ".env") && (fileName == ".env" || (len(fileName) > 4 && fileName[4] == '.'))
}

func isExampleFile(fileName string) bool {
	return strings.HasPrefix(fileName, ".env") && strings.HasSuffix(fileName, ".example")
}
