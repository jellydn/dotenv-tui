// Package parser provides functions to parse and write .env files.
package parser

import (
"bufio"
"fmt"
"io"
"strings"
)

// Entry represents a line in a .env file
type Entry interface{}

// KeyValue represents a KEY=VALUE entry
type KeyValue struct {
Key      string
Value    string
Quoted   string // "", "\"", or "'"
Exported bool   // true if prefixed with 'export'
}

// Comment represents a comment line
type Comment struct {
Text string
}

// BlankLine represents an empty line
type BlankLine struct{}

// Parse reads a .env file and returns ordered entries
func Parse(reader io.Reader) ([]Entry, error) {
var entries []Entry
scanner := bufio.NewScanner(reader)
// Increase buffer to 1MB to handle large values
scanner.Buffer(make([]byte, 1024), 1024*1024)

var accumulated string
var inQuote rune // 0 if not in quote, '"' or '\'' if inside quote

for scanner.Scan() {
line := scanner.Text()

// If we're accumulating a multiline value
if inQuote != 0 {
accumulated += "\n" + line
// Check if this line closes the quote
if strings.ContainsRune(line, inQuote) {
// Find the closing quote (it should be the last occurrence)
idx := strings.LastIndexByte(line, byte(inQuote))
if idx != -1 {
// Quote is closed, process the accumulated value
inQuote = 0
kv, err := parseKeyValue(accumulated)
if err != nil {
return nil, fmt.Errorf("error parsing multiline value %q: %w", accumulated, err)
}
entries = append(entries, kv)
accumulated = ""
}
}
continue
}

// Not in a multiline value, process line normally
line = strings.TrimRight(line, " \t\r\n")

if line == "" {
entries = append(entries, BlankLine{})
continue
}

if strings.HasPrefix(line, "#") {
entries = append(entries, Comment{Text: line})
continue
}

if strings.Contains(line, "=") {
// Check if this line starts a multiline quoted value
quoteStart := findUnclosedQuote(line)
if quoteStart != 0 {
// Start accumulating multiline value
inQuote = quoteStart
accumulated = line
continue
}

// Single-line key-value
kv, err := parseKeyValue(line)
if err != nil {
return nil, fmt.Errorf("error parsing line %q: %w", line, err)
}
entries = append(entries, kv)
continue
}

entries = append(entries, Comment{Text: line})
}

if err := scanner.Err(); err != nil {
return nil, fmt.Errorf("error reading: %w", err)
}

// Check if we ended with an unclosed quote
if inQuote != 0 {
return nil, fmt.Errorf("unclosed quote in multiline value")
}

return entries, nil
}

// countUnescapedQuotes counts the number of unescaped quote characters in a string
func countUnescapedQuotes(s string, quote rune) int {
count := 0
escaped := false
for _, ch := range s {
if ch == '\\' && !escaped {
escaped = true
continue
}
if ch == quote && !escaped {
count++
}
escaped = false
}
return count
}

// findUnclosedQuote checks if a line contains an opening quote without a matching closing quote
// Returns the quote character ('"' or '\'') if unclosed, or 0 if closed or no quotes
func findUnclosedQuote(line string) rune {
// Find the equals sign first
eqIdx := strings.IndexByte(line, '=')
if eqIdx == -1 {
return 0
}

// Check the value part (after =)
valuePart := line[eqIdx+1:]

// Check for double quote
if strings.HasPrefix(valuePart, "\"") {
count := countUnescapedQuotes(valuePart, '"')
// If odd number of quotes, it's unclosed
if count%2 == 1 {
return '"'
}
}

// Check for single quote
if strings.HasPrefix(valuePart, "'") {
count := countUnescapedQuotes(valuePart, '\'')
// If odd number of quotes, it's unclosed
if count%2 == 1 {
return '\''
}
}

return 0
}

// parseKeyValue parses a single key-value line
func parseKeyValue(line string) (KeyValue, error) {
var kv KeyValue

if strings.HasPrefix(line, "export ") {
kv.Exported = true
line = strings.TrimSpace(line[7:])
}

parts := strings.SplitN(line, "=", 2)
if len(parts) != 2 {
return KeyValue{}, fmt.Errorf("invalid key-value format")
}

kv.Key = strings.TrimSpace(parts[0])
value := parts[1]

if len(value) >= 2 {
if (value[0] == '"' && value[len(value)-1] == '"') ||
(value[0] == '\'' && value[len(value)-1] == '\'') {
kv.Quoted = string(value[0])
kv.Value = value[1 : len(value)-1]
} else {
kv.Value = value
}
} else {
kv.Value = value
}

return kv, nil
}

// formatKeyValue converts a KeyValue entry to its string representation.
func formatKeyValue(kv KeyValue) string {
var line string
if kv.Exported {
line = "export "
}
line += kv.Key + "="
if kv.Quoted != "" {
line += kv.Quoted + kv.Value + kv.Quoted
} else {
line += kv.Value
}
return line
}

// Write writes entries to a writer, preserving the original structure
func Write(writer io.Writer, entries []Entry) error {
for _, entry := range entries {
switch e := entry.(type) {
case KeyValue:
if _, err := fmt.Fprintln(writer, formatKeyValue(e)); err != nil {
return err
}

case Comment:
if _, err := fmt.Fprintln(writer, e.Text); err != nil {
return err
}

case BlankLine:
if _, err := fmt.Fprintln(writer); err != nil {
return err
}

default:
return fmt.Errorf("unknown entry type: %T", e)
}
}
return nil
}

// EntryToString converts an Entry to its string representation.
func EntryToString(entry Entry) string {
switch e := entry.(type) {
case KeyValue:
return formatKeyValue(e)
case Comment:
return e.Text
case BlankLine:
return ""
default:
return ""
}
}
