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

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines (but preserve as BlankLine)
		if line == "" {
			entries = append(entries, BlankLine{})
			continue
		}

		// Handle comments
		if strings.HasPrefix(line, "#") {
			entries = append(entries, Comment{Text: line})
			continue
		}

		// Handle key-value pairs
		if strings.Contains(line, "=") {
			kv, err := parseKeyValue(line)
			if err != nil {
				return nil, fmt.Errorf("error parsing line %q: %w", line, err)
			}
			entries = append(entries, kv)
			continue
		}

		// If it's not a comment, key-value, or blank, treat as comment
		entries = append(entries, Comment{Text: line})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading: %w", err)
	}

	return entries, nil
}

// parseKeyValue parses a single key-value line
func parseKeyValue(line string) (KeyValue, error) {
	var kv KeyValue

	// Check for export prefix
	if strings.HasPrefix(line, "export ") {
		kv.Exported = true
		line = strings.TrimSpace(line[7:]) // Remove "export "
	}

	// Split on first =
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return KeyValue{}, fmt.Errorf("invalid key-value format")
	}

	kv.Key = strings.TrimSpace(parts[0])
	value := parts[1]

	// Handle quoted values
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'') {
			kv.Quoted = string(value[0])
			kv.Value = value[1 : len(value)-1] // Remove quotes
		} else {
			kv.Value = value
		}
	} else {
		kv.Value = value
	}

	return kv, nil
}

// Write writes entries to a writer, preserving the original structure
func Write(writer io.Writer, entries []Entry) error {
	for _, entry := range entries {
		switch e := entry.(type) {
		case KeyValue:
			line := ""
			if e.Exported {
				line += "export "
			}
			line += e.Key + "="
			if e.Quoted != "" {
				line += e.Quoted + e.Value + e.Quoted
			} else {
				line += e.Value
			}
			if _, err := fmt.Fprintln(writer, line); err != nil {
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
