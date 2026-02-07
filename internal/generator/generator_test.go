package generator

import (
	"testing"

	"github.com/jellydn/env-man/internal/detector"
	"github.com/jellydn/env-man/internal/parser"
)

func TestGenerateExample(t *testing.T) {
	tests := []struct {
		name     string
		entries  []parser.Entry
		expected []parser.Entry
	}{
		{
			name: "non-secret values preserved",
			entries: []parser.Entry{
				parser.KeyValue{Key: "PORT", Value: "3000"},
				parser.KeyValue{Key: "HOST", Value: "localhost"},
			},
			expected: []parser.Entry{
				parser.KeyValue{Key: "PORT", Value: "3000"},
				parser.KeyValue{Key: "HOST", Value: "localhost"},
			},
		},
		{
			name: "secret values masked",
			entries: []parser.Entry{
				parser.KeyValue{Key: "API_SECRET", Value: "sk_live_123456789"},
				parser.KeyValue{Key: "DB_PASSWORD", Value: "mypassword123"},
			},
			expected: []parser.Entry{
				parser.KeyValue{Key: "API_SECRET", Value: "sk_***"},
				parser.KeyValue{Key: "DB_PASSWORD", Value: "***"},
			},
		},
		{
			name: "mixed secrets and non-secrets",
			entries: []parser.Entry{
				parser.Comment{Text: "# Database configuration"},
				parser.KeyValue{Key: "DB_HOST", Value: "localhost"},
				parser.KeyValue{Key: "DB_PASSWORD", Value: "secret123"},
				parser.BlankLine{},
				parser.KeyValue{Key: "JWT_SECRET", Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"},
				parser.KeyValue{Key: "PORT", Value: "8080"},
			},
			expected: []parser.Entry{
				parser.Comment{Text: "# Database configuration"},
				parser.KeyValue{Key: "DB_HOST", Value: "localhost"},
				parser.KeyValue{Key: "DB_PASSWORD", Value: "***"},
				parser.BlankLine{},
				parser.KeyValue{Key: "JWT_SECRET", Value: "eyJ***"},
				parser.KeyValue{Key: "PORT", Value: "8080"},
			},
		},
		{
			name: "quoted secret values masked correctly",
			entries: []parser.Entry{
				parser.KeyValue{Key: "API_KEY", Value: "ghp_1234567890abcdef", Quoted: "\""},
				parser.KeyValue{Key: "WEBHOOK_SECRET", Value: "whsec_1234567890abcdef", Quoted: "'"},
			},
			expected: []parser.Entry{
				parser.KeyValue{Key: "API_KEY", Value: "ghp_***", Quoted: ""},
				parser.KeyValue{Key: "WEBHOOK_SECRET", Value: "***", Quoted: ""},
			},
		},
		{
			name: "exported secrets masked",
			entries: []parser.Entry{
				parser.KeyValue{Key: "SECRET_KEY", Value: "supersecret", Exported: true},
				parser.KeyValue{Key: "PUBLIC_KEY", Value: "publicvalue", Exported: true},
			},
			expected: []parser.Entry{
				parser.KeyValue{Key: "SECRET_KEY", Value: "***", Exported: true},
				parser.KeyValue{Key: "PUBLIC_KEY", Value: "***", Exported: true}, // PUBLIC_KEY contains "KEY" so it's a secret
			},
		},
		{
			name: "format-hint placeholders",
			entries: []parser.Entry{
				parser.KeyValue{Key: "STRIPE_KEY", Value: "sk_live_123456789"},
				parser.KeyValue{Key: "GITHUB_TOKEN", Value: "ghp_abcdef123456"},
				parser.KeyValue{Key: "DATABASE_URL", Value: "postgres://user:pass@localhost:5432/db"},
				parser.KeyValue{Key: "JWT_TOKEN", Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"},
			},
			expected: []parser.Entry{
				parser.KeyValue{Key: "STRIPE_KEY", Value: "sk_***"},
				parser.KeyValue{Key: "GITHUB_TOKEN", Value: "ghp_***"},
				parser.KeyValue{Key: "DATABASE_URL", Value: "***://***"}, // postgres:// is not http/https
				parser.KeyValue{Key: "JWT_TOKEN", Value: "eyJ***"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateExample(tt.entries)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d entries, got %d", len(tt.expected), len(result))
				return
			}

			for i, expectedEntry := range tt.expected {
				resultEntry := result[i]

				switch expected := expectedEntry.(type) {
				case parser.KeyValue:
					resultKV, ok := resultEntry.(parser.KeyValue)
					if !ok {
						t.Errorf("Expected KeyValue at index %d, got %T", i, resultEntry)
						continue
					}

					if resultKV.Key != expected.Key {
						t.Errorf("Expected key %q at index %d, got %q", expected.Key, i, resultKV.Key)
					}

					if resultKV.Value != expected.Value {
						t.Errorf("Expected value %q for key %q at index %d, got %q", expected.Value, expected.Key, i, resultKV.Value)
					}

					if resultKV.Quoted != expected.Quoted {
						t.Errorf("Expected quoted %q for key %q at index %d, got %q", expected.Quoted, expected.Key, i, resultKV.Quoted)
					}

					if resultKV.Exported != expected.Exported {
						t.Errorf("Expected exported %t for key %q at index %d, got %t", expected.Exported, expected.Key, i, resultKV.Exported)
					}

				case parser.Comment:
					resultComment, ok := resultEntry.(parser.Comment)
					if !ok {
						t.Errorf("Expected Comment at index %d, got %T", i, resultEntry)
						continue
					}

					if resultComment.Text != expected.Text {
						t.Errorf("Expected comment text %q at index %d, got %q", expected.Text, i, resultComment.Text)
					}

				case parser.BlankLine:
					if _, ok := resultEntry.(parser.BlankLine); !ok {
						t.Errorf("Expected BlankLine at index %d, got %T", i, resultEntry)
					}
				}
			}
		})
	}
}

// Integration test with detector package
func TestGenerateExampleIntegration(t *testing.T) {
	// Test that GenerateExample correctly uses detector.IsSecret and detector.GeneratePlaceholder
	entries := []parser.Entry{
		parser.KeyValue{Key: "PORT", Value: "3000"},                           // Non-secret
		parser.KeyValue{Key: "API_SECRET", Value: "sk_live_1234567890abcdef"}, // Secret
		parser.KeyValue{Key: "GITHUB_TOKEN", Value: "ghp_1234567890abcdef"},   // Secret
		parser.KeyValue{Key: "HOST", Value: "localhost"},                      // Non-secret
		parser.KeyValue{Key: "JWT_TOKEN", Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"}, // Secret
	}

	result := GenerateExample(entries)

	// Verify the detector functions are being used correctly
	testCases := []struct {
		key            string
		expectedValue  string
		expectedSecret bool
	}{
		{"PORT", "3000", false},
		{"API_SECRET", "sk_***", true},
		{"GITHUB_TOKEN", "ghp_***", true},
		{"HOST", "localhost", false},
		{"JWT_TOKEN", "eyJ***", true},
	}

	for i, tc := range testCases {
		entry := result[i]
		kv, ok := entry.(parser.KeyValue)
		if !ok {
			t.Errorf("Expected KeyValue for key %s, got %T", tc.key, entry)
			continue
		}

		if kv.Key != tc.key {
			t.Errorf("Expected key %s, got %s", tc.key, kv.Key)
		}

		if kv.Value != tc.expectedValue {
			t.Errorf("Expected value %s for key %s, got %s", tc.expectedValue, tc.key, kv.Value)
		}

		// Verify the detector functions give the same result
		originalEntry := entries[i]
		originalKV, ok := originalEntry.(parser.KeyValue)
		if !ok {
			continue
		}

		isSecret := detector.IsSecret(originalKV.Key, originalKV.Value)
		if isSecret != tc.expectedSecret {
			t.Errorf("Detector mismatch for key %s: expected secret %t, got %t", tc.key, tc.expectedSecret, isSecret)
		}

		if isSecret {
			expectedPlaceholder := detector.GeneratePlaceholder(originalKV.Key, originalKV.Value)
			if kv.Value != expectedPlaceholder {
				t.Errorf("Placeholder mismatch for key %s: expected %s, got %s", tc.key, expectedPlaceholder, kv.Value)
			}
		}
	}
}

// Benchmark tests
func BenchmarkGenerateExample(b *testing.B) {
	// Create a large sample of entries
	var entries []parser.Entry
	for i := 0; i < 1000; i++ {
		switch i % 3 {
		case 0:
			entries = append(entries, parser.KeyValue{Key: "VAR_" + string(rune(i)), Value: "value"})
		case 1:
			entries = append(entries, parser.KeyValue{Key: "SECRET_" + string(rune(i)), Value: "sk_live_123456789"})
		default:
			entries = append(entries, parser.BlankLine{})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateExample(entries)
	}
}
