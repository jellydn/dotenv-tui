// Package backup provides utilities for creating backup files.
package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// CreateBackup creates a timestamped backup of the file at the given path.
// Returns the backup file path on success, or an error if the backup fails.
// If the source file does not exist, returns empty string and no error.
func CreateBackup(path string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File doesn't exist, no backup needed
		return "", nil
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102150405")
	backupPath := fmt.Sprintf("%s.bak.%s", path, timestamp)

	// Copy the original file to backup
	if err := copyFile(path, backupPath); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	return backupPath, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = sourceFile.Close() }()

	// Get source file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	destFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return destFile.Close()
}

// CreateBackupWithFS creates a backup using a FileSystem interface for testing.
type FileSystem interface {
	Stat(name string) (os.FileInfo, error)
	Open(name string) (io.ReadCloser, error)
	Create(name string) (io.WriteCloser, error)
}

// CreateBackupWithFS creates a timestamped backup using the provided filesystem interface.
func CreateBackupWithFS(path string, fs FileSystem) (string, error) {
	// Check if file exists
	if _, err := fs.Stat(path); os.IsNotExist(err) {
		// File doesn't exist, no backup needed
		return "", nil
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102150405")
	backupPath := fmt.Sprintf("%s.bak.%s", path, timestamp)

	// Copy the file
	srcFile, err := fs.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	destFile, err := fs.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer func() { _ = destFile.Close() }()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	if err := destFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close backup file: %w", err)
	}

	return backupPath, nil
}

// GetBackupPath generates a backup path for the given file.
// This is useful for testing or displaying the backup path without creating it.
func GetBackupPath(path string, timestamp time.Time) string {
	ts := timestamp.Format("20060102150405")
	return fmt.Sprintf("%s.bak.%s", path, ts)
}

// GetBackupDir returns the directory where backups would be stored.
func GetBackupDir(path string) string {
	return filepath.Dir(path)
}
