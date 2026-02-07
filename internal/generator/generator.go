// Package generator creates .env.example files by masking secrets.
package generator

import (
	"github.com/jellydn/dotenv-tui/internal/detector"
	"github.com/jellydn/dotenv-tui/internal/parser"
)

// GenerateExample creates a .env.example from .env entries by masking secrets
func GenerateExample(entries []parser.Entry) []parser.Entry {
	var result []parser.Entry

	for _, entry := range entries {
		switch e := entry.(type) {
		case parser.KeyValue:
			if detector.IsSecret(e.Key, e.Value) {
				placeholder := detector.GeneratePlaceholder(e.Key, e.Value)
				newKV := parser.KeyValue{
					Key:      e.Key,
					Value:    placeholder,
					Quoted:   "",
					Exported: e.Exported,
				}
				result = append(result, newKV)
			} else {
				result = append(result, e)
			}

		case parser.Comment, parser.BlankLine:
			result = append(result, e)

		default:
			result = append(result, e)
		}
	}

	return result
}
