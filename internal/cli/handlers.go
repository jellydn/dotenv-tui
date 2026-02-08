// Package cli contains command-line interface handlers for dotenv-tui.
package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jellydn/dotenv-tui/internal/backup"
	"github.com/jellydn/dotenv-tui/internal/generator"
	"github.com/jellydn/dotenv-tui/internal/parser"
	"github.com/jellydn/dotenv-tui/internal/scanner"
)

// EntryProcessor is a function that processes entries from a .env file.
type EntryProcessor func([]parser.Entry) []parser.Entry

// FileSystem defines file operations for testing.
type FileSystem interface {
	Open(name string) (io.ReadCloser, error)
	Stat(name string) (os.FileInfo, error)
	Create(name string) (io.WriteCloser, error)
}

// DirScanner defines directory scanning operations for testing.
type DirScanner interface {
	Scan(root string) ([]string, error)
	ScanExamples(root string) ([]string, error)
}

// RealFileSystem is the default filesystem implementation.
type RealFileSystem struct{}

// Open implements FileSystem.Open.
func (RealFileSystem) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

// Stat implements FileSystem.Stat.
func (RealFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// Create implements FileSystem.Create.
func (RealFileSystem) Create(name string) (io.WriteCloser, error) {
	return os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
}

// RealDirScanner is the default scanner implementation using the scanner package.
type RealDirScanner struct{}

// Scan implements DirScanner.Scan.
func (RealDirScanner) Scan(root string) ([]string, error) {
	return scanner.Scan(root)
}

// ScanExamples implements DirScanner.ScanExamples.
func (RealDirScanner) ScanExamples(root string) ([]string, error) {
	return scanner.ScanExamples(root)
}

// GenerateFile generates a file from an input file, processing entries with the provided function.
func GenerateFile(inputPath string, force bool, createBackup bool, outputFilename string, processEntries EntryProcessor, parseErrMsg string, fs FileSystem, out io.Writer) error {
	file, err := fs.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer func() { _ = file.Close() }()

	entries, err := parser.Parse(file)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", parseErrMsg, err)
	}

	processedEntries := processEntries(entries)

	outputPath := filepath.Join(filepath.Dir(inputPath), outputFilename)

	if _, err := fs.Stat(outputPath); err == nil && !force {
		return fmt.Errorf("%s already exists. Use --force to overwrite", outputPath)
	}

	// Create backup if file exists and backups are enabled
	if createBackup {
		backupPath, err := backup.CreateBackupWithFS(outputPath, fs)
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		if backupPath != "" {
			_, _ = fmt.Fprintf(out, "Backup created: %s\n", backupPath)
		}
	}

	outFile, err := fs.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}

	if err := parser.Write(outFile, processedEntries); err != nil {
		_ = outFile.Close()
		return fmt.Errorf("failed to write output file: %w", err)
	}

	if err := outFile.Close(); err != nil {
		return fmt.Errorf("failed to close output file: %w", err)
	}

	_, _ = fmt.Fprintf(out, "Generated %s\n", outputPath)
	return nil
}

// GenerateExampleFile generates a .env.example file from a .env file.
func GenerateExampleFile(inputPath string, force bool, createBackup bool, fs FileSystem, out io.Writer) error {
	return GenerateFile(inputPath, force, createBackup, ".env.example", generator.GenerateExample, ".env file", fs, out)
}

// GenerateEnvFile generates a .env file from a .env.example file.
func GenerateEnvFile(inputPath string, force bool, createBackup bool, fs FileSystem, out io.Writer) error {
	return GenerateFile(inputPath, force, createBackup, ".env", func(entries []parser.Entry) []parser.Entry {
		return entries
	}, ".env.example file", fs, out)
}

// ScanAndList scans a directory for .env files and lists them.
func ScanAndList(dir string, sc DirScanner, out io.Writer) error {
	if dir == "" {
		dir = "."
	}

	files, err := sc.Scan(dir)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(files) == 0 {
		_, _ = fmt.Fprintln(out, "No .env files found")
		return nil
	}

	_, _ = fmt.Fprintf(out, "Found %d .env file(s):\n", len(files))
	for _, file := range files {
		_, _ = fmt.Fprintf(out, "  %s\n", file)
	}

	return nil
}

// GenerateAllEnvFiles generates .env files from all .env.example files.
func GenerateAllEnvFiles(force bool, createBackup bool, fs FileSystem, sc DirScanner, in io.Reader, out io.Writer) error {
	exampleFiles, err := sc.ScanExamples(".")
	if err != nil {
		return fmt.Errorf("failed to scan for .env.example files: %w", err)
	}

	if len(exampleFiles) == 0 {
		return fmt.Errorf("no .env.example files found")
	}

	_, _ = fmt.Fprintf(out, "Found %d .env.example file(s):\n", len(exampleFiles))
	for _, file := range exampleFiles {
		_, _ = fmt.Fprintf(out, "  %s\n", file)
	}

	var generated, skipped int
	for _, exampleFile := range exampleFiles {
		if err := ProcessExampleFile(exampleFile, force, createBackup, &generated, &skipped, fs, in, out); err != nil {
			return err
		}
	}

	_, _ = fmt.Fprintf(out, "Done: %d generated, %d skipped\n", generated, skipped)
	return nil
}

// ProcessExampleFile processes a single .env.example file and generates a .env file.
func ProcessExampleFile(exampleFile string, force bool, createBackup bool, generated, skipped *int, fs FileSystem, in io.Reader, out io.Writer) error {
	outputPath := strings.TrimSuffix(exampleFile, ".example")

	entries, err := parseAndClose(exampleFile, fs)
	if err != nil {
		return err
	}

	if !force && fileExists(fs, outputPath) {
		confirmed, err := confirmOverwrite(out, outputPath, in)
		if err != nil {
			return err
		}
		if !confirmed {
			_, _ = fmt.Fprintf(out, "Skipped %s\n", outputPath)
			*skipped++
			return nil
		}
	}

	// Create backup if file exists and backups are enabled
	if createBackup {
		backupPath, err := backup.CreateBackupWithFS(outputPath, fs)
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		if backupPath != "" {
			_, _ = fmt.Fprintf(out, "Backup created: %s\n", backupPath)
		}
	}

	if err := writeEntries(outputPath, fs, entries); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(out, "Generated %s\n", outputPath)
	*generated++
	return nil
}

func fileExists(fs FileSystem, path string) bool {
	_, err := fs.Stat(path)
	return err == nil
}

func confirmOverwrite(out io.Writer, path string, in io.Reader) (bool, error) {
	_, _ = fmt.Fprintf(out, "%s already exists. Overwrite? [y/N] ", path)
	reader := bufio.NewReader(in)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}
	response = strings.TrimSpace(response)
	return response == "y" || response == "Y", nil
}

func parseAndClose(path string, fs FileSystem) ([]parser.Entry, error) {
	file, err := fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	entries, err := parser.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return entries, nil
}

func writeEntries(path string, fs FileSystem, entries []parser.Entry) error {
	outFile, err := fs.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", path, err)
	}

	if err := parser.Write(outFile, entries); err != nil {
		_ = outFile.Close()
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	if err := outFile.Close(); err != nil {
		return fmt.Errorf("failed to close %s: %w", path, err)
	}

	return nil
}
