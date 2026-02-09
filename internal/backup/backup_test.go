package backup

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// mockFileSystem is a mock implementation of FileSystem for testing.
type mockFileSystem struct {
	files       map[string]string
	modes       map[string]os.FileMode
	createError error
	openError   error
	statError   error
}

func newMockFileSystem() *mockFileSystem {
	return &mockFileSystem{
		files: make(map[string]string),
		modes: make(map[string]os.FileMode),
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
	mode := m.modes[name]
	if mode == 0 {
		mode = 0600 // default mode
	}
	return mockFileInfo{name: name, mode: mode}, nil
}

func (m *mockFileSystem) CreateWithMode(name string, mode os.FileMode) (io.WriteCloser, error) {
	if m.createError != nil {
		return nil, m.createError
	}
	writer := &mockWriteCloser{
		buffer: &bytes.Buffer{},
		mode:   mode,
		onClose: func(content string) {
			m.files[name] = content
			m.modes[name] = mode
		},
	}
	return writer, nil
}

type mockWriteCloser struct {
	buffer  *bytes.Buffer
	onClose func(string)
	closed  bool
	mode    os.FileMode
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
	mode os.FileMode
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return 0 }
func (m mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m mockFileInfo) IsDir() bool        { return false }
func (m mockFileInfo) Sys() any           { return nil }

func TestCreateBackupWithFS(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		fileContent string
		fileMode    os.FileMode
		fileExists  bool
		wantBackup  bool
		wantErr     bool
	}{
		{
			name:        "creates backup for existing file",
			path:        "/test/.env",
			fileContent: "API_KEY=secret123\nPORT=3000\n",
			fileMode:    0600,
			fileExists:  true,
			wantBackup:  true,
			wantErr:     false,
		},
		{
			name:       "no backup for non-existent file",
			path:       "/test/.env",
			fileExists: false,
			wantBackup: false,
			wantErr:    false,
		},
		{
			name:        "preserves file permissions",
			path:        "/test/.env",
			fileContent: "API_KEY=secret123\nPORT=3000\n",
			fileMode:    0644,
			fileExists:  true,
			wantBackup:  true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newMockFileSystem()
			if tt.fileExists {
				fs.files[tt.path] = tt.fileContent
				fs.modes[tt.path] = tt.fileMode
			}

			backupPath, err := CreateBackupWithFS(tt.path, fs)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if tt.wantBackup {
					if backupPath == "" {
						t.Errorf("expected backup path but got empty string")
					}
					if !strings.HasPrefix(backupPath, tt.path+".bak.") {
						t.Errorf("backup path = %q, want prefix %q", backupPath, tt.path+".bak.")
					}
					backupContent, ok := fs.files[backupPath]
					if !ok {
						t.Errorf("backup file not created at %s", backupPath)
					}
					if backupContent != tt.fileContent {
						t.Errorf("backup content = %q, want %q", backupContent, tt.fileContent)
					}
					backupMode, ok := fs.modes[backupPath]
					if !ok {
						t.Errorf("backup mode not found")
					}
					if backupMode != tt.fileMode {
						t.Errorf("backup mode = %v, want %v", backupMode, tt.fileMode)
					}
				} else if backupPath != "" {
					t.Errorf("expected no backup but got %q", backupPath)
				}
			}
		})
	}
}

func TestGetBackupPath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		timestamp time.Time
		want      string
	}{
		{
			name:      "generates correct backup path with nanoseconds",
			path:      "/test/.env",
			timestamp: time.Date(2026, 2, 8, 10, 30, 45, 123456789, time.UTC),
			want:      "/test/.env.bak.20260208103045.123456789",
		},
		{
			name:      "handles relative path without nanoseconds",
			path:      ".env",
			timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			want:      ".env.bak.20260101000000", // Go doesn't add decimal when nanos are 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBackupPath(tt.path, tt.timestamp)
			if got != tt.want {
				t.Errorf("GetBackupPath() = = %q, want %q", got, tt.want)
			}
		})
	}
}
