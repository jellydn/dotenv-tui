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
	preview       tui.PreviewModel
	form          tui.FormModel
}

type screen int

const (
	menuScreen screen = iota
	pickerScreen
	previewScreen
	formScreen
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
	case previewScreen:
		return updatePreview(msg, m)
	case formScreen:
		return updateForm(msg, m)
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
		// If generating .env.example, transition to preview
		if msg.Mode == tui.GenerateExample && len(msg.Selected) > 0 {
			m.currentScreen = previewScreen
			// For now, just use the first selected file
			return m, tui.NewPreviewModel(msg.Selected[0], nil)
		}
		// If generating .env, transition to form (assuming first selection is .env.example)
		if msg.Mode == tui.GenerateEnv && len(msg.Selected) > 0 {
			m.currentScreen = formScreen
			return m, tui.NewFormModel(msg.Selected[0])
		}
		// For other cases or no selection, go back to menu
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

func updatePreview(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	previewModel, previewCmd := m.preview.Update(msg)
	m.preview = previewModel.(tui.PreviewModel)
	cmd = previewCmd

	switch msg := msg.(type) {
	case tui.PreviewFinishedMsg:
		// Go back to menu after preview is finished
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

func updateForm(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	formModel, formCmd := m.form.Update(msg)
	m.form = formModel.(tui.FormModel)
	cmd = formCmd

	switch msg := msg.(type) {
	case tui.FormFinishedMsg:
		// Go back to menu after form is finished
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
	case previewScreen:
		return m.preview.View()
	case formScreen:
		return m.form.View()
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
