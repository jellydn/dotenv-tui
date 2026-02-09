package parser

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// compareEntries compares two entry slices and reports differences.
func compareEntries(t *testing.T, got, want []Entry) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("got %d entries, expected %d", len(got), len(want))
	}

	for i := range got {
		gotEntry, wantEntry := got[i], want[i]

		switch g := gotEntry.(type) {
		case KeyValue:
			w, ok := wantEntry.(KeyValue)
			if !ok {
				t.Errorf("entry %d is KeyValue, expected %T", i, wantEntry)
				continue
			}
			if g.Key != w.Key || g.Value != w.Value || g.Quoted != w.Quoted || g.Exported != w.Exported {
				t.Errorf("entry %d: got %+v, want %+v", i, g, w)
			}

		case Comment:
			w, ok := wantEntry.(Comment)
			if !ok {
				t.Errorf("entry %d is Comment, expected %T", i, wantEntry)
				continue
			}
			if g.Text != w.Text {
				t.Errorf("entry %d: got %q, want %q", i, g.Text, w.Text)
			}

		case BlankLine:
			if _, ok := wantEntry.(BlankLine); !ok {
				t.Errorf("entry %d is BlankLine, expected %T", i, wantEntry)
			}
		}
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Entry
	}{
		{
			name:  "simple key-value",
			input: "KEY=value\n",
			expected: []Entry{
				KeyValue{Key: "KEY", Value: "value", Quoted: "", Exported: false},
			},
		},
		{
			name: "quoted values",
			input: `KEY1="value with spaces"
KEY2='single quoted'
KEY3=unquoted
`,
			expected: []Entry{
				KeyValue{Key: "KEY1", Value: "value with spaces", Quoted: "\"", Exported: false},
				KeyValue{Key: "KEY2", Value: "single quoted", Quoted: "'", Exported: false},
				KeyValue{Key: "KEY3", Value: "unquoted", Quoted: "", Exported: false},
			},
		},
		{
			name:  "export prefix",
			input: "export KEY=value\n",
			expected: []Entry{
				KeyValue{Key: "KEY", Value: "value", Quoted: "", Exported: true},
			},
		},
		{
			name: "comments and blank lines",
			input: `# This is a comment

KEY=value
# Another comment
`,
			expected: []Entry{
				Comment{Text: "# This is a comment"},
				BlankLine{},
				KeyValue{Key: "KEY", Value: "value", Quoted: "", Exported: false},
				Comment{Text: "# Another comment"},
			},
		},
		{
			name: "complex example",
			input: `# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER="admin"
DB_PASSWORD='secret'

# API settings
export API_KEY=sk_live_12345
API_URL=https://api.example.com
`,
			expected: []Entry{
				Comment{Text: "# Database configuration"},
				KeyValue{Key: "DB_HOST", Value: "localhost", Quoted: "", Exported: false},
				KeyValue{Key: "DB_PORT", Value: "5432", Quoted: "", Exported: false},
				KeyValue{Key: "DB_USER", Value: "admin", Quoted: "\"", Exported: false},
				KeyValue{Key: "DB_PASSWORD", Value: "secret", Quoted: "'", Exported: false},
				BlankLine{},
				Comment{Text: "# API settings"},
				KeyValue{Key: "API_KEY", Value: "sk_live_12345", Quoted: "", Exported: true},
				KeyValue{Key: "API_URL", Value: "https://api.example.com", Quoted: "", Exported: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			entries, err := Parse(reader)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			compareEntries(t, entries, tt.expected)
		})
	}
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name     string
		entries  []Entry
		expected string
	}{
		{
			name: "simple key-value",
			entries: []Entry{
				KeyValue{Key: "KEY", Value: "value", Quoted: "", Exported: false},
			},
			expected: "KEY=value\n",
		},
		{
			name: "quoted values",
			entries: []Entry{
				KeyValue{Key: "KEY1", Value: "value with spaces", Quoted: "\"", Exported: false},
				KeyValue{Key: "KEY2", Value: "single quoted", Quoted: "'", Exported: false},
			},
			expected: `KEY1="value with spaces"
KEY2='single quoted'
`,
		},
		{
			name: "export prefix",
			entries: []Entry{
				KeyValue{Key: "KEY", Value: "value", Quoted: "", Exported: true},
			},
			expected: "export KEY=value\n",
		},
		{
			name: "comments and blank lines",
			entries: []Entry{
				Comment{Text: "# This is a comment"},
				BlankLine{},
				KeyValue{Key: "KEY", Value: "value", Quoted: "", Exported: false},
			},
			expected: `# This is a comment

KEY=value
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var builder strings.Builder
			err := Write(&builder, tt.entries)
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			if builder.String() != tt.expected {
				t.Errorf("Write() = %q, expected %q", builder.String(), tt.expected)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	input := `# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER="admin"
DB_PASSWORD='secret'

# API settings
export API_KEY=sk_live_12345
API_URL=https://api.example.com
`

	reader := strings.NewReader(input)
	entries, err := Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	var builder strings.Builder
	err = Write(&builder, entries)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if builder.String() != input {
		t.Errorf("Round trip failed:\nGot:\n%s\nExpected:\n%s", builder.String(), input)
	}
}

func TestParseMultiline(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Entry
	}{
		{
			name: "double-quoted multiline value",
			input: `KEY="line1
line2
line3"
`,
			expected: []Entry{
				KeyValue{Key: "KEY", Value: "line1\nline2\nline3", Quoted: "\"", Exported: false},
			},
		},
		{
			name: "single-quoted multiline value",
			input: `KEY='line1
line2'
`,
			expected: []Entry{
				KeyValue{Key: "KEY", Value: "line1\nline2", Quoted: "'", Exported: false},
			},
		},
		{
			name: "multiline with escaped newlines",
			input: `KEY="line1\nline2\nline3"
`,
			expected: []Entry{
				KeyValue{Key: "KEY", Value: `line1\nline2\nline3`, Quoted: "\"", Exported: false},
			},
		},
		{
			name: "multiline value with equals sign",
			input: `KEY="line1
line2=value
line3"
`,
			expected: []Entry{
				KeyValue{Key: "KEY", Value: "line1\nline2=value\nline3", Quoted: "\"", Exported: false},
			},
		},
		{
			name: "multiline with export prefix",
			input: `export KEY="line1
line2"
`,
			expected: []Entry{
				KeyValue{Key: "KEY", Value: "line1\nline2", Quoted: "\"", Exported: true},
			},
		},
		{
			name: "mixed single and multiline",
			input: `KEY1=simple
KEY2="multiline
value"
KEY3=another
`,
			expected: []Entry{
				KeyValue{Key: "KEY1", Value: "simple", Quoted: "", Exported: false},
				KeyValue{Key: "KEY2", Value: "multiline\nvalue", Quoted: "\"", Exported: false},
				KeyValue{Key: "KEY3", Value: "another", Quoted: "", Exported: false},
			},
		},
		{
			name: "multiline with comment after",
			input: `KEY="multiline
value"
# Comment after
`,
			expected: []Entry{
				KeyValue{Key: "KEY", Value: "multiline\nvalue", Quoted: "\"", Exported: false},
				Comment{Text: "# Comment after"},
			},
		},
		{
			name: "multiline value with blank lines inside",
			input: `KEY="line1

line3"
`,
			expected: []Entry{
				KeyValue{Key: "KEY", Value: "line1\n\nline3", Quoted: "\"", Exported: false},
			},
		},
		{
			name: "multiline with escaped quotes",
			input: `KEY="line1
line2 has a \"quoted\" word
line3"
`,
			expected: []Entry{
				KeyValue{Key: "KEY", Value: "line1\nline2 has a \\\"quoted\\\" word\nline3", Quoted: "\"", Exported: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			entries, err := Parse(reader)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			compareEntries(t, entries, tt.expected)
		})
	}
}

func TestMultilineRoundTrip(t *testing.T) {
	input := `KEY1="line1
line2
line3"
KEY2=simple
KEY3='multiline
with single quotes'
`

	reader := strings.NewReader(input)
	entries, err := Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	var builder strings.Builder
	err = Write(&builder, entries)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if builder.String() != input {
		t.Errorf("Round trip failed:\nGot:\n%s\nExpected:\n%s", builder.String(), input)
	}
}

func TestParseMultilineTestdata(t *testing.T) {
	// Get the testdata directory relative to the test file
	_, filename, _, _ := runtime.Caller(0)
	testdataPath := filepath.Join(filepath.Dir(filename), "..", "..", "testdata", ".env.multiline")

	file, err := os.Open(testdataPath)
	if err != nil {
		t.Fatalf("Failed to open testdata file: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := file.Close(); closeErr != nil {
			t.Fatalf("Close() error = %v", closeErr)
		}
	})

	entries, err := Parse(file)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Count entries by type
	var kvCount, commentCount, blankCount int
	for _, entry := range entries {
		switch entry.(type) {
		case KeyValue:
			kvCount++
		case Comment:
			commentCount++
		case BlankLine:
			blankCount++
		}
	}

	t.Logf("Parsed .env.multiline: %d key-values, %d comments, %d blank lines",
		kvCount, commentCount, blankCount)

	// Verify we have expected counts (at least)
	if kvCount < 9 {
		t.Errorf("Expected at least 9 key-value entries, got %d", kvCount)
	}

	// Test round-trip
	var builder strings.Builder
	err = Write(&builder, entries)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Parse the written output again
	reEntries, err := Parse(strings.NewReader(builder.String()))
	if err != nil {
		t.Fatalf("Parse() of written output error = %v", err)
	}

	if len(reEntries) != len(entries) {
		t.Errorf("Round-trip changed number of entries: %d -> %d", len(entries), len(reEntries))
	}
}
