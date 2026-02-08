// Package cli contains command-line interface handlers for dotenv-tui.
package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

// RealFileSystem is the default filesystem implementation.
type RealFileSystem struct{}

// Open opens a file for reading.
func (RealFileSystem) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

// Stat returns file information.
func (RealFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// Create creates a file for writing.
func (RealFileSystem) Create(name string) (io.WriteCloser, error) {
	return os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
}

// GenerateFile generates a file from an input file, processing entries with the provided function.
func GenerateFile(inputPath string, force bool, outputFilename string, processEntries EntryProcessor, parseErrMsg string, fs FileSystem, out io.Writer) error {
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

	outFile, err := fs.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	if err := parser.Write(outFile, processedEntries); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Fprintf(out, "Generated %s\n", outputPath)
	return nil
}

// GenerateExampleFile generates a .env.example file from a .env file.
func GenerateExampleFile(inputPath string, force bool, fs FileSystem, out io.Writer) error {
	return GenerateFile(inputPath, force, ".env.example", generator.GenerateExample, ".env file", fs, out)
}

// GenerateEnvFile generates a .env file from a .env.example file.
func GenerateEnvFile(inputPath string, force bool, fs FileSystem, out io.Writer) error {
	return GenerateFile(inputPath, force, ".env", func(entries []parser.Entry) []parser.Entry {
		return entries
	}, ".env.example file", fs, out)
}

// ScanAndList scans a directory for .env files and lists them.
func ScanAndList(dir string, out io.Writer) error {
	if dir == "" {
		dir = "."
	}

	files, err := scanner.Scan(dir)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(files) == 0 {
		fmt.Fprintln(out, "No .env files found")
		return nil
	}

	fmt.Fprintf(out, "Found %d .env file(s):\n", len(files))
	for _, file := range files {
		fmt.Fprintf(out, "  %s\n", file)
	}

	return nil
}

// GenerateAllEnvFiles generates .env files from all .env.example files.
func GenerateAllEnvFiles(force bool, fs FileSystem, in io.Reader, out io.Writer) error {
	exampleFiles, err := scanner.ScanExamples(".")
	if err != nil {
		return fmt.Errorf("failed to scan for .env.example files: %w", err)
	}

	if len(exampleFiles) == 0 {
		return fmt.Errorf("no .env.example files found")
	}

	fmt.Fprintf(out, "Found %d .env.example file(s):\n", len(exampleFiles))
	for _, file := range exampleFiles {
		fmt.Fprintf(out, "  %s\n", file)
	}

	var generated, skipped int
	for _, exampleFile := range exampleFiles {
		if err := ProcessExampleFile(exampleFile, force, &generated, &skipped, fs, in, out); err != nil {
			return err
		}
	}

	fmt.Fprintf(out, "Done: %d generated, %d skipped\n", generated, skipped)
	return nil
}

// ProcessExampleFile processes a single .env.example file and generates a .env file.
func ProcessExampleFile(exampleFile string, force bool, generated, skipped *int, fs FileSystem, in io.Reader, out io.Writer) error {
	outputPath := strings.TrimSuffix(exampleFile, ".example")

	file, err := fs.Open(exampleFile)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", exampleFile, err)
	}

	entries, err := parser.Parse(file)
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to parse %s: %w", exampleFile, err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close %s: %w", exampleFile, err)
	}

	if _, err := fs.Stat(outputPath); err == nil && !force {
		fmt.Fprintf(out, "%s already exists. Overwrite? [y/N] ", outputPath)
		reader := bufio.NewReader(in)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}
		response = strings.TrimSpace(response)

		if response != "y" && response != "Y" {
			fmt.Fprintf(out, "Skipped %s\n", outputPath)
			*skipped++
			return nil
		}
	}

	outFile, err := fs.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", outputPath, err)
	}

	if err := parser.Write(outFile, entries); err != nil {
		_ = outFile.Close()
		return fmt.Errorf("failed to write %s: %w", outputPath, err)
	}

	if err := outFile.Close(); err != nil {
		return fmt.Errorf("failed to close %s: %w", outputPath, err)
	}

	fmt.Fprintf(out, "Generated %s\n", outputPath)
	*generated++
	return nil
}
