package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jellydn/env-man/internal/parser"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FormField struct {
	Key           string
	Value         string
	Placeholder   string
	Input         textinput.Model
	IsPlaceholder bool
}

type FormModel struct {
	fields     []FormField
	cursor     int
	scroll     int
	filePath   string
	confirmed  bool
	errorMsg   string
	successMsg string
}

type FormFinishedMsg struct {
	Success bool
	Error   string
}

type formInitMsg struct {
	fields   []FormField
	filePath string
}

func NewFormModel(exampleFilePath string) tea.Cmd {
	return func() tea.Msg {
		// Read the .env.example file
		file, err := os.Open(exampleFilePath)
		if err != nil {
			return formInitMsg{
				filePath: exampleFilePath,
				fields:   []FormField{},
			}
		}
		defer func() { _ = file.Close() }()

		entries, err := parser.Parse(file)
		if err != nil {
			return formInitMsg{
				filePath: exampleFilePath,
				fields:   []FormField{},
			}
		}

		// Generate initial entries to get proper structure
		generated := entries

		var fields []FormField
		for _, entry := range generated {
			if kv, ok := entry.(parser.KeyValue); ok {
				// Check if the value looks like a placeholder
				isPlaceholder := isPlaceholderValue(kv.Value)
				var placeholder string
				var value string

				if isPlaceholder {
					placeholder = generateHint(kv.Key, kv.Value)
					value = ""
				} else {
					value = kv.Value
					placeholder = ""
				}

				// Create text input
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
			fields:   fields,
			filePath: exampleFilePath,
		}
	}
}

func isPlaceholderValue(value string) bool {
	lower := strings.ToLower(value)
	placeholderPatterns := []string{"your_", "_here", "placeholder"}
	for _, pattern := range placeholderPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	// Example without URLs
	if strings.Contains(lower, "example") && !strings.Contains(lower, "://") {
		return true
	}

	// Generic and format-specific placeholders
	if value == "***" {
		return true
	}

	// Prefix-based placeholders
	placeholderPrefixes := []string{"sk_", "ghp_", "eyJ"}
	for _, prefix := range placeholderPrefixes {
		if strings.HasPrefix(value, prefix) && strings.HasSuffix(value, "***") {
			return true
		}
	}

	return false
}

func generateHint(key, value string) string {
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

func (m FormModel) Init() tea.Cmd {
	return nil
}

// moveCursor moves the cursor and updates scroll position
const visibleFields = 7

func (m *FormModel) moveCursor(newCursor int) {
	m.fields[m.cursor].Input.Blur()
	m.cursor = newCursor
	m.fields[m.cursor].Input.Focus()

	// Adjust scroll to keep cursor visible
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	} else if m.cursor >= m.scroll+visibleFields {
		m.scroll = m.cursor - visibleFields + 1
	}
}

func (m FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case formInitMsg:
		m.fields = msg.fields
		m.filePath = msg.filePath

		// Set focus on first field if there are any
		if len(m.fields) > 0 {
			m.fields[0].Input.Focus()
		}
		return m, nil

	case tea.KeyMsg:
		if m.confirmed {
			if msg.String() == "q" || msg.String() == "esc" {
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.moveCursor(m.cursor - 1)
			}
		case "down", "j":
			if m.cursor < len(m.fields)-1 {
				m.moveCursor(m.cursor + 1)
			}
		case "tab":
			if m.cursor < len(m.fields)-1 {
				m.moveCursor(m.cursor + 1)
			}
		case "shift+tab":
			if m.cursor > 0 {
				m.moveCursor(m.cursor - 1)
			}
		case "ctrl+s":
			return m, m.saveForm()
		case "enter":
			if m.cursor == len(m.fields)-1 {
				// Enter on last field submits
				return m, m.saveForm()
			}
			// Otherwise move to next field
			if m.cursor < len(m.fields)-1 {
				m.moveCursor(m.cursor + 1)
			}
		case "q", "esc":
			return m, func() tea.Msg {
				return FormFinishedMsg{Success: false, Error: "cancelled"}
			}
		}
	}

	// Update the currently focused field
	if len(m.fields) > 0 {
		updatedInput, cmd := m.fields[m.cursor].Input.Update(msg)
		m.fields[m.cursor].Input = updatedInput
		return m, cmd
	}

	return m, nil
}

func (m FormModel) saveForm() tea.Cmd {
	return func() tea.Msg {
		// Determine output file path (.env)
		outputPath := filepath.Join(filepath.Dir(m.filePath), ".env")

		// Create entries from form data
		var entries []parser.Entry
		for _, field := range m.fields {
			value := field.Input.Value()
			if value == "" {
				value = field.Input.Placeholder
			}
			entries = append(entries, parser.KeyValue{
				Key:   field.Key,
				Value: value,
			})
		}

		// Write to file
		file, err := os.Create(outputPath)
		if err != nil {
			return FormFinishedMsg{Success: false, Error: fmt.Sprintf("Failed to create file: %v", err)}
		}
		defer func() { _ = file.Close() }()

		if err := parser.Write(file, entries); err != nil {
			return FormFinishedMsg{Success: false, Error: fmt.Sprintf("Failed to write file: %v", err)}
		}

		return FormFinishedMsg{Success: true}
	}
}

func (m FormModel) View() string {
	if m.confirmed {
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
				Render("Press q to quit")

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

		message := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Render(m.successMsg)

		help := lipgloss.NewStyle().
			Faint(true).
			Render("Press q to quit")

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

	subtitle := lipgloss.NewStyle().
		Faint(true).
		Render(fmt.Sprintf("Editing: %s", m.filePath))

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
		Render("↑/k: up • ↓/j: down • Tab: next • Shift+Tab: previous • Enter: submit • Ctrl+S: save • q/esc: cancel")

	return fmt.Sprintf(
		"\n%s\n%s\n\n%s\n\n%s\n",
		title,
		subtitle,
		form.String(),
		help,
	)
}
