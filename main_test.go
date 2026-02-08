// Package main provides CLI entry point tests.
package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jellydn/dotenv-tui/internal/tui"
)

func TestUpdateFormTracksSavedFiles(t *testing.T) {
	// Arrange
	m := model{
		currentScreen: formScreen,
		fileList:      []string{"/test/file1.env", "/test/file2.env", "/test/file3.env"},
		fileIndex:     0,
		savedFiles:    make(map[int]bool),
		form:          tui.FormModel{},
	}
	msg := tui.FormSavedMsg{Success: true, Error: ""}

	// Act
	newModel, _ := updateForm(msg, m)
	newModelTyped, ok := newModel.(model)
	if !ok {
		t.Fatalf("updateForm() should return model type")
	}

	// Assert
	if !newModelTyped.savedFiles[0] {
		t.Errorf("updateForm() should mark file index 0 as saved after successful save")
	}

	if len(newModelTyped.savedFiles) != 1 {
		t.Errorf("updateForm() savedFiles length = %d, expected 1", len(newModelTyped.savedFiles))
	}
}

func TestUpdateFormDoesNotTrackFailedSaves(t *testing.T) {
	// Arrange
	m := model{
		currentScreen: formScreen,
		fileList:      []string{"/test/file1.env"},
		fileIndex:     0,
		savedFiles:    make(map[int]bool),
		form:          tui.FormModel{},
	}
	msg := tui.FormSavedMsg{Success: false, Error: "permission denied"}

	// Act
	newModel, _ := updateForm(msg, m)
	newModelTyped, ok := newModel.(model)
	if !ok {
		t.Fatalf("updateForm() should return model type")
	}

	// Assert
	if newModelTyped.savedFiles[0] {
		t.Errorf("updateForm() should not mark file as saved after failed save")
	}

	if len(newModelTyped.savedFiles) != 0 {
		t.Errorf("updateForm() savedFiles length = %d, expected 0", len(newModelTyped.savedFiles))
	}
}

func TestUpdateFormReturnsToMenuOnEnter(t *testing.T) {
	// Arrange
	m := model{
		currentScreen: formScreen,
		fileList:      []string{"/test/file1.env"},
		fileIndex:     0,
		savedFiles:    make(map[int]bool),
		form:          tui.FormModel{},
	}
	msg := tui.FormFinishedMsg{Success: true, Error: "", Dir: 0}

	// Act
	newModel, _ := updateForm(msg, m)
	newModelTyped := newModel.(model)

	// Assert
	if newModelTyped.currentScreen != menuScreen {
		t.Errorf("updateForm() with Dir=0 should return to menuScreen, got %v", newModelTyped.currentScreen)
	}
}

func TestUpdateFormReturnsToMenuWhenAllFilesSaved(t *testing.T) {
	// Arrange
	m := model{
		currentScreen: formScreen,
		fileList:      []string{"/test/file1.env", "/test/file2.env", "/test/file3.env"},
		fileIndex:     0,
		savedFiles:    map[int]bool{0: true, 1: true, 2: true},
		form:          tui.FormModel{},
	}
	msg := tui.FormFinishedMsg{Success: true, Error: "", Dir: 1}

	// Act
	newModel, _ := updateForm(msg, m)
	newModelTyped := newModel.(model)

	// Assert
	if newModelTyped.currentScreen != menuScreen {
		t.Errorf("updateForm() when all files saved should return to menuScreen, got %v", newModelTyped.currentScreen)
	}
}

func TestUpdateFormNavigatesToNextUnsavedFile(t *testing.T) {
	tests := []struct {
		name           string
		fileList       []string
		initialIndex   int
		savedFiles     map[int]bool
		direction      int
		expectedIndex  int
		expectedScreen screen
	}{
		{
			name:           "moves to next file when current is saved",
			fileList:       []string{"f1", "f2", "f3"},
			initialIndex:   0,
			savedFiles:     map[int]bool{0: true},
			direction:      1,
			expectedIndex:  1,
			expectedScreen: formScreen,
		},
		{
			name:           "skips saved files moving forward",
			fileList:       []string{"f1", "f2", "f3", "f4"},
			initialIndex:   0,
			savedFiles:     map[int]bool{0: true, 1: true},
			direction:      1,
			expectedIndex:  2,
			expectedScreen: formScreen,
		},
		{
			name:           "wraps around to find unsaved file",
			fileList:       []string{"f1", "f2", "f3"},
			initialIndex:   2,
			savedFiles:     map[int]bool{2: true, 0: true},
			direction:      1,
			expectedIndex:  1,
			expectedScreen: formScreen,
		},
		{
			name:           "returns to menu when all files are saved",
			fileList:       []string{"f1", "f2", "f3"},
			initialIndex:   0,
			savedFiles:     map[int]bool{0: true, 1: true, 2: true},
			direction:      1,
			expectedIndex:  0,
			expectedScreen: menuScreen,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			m := model{
				currentScreen: formScreen,
				fileList:      tt.fileList,
				fileIndex:     tt.initialIndex,
				savedFiles:    tt.savedFiles,
				form:          tui.FormModel{},
			}
			msg := tui.FormFinishedMsg{Success: true, Error: "", Dir: tt.direction}

			// Act
			newModel, _ := updateForm(msg, m)
			newModelTyped := newModel.(model)

			// Assert
			if newModelTyped.currentScreen != tt.expectedScreen {
				t.Errorf("updateForm() screen = %v, expected %v", newModelTyped.currentScreen, tt.expectedScreen)
			}

			if newModelTyped.currentScreen == formScreen && newModelTyped.fileIndex != tt.expectedIndex {
				t.Errorf("updateForm() fileIndex = %d, expected %d", newModelTyped.fileIndex, tt.expectedIndex)
			}
		})
	}
}

func TestReturnToMenu(t *testing.T) {
	// Arrange
	m := model{
		currentScreen: formScreen,
		form:          tui.FormModel{},
		menu:          tui.MenuModel{},
	}

	// Act
	newModel := returnToMenu(m)
	newModelTyped := newModel.(model)

	// Assert
	if newModelTyped.currentScreen != menuScreen {
		t.Errorf("returnToMenu() should set currentScreen to menuScreen, got %v", newModelTyped.currentScreen)
	}
}

func TestInitialModel(t *testing.T) {
	// Act
	m := initialModel()

	// Assert
	if m.currentScreen != menuScreen {
		t.Errorf("initialModel() should start at menuScreen, got %v", m.currentScreen)
	}

	if m.fileList != nil {
		t.Errorf("initialModel() fileList should be nil, got %v", m.fileList)
	}

	if m.savedFiles != nil {
		t.Errorf("initialModel() savedFiles should be nil, got %v", m.savedFiles)
	}
}

func TestModelUpdateWindowSize(t *testing.T) {
	// Arrange
	m := initialModel()
	msg := tea.WindowSizeMsg{Height: 42}

	// Act
	newModel, _ := m.Update(msg)
	newModelTyped, ok := newModel.(model)
	if !ok {
		t.Fatalf("Update() should return model type")
	}

	// Assert
	if newModelTyped.windowHeight != 42 {
		t.Errorf("Update(WindowSizeMsg) windowHeight = %d, expected 42", newModelTyped.windowHeight)
	}
}
