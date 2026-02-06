package generator

import (
	"dotenv-tui/internal/detector"
	"dotenv-tui/internal/parser"
)

// GenerateExample creates a .env.example from .env entries by masking secrets
func GenerateExample(entries []parser.Entry) []parser.Entry {
	var result []parser.Entry

	for _, entry := range entries {
		switch e := entry.(type) {
		case parser.KeyValue:
			// Check if this is a secret
			if detector.IsSecret(e.Key, e.Value) {
				// Replace with placeholder
				placeholder := detector.GeneratePlaceholder(e.Key, e.Value)
				newKV := parser.KeyValue{
					Key:      e.Key,
					Value:    placeholder,
					Quoted:   "", // Placeholders are not quoted
					Exported: e.Exported,
				}
				result = append(result, newKV)
			} else {
				// Keep non-secret values as-is
				result = append(result, e)
			}

		case parser.Comment:
			// Preserve comments as-is
			result = append(result, e)

		case parser.BlankLine:
			// Preserve blank lines as-is
			result = append(result, e)

		default:
			// Preserve any unknown entry types as-is
			result = append(result, e)
		}
	}

	return result
}

// GenerateEnv creates a .env from .env.example entries by copying them
// This is for non-interactive mode where we just copy entries as-is
func GenerateEnv(entries []parser.Entry) []parser.Entry {
	return append([]parser.Entry(nil), entries...)
}
