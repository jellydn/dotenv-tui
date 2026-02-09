// Package tui provides Bubble Tea components for the terminal UI.
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jellydn/dotenv-tui/internal/backup"
	"github.com/jellydn/dotenv-tui/internal/parser"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormField represents a single field in the form for user input.
type FormField struct {
	Key           string
	Value         string
	Placeholder   string
	Input         textinput.Model
	IsPlaceholder bool
}

// FormModel is the Bubble Tea model for the interactive form component.
type FormModel struct {
	fields          []FormField
	originalEntries []parser.Entry
	cursor          int
	scroll          int
	filePath        string
	confirmed       bool
	errorMsg        string
	fileIndex       int
	totalFiles      int
	savedFiles      map[int]bool
	enableBackup    bool
}

// FormSavedMsg signals the form save operation has completed.
type FormSavedMsg struct {
	Success bool
	Error   string
}

// FormFinishedMsg signals the form has completed with success status.
type FormFinishedMsg struct {
	Success bool
	Error   string
	Dir     int
}

type formInitMsg struct {
	fields          []FormField
	originalEntries []parser.Entry
	filePath        string
	fileIndex       int
	totalFiles      int
	savedFiles      map[int]bool
	enableBackup    bool
}

// NewFormModel creates a new form model for collecting environment variables.
func NewFormModel(exampleFilePath string, fileIndex, totalFiles int, savedFiles map[int]bool, enableBackup bool) tea.Cmd {
	return func() tea.Msg {
		file, err := os.Open(exampleFilePath)
		if err != nil {
			return formInitMsg{
				filePath:     exampleFilePath,
				fields:       []FormField{},
				fileIndex:    fileIndex,
				totalFiles:   totalFiles,
				savedFiles:   savedFiles,
				enableBackup: enableBackup,
			}
		}
		defer func() { _ = file.Close() }()

		entries, err := parser.Parse(file)
		if err != nil {
			return formInitMsg{
				filePath:     exampleFilePath,
				fields:       []FormField{},
				fileIndex:    fileIndex,
				totalFiles:   totalFiles,
				savedFiles:   savedFiles,
				enableBackup: enableBackup,
			}
		}

		var fields []FormField
		for _, entry := range entries {
			if kv, ok := entry.(parser.KeyValue); ok {
				isPlaceholder := isPlaceholderValue(kv.Value)
				var placeholder, value string

				if isPlaceholder {
					placeholder = generateHint(kv.Key, kv.Value)
				} else {
					value = kv.Value
				}

				input := textinput.New()
				input.SetValue(value)
				input.Placeholder = placeholder
				input.Width = 50

				fields = append(fields, FormField{
					Key:           kv.Key,
					Value:         value,
					Placeholder:   placeholder,
					Input:         input,
					IsPlaceholder: isPlaceholder,
				})
			}
		}

		return formInitMsg{
			fields:          fields,
			originalEntries: entries,
			filePath:        exampleFilePath,
			fileIndex:       fileIndex,
			totalFiles:      totalFiles,
			savedFiles:      savedFiles,
			enableBackup:    enableBackup,
		}
	}
}

// isPlaceholderValue returns true if the value appears to be a placeholder.
// It checks for common placeholder patterns like *** suffix, "your_*" prefix,
// and words like "placeholder" or "example" in the value.
func isPlaceholderValue(value string) bool {
	if strings.HasSuffix(value, "***") {
		return true
	}

	lower := strings.ToLower(value)
	placeholderPatterns := []string{"your_", "_here", "placeholder"}
	for _, pattern := range placeholderPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	if strings.Contains(lower, "example") && !strings.Contains(lower, "://") {
		return true
	}

	return false
}

// generateHint creates a context-aware placeholder hint for a given key.
// It maps common key patterns to appropriate user-friendly hints.
func generateHint(key, _ string) string {
	lowerKey := strings.ToLower(key)

	// Map patterns to hints
	hintMap := []struct {
		patterns []string
		hint     string
	}{
		{[]string{"api", "key"}, "Enter your API key"},
		{[]string{"secret"}, "Enter your secret"},
		{[]string{"token"}, "Enter your token"},
		{[]string{"password", "pass"}, "Enter your password"},
		{[]string{"url", "uri"}, "Enter URL (e.g., https://example.com)"},
		{[]string{"port"}, "Enter port number (e.g., 3000)"},
		{[]string{"host"}, "Enter host (e.g., localhost)"},
		{[]string{"database", "db"}, "Enter database connection string"},
	}

	for _, entry := range hintMap {
		for _, pattern := range entry.patterns {
			if strings.Contains(lowerKey, pattern) {
				return entry.hint
			}
		}
	}

	return "Enter value for " + key
}

// Init initializes the form model.
func (m FormModel) Init() tea.Cmd {
	return nil
}

// moveCursor moves the cursor and updates scroll position
const visibleFields = 7

const (
	directionUp   = -1
	directionDown = 1
)

// moveCursor moves the cursor to a new position and updates the scroll offset
// to keep the cursor visible within the visible fields window.
func (m *FormModel) moveCursor(newCursor int) {
	m.fields[m.cursor].Input.Blur()
	m.cursor = newCursor
	m.fields[m.cursor].Input.Focus()

	if m.cursor < m.scroll {
		m.scroll = m.cursor
	} else if m.cursor >= m.scroll+visibleFields {
		m.scroll = m.cursor - visibleFields + 1
	}
}

// moveCursorByDirection moves the cursor by the specified direction (up or down).
// It clamps the movement to stay within the bounds of available fields.
func (m *FormModel) moveCursorByDirection(dir int) {
	newCursor := m.cursor + dir
	if newCursor >= 0 && newCursor < len(m.fields) {
		m.moveCursor(newCursor)
	}
}

// Update handles messages and updates the form model.
func (m FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case formInitMsg:
		m.fields = msg.fields
		m.originalEntries = msg.originalEntries
		m.filePath = msg.filePath
		m.fileIndex = msg.fileIndex
		m.totalFiles = msg.totalFiles
		m.savedFiles = msg.savedFiles
		m.enableBackup = msg.enableBackup
		m.cursor = 0
		m.scroll = 0
		m.confirmed = false
		m.errorMsg = ""

		if len(m.fields) > 0 {
			m.fields[0].Input.Focus()
		}
		return m, nil

	case FormSavedMsg:
		m.confirmed = true
		if msg.Success {
			m.errorMsg = ""
		} else {
			m.errorMsg = msg.Error
		}
		return m, nil

	case tea.KeyMsg:
		if m.confirmed {
			switch msg.String() {
			case "tab":
				if m.totalFiles > 1 {
					return m, func() tea.Msg {
						return FormFinishedMsg{Success: m.errorMsg == "", Error: m.errorMsg, Dir: 1}
					}
				}
			case "enter", "q", "esc":
				return m, func() tea.Msg {
					return FormFinishedMsg{Success: m.errorMsg == "", Error: m.errorMsg, Dir: 0}
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "up", "shift+tab":
			m.moveCursorByDirection(directionUp)
		case "down", "tab":
			m.moveCursorByDirection(directionDown)
		case "enter":
			if m.cursor == len(m.fields)-1 {
				return m, m.saveForm()
			}
			m.moveCursorByDirection(directionDown)
		case "esc":
			return m, func() tea.Msg {
				return FormFinishedMsg{Success: false, Error: "cancelled", Dir: 0}
			}
		}
	}

	// Update the currently focused field
	if len(m.fields) > 0 && m.cursor >= 0 && m.cursor < len(m.fields) {
		updatedInput, cmd := m.fields[m.cursor].Input.Update(msg)
		m.fields[m.cursor].Input = updatedInput
		return m, cmd
	}

	return m, nil
}

// saveForm processes the form fields and writes the resulting .env file.
// It returns a command that emits a FormSavedMsg upon completion.
func (m FormModel) saveForm() tea.Cmd {
	return func() tea.Msg {
		outputPath := filepath.Join(filepath.Dir(m.filePath), ".env")

		fieldIndex := 0
		var entries []parser.Entry
		for _, entry := range m.originalEntries {
			switch e := entry.(type) {
			case parser.KeyValue:
				if fieldIndex < len(m.fields) {
					newValue := m.fields[fieldIndex].Input.Value()
					entries = append(entries, parser.KeyValue{
						Key:      e.Key,
						Value:    newValue,
						Quoted:   e.Quoted,
						Exported: e.Exported,
					})
					fieldIndex++
				}
			case parser.Comment, parser.BlankLine:
				entries = append(entries, e)
			}
		}

		if m.enableBackup {
			if _, err := os.Stat(outputPath); err == nil {
				if _, err := backup.CreateBackup(outputPath); err != nil {
					return FormSavedMsg{Success: false, Error: fmt.Sprintf("Failed to create backup: %v", err)}
				}
			}
		}

		file, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return FormSavedMsg{Success: false, Error: fmt.Sprintf("Failed to create file: %v", err)}
		}
		defer func() { _ = file.Close() }()

		if err := parser.Write(file, entries); err != nil {
			return FormSavedMsg{Success: false, Error: fmt.Sprintf("Failed to write file: %v", err)}
		}

		return FormSavedMsg{Success: true}
	}
}

// View renders the form UI.
func (m FormModel) View() string {
	if m.confirmed {
		allDone := len(m.savedFiles) >= m.totalFiles
		var helpText string
		switch {
		case allDone:
			helpText = "All files saved! Enter: done"
		case m.totalFiles > 1:
			remaining := m.totalFiles - len(m.savedFiles)
			helpText = fmt.Sprintf("Tab: next file • Enter: done (%d remaining)", remaining)
		default:
			helpText = "Enter: done"
		}

		if m.errorMsg != "" {
			title := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF5F56")).
				Bold(true).
				Render("Error")

			message := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Render(m.errorMsg)

			help := lipgloss.NewStyle().
				Faint(true).
				Render(helpText)

			return fmt.Sprintf(
				"\n%s\n\n%s\n\n%s\n",
				title,
				message,
				help,
			)
		}

		title := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5FAF5F")).
			Bold(true).
			Render("Success!")

		outputPath := filepath.Join(filepath.Dir(m.filePath), ".env")
		message := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Render(fmt.Sprintf("Successfully wrote %s", outputPath))

		help := lipgloss.NewStyle().
			Faint(true).
			Render(helpText)

		return fmt.Sprintf(
			"\n%s\n\n%s\n\n%s\n",
			title,
			message,
			help,
		)
	}

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true).
		Render("Edit Environment Variables")

	savedCount := len(m.savedFiles)
	positionText := fmt.Sprintf("[%d/%d] %s  (%d/%d saved)", m.fileIndex+1, m.totalFiles, m.filePath, savedCount, m.totalFiles)
	subtitle := lipgloss.NewStyle().
		Faint(true).
		Render(positionText)

	var form strings.Builder

	// Calculate visible area
	visibleStart := m.scroll
	visibleEnd := m.scroll + visibleFields
	if visibleEnd > len(m.fields) {
		visibleEnd = len(m.fields)
	}

	for i := visibleStart; i < visibleEnd; i++ {
		field := m.fields[i]

		// Field label
		var label string
		if i == m.cursor {
			label = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true).
				Render(field.Key + ":")
		} else {
			label = field.Key + ":"
		}

		// Input field
		input := field.Input.View()

		// Add hint text for placeholder fields if empty
		if field.IsPlaceholder && field.Input.Value() == "" && field.Input.Placeholder != "" {
			hint := lipgloss.NewStyle().
				Faint(true).
				Italic(true).
				Render("  (" + field.Input.Placeholder + ")")
			form.WriteString(fmt.Sprintf("%s\n%s\n%s\n", label, input, hint))
		} else {
			form.WriteString(fmt.Sprintf("%s\n%s\n", label, input))
		}
	}

	// Scroll indicator
	if len(m.fields) > visibleFields {
		scrollInfo := lipgloss.NewStyle().
			Faint(true).
			Render(fmt.Sprintf("Showing %d-%d of %d fields", visibleStart+1, visibleEnd, len(m.fields)))
		form.WriteString("\n" + scrollInfo + "\n")
	}

	help := lipgloss.NewStyle().
		Faint(true).
		Render("↑: up • ↓: down • Tab: next • Shift+Tab: prev • Enter: next/submit • Esc: cancel")

	return fmt.Sprintf(
		"\n%s\n%s\n\n%s\n\n%s\n",
		title,
		subtitle,
		form.String(),
		help,
	)
}
