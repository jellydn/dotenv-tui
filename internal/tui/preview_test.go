package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jellydn/dotenv-tui/internal/parser"

	tea "github.com/charmbracelet/bubbletea"
)

func TestEntryToString(t *testing.T) {
	tests := []struct {
		name     string
		entry    parser.Entry
		expected string
	}{
		{
			name:     "simple key-value",
			entry:    parser.KeyValue{Key: "KEY", Value: "value", Quoted: "", Exported: false},
			expected: "KEY=value",
		},
		{
			name:     "key-value with export prefix",
			entry:    parser.KeyValue{Key: "KEY", Value: "value", Quoted: "", Exported: true},
			expected: "export KEY=value",
		},
		{
			name:     "key-value with double quotes",
			entry:    parser.KeyValue{Key: "KEY", Value: "value with spaces", Quoted: "\"", Exported: false},
			expected: `KEY="value with spaces"`,
		},
		{
			name:     "key-value with single quotes",
			entry:    parser.KeyValue{Key: "KEY", Value: "value", Quoted: "'", Exported: false},
			expected: "KEY='value'",
		},
		{
			name:     "export with quoted value",
			entry:    parser.KeyValue{Key: "API_KEY", Value: "secret", Quoted: "'", Exported: true},
			expected: "export API_KEY='secret'",
		},
		{
			name:     "comment entry",
			entry:    parser.Comment{Text: "# This is a comment"},
			expected: "# This is a comment",
		},
		{
			name:     "blank line entry",
			entry:    parser.BlankLine{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.EntryToString(tt.entry)

			if result != tt.expected {
				t.Errorf("parser.EntryToString() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestPreviewModelInit(t *testing.T) {
	model := PreviewModel{}

	cmd := model.Init()

	if cmd != nil {
		t.Errorf("PreviewModel.Init() should return nil, got %v", cmd)
	}
}

func TestPreviewModelUpdateWithInitMsg(t *testing.T) {
	model := PreviewModel{}
	files := []filePreview{
		{
			filePath:   "/test/.env",
			outputPath: "/test/.env.example",
			diffLines:  []string{"- KEY=value", "+ KEY=placeholder"},
			generatedEntries: []parser.Entry{
				parser.KeyValue{Key: "KEY", Value: "placeholder", Quoted: "", Exported: false},
			},
		},
	}

	initMsg := previewInitMsg{files: files}

	newModel, cmd := model.Update(initMsg)

	newPreviewModel, ok := newModel.(PreviewModel)
	if !ok {
		t.Fatalf("Update() did not return PreviewModel")
	}

	if len(newPreviewModel.files) != 1 {
		t.Errorf("Update() files count = %d, expected 1", len(newPreviewModel.files))
	}

	if newPreviewModel.files[0].outputPath != "/test/.env.example" {
		t.Errorf("Update() outputPath = %q, expected %q", newPreviewModel.files[0].outputPath, "/test/.env.example")
	}

	if cmd != nil {
		t.Errorf("Update() should return nil command, got %v", cmd)
	}
}

func TestPreviewModelUpdateNavigation(t *testing.T) {
	diffLines := []string{"line 1", "line 2", "line 3", "line 4", "line 5"}
	model := PreviewModel{
		files: []filePreview{
			{diffLines: diffLines},
		},
		currentFile:  0,
		cursor:       1,
		scrollOffset: 0,
	}

	tests := []struct {
		name           string
		initialCursor  int
		initialScroll  int
		keyMsg         tea.KeyMsg
		expectedCursor int
		expectedScroll int
	}{
		{
			name:           "down key moves cursor down",
			initialCursor:  1,
			initialScroll:  0,
			keyMsg:         tea.KeyMsg{Type: tea.KeyDown},
			expectedCursor: 2,
			expectedScroll: 0,
		},
		{
			name:           "j key moves cursor down",
			initialCursor:  1,
			initialScroll:  0,
			keyMsg:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			expectedCursor: 2,
			expectedScroll: 0,
		},
		{
			name:           "up key moves cursor up",
			initialCursor:  1,
			initialScroll:  0,
			keyMsg:         tea.KeyMsg{Type: tea.KeyUp},
			expectedCursor: 0,
			expectedScroll: 0,
		},
		{
			name:           "k key moves cursor up",
			initialCursor:  1,
			initialScroll:  0,
			keyMsg:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			expectedCursor: 0,
			expectedScroll: 0,
		},
		{
			name:           "up key at top stays at top",
			initialCursor:  0,
			initialScroll:  0,
			keyMsg:         tea.KeyMsg{Type: tea.KeyUp},
			expectedCursor: 0,
			expectedScroll: 0,
		},
		{
			name:           "down key at bottom stays at bottom",
			initialCursor:  4,
			initialScroll:  0,
			keyMsg:         tea.KeyMsg{Type: tea.KeyDown},
			expectedCursor: 4,
			expectedScroll: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.cursor = tt.initialCursor
			model.scrollOffset = tt.initialScroll

			newModel, cmd := model.Update(tt.keyMsg)

			newPreviewModel, ok := newModel.(PreviewModel)
			if !ok {
				t.Fatalf("Update() did not return PreviewModel")
			}

			if newPreviewModel.cursor != tt.expectedCursor {
				t.Errorf("Update() cursor = %d, expected %d", newPreviewModel.cursor, tt.expectedCursor)
			}

			if newPreviewModel.scrollOffset != tt.expectedScroll {
				t.Errorf("Update() scrollOffset = %d, expected %d", newPreviewModel.scrollOffset, tt.expectedScroll)
			}

			if cmd != nil {
				t.Errorf("Update() should return nil command, got %v", cmd)
			}
		})
	}
}

func TestPreviewModelUpdateEnterWritesAll(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test.env.example")

	model := PreviewModel{
		files: []filePreview{
			{
				filePath:   "/test/.env",
				outputPath: outputPath,
				diffLines:  []string{"line 1", "line 2"},
				generatedEntries: []parser.Entry{
					parser.KeyValue{Key: "TEST", Value: "value", Quoted: "", Exported: false},
				},
			},
		},
		currentFile:  0,
		cursor:       0,
		scrollOffset: 0,
	}

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := model.Update(enterKey)

	newPreviewModel, ok := newModel.(PreviewModel)
	if !ok {
		t.Fatalf("Update() did not return PreviewModel")
	}

	if !newPreviewModel.written {
		t.Errorf("Update() enter key should set written to true")
	}

	if cmd != nil {
		t.Errorf("Update() should return nil command after enter, got %v", cmd)
	}

	// Verify file was written
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Expected file to be written at %s", outputPath)
	}
}

func TestPreviewModelUpdateQuit(t *testing.T) {
	tests := []struct {
		name   string
		keyMsg tea.KeyMsg
	}{
		{name: "q key returns quit command", keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
		{name: "esc key returns quit command", keyMsg: tea.KeyMsg{Type: tea.KeyEsc}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := PreviewModel{
				files: []filePreview{
					{diffLines: []string{"line 1"}},
				},
				currentFile:  0,
				cursor:       0,
				scrollOffset: 0,
			}

			_, cmd := model.Update(tt.keyMsg)

			if cmd == nil {
				t.Errorf("Update(%q) should return a command", tt.name)
				return
			}

			msg := cmd()
			_, ok := msg.(PreviewFinishedMsg)
			if !ok {
				t.Fatalf("Command did not return PreviewFinishedMsg")
			}
		})
	}
}

func TestPreviewModelTabSwitchesFile(t *testing.T) {
	model := PreviewModel{
		files: []filePreview{
			{filePath: "file1.env", diffLines: []string{"line 1"}},
			{filePath: "file2.env", diffLines: []string{"line 2"}},
		},
		currentFile:  0,
		cursor:       0,
		scrollOffset: 0,
	}

	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	newModel, _ := model.Update(tabKey)

	newPreviewModel := newModel.(PreviewModel)
	if newPreviewModel.currentFile != 1 {
		t.Errorf("Tab should switch to file 1, got %d", newPreviewModel.currentFile)
	}

	// Tab again should wrap to file 0
	newModel, _ = newPreviewModel.Update(tabKey)
	newPreviewModel = newModel.(PreviewModel)
	if newPreviewModel.currentFile != 0 {
		t.Errorf("Tab should wrap to file 0, got %d", newPreviewModel.currentFile)
	}
}

func TestPreviewModelShiftTabSwitchesFile(t *testing.T) {
	model := PreviewModel{
		files: []filePreview{
			{filePath: "file1.env", diffLines: []string{"line 1"}},
			{filePath: "file2.env", diffLines: []string{"line 2"}},
		},
		currentFile:  0,
		cursor:       0,
		scrollOffset: 0,
	}

	shiftTabKey := tea.KeyMsg{Type: tea.KeyShiftTab}
	newModel, _ := model.Update(shiftTabKey)

	newPreviewModel := newModel.(PreviewModel)
	if newPreviewModel.currentFile != 1 {
		t.Errorf("Shift+Tab from 0 should wrap to file 1, got %d", newPreviewModel.currentFile)
	}
}

func TestPreviewModelWrittenStateEnterFinishes(t *testing.T) {
	model := PreviewModel{
		files: []filePreview{
			{filePath: "file1.env", diffLines: []string{"line 1"}},
		},
		currentFile: 0,
		written:     true,
		writeResults: []writeResult{
			{OutputPath: "/test/.env.example", Success: true},
		},
	}

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := model.Update(enterKey)

	if cmd == nil {
		t.Fatal("Enter in written state should return a command")
	}

	msg := cmd()
	finishedMsg, ok := msg.(PreviewFinishedMsg)
	if !ok {
		t.Fatalf("Command did not return PreviewFinishedMsg")
	}

	if len(finishedMsg.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(finishedMsg.Results))
	}
}

func TestEntryToStringComplex(t *testing.T) {
	tests := []struct {
		name     string
		entry    parser.Entry
		contains string
	}{
		{
			name:     "key-value contains equals sign",
			entry:    parser.KeyValue{Key: "DB_HOST", Value: "localhost", Quoted: "", Exported: false},
			contains: "=",
		},
		{
			name:     "comment starts with hash",
			entry:    parser.Comment{Text: "# Database config"},
			contains: "#",
		},
		{
			name:     "quoted value preserves quotes",
			entry:    parser.KeyValue{Key: "URL", Value: "https://example.com", Quoted: "\"", Exported: false},
			contains: "\"",
		},
		{
			name:     "export prefix is included",
			entry:    parser.KeyValue{Key: "KEY", Value: "value", Quoted: "", Exported: true},
			contains: "export",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.EntryToString(tt.entry)

			if !strings.Contains(result, tt.contains) {
				t.Errorf("parser.EntryToString() result %q does not contain %q", result, tt.contains)
			}
		})
	}
}
