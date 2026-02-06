package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPickerModelInit(t *testing.T) {
	// Arrange
	model := PickerModel{
		files:    []string{"file1.env", "file2.env"},
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
		files:    []string{".env", "test/.env"},
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

	if len(newPickerModel.files) != 2 {
		t.Errorf("Update() files count = %d, expected 2", len(newPickerModel.files))
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
		initialFiles   []string
		keyMsg         tea.KeyMsg
		expectedCursor int
	}{
		{
			name:           "down key moves cursor down",
			initialCursor:  0,
			initialFiles:   []string{"file1.env", "file2.env", "file3.env"},
			keyMsg:         tea.KeyMsg{Type: tea.KeyDown},
			expectedCursor: 1,
		},
		{
			name:           "j key moves cursor down",
			initialCursor:  0,
			initialFiles:   []string{"file1.env", "file2.env"},
			keyMsg:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			expectedCursor: 1,
		},
		{
			name:           "up key moves cursor up",
			initialCursor:  1,
			initialFiles:   []string{"file1.env", "file2.env"},
			keyMsg:         tea.KeyMsg{Type: tea.KeyUp},
			expectedCursor: 0,
		},
		{
			name:           "k key moves cursor up",
			initialCursor:  1,
			initialFiles:   []string{"file1.env", "file2.env"},
			keyMsg:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			expectedCursor: 0,
		},
		{
			name:           "up key at top stays at top",
			initialCursor:  0,
			initialFiles:   []string{"file1.env", "file2.env"},
			keyMsg:         tea.KeyMsg{Type: tea.KeyUp},
			expectedCursor: 0,
		},
		{
			name:           "down key at bottom stays at bottom",
			initialCursor:  1,
			initialFiles:   []string{"file1.env", "file2.env"},
			keyMsg:         tea.KeyMsg{Type: tea.KeyDown},
			expectedCursor: 1,
		},
		{
			name:           "down key with single file stays at 0",
			initialCursor:  0,
			initialFiles:   []string{"file1.env"},
			keyMsg:         tea.KeyMsg{Type: tea.KeyDown},
			expectedCursor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			model := PickerModel{
				files:    tt.initialFiles,
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
		files:    []string{"file1.env", "file2.env"},
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
		files:    []string{"file1.env", "file2.env"},
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
		files:    []string{".env", "test/.env", "prod/.env"},
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

	// Execute the command to get the message
	msg := cmd()
	finishedMsg, ok := msg.(PickerFinishedMsg)
	if !ok {
		t.Fatalf("Command did not return PickerFinishedMsg")
	}

	// Should have 2 selected files
	if len(finishedMsg.Selected) != 2 {
		t.Errorf("PickerFinishedMsg.Selected = %d files, expected 2", len(finishedMsg.Selected))
	}

	// Should contain the correct files (order may vary due to map iteration)
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

	if finishedMsg.Mode != GenerateExample {
		t.Errorf("PickerFinishedMsg.Mode = %v, expected %v", finishedMsg.Mode, GenerateExample)
	}
}

func TestPickerModelUpdateEnterWithNoSelection(t *testing.T) {
	// Arrange
	model := PickerModel{
		files:    []string{".env"},
		selected: map[int]bool{0: false},
		cursor:   0,
		mode:     GenerateEnv,
	}

	// Act
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := model.Update(enterKey)

	// Assert
	if cmd == nil {
		t.Errorf("Update() enter key should return a command")
		return
	}

	// Execute the command
	msg := cmd()
	finishedMsg, ok := msg.(PickerFinishedMsg)
	if !ok {
		t.Fatalf("Command did not return PickerFinishedMsg")
	}

	// Should have 0 selected files
	if len(finishedMsg.Selected) != 0 {
		t.Errorf("PickerFinishedMsg.Selected = %d files, expected 0", len(finishedMsg.Selected))
	}
}

func TestPickerModelUpdateEmptyFilesNavigation(t *testing.T) {
	// Arrange
	model := PickerModel{
		files:    []string{},
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
				files:    []string{".env"},
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
