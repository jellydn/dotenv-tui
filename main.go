package main

import (
	"fmt"
	"os"

	"dotenv-tui/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	currentScreen screen
	menu          tui.MenuModel
	picker        tui.PickerModel
}

type screen int

const (
	menuScreen screen = iota
	pickerScreen
)

func initialModel() model {
	return model{
		currentScreen: menuScreen,
		menu:          tui.NewMenuModel(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.currentScreen {
	case menuScreen:
		return updateMenu(msg, m)
	case pickerScreen:
		return updatePicker(msg, m)
	}
	return m, nil
}

func updateMenu(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	menuModel, menuCmd := m.menu.Update(msg)
	m.menu = menuModel.(tui.MenuModel)
	cmd = menuCmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" || msg.String() == " " {
			// Transition to picker
			m.currentScreen = pickerScreen
			return m, tui.NewPickerModel(m.menu.Choice(), ".")
		}
	}

	return m, cmd
}

func updatePicker(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	pickerModel, pickerCmd := m.picker.Update(msg)
	m.picker = pickerModel.(tui.PickerModel)
	cmd = pickerCmd

	switch msg := msg.(type) {
	case tui.PickerFinishedMsg:
		// For now, just go back to menu. In future stories, this will transition to preview/form
		m.currentScreen = menuScreen
		m.menu = tui.NewMenuModel()
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "esc" {
			// Return to menu
			m.currentScreen = menuScreen
			m.menu = tui.NewMenuModel()
			return m, nil
		}
	}

	return m, cmd
}

func (m model) View() string {
	switch m.currentScreen {
	case menuScreen:
		return m.menu.View()
	case pickerScreen:
		return m.picker.View()
	default:
		return ""
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}
