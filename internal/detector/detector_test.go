package detector

import (
	"testing"
)

func TestIsSecret(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected bool
	}{
		// Secret key patterns
		{
			name:     "secret key with SECRET",
			key:      "DATABASE_SECRET",
			value:    "somevalue",
			expected: true,
		},
		{
			name:     "secret key with KEY",
			key:      "API_KEY",
			value:    "somevalue",
			expected: true,
		},
		{
			name:     "secret key with TOKEN",
			key:      "AUTH_TOKEN",
			value:    "somevalue",
			expected: true,
		},
		{
			name:     "secret key with PASSWORD",
			key:      "DB_PASSWORD",
			value:    "somevalue",
			expected: true,
		},
		{
			name:     "secret key with PASS",
			key:      "PASS",
			value:    "somevalue",
			expected: true,
		},
		{
			name:     "secret key with AUTH",
			key:      "BASIC_AUTH",
			value:    "somevalue",
			expected: true,
		},
		{
			name:     "secret key with CREDENTIAL",
			key:      "AWS_CREDENTIAL",
			value:    "somevalue",
			expected: true,
		},
		{
			name:     "secret key with PRIVATE",
			key:      "PRIVATE_KEY",
			value:    "somevalue",
			expected: true,
		},
		{
			name:     "secret key with API_KEY",
			key:      "STRIPE_API_KEY",
			value:    "somevalue",
			expected: true,
		},
		{
			name:     "secret key with ACCESS_KEY",
			key:      "AWS_ACCESS_KEY",
			value:    "somevalue",
			expected: true,
		},

		// Secret value patterns
		{
			name:     "base64 string longer than 20 chars",
			key:      "CONFIG",
			value:    "VGhpcyBpcyBhIGxvbmcgYmFzZTY0IGVuY29kZWQgc3RyaW5nIHRoYXQgaXMgZGVmaW5pdGVseSBsb25nZXIgdGhhbiAyMCBjaGFycyBhbmQgc2hvdWxkIGJlIGRldGVjdGVkIGFzIGEgc2VjcmV0",
			expected: true,
		},
		{
			name:     "JWT token",
			key:      "TOKEN",
			value:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expected: true,
		},
		{
			name:     "URL with user:pass@ pattern",
			key:      "DATABASE_URL",
			value:    "postgres://user:password@localhost:5432/db",
			expected: true,
		},
		{
			name:     "hex string longer than 32 chars",
			key:      "HASH",
			value:    "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
			expected: true,
		},

		// Common non-secrets
		{
			name:     "PORT is not a secret",
			key:      "PORT",
			value:    "3000",
			expected: false,
		},
		{
			name:     "HOST is not a secret",
			key:      "HOST",
			value:    "localhost",
			expected: false,
		},
		{
			name:     "NODE_ENV is not a secret",
			key:      "NODE_ENV",
			value:    "development",
			expected: false,
		},
		{
			name:     "APP_NAME is not a secret",
			key:      "APP_NAME",
			value:    "myapp",
			expected: false,
		},
		{
			name:     "DEBUG is not a secret",
			key:      "DEBUG",
			value:    "true",
			expected: false,
		},
		{
			name:     "LOG_LEVEL is not a secret",
			key:      "LOG_LEVEL",
			value:    "info",
			expected: false,
		},

		// Edge cases
		{
			name:     "empty value",
			key:      "SECRET_KEY",
			value:    "",
			expected: false,
		},
		{
			name:     "short base64 string",
			key:      "CONFIG",
			value:    "aGVsbG8=", // 8 chars, should not be detected as secret
			expected: false,
		},
		{
			name:     "short hex string",
			key:      "HASH",
			value:    "a1b2c3d4", // 8 chars, should not be detected as secret
			expected: false,
		},
		{
			name:     "non-secret key with secret-like value",
			key:      "CONFIG",
			value:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSecret(tt.key, tt.value)
			if result != tt.expected {
				t.Errorf("IsSecret(%q, %q) = %v; want %v", tt.key, tt.value, result, tt.expected)
			}
		})
	}
}

func TestGeneratePlaceholder(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{
			name:     "JWT token",
			key:      "TOKEN",
			value:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expected: "eyJ***",
		},
		{
			name:     "HTTP URL with credentials",
			key:      "DATABASE_URL",
			value:    "http://user:pass@localhost:5432/db",
			expected: "http://***",
		},
		{
			name:     "HTTPS URL with credentials",
			key:      "DATABASE_URL",
			value:    "https://user:pass@localhost:5432/db",
			expected: "https://***",
		},
		{
			name:     "Stripe live key",
			key:      "STRIPE_SECRET_KEY",
			value:    "sk_live_1234567890abcdef",
			expected: "sk_***",
		},
		{
			name:     "Stripe test key",
			key:      "STRIPE_SECRET_KEY",
			value:    "sk_test_1234567890abcdef",
			expected: "sk_***",
		},
		{
			name:     "Stripe restricted live key",
			key:      "STRIPE_RESTRICTED_KEY",
			value:    "rk_live_1234567890abcdef",
			expected: "rk_***",
		},
		{
			name:     "Stripe restricted test key",
			key:      "STRIPE_RESTRICTED_KEY",
			value:    "rk_test_1234567890abcdef",
			expected: "rk_***",
		},
		{
			name:     "GitHub personal access token",
			key:      "GITHUB_TOKEN",
			value:    "ghp_1234567890abcdef",
			expected: "ghp_***",
		},
		{
			name:     "GitHub OAuth token",
			key:      "GITHUB_TOKEN",
			value:    "gho_1234567890abcdef",
			expected: "gho_***",
		},
		{
			name:     "GitHub user token",
			key:      "GITHUB_TOKEN",
			value:    "ghu_1234567890abcdef",
			expected: "ghu_***",
		},
		{
			name:     "GitHub App installation token",
			key:      "GITHUB_TOKEN",
			value:    "ghs_1234567890abcdef",
			expected: "ghs_***",
		},
		{
			name:     "GitHub fine-grained PAT",
			key:      "GITHUB_TOKEN",
			value:    "github_pat_1234567890abcdef",
			expected: "github_pat_***",
		},
		{
			name:     "Stripe publishable key",
			key:      "STRIPE_PUBLISHABLE_KEY",
			value:    "pk_live_1234567890abcdef",
			expected: "pk_***",
		},
		{
			name:     "Slack bot token",
			key:      "SLACK_BOT_TOKEN",
			value:    "xoxb-1234567890-abcdef",
			expected: "xox***",
		},
		{
			name:     "Slack user token",
			key:      "SLACK_USER_TOKEN",
			value:    "xoxp-1234567890-abcdef",
			expected: "xox***",
		},
		{
			name:     "Slack app-level token",
			key:      "SLACK_APP_TOKEN",
			value:    "xoxa-1234567890-abcdef",
			expected: "xox***",
		},
		{
			name:     "Google OAuth token",
			key:      "GOOGLE_TOKEN",
			value:    "ya29.1234567890-abcdef",
			expected: "ya29.***",
		},
		{
			name:     "Stripe webhook secret",
			key:      "STRIPE_WEBHOOK_SECRET",
			value:    "whsec_1234567890abcdef",
			expected: "whsec_***",
		},
		{
			name:     "AWS access key",
			key:      "AWS_ACCESS_KEY",
			value:    "AKIAIOSFODNN7EXAMPLE",
			expected: "akia***",
		},
		{
			name:     "AWS IAM access key ID",
			key:      "AWS_ACCESS_KEY_ID",
			value:    "AKIAIEXAMPLE12345678",
			expected: "akia***",
		},
		{
			name:     "age encryption key",
			key:      "AGE_SECRET_KEY",
			value:    "age-secret-key-1234567890abcdef",
			expected: "age-***",
		},
		{
			name:     "SSH RSA key",
			key:      "SSH_KEY",
			value:    "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7...",
			expected: "ssh-rsa-***",
		},
		{
			name:     "SSH ED25519 key",
			key:      "SSH_KEY",
			value:    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIG...",
			expected: "ssh-ed25519-***",
		},
		{
			name:     "generic secret",
			key:      "SECRET",
			value:    "some-random-secret-value",
			expected: "***",
		},
		{
			name:     "empty value",
			key:      "SECRET",
			value:    "",
			expected: "***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GeneratePlaceholder(tt.key, tt.value)
			if result != tt.expected {
				t.Errorf("GeneratePlaceholder(%q, %q) = %v; want %v", tt.key, tt.value, result, tt.expected)
			}
		})
	}
}

func TestIsSecretKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"uppercase secret", "DATABASE_SECRET", true},
		{"lowercase secret", "database_secret", true},
		{"mixed case secret", "Database_Secret", true},
		{"contains secret", "MY_SECRET_KEY", true},
		{"contains key", "API_KEY", true},
		{"contains token", "AUTH_TOKEN", true},
		{"contains password", "USER_PASSWORD", true},
		{"contains pass", "DB_PASS", true},
		{"contains auth", "BASIC_AUTH", true},
		{"contains credential", "AWS_CREDENTIAL", true},
		{"contains private", "PRIVATE_KEY", true},
		{"contains api_key", "STRIPE_API_KEY", true},
		{"contains access_key", "AWS_ACCESS_KEY", true},
		{"non-secret", "PORT", false},
		{"non-secret", "HOST", false},
		{"non-secret", "DEBUG", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSecretKey(tt.key)
			if result != tt.expected {
				t.Errorf("isSecretKey(%q) = %v; want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestIsSecretValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"long base64", "VGhpcyBpcyBhIGxvbmcgYmFzZTY0IGVuY29kZWQgc3RyaW5nIHRoYXQgaXMgZGVmaW5pdGVseSBsb25nZXIgdGhhbiAyMCBjaGFycyBhbmQgc2hvdWxkIGJlIGRldGVjdGVkIGFzIGEgc2VjcmV0", true},
		{"short base64", "aGVsbG8=", false},
		{"JWT token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c", true},
		{"short JWT", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", false},
		{"URL with credentials", "postgres://user:password@localhost:5432/db", true},
		{"URL without credentials", "postgres://localhost:5432/db", false},
		{"long hex", "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2", true},
		{"short hex", "a1b2c3d4", false},
		{"empty string", "", false},
		{"regular text", "hello world", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSecretValue(tt.value)
			if result != tt.expected {
				t.Errorf("isSecretValue(%q) = %v; want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestIsCommonNonSecret(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"PORT", "PORT", true},
		{"HOST", "HOST", true},
		{"NODE_ENV", "NODE_ENV", true},
		{"APP_NAME", "APP_NAME", true},
		{"DEBUG", "DEBUG", true},
		{"LOG_LEVEL", "LOG_LEVEL", true},
		{"ENV", "ENV", true},
		{"ENVIRONMENT", "ENVIRONMENT", true},
		{"VERSION", "VERSION", true},
		{"LANG", "LANG", true},
		{"TIMEZONE", "TIMEZONE", true},
		{"REGION", "REGION", true},
		{"ENDPOINT", "ENDPOINT", true},
		{"URL", "URL", true},
		{"URI", "URI", true},
		{"DOMAIN", "DOMAIN", true},
		{"SERVER", "SERVER", true},
		{"CLUSTER", "CLUSTER", true},
		{"secret key", "SECRET_KEY", false},
		{"api key", "API_KEY", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCommonNonSecret(tt.key)
			if result != tt.expected {
				t.Errorf("isCommonNonSecret(%q) = %v; want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestIsBase64(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"valid base64", "aGVsbG8gd29ybGQ=", true},
		{"valid base64 with newlines", "aGVsbG8gd29ybGQ=\naGVsbG8gd29ybGQ=", false},
		{"valid base64 with spaces", "aGVsbG8gd29ybGQ= aGVsbG8gd29ybGQ=", false},
		{"invalid base64 - wrong length", "aGVsbG8gd29ybGQ", false},
		{"invalid base64 - invalid chars", "hello!world", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBase64(tt.value)
			if result != tt.expected {
				t.Errorf("isBase64(%q) = %v; want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestIsHex(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"lowercase hex", "a1b2c3d4", true},
		{"uppercase hex", "A1B2C3D4", true},
		{"mixed case hex", "a1B2c3D4", true},
		{"with numbers", "1234567890abcdef", true},
		{"invalid hex chars", "g1h2i3j4", false},
		{"empty string", "", false},
		{"contains non-hex", "a1b2c3d4x", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHex(tt.value)
			if result != tt.expected {
				t.Errorf("isHex(%q) = %v; want %v", tt.value, result, tt.expected)
			}
		})
	}
}
