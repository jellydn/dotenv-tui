package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jellydn/dotenv-tui/internal/parser"
)

func TestIsPlaceholderValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{
			name:     "value with your_ prefix is placeholder",
			value:    "your_api_key_here",
			expected: true,
		},
		{
			name:     "value with _here suffix is placeholder",
			value:    "secret_here",
			expected: true,
		},
		{
			name:     "value containing placeholder word",
			value:    "this_is_a_placeholder_value",
			expected: true,
		},
		{
			name:     "value containing example word",
			value:    "example_database_url",
			expected: true,
		},
		{
			name:     "value is asterisks",
			value:    "***",
			expected: true,
		},
		{
			name:     "value is sk_ prefix with *** suffix",
			value:    "sk_test_***",
			expected: true,
		},
		{
			name:     "value is ghp_ prefix with *** suffix",
			value:    "ghp_12345***",
			expected: true,
		},
		{
			name:     "value is JWT prefix with *** suffix",
			value:    "eyJhbGciOi***",
			expected: true,
		},
		{
			name:     "normal value is not placeholder",
			value:    "actual_value_123",
			expected: false,
		},
		{
			name:     "URL is not placeholder",
			value:    "https://api.example.com",
			expected: false,
		},
		{
			name:     "empty string is not placeholder",
			value:    "",
			expected: false,
		},
		{
			name:     "database connection string is not placeholder",
			value:    "postgresql://localhost:5432/db",
			expected: false,
		},
		{
			name:     "port number is not placeholder",
			value:    "5432",
			expected: false,
		},
		{
			name:     "localhost is not placeholder",
			value:    "localhost",
			expected: false,
		},
		{
			name:     "case insensitive - YOUR_KEY_HERE",
			value:    "YOUR_KEY_HERE",
			expected: true,
		},
		{
			name:     "case insensitive - YOUR_API_KEY",
			value:    "YOUR_API_KEY",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: Value is provided in test case

			// Act
			result := isPlaceholderValue(tt.value)

			// Assert
			if result != tt.expected {
				t.Errorf("isPlaceholderValue(%q) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestGenerateHint(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{
			name:     "API key returns API key hint",
			key:      "API_KEY",
			value:    "sk_***",
			expected: "Enter your API key",
		},
		{
			name:     "secret returns secret hint",
			key:      "DB_SECRET",
			value:    "***",
			expected: "Enter your secret",
		},
		{
			name:     "token returns token hint",
			key:      "AUTH_TOKEN",
			value:    "your_token_here",
			expected: "Enter your token",
		},
		{
			name:     "password returns password hint",
			key:      "DB_PASSWORD",
			value:    "***",
			expected: "Enter your password",
		},
		{
			name:     "pass shorthand returns password hint",
			key:      "DB_PASS",
			value:    "***",
			expected: "Enter your password",
		},
		{
			name:     "url returns URL hint",
			key:      "DATABASE_URL",
			value:    "your_url_here",
			expected: "Enter URL (e.g., https://example.com)",
		},
		{
			name:     "uri returns URL hint",
			key:      "REDIS_URI",
			value:    "placeholder",
			expected: "Enter URL (e.g., https://example.com)",
		},
		{
			name:     "port returns port hint",
			key:      "DB_PORT",
			value:    "example",
			expected: "Enter port number (e.g., 3000)",
		},
		{
			name:     "host returns host hint",
			key:      "DB_HOST",
			value:    "localhost",
			expected: "Enter host (e.g., localhost)",
		},
		{
			name:     "database keyword returns database hint",
			key:      "DATABASE",
			value:    "your_database_here",
			expected: "Enter database connection string",
		},
		{
			name:     "db prefix returns database hint",
			key:      "DB_NAME",
			value:    "***",
			expected: "Enter database connection string",
		},
		{
			name:     "unknown key returns generic hint with key name",
			key:      "UNKNOWN_CONFIG",
			value:    "placeholder",
			expected: "Enter value for UNKNOWN_CONFIG",
		},
		{
			name:     "case insensitive - api in ApiKey",
			key:      "ApiKey",
			value:    "***",
			expected: "Enter your API key",
		},
		{
			name:     "case insensitive - password in PASSWORD",
			key:      "PASSWORD",
			value:    "***",
			expected: "Enter your password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: Key and value provided in test case

			// Act
			result := generateHint(tt.key, tt.value)

			// Assert
			if result != tt.expected {
				t.Errorf("generateHint(%q, %q) = %q, expected %q", tt.key, tt.value, result, tt.expected)
			}
		})
	}
}

func TestFormModelInit(t *testing.T) {
	// Arrange
	model := FormModel{fields: []FormField{}, cursor: 0}

	// Act
	cmd := model.Init()

	// Assert
	if cmd != nil {
		t.Errorf("FormModel.Init() should return nil, got %v", cmd)
	}
}

func TestFormModelInitEmptyFields(t *testing.T) {
	// Arrange
	model := FormModel{
		fields:    []FormField{},
		cursor:    0,
		filePath:  "/test/.env.example",
		confirmed: false,
		errorMsg:  "",
	}

	// Act
	cmd := model.Init()

	// Assert
	if cmd != nil {
		t.Errorf("FormModel.Init() with empty fields should return nil, got %v", cmd)
	}
}

func TestFormModelUpdateWithFormSavedMsg(t *testing.T) {
	tests := []struct {
		name            string
		confirmedBefore bool
		errorMsgBefore  string
		success         bool
		error           string
		wantConfirmed   bool
		wantErrorMsg    string
	}{
		{
			name:            "successful save marks form as confirmed with no error",
			confirmedBefore: false,
			errorMsgBefore:  "",
			success:         true,
			error:           "",
			wantConfirmed:   true,
			wantErrorMsg:    "",
		},
		{
			name:            "failed save marks form as confirmed with error message",
			confirmedBefore: false,
			errorMsgBefore:  "",
			success:         false,
			error:           "permission denied",
			wantConfirmed:   true,
			wantErrorMsg:    "permission denied",
		},
		{
			name:            "error is cleared when subsequent save succeeds",
			confirmedBefore: true,
			errorMsgBefore:  "previous error",
			success:         true,
			error:           "",
			wantConfirmed:   true,
			wantErrorMsg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			model := FormModel{
				fields:    []FormField{{Key: "TEST"}},
				cursor:    0,
				confirmed: tt.confirmedBefore,
				errorMsg:  tt.errorMsgBefore,
			}

			// Act
			msg := FormSavedMsg{Success: tt.success, Error: tt.error}
			newModel, cmd := model.Update(msg)

			// Assert
			newFormModel, ok := newModel.(FormModel)
			if !ok {
				t.Fatalf("Update() did not return FormModel")
			}

			if newFormModel.confirmed != tt.wantConfirmed {
				t.Errorf("Update(FormSavedMsg) confirmed = %v, expected %v", newFormModel.confirmed, tt.wantConfirmed)
			}

			if newFormModel.errorMsg != tt.wantErrorMsg {
				t.Errorf("Update(FormSavedMsg) errorMsg = %q, expected %q", newFormModel.errorMsg, tt.wantErrorMsg)
			}

			if cmd != nil {
				t.Errorf("Update(FormSavedMsg) should not return command, got %v", cmd)
			}
		})
	}
}

func TestFormModelNavigationAfterConfirmation(t *testing.T) {
	tests := []struct {
		name       string
		totalFiles int
		keyMsg     string
		wantDir    int
		wantCmd    bool
	}{
		{
			name:       "tab key moves to next file when multiple files exist",
			totalFiles: 3,
			keyMsg:     "tab",
			wantDir:    1,
			wantCmd:    true,
		},
		{
			name:       "tab key does nothing when only one file",
			totalFiles: 1,
			keyMsg:     "tab",
			wantDir:    0,
			wantCmd:    false,
		},
		{
			name:       "enter key finishes with dir 0",
			totalFiles: 3,
			keyMsg:     "enter",
			wantDir:    0,
			wantCmd:    true,
		},
		{
			name:       "q key finishes with dir 0",
			totalFiles: 3,
			keyMsg:     "q",
			wantDir:    0,
			wantCmd:    true,
		},
		{
			name:       "esc key finishes with dir 0",
			totalFiles: 3,
			keyMsg:     "esc",
			wantDir:    0,
			wantCmd:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			model := FormModel{
				fields:     []FormField{{Key: "TEST"}},
				cursor:     0,
				confirmed:  true,
				totalFiles: tt.totalFiles,
				savedFiles: make(map[int]bool),
			}

			// Act
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.keyMsg)}
			switch tt.keyMsg {
			case "tab":
				msg.Type = tea.KeyTab
			case "enter":
				msg.Type = tea.KeyEnter
			case "esc":
				msg.Type = tea.KeyEsc
			}

			newModel, cmd := model.Update(msg)

			// Assert
			_, ok := newModel.(FormModel)
			if !ok {
				t.Fatalf("Update() did not return FormModel")
			}

			if tt.wantCmd && cmd == nil {
				t.Errorf("Update(%q) should return command, got nil", tt.keyMsg)
			}

			if !tt.wantCmd && cmd != nil {
				t.Errorf("Update(%q) should not return command, got %v", tt.keyMsg, cmd)
			}

			if tt.wantCmd {
				// Execute the command to get the message
				msg := cmd()
				finishedMsg, ok := msg.(FormFinishedMsg)
				if !ok {
					t.Fatalf("Command did not return FormFinishedMsg")
				}

				if finishedMsg.Dir != tt.wantDir {
					t.Errorf("Update(%q) FormFinishedMsg.Dir = %d, expected %d", tt.keyMsg, finishedMsg.Dir, tt.wantDir)
				}
			}
		})
	}
}

func TestFormModelViewInitProgressDisplay(t *testing.T) {
	tests := []struct {
		name            string
		fileIndex       int
		totalFiles      int
		savedFiles      map[int]bool
		confirmed       bool
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:       "single file shows no saved count",
			fileIndex:  0,
			totalFiles: 1,
			savedFiles: map[int]bool{},
			confirmed:  false,
			wantContains: []string{
				"[1/1]",
				"(0/1 saved)",
			},
			wantNotContains: []string{},
		},
		{
			name:       "multiple files shows saved progress",
			fileIndex:  1,
			totalFiles: 3,
			savedFiles: map[int]bool{0: true},
			confirmed:  false,
			wantContains: []string{
				"[2/3]",
				"(1/3 saved)",
			},
			wantNotContains: []string{},
		},
		{
			name:       "all files saved shows completion",
			fileIndex:  2,
			totalFiles: 3,
			savedFiles: map[int]bool{0: true, 1: true, 2: true},
			confirmed:  true,
			wantContains: []string{
				"All files saved!",
			},
			wantNotContains: []string{},
		},
		{
			name:       "partial save shows remaining count",
			fileIndex:  0,
			totalFiles: 5,
			savedFiles: map[int]bool{0: true, 2: true},
			confirmed:  true,
			wantContains: []string{
				"(3 remaining)",
			},
			wantNotContains: []string{
				"All files saved!",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			model := FormModel{
				fields:     []FormField{{Key: "TEST", Input: textinput.New()}},
				cursor:     0,
				filePath:   "/test/.env.example",
				fileIndex:  tt.fileIndex,
				totalFiles: tt.totalFiles,
				savedFiles: tt.savedFiles,
				confirmed:  tt.confirmed,
			}
			model.fields[0].Input.Focus()

			// Act
			view := model.View()

			// Assert
			for _, substr := range tt.wantContains {
				if !contains(view, substr) {
					t.Errorf("View() should contain %q, got:\n%s", substr, view)
				}
			}

			for _, substr := range tt.wantNotContains {
				if contains(view, substr) {
					t.Errorf("View() should not contain %q, got:\n%s", substr, view)
				}
			}
		})
	}
}

func TestFormModelInitWithSavedFiles(t *testing.T) {
	// Arrange
	input := textinput.New()
	msg := formInitMsg{
		fields:          []FormField{{Key: "TEST", Input: input}},
		originalEntries: []parser.Entry{},
		filePath:        "/test/.env.example",
		fileIndex:       1,
		totalFiles:      3,
		savedFiles:      map[int]bool{0: true, 2: true},
	}
	model := FormModel{}

	// Act
	newModel, cmd := model.Update(msg)

	// Assert
	newFormModel, ok := newModel.(FormModel)
	if !ok {
		t.Fatalf("Update() did not return FormModel")
	}

	if newFormModel.fileIndex != 1 {
		t.Errorf("Update(formInitMsg) fileIndex = %d, expected 1", newFormModel.fileIndex)
	}

	if newFormModel.totalFiles != 3 {
		t.Errorf("Update(formInitMsg) totalFiles = %d, expected 3", newFormModel.totalFiles)
	}

	if len(newFormModel.savedFiles) != 2 {
		t.Errorf("Update(formInitMsg) savedFiles length = %d, expected 2", len(newFormModel.savedFiles))
	}

	if !newFormModel.savedFiles[0] || !newFormModel.savedFiles[2] {
		t.Errorf("Update(formInitMsg) savedFiles should contain 0 and 2, got %v", newFormModel.savedFiles)
	}

	if cmd != nil {
		t.Errorf("Update(formInitMsg) should not return command, got %v", cmd)
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
