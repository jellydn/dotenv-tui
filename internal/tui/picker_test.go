package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPickerModelInit(t *testing.T) {
	// Arrange
	model := PickerModel{
		items: []pickerItem{
			{text: "file1.env", filePath: "file1.env", isHeader: false},
			{text: "file2.env", filePath: "file2.env", isHeader: false},
		},
		selected: map[int]bool{0: true, 1: false},
		cursor:   0,
		mode:     GenerateExample,
		rootDir:  "/test",
	}

	// Act
	cmd := model.Init()

	// Assert
	if cmd != nil {
		t.Errorf("PickerModel.Init() should return nil, got %v", cmd)
	}
}

func TestPickerModelUpdateWithInitMsg(t *testing.T) {
	// Arrange
	model := PickerModel{}
	initMsg := pickerInitMsg{
		items: []pickerItem{
			{text: ".env", filePath: ".env", isHeader: false},
			{text: "test/.env", filePath: "test/.env", isHeader: false},
		},
		selected: map[int]bool{0: true, 1: true},
		mode:     GenerateEnv,
		rootDir:  "/project",
	}

	// Act
	newModel, cmd := model.Update(initMsg)

	// Assert
	newPickerModel, ok := newModel.(PickerModel)
	if !ok {
		t.Fatalf("Update() did not return PickerModel")
	}

	if len(newPickerModel.items) != 2 {
		t.Errorf("Update() items count = %d, expected 2", len(newPickerModel.items))
	}

	if len(newPickerModel.selected) != 2 {
		t.Errorf("Update() selected count = %d, expected 2", len(newPickerModel.selected))
	}

	if newPickerModel.mode != GenerateEnv {
		t.Errorf("Update() mode = %v, expected %v", newPickerModel.mode, GenerateEnv)
	}

	if newPickerModel.rootDir != "/project" {
		t.Errorf("Update() rootDir = %q, expected %q", newPickerModel.rootDir, "/project")
	}

	if cmd != nil {
		t.Errorf("Update() should return nil command, got %v", cmd)
	}
}

func TestPickerModelUpdateNavigation(t *testing.T) {
	tests := []struct {
		name           string
		initialCursor  int
		initialItems   []pickerItem
		keyMsg         tea.KeyMsg
		expectedCursor int
	}{
		{
			name:           "down key moves cursor down",
			initialCursor:  0,
			initialItems:   []pickerItem{{text: "file1.env", filePath: "file1.env", isHeader: false}, {text: "file2.env", filePath: "file2.env", isHeader: false}, {text: "file3.env", filePath: "file3.env", isHeader: false}},
			keyMsg:         tea.KeyMsg{Type: tea.KeyDown},
			expectedCursor: 1,
		},
		{
			name:           "j key moves cursor down",
			initialCursor:  0,
			initialItems:   []pickerItem{{text: "file1.env", filePath: "file1.env", isHeader: false}, {text: "file2.env", filePath: "file2.env", isHeader: false}},
			keyMsg:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			expectedCursor: 1,
		},
		{
			name:           "up key moves cursor up",
			initialCursor:  1,
			initialItems:   []pickerItem{{text: "file1.env", filePath: "file1.env", isHeader: false}, {text: "file2.env", filePath: "file2.env", isHeader: false}},
			keyMsg:         tea.KeyMsg{Type: tea.KeyUp},
			expectedCursor: 0,
		},
		{
			name:           "k key moves cursor up",
			initialCursor:  1,
			initialItems:   []pickerItem{{text: "file1.env", filePath: "file1.env", isHeader: false}, {text: "file2.env", filePath: "file2.env", isHeader: false}},
			keyMsg:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			expectedCursor: 0,
		},
		{
			name:           "up key at top stays at top",
			initialCursor:  0,
			initialItems:   []pickerItem{{text: "file1.env", filePath: "file1.env", isHeader: false}, {text: "file2.env", filePath: "file2.env", isHeader: false}},
			keyMsg:         tea.KeyMsg{Type: tea.KeyUp},
			expectedCursor: 0,
		},
		{
			name:           "down key at bottom stays at bottom",
			initialCursor:  1,
			initialItems:   []pickerItem{{text: "file1.env", filePath: "file1.env", isHeader: false}, {text: "file2.env", filePath: "file2.env", isHeader: false}},
			keyMsg:         tea.KeyMsg{Type: tea.KeyDown},
			expectedCursor: 1,
		},
		{
			name:           "down key with single file stays at 0",
			initialCursor:  0,
			initialItems:   []pickerItem{{text: "file1.env", filePath: "file1.env", isHeader: false}},
			keyMsg:         tea.KeyMsg{Type: tea.KeyDown},
			expectedCursor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			model := PickerModel{
				items:    tt.initialItems,
				selected: make(map[int]bool),
				cursor:   tt.initialCursor,
			}

			// Act
			newModel, cmd := model.Update(tt.keyMsg)

			// Assert
			newPickerModel, ok := newModel.(PickerModel)
			if !ok {
				t.Fatalf("Update() did not return PickerModel")
			}

			if newPickerModel.cursor != tt.expectedCursor {
				t.Errorf("Update() cursor = %d, expected %d", newPickerModel.cursor, tt.expectedCursor)
			}

			if cmd != nil {
				t.Errorf("Update() should return nil command, got %v", cmd)
			}
		})
	}
}

func TestPickerModelUpdateToggleSelection(t *testing.T) {
	// Arrange
	model := PickerModel{
		items: []pickerItem{
			{text: "file1.env", filePath: "file1.env", isHeader: false},
			{text: "file2.env", filePath: "file2.env", isHeader: false},
		},
		selected: map[int]bool{0: true, 1: false},
		cursor:   1,
	}

	// Act
	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	newModel, cmd := model.Update(spaceKey)

	// Assert
	newPickerModel, ok := newModel.(PickerModel)
	if !ok {
		t.Fatalf("Update() did not return PickerModel")
	}

	// Selection should be toggled from false to true
	if !newPickerModel.selected[1] {
		t.Errorf("Update() space key should toggle selection, expected index 1 to be true")
	}

	if cmd != nil {
		t.Errorf("Update() should return nil command, got %v", cmd)
	}
}

func TestPickerModelUpdateToggleSelectionFromTrue(t *testing.T) {
	// Arrange
	model := PickerModel{
		items: []pickerItem{
			{text: "file1.env", filePath: "file1.env", isHeader: false},
			{text: "file2.env", filePath: "file2.env", isHeader: false},
		},
		selected: map[int]bool{0: true, 1: true},
		cursor:   1,
	}

	// Act
	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	newModel, cmd := model.Update(spaceKey)

	// Assert
	newPickerModel, ok := newModel.(PickerModel)
	if !ok {
		t.Fatalf("Update() did not return PickerModel")
	}

	// Selection should be toggled from true to false
	if newPickerModel.selected[1] {
		t.Errorf("Update() space key should toggle selection, expected index 1 to be false")
	}

	if cmd != nil {
		t.Errorf("Update() should return nil command, got %v", cmd)
	}
}

func TestPickerModelUpdateEnterWithSelection(t *testing.T) {
	// Arrange
	model := PickerModel{
		items: []pickerItem{
			{text: ".env", filePath: ".env", isHeader: false},
			{text: "test/.env", filePath: "test/.env", isHeader: false},
			{text: "prod/.env", filePath: "prod/.env", isHeader: false},
		},
		selected: map[int]bool{0: true, 1: false, 2: true},
		cursor:   1,
		mode:     GenerateExample,
	}

	// Act
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := model.Update(enterKey)

	// Assert
	if cmd == nil {
		t.Errorf("Update() enter key should return a command")
		return
	}

	// Execute command to get message
	msg := cmd()
	finishedMsg, ok := msg.(PickerFinishedMsg)
	if !ok {
		t.Fatalf("Command did not return PickerFinishedMsg")
	}

	// Should have 2 selected files
	if len(finishedMsg.Selected) != 2 {
		t.Errorf("PickerFinishedMsg.Selected = %d files, expected 2", len(finishedMsg.Selected))
	}

	// Should contain() correct files (order may vary due to map iteration)
	expectedFiles := map[string]bool{
		".env":      true,
		"prod/.env": true,
	}

	if len(finishedMsg.Selected) != 2 {
		t.Errorf("Expected 2 selected files, got %d", len(finishedMsg.Selected))
	}

	for _, file := range finishedMsg.Selected {
		if !expectedFiles[file] {
			t.Errorf("Unexpected selected file: %q", file)
		}
	}
}

func TestGroupFilesByDirectory(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected []pickerItem
	}{
		{
			name:  "files in different directories",
			files: []string{"apps/api/.env", "apps/web/.env", "services/auth/.env", "packages/db/.env"},
			expected: []pickerItem{
				{text: "apps/api", filePath: "", isHeader: true},
				{text: "apps/api/.env", filePath: "apps/api/.env", isHeader: false},
				{text: "apps/web", filePath: "", isHeader: true},
				{text: "apps/web/.env", filePath: "apps/web/.env", isHeader: false},
				{text: "packages/db", filePath: "", isHeader: true},
				{text: "packages/db/.env", filePath: "packages/db/.env", isHeader: false},
				{text: "services/auth", filePath: "", isHeader: true},
				{text: "services/auth/.env", filePath: "services/auth/.env", isHeader: false},
			},
		},
		{
			name:  "files in current directory",
			files: []string{".env", ".env.local"},
			expected: []pickerItem{
				{text: "Current Directory", filePath: "", isHeader: true},
				{text: ".env", filePath: ".env", isHeader: false},
				{text: ".env.local", filePath: ".env.local", isHeader: false},
			},
		},
		{
			name:     "empty list",
			files:    []string{},
			expected: []pickerItem{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := groupFilesByDirectory(tt.files)

			// Assert
			if len(result) != len(tt.expected) {
				t.Errorf("groupFilesByDirectory() returned %d items, expected %d", len(result), len(tt.expected))
				return
			}

			for i, item := range result {
				if i >= len(tt.expected) {
					t.Errorf("Result has more items than expected")
					break
				}
				expected := tt.expected[i]
				if item.text != expected.text || item.filePath != expected.filePath || item.isHeader != expected.isHeader {
					t.Errorf("Item %d mismatch:\n  got:      {text: %q, filePath: %q, isHeader: %v}\n  expected: {text: %q, filePath: %q, isHeader: %v}",
						i, item.text, item.filePath, item.isHeader, expected.text, expected.filePath, expected.isHeader)
				}
			}
		})
	}
}

func TestPickerModelUpdateEnterWithNoSelection(t *testing.T) {
	// Arrange
	model := PickerModel{
		items: []pickerItem{
			{text: ".env", filePath: ".env", isHeader: false},
		},
		selected: map[int]bool{0: false},
		cursor:   0,
		mode:     GenerateEnv,
	}

	// Act
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := model.Update(enterKey)

	// Assert - when no files are selected, Enter should do nothing (return nil command)
	if cmd != nil {
		t.Errorf("Update() enter key with no selection should return nil command, got %v", cmd)
	}
}

func TestPickerModelUpdateEmptyFilesNavigation(t *testing.T) {
	// Arrange
	model := PickerModel{
		items:    []pickerItem{},
		selected: map[int]bool{},
		cursor:   0,
	}

	tests := []struct {
		name   string
		keyMsg tea.KeyMsg
	}{
		{name: "down key with empty files", keyMsg: tea.KeyMsg{Type: tea.KeyDown}},
		{name: "up key with empty files", keyMsg: tea.KeyMsg{Type: tea.KeyUp}},
		{name: "space key with empty files", keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			newModel, cmd := model.Update(tt.keyMsg)

			// Assert
			newPickerModel, ok := newModel.(PickerModel)
			if !ok {
				t.Fatalf("Update() did not return PickerModel")
			}

			// Cursor should remain at 0
			if newPickerModel.cursor != 0 {
				t.Errorf("Update() cursor = %d, expected 0", newPickerModel.cursor)
			}

			if cmd != nil {
				t.Errorf("Update() should return nil command, got %v", cmd)
			}
		})
	}
}

func TestPickerModelUpdateQuit(t *testing.T) {
	tests := []struct {
		name   string
		keyMsg tea.KeyMsg
	}{
		{name: "q key returns nil command", keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
		{name: "esc key returns nil command", keyMsg: tea.KeyMsg{Type: tea.KeyEsc}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			model := PickerModel{
				items: []pickerItem{
					{text: ".env", filePath: ".env", isHeader: false},
				},
				selected: map[int]bool{0: true},
				cursor:   0,
			}

			// Act
			newModel, cmd := model.Update(tt.keyMsg)

			// Assert
			_, ok := newModel.(PickerModel)
			if !ok {
				t.Fatalf("Update() did not return PickerModel")
			}

			// q and esc should return nil (handled by main)
			if cmd != nil {
				t.Errorf("Update(%q) should return nil command, got %v", tt.name, cmd)
			}
		})
	}
}
