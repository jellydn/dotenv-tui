// Package tui provides Bubble Tea components for the terminal UI.
package tui

import (
	"github.com/jellydn/dotenv-tui/internal/scanner"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PickerModel is the Bubble Tea model for selecting .env files.
type PickerModel struct {
	files    []string
	selected map[int]bool
	cursor   int
	mode     MenuChoice
	rootDir  string
}

// PickerFinishedMsg signals file selection is complete.
type PickerFinishedMsg struct {
	Selected []string
	Mode     MenuChoice
}

// NewPickerModel creates a file picker for selecting .env files.
func NewPickerModel(mode MenuChoice, rootDir string) tea.Cmd {
	files, err := scanner.Scan(rootDir)
	if err != nil {
		// Return empty list if scan fails - we could handle this better in a real app
		files = []string{}
	}

	selected := make(map[int]bool)
	for i := range files {
		selected[i] = true // Select all files by default
	}

	return func() tea.Msg {
		return pickerInitMsg{
			files:    files,
			selected: selected,
			mode:     mode,
			rootDir:  rootDir,
		}
	}
}

type pickerInitMsg struct {
	files    []string
	selected map[int]bool
	mode     MenuChoice
	rootDir  string
}

// Init initializes the picker model.
func (m PickerModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the picker model.
func (m PickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pickerInitMsg:
		m.files = msg.files
		m.selected = msg.selected
		m.mode = msg.mode
		m.rootDir = msg.rootDir
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
			}
		case " ":
			if len(m.files) > 0 {
				m.selected[m.cursor] = !m.selected[m.cursor]
			}
		case "a":
			// Toggle select all
			if len(m.files) > 0 {
				allSelected := true
				for i := range m.files {
					if !m.selected[i] {
						allSelected = false
						break
					}
				}
				for i := range m.files {
					m.selected[i] = !allSelected
				}
			}
		case "enter":
			// Collect selected files (iterate in order to ensure deterministic output)
			var selectedFiles []string
			for i := 0; i < len(m.files); i++ {
				if m.selected[i] {
					selectedFiles = append(selectedFiles, m.files[i])
				}
			}
			return m, func() tea.Msg {
				return PickerFinishedMsg{
					Selected: selectedFiles,
					Mode:     m.mode,
				}
			}
		case "q", "esc":
			// Let main handle the screen transition
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the file picker UI.
func (m PickerModel) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("Select .env files")

	if len(m.files) == 0 {
		noFiles := lipgloss.NewStyle().
			Faint(true).
			Render("No .env files found in current directory")
		return "\n" + title + "\n\n" + noFiles + "\n\nPress q to return to menu"
	}

	var list string

	// Show indicator if only one file found
	if len(m.files) == 1 {
		singleFileIndicator := lipgloss.NewStyle().
			Faint(true).
			Render("(only 1 file found)")
		list += singleFileIndicator + "\n\n"
	}

	for i, file := range m.files {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		checkbox := "[ ]"
		if m.selected[i] {
			checkbox = "[x]"
		}

		style := lipgloss.NewStyle()
		if i == m.cursor {
			style = style.Foreground(lipgloss.Color("#7D56F4")).Bold(true)
		}

		list += style.Render(cursor+" "+checkbox+" "+file) + "\n"
	}

	help := lipgloss.NewStyle().
		Faint(true).
		Render("↑/k: up • ↓/j: down • Space: toggle • a: all • Enter: confirm • q: back")

	return "\n" + title + "\n\n" + list + "\n" + help + "\n"
}
