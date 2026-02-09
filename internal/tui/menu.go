// Package tui provides Bubble Tea components for the terminal UI.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MenuChoice represents the user's selection in the main menu.
type MenuChoice int

const (
	// GenerateExample creates .env.example from .env files.
	GenerateExample MenuChoice = iota
	// GenerateEnv creates .env files from .env.example.
	GenerateEnv
)

// MenuModel is the Bubble Tea model for the main menu.
type MenuModel struct {
	choice       MenuChoice
	enableBackup bool
}

// NewMenuModel creates a new menu model with default selection.
func NewMenuModel() MenuModel {
	return MenuModel{
		choice:       GenerateExample,
		enableBackup: true,
	}
}

// Choice returns the current menu choice.
func (m MenuModel) Choice() MenuChoice {
	return m.choice
}

// EnableBackup returns whether backups are enabled.
func (m MenuModel) EnableBackup() bool {
	return m.enableBackup
}

// Init initializes the menu model.
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the menu model.
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if ok {
		switch keyMsg.String() {
		case "up", "k":
			if m.choice > GenerateExample {
				m.choice--
			}
		case "down", "j":
			if m.choice < GenerateEnv {
				m.choice++
			}
		case "b":
			m.enableBackup = !m.enableBackup
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter", " ":
			return m, nil
		}
	}
	return m, nil
}

// View renders the menu UI.
func (m MenuModel) View() string {
	logo := Logo()
	wordmark := Wordmark()

	header := lipgloss.JoinHorizontal(lipgloss.Top, logo, "  "+wordmark)

	choices := []string{
		"Generate .env.example from .env",
		"Generate .env from .env.example",
	}

	var renderedChoices string
	for i, choice := range choices {
		cursor := " "
		if MenuChoice(i) == m.choice {
			cursor = ">"
			renderedChoices += lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true).
				Render(cursor+" "+choice) + "\n"
		} else {
			renderedChoices += cursor + " " + choice + "\n"
		}
	}

	var backupStatus string
	if m.enableBackup {
		backupStatus = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Render("[B] Backup: ON")
	} else {
		backupStatus = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F56")).
			Render("[B] Backup: OFF")
	}

	help := lipgloss.NewStyle().
		Faint(true).
		Render("↑/k: up • ↓/j: down • b: toggle backup • Enter: select • q: quit")

	return "\n" + header + "\n\n" + renderedChoices + "\n" + backupStatus + "\n\n" + help + "\n"
}
