package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuChoice int

const (
	generateExample menuChoice = iota
	generateEnv
)

type menuModel struct {
	choice menuChoice
}

func NewMenuModel() menuModel {
	return menuModel{
		choice: generateExample,
	}
}

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.choice > generateExample {
				m.choice--
			}
		case "down", "j":
			if m.choice < generateEnv {
				m.choice++
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter", " ":
			// For now, just quit. In future stories, this will transition to file picker
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m menuModel) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("dotenv-tui")

	choices := []string{
		"Generate .env.example from .env",
		"Generate .env from .env.example",
	}

	var renderedChoices string
	for i, choice := range choices {
		cursor := " "
		if menuChoice(i) == m.choice {
			cursor = ">"
			renderedChoices += lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true).
				Render(cursor + " " + choice + "\n")
		} else {
			renderedChoices += cursor + " " + choice + "\n"
		}
	}

	help := lipgloss.NewStyle().
		Faint(true).
		Render("↑/k: up • ↓/j: down • Enter: select • q: quit")

	return "\n" + title + "\n\n" + renderedChoices + "\n" + help + "\n"
}
