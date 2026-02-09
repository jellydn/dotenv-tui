package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jellydn/dotenv-tui/internal/parser"
)

// mockFileSystem is a mock implementation of FileSystem for testing.
type mockFileSystem struct {
	files       map[string]string
	createError error
	openError   error
	statError   error
}

func newMockFileSystem() *mockFileSystem {
	return &mockFileSystem{
		files: make(map[string]string),
	}
}

func (m *mockFileSystem) Open(name string) (io.ReadCloser, error) {
	if m.openError != nil {
		return nil, m.openError
	}
	content, ok := m.files[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return io.NopCloser(strings.NewReader(content)), nil
}

func (m *mockFileSystem) Stat(name string) (os.FileInfo, error) {
	if m.statError != nil {
		return nil, m.statError
	}
	_, ok := m.files[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	// Return a simple mock file info
	return mockFileInfo{name: name}, nil
}

func (m *mockFileSystem) Create(name string) (io.WriteCloser, error) {
	return m.CreateWithMode(name, 0600)
}

func (m *mockFileSystem) CreateWithMode(name string, mode os.FileMode) (io.WriteCloser, error) {
	if m.createError != nil {
		return nil, m.createError
	}
	writer := &mockWriteCloser{
		buffer: &bytes.Buffer{},
		onClose: func(content string) {
			m.files[name] = content
		},
	}
	return writer, nil
}

type mockWriteCloser struct {
	buffer  *bytes.Buffer
	onClose func(string)
	closed  bool
}

func (m *mockWriteCloser) Write(p []byte) (n int, err error) {
	return m.buffer.Write(p)
}

func (m *mockWriteCloser) Close() error {
	if m.closed {
		return fmt.Errorf("already closed")
	}
	m.closed = true
	if m.onClose != nil {
		m.onClose(m.buffer.String())
	}
	return nil
}

type mockFileInfo struct {
	name string
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return 0 }
func (m mockFileInfo) Mode() os.FileMode  { return 0 }
func (m mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m mockFileInfo) IsDir() bool        { return false }
func (m mockFileInfo) Sys() interface{}   { return nil }

type mockDirScanner struct {
	scanFiles    []string
	scanErr      error
	exampleFiles []string
	exampleErr   error
}

func (m *mockDirScanner) Scan(_ string) ([]string, error) {
	return m.scanFiles, m.scanErr
}

func (m *mockDirScanner) ScanExamples(_ string) ([]string, error) {
	return m.exampleFiles, m.exampleErr
}

func TestGenerateExampleFile(t *testing.T) {
	tests := []struct {
		name           string
		inputContent   string
		force          bool
		existingOutput bool
		wantErr        bool
		errContains    string
		wantOutput     string
	}{
		{
			name:         "successful generation",
			inputContent: "API_KEY=secret123\nPORT=3000\n",
			force:        false,
			wantErr:      false,
			wantOutput:   "API_KEY=***\nPORT=3000\n",
		},
		{
			name:           "file exists without force",
			inputContent:   "KEY=value\n",
			force:          false,
			existingOutput: true,
			wantErr:        true,
			errContains:    "already exists",
		},
		{
			name:           "file exists with force",
			inputContent:   "API_KEY=secret123\n",
			force:          true,
			existingOutput: true,
			wantErr:        false,
			wantOutput:     "API_KEY=***\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newMockFileSystem()
			fs.files["/test/.env"] = tt.inputContent
			if tt.existingOutput {
				fs.files["/test/.env.example"] = "existing content"
			}

			var out bytes.Buffer
			err := GenerateExampleFile("/test/.env", tt.force, true, false, fs, &out)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want substring %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.wantOutput != "" {
					got := fs.files["/test/.env.example"]
					if got != tt.wantOutput {
						t.Errorf("output = %q, want %q", got, tt.wantOutput)
					}
				}
			}
		})
	}
}

func TestGenerateEnvFile(t *testing.T) {
	tests := []struct {
		name           string
		inputContent   string
		force          bool
		existingOutput bool
		wantErr        bool
		errContains    string
		wantOutput     string
	}{
		{
			name:         "successful generation from example",
			inputContent: "API_KEY=***\nPORT=3000\n",
			force:        false,
			wantErr:      false,
			wantOutput:   "API_KEY=***\nPORT=3000\n",
		},
		{
			name:           "file exists without force",
			inputContent:   "KEY=value\n",
			force:          false,
			existingOutput: true,
			wantErr:        true,
			errContains:    "already exists",
		},
		{
			name:           "file exists with force",
			inputContent:   "KEY=value\n",
			force:          true,
			existingOutput: true,
			wantErr:        false,
			wantOutput:     "KEY=value\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newMockFileSystem()
			fs.files["/test/.env.example"] = tt.inputContent
			if tt.existingOutput {
				fs.files["/test/.env"] = "existing content"
			}

			var out bytes.Buffer
			err := GenerateEnvFile("/test/.env.example", tt.force, true, false, fs, &out)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want substring %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.wantOutput != "" {
					got := fs.files["/test/.env"]
					if got != tt.wantOutput {
						t.Errorf("output = %q, want %q", got, tt.wantOutput)
					}
				}
			}
		})
	}
}

func TestScanAndList(t *testing.T) {
	tests := []struct {
		name       string
		dir        string
		scanFiles  []string
		scanErr    error
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "empty directory",
			dir:        ".",
			scanFiles:  nil,
			wantOutput: "No .env files found",
		},
		{
			name:       "found files",
			dir:        ".",
			scanFiles:  []string{".env", "sub/.env.local"},
			wantOutput: "Found 2 .env file(s):",
		},
		{
			name:    "scan error",
			dir:     ".",
			scanErr: fmt.Errorf("permission denied"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &mockDirScanner{scanFiles: tt.scanFiles, scanErr: tt.scanErr}
			var out bytes.Buffer
			err := ScanAndList(tt.dir, sc, &out)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			got := out.String()
			if !strings.Contains(got, tt.wantOutput) {
				t.Errorf("output = %q, want substring %q", got, tt.wantOutput)
			}
		})
	}
}

func TestProcessExampleFile(t *testing.T) {
	tests := []struct {
		name          string
		inputContent  string
		force         bool
		userInput     string
		existingFile  bool
		wantGenerated int
		wantSkipped   int
		wantErr       bool
	}{
		{
			name:          "generate new file",
			inputContent:  "KEY=value\n",
			force:         false,
			existingFile:  false,
			wantGenerated: 1,
			wantSkipped:   0,
			wantErr:       false,
		},
		{
			name:          "force overwrite existing file",
			inputContent:  "KEY=value\n",
			force:         true,
			existingFile:  true,
			wantGenerated: 1,
			wantSkipped:   0,
			wantErr:       false,
		},
		{
			name:          "user says yes to overwrite",
			inputContent:  "KEY=value\n",
			force:         false,
			userInput:     "y\n",
			existingFile:  true,
			wantGenerated: 1,
			wantSkipped:   0,
			wantErr:       false,
		},
		{
			name:          "user says no to overwrite",
			inputContent:  "KEY=value\n",
			force:         false,
			userInput:     "n\n",
			existingFile:  true,
			wantGenerated: 0,
			wantSkipped:   1,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newMockFileSystem()
			fs.files["/test/.env.example"] = tt.inputContent
			if tt.existingFile {
				fs.files["/test/.env"] = "existing"
			}

			var out bytes.Buffer
			in := strings.NewReader(tt.userInput)
			generated, skipped := 0, 0

			err := ProcessExampleFile("/test/.env.example", tt.force, true, &generated, &skipped, fs, in, &out)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if generated != tt.wantGenerated {
					t.Errorf("generated = %d, want %d", generated, tt.wantGenerated)
				}
				if skipped != tt.wantSkipped {
					t.Errorf("skipped = %d, want %d", skipped, tt.wantSkipped)
				}
			}
		})
	}
}

func TestGenerateFile(t *testing.T) {
	tests := []struct {
		name           string
		inputContent   string
		force          bool
		outputFilename string
		existingOutput bool
		wantErr        bool
		errContains    string
	}{
		{
			name:           "successful generation",
			inputContent:   "KEY=value\n",
			force:          false,
			outputFilename: "output.env",
			wantErr:        false,
		},
		{
			name:           "existing file without force",
			inputContent:   "KEY=value\n",
			force:          false,
			outputFilename: "output.env",
			existingOutput: true,
			wantErr:        true,
			errContains:    "already exists",
		},
		{
			name:           "invalid input file",
			inputContent:   "",
			force:          false,
			outputFilename: "output.env",
			wantErr:        true,
			errContains:    "failed to open input file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newMockFileSystem()
			if tt.inputContent != "" {
				fs.files["/test/input.env"] = tt.inputContent
			}
			if tt.existingOutput {
				fs.files["/test/"+tt.outputFilename] = "existing"
			}

			var out bytes.Buffer
			processEntries := func(entries []parser.Entry) []parser.Entry {
				return entries
			}

			inputPath := "/test/input.env"
			if tt.inputContent == "" {
				inputPath = "/test/nonexistent.env"
			}

			err := GenerateFile(inputPath, tt.force, true, false, tt.outputFilename, processEntries, "test file", fs, &out)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want substring %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestDryRunGenerateExampleFile(t *testing.T) {
	tests := []struct {
		name           string
		inputContent   string
		existingOutput bool
		wantErr        bool
		wantInOutput   []string
	}{
		{
			name:           "dry-run preview for new file",
			inputContent:   "API_KEY=secret123\nPORT=3000\n",
			existingOutput: false,
			wantErr:        false,
			wantInOutput:   []string{"DRY RUN PREVIEW", ".env.example", "Would CREATE new file", "API_KEY=***", "PORT=3000"},
		},
		{
			name:           "dry-run preview for overwrite",
			inputContent:   "API_KEY=secret123\n",
			existingOutput: true,
			wantErr:        false,
			wantInOutput:   []string{"DRY RUN PREVIEW", ".env.example", "Would OVERWRITE existing file", "API_KEY=***"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newMockFileSystem()
			fs.files["/test/.env"] = tt.inputContent
			if tt.existingOutput {
				fs.files["/test/.env.example"] = "existing content"
			}

			var out bytes.Buffer
			err := GenerateExampleFile("/test/.env", false, false, true, fs, &out)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Verify file was NOT written in dry-run mode
			if _, exists := fs.files["/test/.env.example"]; !tt.existingOutput && exists {
				t.Errorf("file should not be created in dry-run mode")
			}

			// Verify output contains expected strings
			outputStr := out.String()
			for _, want := range tt.wantInOutput {
				if !strings.Contains(outputStr, want) {
					t.Errorf("output missing expected string %q, got:\n%s", want, outputStr)
				}
			}
		})
	}
}

func TestDryRunGenerateEnvFile(t *testing.T) {
	tests := []struct {
		name           string
		inputContent   string
		existingOutput bool
		wantErr        bool
		wantInOutput   []string
	}{
		{
			name:           "dry-run preview for new env file",
			inputContent:   "API_KEY=***\nPORT=3000\n",
			existingOutput: false,
			wantErr:        false,
			wantInOutput:   []string{"DRY RUN PREVIEW", ".env", "Would CREATE new file", "API_KEY=***", "PORT=3000"},
		},
		{
			name:           "dry-run preview for overwrite env file",
			inputContent:   "DATABASE_URL=***\n",
			existingOutput: true,
			wantErr:        false,
			wantInOutput:   []string{"DRY RUN PREVIEW", ".env", "Would OVERWRITE existing file", "DATABASE_URL=***"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newMockFileSystem()
			fs.files["/test/.env.example"] = tt.inputContent
			if tt.existingOutput {
				fs.files["/test/.env"] = "existing content"
			}

			var out bytes.Buffer
			err := GenerateEnvFile("/test/.env.example", false, false, true, fs, &out)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Verify file was NOT written in dry-run mode
			if _, exists := fs.files["/test/.env"]; !tt.existingOutput && exists {
				t.Errorf("file should not be created in dry-run mode")
			}

			// Verify output contains expected strings
			outputStr := out.String()
			for _, want := range tt.wantInOutput {
				if !strings.Contains(outputStr, want) {
					t.Errorf("output missing expected string %q, got:\n%s", want, outputStr)
				}
			}
		})
	}
}

func TestDryRunGenerateAllEnvFiles(t *testing.T) {
	tests := []struct {
		name         string
		exampleFiles []string
		wantErr      bool
		wantInOutput []string
	}{
		{
			name:         "dry-run preview for multiple files",
			exampleFiles: []string{"/test/.env.example", "/test/dir/.env.example"},
			wantErr:      false,
			wantInOutput: []string{"DRY RUN MODE", "DRY RUN PREVIEW", "/test/.env", "/test/dir/.env"},
		},
		{
			name:         "dry-run with no example files",
			exampleFiles: []string{},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newMockFileSystem()
			for _, file := range tt.exampleFiles {
				fs.files[file] = "KEY=value\n"
			}

			sc := &mockDirScanner{exampleFiles: tt.exampleFiles}

			var out bytes.Buffer
			err := GenerateAllEnvFiles(false, false, true, fs, sc, strings.NewReader(""), &out)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Verify no .env files were created in dry-run mode
			for _, exampleFile := range tt.exampleFiles {
				envFile := strings.TrimSuffix(exampleFile, ".example")
				if _, exists := fs.files[envFile]; exists {
					t.Errorf("file %s should not be created in dry-run mode", envFile)
				}
			}

			// Verify output contains expected strings
			outputStr := out.String()
			for _, want := range tt.wantInOutput {
				if !strings.Contains(outputStr, want) {
					t.Errorf("output missing expected string %q, got:\n%s", want, outputStr)
				}
			}
		})
	}
}
