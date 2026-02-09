// Package backup provides utilities for creating backup files.
package backup

import (
	"fmt"
	"io"
	"os"
	"time"
)

// CreateBackup creates a timestamped backup of the file at the given path.
// Returns the backup file path on success, or an error if the backup fails.
// If the source file does not exist, returns empty string and no error.
// Preserves the original file's permissions.
func CreateBackup(path string) (string, error) {
	return CreateBackupWithFS(path, realFS{})
}

// realFS is the default filesystem implementation.
type realFS struct{}

func (realFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
func (realFS) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}
func (realFS) CreateWithMode(name string, mode os.FileMode) (io.WriteCloser, error) {
	return os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
}

// FileSystem defines file operations for testing.
type FileSystem interface {
	Stat(name string) (os.FileInfo, error)
	Open(name string) (io.ReadCloser, error)
	CreateWithMode(name string, mode os.FileMode) (io.WriteCloser, error)
}

// CreateBackupWithFS creates a timestamped backup using the provided filesystem interface.
// Preserves the original file's permissions.
func CreateBackupWithFS(path string, fs FileSystem) (string, error) {
	info, err := fs.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, no backup needed
			return "", nil
		}
		return "", fmt.Errorf("failed to stat source file: %w", err)
	}

	timestamp := time.Now().Format("20060102150405.999999999")
	backupPath := fmt.Sprintf("%s.bak.%s", path, timestamp)

	srcFile, err := fs.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	destFile, err := fs.CreateWithMode(backupPath, info.Mode())
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}

	if _, err := io.Copy(destFile, srcFile); err != nil {
		_ = destFile.Close()
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
	ts := timestamp.Format("20060102150405.999999999")
	return fmt.Sprintf("%s.bak.%s", path, ts)
}
