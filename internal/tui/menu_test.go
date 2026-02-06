package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewMenuModel(t *testing.T) {
	// Arrange: No preconditions needed

	// Act
	model := NewMenuModel()

	// Assert
	if model.choice != GenerateExample {
		t.Errorf("NewMenuModel() choice = %v, expected %v", model.choice, GenerateExample)
	}
}

func TestMenuModelChoice(t *testing.T) {
	tests := []struct {
		name     string
		choice   MenuChoice
		expected MenuChoice
	}{
		{
			name:     "returns GenerateExample choice",
			choice:   GenerateExample,
			expected: GenerateExample,
		},
		{
			name:     "returns GenerateEnv choice",
			choice:   GenerateEnv,
			expected: GenerateEnv,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			model := MenuModel{choice: tt.choice}

			// Act
			result := model.Choice()

			// Assert
			if result != tt.expected {
				t.Errorf("MenuModel.Choice() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMenuModelInit(t *testing.T) {
	// Arrange
	model := NewMenuModel()

	// Act
	cmd := model.Init()

	// Assert
	if cmd != nil {
		t.Errorf("MenuModel.Init() should return nil, got %v", cmd)
	}
}

func TestMenuModelUpdateNavigation(t *testing.T) {
	tests := []struct {
		name          string
		initialChoice MenuChoice
		keyMsg        string
		expectedChoice MenuChoice
	}{
		{
			name:          "up key from GenerateEnv moves to GenerateExample",
			initialChoice: GenerateEnv,
			keyMsg:        "up",
			expectedChoice: GenerateExample,
		},
		{
			name:          "k key from GenerateEnv moves to GenerateExample",
			initialChoice: GenerateEnv,
			keyMsg:        "k",
			expectedChoice: GenerateExample,
		},
		{
			name:          "down key from GenerateExample moves to GenerateEnv",
			initialChoice: GenerateExample,
			keyMsg:        "down",
			expectedChoice: GenerateEnv,
		},
		{
			name:          "j key from GenerateExample moves to GenerateEnv",
			initialChoice: GenerateExample,
			keyMsg:        "j",
			expectedChoice: GenerateEnv,
		},
		{
			name:          "up key at GenerateExample stays at GenerateExample",
			initialChoice: GenerateExample,
			keyMsg:        "up",
			expectedChoice: GenerateExample,
		},
		{
			name:          "down key at GenerateEnv stays at GenerateEnv",
			initialChoice: GenerateEnv,
			keyMsg:        "down",
			expectedChoice: GenerateEnv,
		},
		{
			name:          "enter key does not change choice",
			initialChoice: GenerateExample,
			keyMsg:        "enter",
			expectedChoice: GenerateExample,
		},
		{
			name:          "space key does not change choice",
			initialChoice: GenerateEnv,
			keyMsg:        " ",
			expectedChoice: GenerateEnv,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			model := MenuModel{choice: tt.initialChoice}
			msg := tea.KeyMsg{Type: tea.KeyKeyPress, Runes: []rune{}, Alt: false}

			// Set key type based on message string
			switch tt.keyMsg {
			case "up":
				msg.Type = tea.KeyUp
			case "down":
				msg.Type = tea.KeyDown
			case "k":
				msg.Type = tea.KeyRunes
				msg.Runes = []rune{'k'}
			case "j":
				msg.Type = tea.KeyRunes
				msg.Runes = []rune{'j'}
			case "enter":
				msg.Type = tea.KeyEnter
			case " ":
				msg.Type = tea.KeyRunes
				msg.Runes = []rune{' '}
			}

			// Act
			newModel, cmd := model.Update(msg)

			// Assert
			newMenuModel, ok := newModel.(MenuModel)
			if !ok {
				t.Fatalf("Update() did not return MenuModel")
			}

			if newMenuModel.choice != tt.expectedChoice {
				t.Errorf("Update(%q) choice = %v, expected %v", tt.keyMsg, newMenuModel.choice, tt.expectedChoice)
			}

			// Verify no command is returned for navigation keys
			if cmd != nil {
				t.Errorf("Update(%q) should not return a command, got %v", tt.keyMsg, cmd)
			}
		})
	}
}

func TestMenuModelUpdateQuit(t *testing.T) {
	tests := []struct {
		name   string
		keyMsg tea.KeyMsg
	}{
		{
			name:   "q key returns quit command",
			keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
		},
		{
			name:   "ctrl+c returns quit command",
			keyMsg: tea.KeyMsg{Type: tea.KeyCtrlC},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			model := NewMenuModel()

			// Act
			_, cmd := model.Update(tt.keyMsg)

			// Assert
			if cmd == nil {
				t.Errorf("Update(%q) should return quit command", tt.name)
			} else if cmd != tea.Quit {
				t.Errorf("Update(%q) should return tea.Quit, got %v", tt.name, cmd)
			}
		})
	}
}

func TestMenuModelUpdateUnknownMessage(t *testing.T) {
	// Arrange
	model := NewMenuModel()
	initialChoice := model.choice
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}

	// Act
	newModel, cmd := model.Update(msg)

	// Assert
	newMenuModel, ok := newModel.(MenuModel)
	if !ok {
		t.Fatalf("Update() did not return MenuModel")
	}

	if newMenuModel.choice != initialChoice {
		t.Errorf("Update(unknown key) changed choice from %v to %v", initialChoice, newMenuModel.choice)
	}

	if cmd != nil {
		t.Errorf("Update(unknown key) should return nil command, got %v", cmd)
	}
}
