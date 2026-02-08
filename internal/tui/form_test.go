package tui

import (
	"testing"
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
