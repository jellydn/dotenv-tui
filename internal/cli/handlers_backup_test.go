package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenerateExampleFileWithBackup(t *testing.T) {
	tests := []struct {
		name             string
		inputContent     string
		existingOutput   string
		createBackup     bool
		force            bool
		wantBackupCreated bool
	}{
		{
			name:              "creates backup when overwriting with backup enabled",
			inputContent:      "API_KEY=secret123\n",
			existingOutput:    "OLD_KEY=oldvalue\n",
			createBackup:      true,
			force:             true,
			wantBackupCreated: true,
		},
		{
			name:              "no backup when overwriting with backup disabled",
			inputContent:      "API_KEY=secret123\n",
			existingOutput:    "OLD_KEY=oldvalue\n",
			createBackup:      false,
			force:             true,
			wantBackupCreated: false,
		},
		{
			name:              "no backup when file doesn't exist",
			inputContent:      "API_KEY=secret123\n",
			createBackup:      true,
			force:             false,
			wantBackupCreated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newMockFileSystem()
			fs.files["/test/.env"] = tt.inputContent
			if tt.existingOutput != "" {
				fs.files["/test/.env.example"] = tt.existingOutput
			}

			var out bytes.Buffer
			err := GenerateExampleFile("/test/.env", tt.force, tt.createBackup, fs, &out)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check if backup was created
			backupExists := false
			for path := range fs.files {
				if strings.HasPrefix(path, "/test/.env.example.bak.") {
					backupExists = true
					// Verify backup content matches original
					if tt.existingOutput != "" {
						backupContent := fs.files[path]
						if backupContent != tt.existingOutput {
							t.Errorf("backup content = %q, want %q", backupContent, tt.existingOutput)
						}
					}
				}
			}

			if backupExists != tt.wantBackupCreated {
				t.Errorf("backup created = %v, want %v", backupExists, tt.wantBackupCreated)
			}

			// Verify backup message in output
			outputStr := out.String()
			if tt.wantBackupCreated {
				if !strings.Contains(outputStr, "Backup created:") {
					t.Errorf("expected backup message in output, got: %s", outputStr)
				}
			}
		})
	}
}

func TestGenerateEnvFileWithBackup(t *testing.T) {
	tests := []struct {
		name              string
		inputContent      string
		existingOutput    string
		createBackup      bool
		force             bool
		wantBackupCreated bool
	}{
		{
			name:              "creates backup when overwriting .env with backup enabled",
			inputContent:      "API_KEY=***\nPORT=3000\n",
			existingOutput:    "OLD_API_KEY=real_secret\nPORT=8080\n",
			createBackup:      true,
			force:             true,
			wantBackupCreated: true,
		},
		{
			name:              "no backup when overwriting .env with backup disabled",
			inputContent:      "API_KEY=***\n",
			existingOutput:    "OLD_API_KEY=real_secret\n",
			createBackup:      false,
			force:             true,
			wantBackupCreated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newMockFileSystem()
			fs.files["/test/.env.example"] = tt.inputContent
			if tt.existingOutput != "" {
				fs.files["/test/.env"] = tt.existingOutput
			}

			var out bytes.Buffer
			err := GenerateEnvFile("/test/.env.example", tt.force, tt.createBackup, fs, &out)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check if backup was created
			backupExists := false
			for path := range fs.files {
				if strings.HasPrefix(path, "/test/.env.bak.") {
					backupExists = true
					// Verify backup content matches original
					if tt.existingOutput != "" {
						backupContent := fs.files[path]
						if backupContent != tt.existingOutput {
							t.Errorf("backup content = %q, want %q", backupContent, tt.existingOutput)
						}
					}
				}
			}

			if backupExists != tt.wantBackupCreated {
				t.Errorf("backup created = %v, want %v", backupExists, tt.wantBackupCreated)
			}
		})
	}
}

func TestProcessExampleFileWithBackup(t *testing.T) {
	tests := []struct {
		name              string
		inputContent      string
		existingOutput    string
		createBackup      bool
		force             bool
		wantBackupCreated bool
	}{
		{
			name:              "creates backup when processing with backup enabled",
			inputContent:      "KEY=value\n",
			existingOutput:    "OLD_KEY=oldvalue\n",
			createBackup:      true,
			force:             true,
			wantBackupCreated: true,
		},
		{
			name:              "no backup when processing with backup disabled",
			inputContent:      "KEY=value\n",
			existingOutput:    "OLD_KEY=oldvalue\n",
			createBackup:      false,
			force:             true,
			wantBackupCreated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newMockFileSystem()
			fs.files["/test/.env.example"] = tt.inputContent
			if tt.existingOutput != "" {
				fs.files["/test/.env"] = tt.existingOutput
			}

			var out bytes.Buffer
			in := strings.NewReader("")
			generated, skipped := 0, 0

			err := ProcessExampleFile("/test/.env.example", tt.force, tt.createBackup, &generated, &skipped, fs, in, &out)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check if backup was created
			backupExists := false
			for path := range fs.files {
				if strings.HasPrefix(path, "/test/.env.bak.") {
					backupExists = true
					// Verify backup content
					if tt.existingOutput != "" {
						backupContent := fs.files[path]
						if backupContent != tt.existingOutput {
							t.Errorf("backup content = %q, want %q", backupContent, tt.existingOutput)
						}
					}
				}
			}

			if backupExists != tt.wantBackupCreated {
				t.Errorf("backup created = %v, want %v", backupExists, tt.wantBackupCreated)
			}

			// Verify backup message in output
			outputStr := out.String()
			if tt.wantBackupCreated {
				if !strings.Contains(outputStr, "Backup created:") {
					t.Errorf("expected backup message in output, got: %s", outputStr)
				}
			}
		})
	}
}
