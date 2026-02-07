// Package scanner recursively finds .env files in project directories.
package scanner

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// Scan recursively finds .env files in a project tree, skipping dependency directories.
// Returns list of .env file paths relative to root.
func Scan(root string) ([]string, error) {
	var envFiles []string

	skipDirs := map[string]bool{
		"node_modules": true,
		".git":         true,
		"vendor":       true,
		"dist":         true,
		"build":        true,
		".next":        true,
		".nuxt":        true,
		"__pycache__":  true,
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip files/dirs that can't be accessed
		}

		// Get relative path
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		// Skip if we're in a directory to skip
		pathParts := strings.Split(relPath, string(filepath.Separator))
		for _, part := range pathParts {
			if skipDirs[part] {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
		}

		// Only check files, not directories
		if d.IsDir() {
			return nil
		}

		fileName := d.Name()
		if isEnvFile(fileName) {
			envFiles = append(envFiles, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning directory: %w", err)
	}

	return envFiles, nil
}

// isEnvFile checks if a filename matches .env patterns but excludes example files
func isEnvFile(fileName string) bool {
	// Skip .env.example and .env.*.example
	if strings.HasSuffix(fileName, ".example") {
		return false
	}

	// Match .env or .env.* (but not .env alone without extension)
	// .env, .env.local, .env.production, etc.
	return strings.HasPrefix(fileName, ".env") && (fileName == ".env" || (len(fileName) > 4 && fileName[4] == '.'))
}
