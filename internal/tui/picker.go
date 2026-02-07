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
	var files []string
	var err error

	// Choose scanner based on mode
	switch mode {
	case GenerateExample:
		files, err = scanner.Scan(rootDir)
	case GenerateEnv:
		files, err = scanner.ScanExamples(rootDir)
	default:
		files, err = scanner.Scan(rootDir)
	}

	if err != nil {
		files = []string{}
	}

	// Initialize with no files selected by default
	selected := make(map[int]bool)
	for i := range files {
		selected[i] = false
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
			var selectedFiles []string
			for i := 0; i < len(m.files); i++ {
				if m.selected[i] {
					selectedFiles = append(selectedFiles, m.files[i])
				}
			}
			// Only proceed if at least one file is selected
			if len(selectedFiles) > 0 {
				return m, func() tea.Msg {
					return PickerFinishedMsg{
						Selected: selectedFiles,
						Mode:     m.mode,
					}
				}
			}
			// If no files selected, do nothing
		case "q", "esc":
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the file picker UI.
func (m PickerModel) View() string {
	var titleText string
	switch m.mode {
	case GenerateExample:
		titleText = "Select .env files"
	case GenerateEnv:
		titleText = "Select .env.example files"
	default:
		titleText = "Select .env files"
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render(titleText)

	if len(m.files) == 0 {
		var noFilesText string
		switch m.mode {
		case GenerateExample:
			noFilesText = "No .env files found in current directory"
		case GenerateEnv:
			noFilesText = "No .env.example files found in current directory"
		default:
			noFilesText = "No .env files found in current directory"
		}
		noFiles := lipgloss.NewStyle().
			Faint(true).
			Render(noFilesText)
		return "\n" + title + "\n\n" + noFiles + "\n\nPress q to return to menu"
	}

	var list string

	if len(m.files) == 1 {
		var fileType string
		switch m.mode {
		case GenerateExample:
			fileType = ".env"
		case GenerateEnv:
			fileType = ".env.example"
		default:
			fileType = ".env"
		}
		singleFileIndicator := lipgloss.NewStyle().
			Faint(true).
			Render("(only 1 " + fileType + " file found)")
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
