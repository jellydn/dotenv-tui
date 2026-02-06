package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MenuChoice int

const (
	GenerateExample MenuChoice = iota
	GenerateEnv
)

type MenuModel struct {
	choice MenuChoice
}

func NewMenuModel() MenuModel {
	return MenuModel{
		choice: GenerateExample,
	}
}

// Choice returns the current menu choice
func (m MenuModel) Choice() MenuChoice {
	return m.choice
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.choice > GenerateExample {
				m.choice--
			}
		case "down", "j":
			if m.choice < GenerateEnv {
				m.choice++
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter", " ":
			// Let main handle the screen transition
			return m, nil
		}
	}
	return m, nil
}

func (m MenuModel) View() string {
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

	help := lipgloss.NewStyle().
		Faint(true).
		Render("↑/k: up • ↓/j: down • Enter: select • q: quit")

	return "\n" + title + "\n\n" + renderedChoices + "\n" + help + "\n"
}
