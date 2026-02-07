// Package tui provides Bubble Tea components for the terminal UI.
package tui

import (
	"path/filepath"
	"sort"

	"github.com/jellydn/dotenv-tui/internal/scanner"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// pickerItem represents an item in the picker list (either a header or a file).
type pickerItem struct {
	text     string
	filePath string // empty for headers
	isHeader bool
}

// PickerModel is the Bubble Tea model for selecting .env files.
type PickerModel struct {
	items    []pickerItem
	selected map[int]bool // only applies to non-header items
	cursor   int
	mode     MenuChoice
	rootDir  string
}

// PickerFinishedMsg signals file selection is complete.
type PickerFinishedMsg struct {
	Selected []string
	Mode     MenuChoice
}

// groupFilesByDirectory groups files by directory and creates picker items with headers.
func groupFilesByDirectory(files []string) []pickerItem {
	// Group files by directory
	dirGroups := make(map[string][]string)
	for _, file := range files {
		dir := filepath.Dir(file)
		if dir == "." {
			dir = "Current Directory"
		}
		dirGroups[dir] = append(dirGroups[dir], file)
	}

	// Sort directories alphabetically
	var dirs []string
	for dir := range dirGroups {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	// Create items with headers and sorted files
	var items []pickerItem
	for _, dir := range dirs {
		// Add header
		items = append(items, pickerItem{
			text:     dir,
			filePath: "",
			isHeader: true,
		})

		// Sort files within the directory alphabetically
		sort.Strings(dirGroups[dir])
		for _, file := range dirGroups[dir] {
			items = append(items, pickerItem{
				text:     file,
				filePath: file,
				isHeader: false,
			})
		}
	}

	return items
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

	// Group files by directory with headers
	items := groupFilesByDirectory(files)

	// Initialize with no files selected by default (only for non-header items)
	selected := make(map[int]bool)
	for i, item := range items {
		if !item.isHeader {
			selected[i] = false
		}
	}

	return func() tea.Msg {
		return pickerInitMsg{
			items:    items,
			selected: selected,
			mode:     mode,
			rootDir:  rootDir,
		}
	}
}

type pickerInitMsg struct {
	items    []pickerItem
	selected map[int]bool
	mode     MenuChoice
	rootDir  string
}

// Init initializes the picker model.
func (m PickerModel) Init() tea.Cmd {
	return nil
}

// findNextSelectableItem finds the next selectable (non-header) item.
func (m PickerModel) findNextSelectableItem(from int, direction int) int {
	for i := from; i >= 0 && i < len(m.items); i += direction {
		if !m.items[i].isHeader {
			return i
		}
	}
	return from // return original if no selectable item found
}

// Update handles messages and updates the picker model.
func (m PickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pickerInitMsg:
		m.items = msg.items
		m.selected = msg.selected
		m.mode = msg.mode
		m.rootDir = msg.rootDir
		// Set cursor to first selectable item
		if len(m.items) > 0 {
			m.cursor = m.findNextSelectableItem(0, 1)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				newCursor := m.cursor - 1
				m.cursor = m.findNextSelectableItem(newCursor, -1)
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				newCursor := m.cursor + 1
				m.cursor = m.findNextSelectableItem(newCursor, 1)
			}
		case " ":
			if len(m.items) > 0 && !m.items[m.cursor].isHeader {
				m.selected[m.cursor] = !m.selected[m.cursor]
			}
		case "a":
			if len(m.items) > 0 {
				allSelected := true
				for i := range m.items {
					if !m.items[i].isHeader && !m.selected[i] {
						allSelected = false
						break
					}
				}
				for i := range m.items {
					if !m.items[i].isHeader {
						m.selected[i] = !allSelected
					}
				}
			}
		case "enter":
			var selectedFiles []string
			for i := 0; i < len(m.items); i++ {
				if !m.items[i].isHeader && m.selected[i] {
					selectedFiles = append(selectedFiles, m.items[i].filePath)
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

	// Count actual files (non-header items)
	fileCount := 0
	for _, item := range m.items {
		if !item.isHeader {
			fileCount++
		}
	}

	if fileCount == 0 {
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

	if fileCount == 1 {
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

	for i, item := range m.items {
		if item.isHeader {
			// Render header with different style
			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Faint(true).
				PaddingLeft(2)
			list += headerStyle.Render(item.text) + "\n"
		} else {
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

			list += style.Render(cursor+" "+checkbox+" "+item.text) + "\n"
		}
	}

	help := lipgloss.NewStyle().
		Faint(true).
		Render("↑/k: up • ↓/j: down • Space: toggle • a: all • Enter: confirm • q: back")

	return "\n" + title + "\n\n" + list + "\n" + help + "\n"
}
