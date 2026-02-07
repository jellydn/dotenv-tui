package parser

import (
	"strings"
	"testing"
)

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

			if len(entries) != len(tt.expected) {
				t.Fatalf("Parse() returned %d entries, expected %d", len(entries), len(tt.expected))
			}

			for i, entry := range entries {
				expected := tt.expected[i]

				switch e := entry.(type) {
				case KeyValue:
					exp, ok := expected.(KeyValue)
					if !ok {
						t.Errorf("Entry %d is KeyValue, expected %T", i, expected)
						continue
					}
					if e.Key != exp.Key || e.Value != exp.Value || e.Quoted != exp.Quoted || e.Exported != exp.Exported {
						t.Errorf("KeyValue %d = %+v, expected %+v", i, e, exp)
					}

				case Comment:
					exp, ok := expected.(Comment)
					if !ok {
						t.Errorf("Entry %d is Comment, expected %T", i, expected)
						continue
					}
					if e.Text != exp.Text {
						t.Errorf("Comment %d = %v, expected %v", i, e.Text, exp.Text)
					}

				case BlankLine:
					_, ok := expected.(BlankLine)
					if !ok {
						t.Errorf("Entry %d is BlankLine, expected %T", i, expected)
					}
				}
			}
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
