// Package upgrade provides functionality for self-updating the dotenv-tui binary.
package upgrade

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	tests := []struct {
		name            string
		expectedOS      string
		expectedArch    string
		expectedWarning string
	}{
		{
			name:         "standard linux amd64",
			expectedOS:   "linux",
			expectedArch: "amd64",
		},
		{
			name:         "standard darwin arm64",
			expectedOS:   "darwin",
			expectedArch: "arm64",
		},
		{
			name:         "standard windows 386",
			expectedOS:   "windows",
			expectedArch: "386",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			osType, arch := detectPlatform()

			validOS := osType == "linux" || osType == "darwin" || osType == "windows"
			validArch := arch == "amd64" || arch == "386" || arch == "arm" || arch == "arm64"

			if !validOS {
				t.Errorf("detectPlatform() osType = %q, expected one of [linux, darwin, windows]", osType)
			}
			if !validArch {
				t.Errorf("detectPlatform() arch = %q, expected one of [amd64, 386, arm, arm64]", arch)
			}
		})
	}
}

func TestDetectPlatformAliases(t *testing.T) {
	tests := []struct {
		name            string
		inputOS         string
		inputArch       string
		expectedOS      string
		expectedArch    string
		expectedWarning string
	}{
		{
			name:         "x86_64 normalized to amd64",
			inputArch:    "x86_64",
			expectedOS:   runtime.GOOS,
			expectedArch: "amd64",
		},
		{
			name:         "aarch64 normalized to arm64",
			inputArch:    "aarch64",
			expectedOS:   runtime.GOOS,
			expectedArch: "arm64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if strings.Contains(tt.name, "x86_64") || strings.Contains(tt.name, "aarch64") {
				osType, arch := detectPlatform()
				if osType == "" || arch == "" {
					t.Error("detectPlatform() returned empty values")
				}
			}
		})
	}
}

func TestGetLatestVersion(t *testing.T) {
	t.Run("successful version fetch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"tag_name": "v1.2.3"}`))
		}))
		defer server.Close()

		original := githubAPIURL
		githubAPIURL = server.URL
		defer func() { githubAPIURL = original }()

		version, err := getLatestVersion()
		if err != nil {
			t.Fatalf("getLatestVersion() unexpected error: %v", err)
		}
		if version != "v1.2.3" {
			t.Errorf("getLatestVersion() = %q, want %q", version, "v1.2.3")
		}
	})

	t.Run("network error", func(t *testing.T) {
		original := githubAPIURL
		githubAPIURL = "http://localhost:1" // connection refused
		defer func() { githubAPIURL = original }()

		_, err := getLatestVersion()
		if err == nil {
			t.Error("getLatestVersion() expected error for network failure, got nil")
		}
	})

	t.Run("non-200 status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		original := githubAPIURL
		githubAPIURL = server.URL
		defer func() { githubAPIURL = original }()

		_, err := getLatestVersion()
		if err == nil {
			t.Error("getLatestVersion() expected error for non-200 status, got nil")
		}
	})

	t.Run("empty tag name", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"tag_name": ""}`))
		}))
		defer server.Close()

		original := githubAPIURL
		githubAPIURL = server.URL
		defer func() { githubAPIURL = original }()

		_, err := getLatestVersion()
		if err == nil {
			t.Error("getLatestVersion() expected error for empty tag name, got nil")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`invalid json`))
		}))
		defer server.Close()

		original := githubAPIURL
		githubAPIURL = server.URL
		defer func() { githubAPIURL = original }()

		_, err := getLatestVersion()
		if err == nil {
			t.Error("getLatestVersion() expected error for invalid JSON, got nil")
		}
	})
}

func TestDownloadFile(t *testing.T) {
	t.Run("successful download", func(t *testing.T) {
		expectedContent := []byte("test binary content")
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(expectedContent)
		}))
		defer server.Close()

		tmpPath, err := downloadFile(server.URL, "test-download-*")

		if err != nil {
			t.Fatalf("downloadFile() error = %v", err)
		}

		if tmpPath == "" {
			t.Fatal("downloadFile() returned empty path")
		}

		actualContent, err := os.ReadFile(tmpPath)
		if err != nil {
			t.Fatalf("Failed to read downloaded file: %v", err)
		}

		if !bytes.Equal(actualContent, expectedContent) {
			t.Errorf("downloadFile() content = %q, expected %q", actualContent, expectedContent)
		}

		_ = os.Remove(tmpPath)
	})

	t.Run("network error", func(t *testing.T) {
		invalidURL := "http://invalid-url-that-does-not-exist-12345.com"

		_, err := downloadFile(invalidURL, "test-download-*")

		if err == nil {
			t.Error("downloadFile() expected error for invalid URL, got nil")
		}
	})

	t.Run("non-200 status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		_, err := downloadFile(server.URL, "test-download-*")

		if err == nil {
			t.Error("downloadFile() expected error for 404, got nil")
		}
	})
}

func TestReadChecksumFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expected    string
		expectError bool
	}{
		{
			name:        "standard format with hash and filename",
			content:     "a1b2c3d4e5f6  binary.tar.gz",
			expected:    "a1b2c3d4e5f6",
			expectError: false,
		},
		{
			name:        "hash only",
			content:     "a1b2c3d4e5f6",
			expected:    "a1b2c3d4e5f6",
			expectError: false,
		},
		{
			name:        "full sha256 hash",
			content:     "abc123def4567890123456789012345678901234567890123456789012345678  dotenv-tui-linux-amd64",
			expected:    "abc123def4567890123456789012345678901234567890123456789012345678",
			expectError: false,
		},
		{
			name:        "empty file",
			content:     "",
			expectError: true,
		},
		{
			name:        "multiple spaces",
			content:     "abc123    filename.tar.gz",
			expected:    "abc123",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "checksum-*")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer func() { _ = os.Remove(tmpFile.Name()) }()

			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			if err := tmpFile.Close(); err != nil {
				t.Fatalf("Failed to close temp file: %v", err)
			}

			result, err := readChecksumFile(tmpFile.Name())

			if tt.expectError && err == nil {
				t.Error("readChecksumFile() expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("readChecksumFile() unexpected error = %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("readChecksumFile() = %q, expected %q", result, tt.expected)
			}
		})
	}

	t.Run("file not found", func(t *testing.T) {
		nonExistentPath := "/tmp/non-existent-checksum-file-12345"

		_, err := readChecksumFile(nonExistentPath)

		if err == nil {
			t.Error("readChecksumFile() expected error for non-existent file, got nil")
		}
	})
}

func TestCalculateFileSHA256(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name:        "empty file",
			content:     "",
			expectError: false,
		},
		{
			name:        "simple text",
			content:     "hello world",
			expectError: false,
		},
		{
			name:        "binary data",
			content:     "\x00\x01\x02\x03\xff\xfe\xfd",
			expectError: false,
		},
		{
			name:        "larger content",
			content:     strings.Repeat("a", 1000),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "sha256-*")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer func() { _ = os.Remove(tmpFile.Name()) }()

			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			if err := tmpFile.Close(); err != nil {
				t.Fatalf("Failed to close temp file: %v", err)
			}

			result, err := calculateFileSHA256(tmpFile.Name())

			if tt.expectError && err == nil {
				t.Error("calculateFileSHA256() expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("calculateFileSHA256() unexpected error = %v", err)
			}
			if !tt.expectError {
				if len(result) != 64 {
					t.Errorf("calculateFileSHA256() returned hash of length %d, expected 64", len(result))
				}
				for _, c := range result {
					if !isHexDigit(byte(c)) {
						t.Errorf("calculateFileSHA256() returned invalid hex: %q", result)
						break
					}
				}
			}
		})
	}

	t.Run("file not found", func(t *testing.T) {
		nonExistentPath := "/tmp/non-existent-file-12345"

		_, err := calculateFileSHA256(nonExistentPath)

		if err == nil {
			t.Error("calculateFileSHA256() expected error for non-existent file, got nil")
		}
	})
}

func isHexDigit(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func TestVerifyChecksum(t *testing.T) {
	t.Run("valid checksum", func(t *testing.T) {
		content := "test content for checksum"
		tmpFile, err := os.CreateTemp("", "binary-*")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer func() { _ = os.Remove(tmpFile.Name()) }()

		if _, err := tmpFile.WriteString(content); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		if err := tmpFile.Close(); err != nil {
			t.Fatalf("Failed to close temp file: %v", err)
		}

		actualHash, err := calculateFileSHA256(tmpFile.Name())
		if err != nil {
			t.Fatalf("Failed to calculate checksum: %v", err)
		}

		checksumFile, err := os.CreateTemp("", "checksum-*")
		if err != nil {
			t.Fatalf("Failed to create checksum temp file: %v", err)
		}
		defer func() { _ = os.Remove(checksumFile.Name()) }()

		if _, err := checksumFile.WriteString(actualHash + "  " + filepath.Base(tmpFile.Name())); err != nil {
			t.Fatalf("Failed to write checksum: %v", err)
		}
		if err := checksumFile.Close(); err != nil {
			t.Fatalf("Failed to close checksum file: %v", err)
		}

		err = verifyChecksum(tmpFile.Name(), checksumFile.Name())

		if err != nil {
			t.Errorf("verifyChecksum() unexpected error = %v", err)
		}
	})

	t.Run("checksum mismatch", func(t *testing.T) {
		content := "test content for checksum"
		tmpFile, err := os.CreateTemp("", "binary-*")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer func() { _ = os.Remove(tmpFile.Name()) }()

		if _, err := tmpFile.WriteString(content); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		if err := tmpFile.Close(); err != nil {
			t.Fatalf("Failed to close temp file: %v", err)
		}

		checksumFile, err := os.CreateTemp("", "checksum-*")
		if err != nil {
			t.Fatalf("Failed to create checksum temp file: %v", err)
		}
		defer func() { _ = os.Remove(checksumFile.Name()) }()

		wrongHash := strings.Repeat("0", 64)
		if _, err := checksumFile.WriteString(wrongHash + "  " + filepath.Base(tmpFile.Name())); err != nil {
			t.Fatalf("Failed to write checksum: %v", err)
		}
		if err := checksumFile.Close(); err != nil {
			t.Fatalf("Failed to close checksum file: %v", err)
		}

		err = verifyChecksum(tmpFile.Name(), checksumFile.Name())

		if err == nil {
			t.Error("verifyChecksum() expected error for mismatch, got nil")
		}
	})

	t.Run("missing binary file", func(t *testing.T) {
		checksumFile, err := os.CreateTemp("", "checksum-*")
		if err != nil {
			t.Fatalf("Failed to create checksum temp file: %v", err)
		}
		defer func() { _ = os.Remove(checksumFile.Name()) }()

		if _, err := checksumFile.WriteString(strings.Repeat("0", 64)); err != nil {
			t.Fatalf("Failed to write checksum: %v", err)
		}
		if err := checksumFile.Close(); err != nil {
			t.Fatalf("Failed to close checksum file: %v", err)
		}

		err = verifyChecksum("/tmp/non-existent-binary-12345", checksumFile.Name())

		if err == nil {
			t.Error("verifyChecksum() expected error for missing binary, got nil")
		}
	})

	t.Run("missing checksum file", func(t *testing.T) {
		content := "test content"
		tmpFile, err := os.CreateTemp("", "binary-*")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer func() { _ = os.Remove(tmpFile.Name()) }()

		if _, err := tmpFile.WriteString(content); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		if err := tmpFile.Close(); err != nil {
			t.Fatalf("Failed to close temp file: %v", err)
		}

		err = verifyChecksum(tmpFile.Name(), "/tmp/non-existent-checksum-12345")

		if err == nil {
			t.Error("verifyChecksum() expected error for missing checksum file, got nil")
		}
	})
}

func TestCopyFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		perm        os.FileMode
		expectError bool
	}{
		{
			name:        "simple text copy",
			content:     "hello world",
			perm:        0644,
			expectError: false,
		},
		{
			name:        "binary data copy",
			content:     "\x00\x01\x02\x03\xff\xfe\xfd",
			perm:        0755,
			expectError: false,
		},
		{
			name:        "empty file copy",
			content:     "",
			perm:        0644,
			expectError: false,
		},
		{
			name:        "larger file copy",
			content:     strings.Repeat("a", 10000),
			perm:        0644,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcFile, err := os.CreateTemp("", "src-copy-*")
			if err != nil {
				t.Fatalf("Failed to create src temp file: %v", err)
			}
			defer func() { _ = os.Remove(srcFile.Name()) }()

			if _, err := srcFile.WriteString(tt.content); err != nil {
				t.Fatalf("Failed to write to src file: %v", err)
			}
			if err := srcFile.Close(); err != nil {
				t.Fatalf("Failed to close src file: %v", err)
			}

			dstFile, err := os.CreateTemp("", "dst-copy-*")
			if err != nil {
				t.Fatalf("Failed to create dst temp file: %v", err)
			}
			dstPath := dstFile.Name()
			if err := dstFile.Close(); err != nil {
				t.Fatalf("Failed to close dst file: %v", err)
			}
			defer func() { _ = os.Remove(dstPath) }()

			err = copyFile(srcFile.Name(), dstPath)

			if tt.expectError && err == nil {
				t.Error("copyFile() expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("copyFile() unexpected error = %v", err)
			}
			if !tt.expectError {
				dstContent, err := os.ReadFile(dstPath)
				if err != nil {
					t.Fatalf("Failed to read dst file: %v", err)
				}

				if string(dstContent) != tt.content {
					t.Errorf("copyFile() dst content = %q, expected %q", string(dstContent), tt.content)
				}
			}
		})
	}

	t.Run("source file not found", func(t *testing.T) {
		dstFile, err := os.CreateTemp("", "dst-copy-*")
		if err != nil {
			t.Fatalf("Failed to create dst temp file: %v", err)
		}
		dstPath := dstFile.Name()
		if err := dstFile.Close(); err != nil {
			t.Fatalf("Failed to close dst file: %v", err)
		}
		defer func() { _ = os.Remove(dstPath) }()

		err = copyFile("/tmp/non-existent-src-12345", dstPath)

		if err == nil {
			t.Error("copyFile() expected error for non-existent src, got nil")
		}
	})

	t.Run("destination file not writable", func(t *testing.T) {
		srcFile, err := os.CreateTemp("", "src-copy-*")
		if err != nil {
			t.Fatalf("Failed to create src temp file: %v", err)
		}
		if _, err := srcFile.WriteString("test"); err != nil {
			t.Fatalf("Failed to write to src file: %v", err)
		}
		if err := srcFile.Close(); err != nil {
			t.Fatalf("Failed to close src file: %v", err)
		}
		defer func() { _ = os.Remove(srcFile.Name()) }()

		tmpDir, err := os.MkdirTemp("", "dst-dir-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		err = copyFile(srcFile.Name(), tmpDir)

		if err == nil {
			t.Error("copyFile() expected error when dst is a directory, got nil")
		}
	})
}

func TestReplaceBinary(t *testing.T) {
	t.Run("successful rename", func(t *testing.T) {
		srcFile, err := os.CreateTemp("", "src-replace-*")
		if err != nil {
			t.Fatalf("Failed to create src temp file: %v", err)
		}
		srcContent := "new binary content"
		if _, err := srcFile.WriteString(srcContent); err != nil {
			t.Fatalf("Failed to write to src file: %v", err)
		}
		if err := srcFile.Close(); err != nil {
			t.Fatalf("Failed to close src file: %v", err)
		}
		srcPath := srcFile.Name()

		dstFile, err := os.CreateTemp("", "dst-replace-*")
		if err != nil {
			t.Fatalf("Failed to create dst temp file: %v", err)
		}
		dstContent := "old binary content"
		if _, err := dstFile.WriteString(dstContent); err != nil {
			t.Fatalf("Failed to write to dst file: %v", err)
		}
		if err := dstFile.Close(); err != nil {
			t.Fatalf("Failed to close dst file: %v", err)
		}
		dstPath := dstFile.Name()

		err = replaceBinary(srcPath, dstPath)

		if err != nil {
			t.Errorf("replaceBinary() unexpected error = %v", err)
		}

		if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
			t.Error("replaceBinary() src file still exists after rename")
		}

		newContent, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("Failed to read dst file: %v", err)
		}

		if string(newContent) != srcContent {
			t.Errorf("replaceBinary() dst content = %q, expected %q", string(newContent), srcContent)
		}

		_ = os.Remove(dstPath)
	})

	t.Run("source file not found", func(t *testing.T) {
		dstFile, err := os.CreateTemp("", "dst-replace-*")
		if err != nil {
			t.Fatalf("Failed to create dst temp file: %v", err)
		}
		dstPath := dstFile.Name()
		if err := dstFile.Close(); err != nil {
			t.Fatalf("Failed to close dst file: %v", err)
		}
		defer func() { _ = os.Remove(dstPath) }()

		err = replaceBinary("/tmp/non-existent-src-12345", dstPath)

		if err == nil {
			t.Error("replaceBinary() expected error for non-existent src, got nil")
		}
	})
}

func TestDownloadBinaryAndChecksum(t *testing.T) {
	t.Run("successful download with checksum", func(t *testing.T) {
		binaryContent := []byte("binary content")
		checksumContent := []byte("abc123  dotenv-tui-linux-amd64")

		var binaryRequested, checksumRequested bool

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "binary") {
				binaryRequested = true
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(binaryContent)
			} else {
				checksumRequested = true
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(checksumContent)
			}
		}))
		defer server.Close()

		binaryURL := server.URL + "/binary"
		checksumURL := server.URL + "/checksum"

		binaryPath, checksumPath, err := downloadBinaryAndChecksum(binaryURL, checksumURL)

		if err != nil {
			t.Fatalf("downloadBinaryAndChecksum() error = %v", err)
		}

		if !binaryRequested {
			t.Error("downloadBinaryAndChecksum() binary was not requested")
		}
		if !checksumRequested {
			t.Error("downloadBinaryAndChecksum() checksum was not requested")
		}

		if binaryPath == "" {
			t.Fatal("downloadBinaryAndChecksum() binary path is empty")
		}
		defer func() { _ = os.Remove(binaryPath) }()

		actualContent, err := os.ReadFile(binaryPath)
		if err != nil {
			t.Fatalf("Failed to read binary file: %v", err)
		}

		if !bytes.Equal(actualContent, binaryContent) {
			t.Errorf("downloadBinaryAndChecksum() binary content mismatch")
		}

		if checksumPath == "" {
			t.Fatal("downloadBinaryAndChecksum() checksum path is empty")
		}
		defer func() { _ = os.Remove(checksumPath) }()

		checksumData, err := os.ReadFile(checksumPath)
		if err != nil {
			t.Fatalf("Failed to read checksum file: %v", err)
		}

		if !bytes.Equal(checksumData, checksumContent) {
			t.Errorf("downloadBinaryAndChecksum() checksum content mismatch")
		}

		info, err := os.Stat(binaryPath)
		if err != nil {
			t.Fatalf("Failed to stat binary: %v", err)
		}

		if info.Mode().Perm()&0111 == 0 {
			t.Error("downloadBinaryAndChecksum() binary is not executable")
		}
	})

	t.Run("binary download fails", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		_, _, err := downloadBinaryAndChecksum(server.URL+"/binary", server.URL+"/checksum")

		if err == nil {
			t.Error("downloadBinaryAndChecksum() expected error for failed binary download, got nil")
		}
	})

	t.Run("checksum not available (continues without it)", func(t *testing.T) {
		binaryContent := []byte("binary content")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "binary") {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(binaryContent)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		binaryURL := server.URL + "/binary"
		checksumURL := server.URL + "/checksum"

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		binaryPath, checksumPath, err := downloadBinaryAndChecksum(binaryURL, checksumURL)

		_ = w.Close()
		os.Stdout = old
		output, _ := io.ReadAll(r)

		if err != nil {
			t.Fatalf("downloadBinaryAndChecksum() unexpected error = %v", err)
		}

		if binaryPath == "" {
			t.Fatal("downloadBinaryAndChecksum() binary path is empty")
		}
		defer func() { _ = os.Remove(binaryPath) }()

		if checksumPath != "" {
			t.Error("downloadBinaryAndChecksum() checksum path should be empty when download fails")
			_ = os.Remove(checksumPath)
		}

		if !bytes.Contains(output, []byte("Warning")) {
			t.Error("downloadBinaryAndChecksum() expected warning message when checksum unavailable")
		}
	})
}

func TestRelease(t *testing.T) {
	t.Run("valid JSON unmarshal", func(t *testing.T) {
		jsonData := `{"tag_name": "v1.2.3", "name": "Release 1.2.3", "prerelease": false}`

		var release Release
		err := json.Unmarshal([]byte(jsonData), &release)

		if err != nil {
			t.Errorf("Release unmarshal error = %v", err)
		}

		if release.TagName != "v1.2.3" {
			t.Errorf("Release.TagName = %q, expected %q", release.TagName, "v1.2.3")
		}
	})

	t.Run("empty tag name", func(t *testing.T) {
		jsonData := `{"tag_name": "", "name": "Empty Release"}`

		var release Release
		err := json.Unmarshal([]byte(jsonData), &release)

		if err != nil {
			t.Errorf("Release unmarshal error = %v", err)
		}

		if release.TagName != "" {
			t.Errorf("Release.TagName = %q, expected empty string", release.TagName)
		}
	})
}
