package tui

import (
	"strings"
	"testing"
)

func TestLogo(t *testing.T) {
	// Arrange: No preconditions needed

	// Act
	result := Logo()

	// Assert
	if result == "" {
		t.Errorf("Logo() returned empty string")
	}

	// Verify the logo contains expected elements
	expectedStrings := []string{
		".env",
		"╭",
		"╮",
		"╰",
		"╯",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Logo() does not contain expected string %q", expected)
		}
	}

	// Verify the logo has multiple lines (it's an ASCII art logo)
	lines := strings.Split(result, "\n")
	if len(lines) < 5 {
		t.Errorf("Logo() should have at least 5 lines, got %d", len(lines))
	}
}

func TestWordmark(t *testing.T) {
	// Arrange: No preconditions needed

	// Act
	result := Wordmark()

	// Assert
	if result == "" {
		t.Errorf("Wordmark() returned empty string")
	}

	// Verify the wordmark contains expected elements
	if !strings.Contains(result, "dotenv-tui") {
		t.Errorf("Wordmark() does not contain 'dotenv-tui'")
	}

	if !strings.Contains(result, "secure .env workflows") {
		t.Errorf("Wordmark() does not contain tagline")
	}

	// Verify the wordmark has exactly 2 lines (title and tagline)
	lines := strings.Split(result, "\n")
	if len(lines) != 2 {
		t.Errorf("Wordmark() should have exactly 2 lines, got %d", len(lines))
	}
}
