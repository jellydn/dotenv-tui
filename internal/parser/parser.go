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

const (
	initialBufferSize = 1024
	maxBufferSize     = 1024 * 1024 // 1MB to handle large multiline values
)

// Parse reads a .env file and returns ordered entries
func Parse(reader io.Reader) ([]Entry, error) {
	var entries []Entry
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, initialBufferSize), maxBufferSize)

	var accumulated string
	var inQuote rune // 0 if not in quote, '"' or '\'' if inside quote

	for scanner.Scan() {
		line := scanner.Text()

		// Trim trailing carriage return to handle CRLF inputs consistently
		line = strings.TrimRight(line, "\r")

		// If we're accumulating a multiline value
		if inQuote != 0 {
			accumulated += "\n" + line
			if isMultilineClosed(accumulated, inQuote) {
				inQuote = 0
				trimmed := strings.TrimRight(accumulated, " \t\r\n")
				kv, err := parseKeyValue(trimmed)
				if err != nil {
					return nil, fmt.Errorf("parsing multiline value %q: %w", trimmed, err)
				}
				entries = append(entries, kv)
				accumulated = ""
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
				return nil, fmt.Errorf("parsing line %q: %w", line, err)
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
		// Extract key name for better error context
		key := "<unknown>"
		if eq := strings.Index(accumulated, "="); eq != -1 {
			key = strings.TrimSpace(accumulated[:eq])
		}

		// Create truncated snippet for error message
		snippet := accumulated
		const maxSnippetLen = 80
		if len(snippet) > maxSnippetLen {
			snippet = snippet[:maxSnippetLen-3] + "..."
		}

		return nil, fmt.Errorf("unclosed %q quote in multiline value for key %q starting with %q",
			string(inQuote), key, snippet)
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

// extractValuePart returns the portion of the line after the equals sign.
// Returns empty string if no equals sign is found.
func extractValuePart(line string) string {
	eqIdx := strings.IndexByte(line, '=')
	if eqIdx == -1 {
		return ""
	}
	return line[eqIdx+1:]
}

// findUnclosedQuote checks if a line contains an opening quote without a matching closing quote.
// Returns the quote character ('"' or '\”) if unclosed, or 0 if closed or no quotes.
func findUnclosedQuote(line string) rune {
	valuePart := extractValuePart(line)
	if valuePart == "" {
		return 0
	}

	// Check for double quote
	if strings.HasPrefix(valuePart, "\"") {
		count := countUnescapedQuotes(valuePart, '"')
		if count%2 == 1 {
			return '"'
		}
	}

	// Check for single quote
	if strings.HasPrefix(valuePart, "'") {
		count := countUnescapedQuotes(valuePart, '\'')
		if count%2 == 1 {
			return '\''
		}
	}

	return 0
}

// isMultilineClosed checks if the accumulated multiline value has a closing quote.
func isMultilineClosed(accumulated string, quote rune) bool {
	valuePart := extractValuePart(accumulated)
	if valuePart == "" || !strings.HasPrefix(valuePart, string(quote)) {
		return false
	}

	count := countUnescapedQuotes(valuePart, quote)
	return count > 0 && count%2 == 0
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

	// Check if value is quoted
	if len(value) >= 2 {
		firstChar, lastChar := value[0], value[len(value)-1]
		if (firstChar == '"' && lastChar == '"') || (firstChar == '\'' && lastChar == '\'') {
			kv.Quoted = string(firstChar)
			kv.Value = value[1 : len(value)-1]
			return kv, nil
		}
	}

	// Unquoted value (or too short to be quoted)
	kv.Value = value
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
