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
	items        []pickerItem
	selected     map[int]bool // only applies to non-header items
	cursor       int
	mode         MenuChoice
	rootDir      string
	windowHeight int
	offset       int // scroll offset (first visible item index)
}

// PickerFinishedMsg signals file selection is complete.
type PickerFinishedMsg struct {
	Selected []string
	Mode     MenuChoice
}

// groupFilesByDirectory organizes files into a list of pickerItem structs,
// grouping them by their parent directory with non-selectable headers.
func groupFilesByDirectory(files []string) []pickerItem {
	dirGroups := make(map[string][]string)
	for _, file := range files {
		dir := filepath.Dir(file)
		if dir == "." {
			dir = "Current Directory"
		}
		dirGroups[dir] = append(dirGroups[dir], file)
	}

	var dirs []string
	for dir := range dirGroups {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	var items []pickerItem
	for _, dir := range dirs {
		items = append(items, pickerItem{
			text:     dir,
			filePath: "",
			isHeader: true,
		})

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

	if mode == GenerateEnv {
		files, err = scanner.ScanExamples(rootDir)
	} else {
		files, err = scanner.Scan(rootDir)
	}

	if err != nil {
		files = []string{}
	}

	items := groupFilesByDirectory(files)

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

// SetWindowHeight sets the terminal height for scroll calculations.
func (m *PickerModel) SetWindowHeight(h int) {
	m.windowHeight = h
}

// Init initializes the picker model.
func (m PickerModel) Init() tea.Cmd {
	return nil
}

const pickerOverheadLines = 6 // title + padding + help + surrounding newlines

func (m PickerModel) visibleLines() int {
	if m.windowHeight <= pickerOverheadLines {
		return len(m.items)
	}
	maxVisible := m.windowHeight - pickerOverheadLines
	if maxVisible > len(m.items) {
		return len(m.items)
	}
	return maxVisible
}

func (m *PickerModel) ensureCursorVisible() {
	visible := m.visibleLines()
	if visible <= 0 {
		return
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
	maxOffset := len(m.items) - visible
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
}

// findNextSelectableItem finds the next item in the given direction
// that is not a header, starting from the given index.
func (m PickerModel) findNextSelectableItem(from int, direction int) int {
	for i := from; i >= 0 && i < len(m.items); i += direction {
		if !m.items[i].isHeader {
			return i
		}
	}
	return from
}

// Update handles messages and updates the picker model.
func (m PickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pickerInitMsg:
		m.items = msg.items
		m.selected = msg.selected
		m.mode = msg.mode
		m.rootDir = msg.rootDir
		if len(m.items) > 0 {
			m.cursor = m.findNextSelectableItem(0, 1)
		}
		m.ensureCursorVisible()
		return m, nil

	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height
		m.ensureCursorVisible()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				newCursor := m.cursor - 1
				m.cursor = m.findNextSelectableItem(newCursor, -1)
				m.ensureCursorVisible()
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				newCursor := m.cursor + 1
				m.cursor = m.findNextSelectableItem(newCursor, 1)
				m.ensureCursorVisible()
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
			if len(selectedFiles) > 0 {
				return m, func() tea.Msg {
					return PickerFinishedMsg{
						Selected: selectedFiles,
						Mode:     m.mode,
					}
				}
			}
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
	titleText := "Select .env files"
	if m.mode == GenerateEnv {
		titleText = "Select .env.example files"
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render(titleText)

	fileCount := 0
	for _, item := range m.items {
		if !item.isHeader {
			fileCount++
		}
	}

	if fileCount == 0 {
		noFilesText := "No .env files found in current directory"
		if m.mode == GenerateEnv {
			noFilesText = "No .env.example files found in current directory"
		}
		noFiles := lipgloss.NewStyle().
			Faint(true).
			Render(noFilesText)
		return "\n" + title + "\n\n" + noFiles + "\n\nPress q to return to menu"
	}

	var list string

	if fileCount == 1 {
		fileType := ".env"
		if m.mode == GenerateEnv {
			fileType = ".env.example"
		}
		singleFileIndicator := lipgloss.NewStyle().
			Faint(true).
			Render("(only 1 " + fileType + " file found)")
		list += singleFileIndicator + "\n\n"
	}

	visible := m.visibleLines()
	end := m.offset + visible
	if end > len(m.items) {
		end = len(m.items)
	}

	faintStyle := lipgloss.NewStyle().Faint(true)

	if m.offset > 0 {
		list += faintStyle.Render("  ↑ more items above") + "\n"
	}

	for i := m.offset; i < end; i++ {
		item := m.items[i]
		if item.isHeader {
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

	if end < len(m.items) {
		list += faintStyle.Render("  ↓ more items below") + "\n"
	}

	help := lipgloss.NewStyle().
		Faint(true).
		Render("↑/k: up • ↓/j: down • Space: toggle • a: all • Enter: confirm • q: back")

	return "\n" + title + "\n\n" + list + "\n" + help + "\n"
}
