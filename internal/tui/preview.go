package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dotenv-tui/internal/generator"
	"dotenv-tui/internal/parser"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PreviewModel struct {
	originalEntries  []parser.Entry
	generatedEntries []parser.Entry
	diffLines        []string
	cursor           int
	scrollOffset     int
	outputPath       string
	confirmed        bool
	success          bool
}

type PreviewFinishedMsg struct {
	Success bool
}

func NewPreviewModel(filePath string, _ []parser.Entry) tea.Cmd {
	return func() tea.Msg {
		// Read the original file
		file, err := os.Open(filePath)
		if err != nil {
			return previewInitMsg{
				originalEntries:  []parser.Entry{},
				generatedEntries: []parser.Entry{},
				diffLines:        []string{fmt.Sprintf("Error reading file: %v", err)},
				outputPath:       filepath.Join(filepath.Dir(filePath), ".env.example"),
			}
		}
		defer func() { _ = file.Close() }()

		// Parse the original file
		originalEntries, err := parser.Parse(file)
		if err != nil {
			return previewInitMsg{
				originalEntries:  []parser.Entry{},
				generatedEntries: []parser.Entry{},
				diffLines:        []string{fmt.Sprintf("Error parsing file: %v", err)},
				outputPath:       filepath.Join(filepath.Dir(filePath), ".env.example"),
			}
		}

		// Generate the .env.example entries
		generatedEntries := generator.GenerateExample(originalEntries)

		// Create diff lines
		var diffLines []string
		for i, orig := range originalEntries {
			if i < len(generatedEntries) {
				origLine := entryToString(orig)
				genLine := entryToString(generatedEntries[i])

				if origLine == genLine {
					// Unchanged line - show in green
					diffLines = append(diffLines, fmt.Sprintf("  %s", origLine))
				} else {
					// Changed line - show old → new in yellow
					diffLines = append(diffLines, fmt.Sprintf("- %s", origLine))
					diffLines = append(diffLines, fmt.Sprintf("+ %s", genLine))
				}
			}
		}

		outputPath := filepath.Join(filepath.Dir(filePath), ".env.example")

		return previewInitMsg{
			originalEntries:  originalEntries,
			generatedEntries: generatedEntries,
			diffLines:        diffLines,
			outputPath:       outputPath,
		}
	}
}

type previewInitMsg struct {
	originalEntries  []parser.Entry
	generatedEntries []parser.Entry
	diffLines        []string
	outputPath       string
}

func entryToString(entry parser.Entry) string {
	switch e := entry.(type) {
	case parser.KeyValue:
		line := ""
		if e.Exported {
			line += "export "
		}
		line += e.Key + "="
		if e.Quoted != "" {
			line += e.Quoted + e.Value + e.Quoted
		} else {
			line += e.Value
		}
		return line
	case parser.Comment:
		return e.Text
	case parser.BlankLine:
		return ""
	default:
		return ""
	}
}

func (m PreviewModel) Init() tea.Cmd {
	return nil
}

func (m PreviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case previewInitMsg:
		m.originalEntries = msg.originalEntries
		m.generatedEntries = msg.generatedEntries
		m.diffLines = msg.diffLines
		m.outputPath = msg.outputPath
		return m, nil

	case tea.KeyMsg:
		if m.confirmed && m.success {
			switch msg.String() {
			case "enter", "q", "esc":
				return m, func() tea.Msg {
					return PreviewFinishedMsg{Success: true}
				}
			}
		} else if m.confirmed {
			switch msg.String() {
			case "y", "Y", "enter":
				// Write the file
				if err := m.writeFile(); err != nil {
					// Handle error - for now just show failure
					m.success = false
					return m, nil
				}
				m.success = true
				return m, nil
			case "n", "N", "q", "esc":
				// Cancel and go back
				return m, func() tea.Msg {
					return PreviewFinishedMsg{Success: false}
				}
			}
		} else {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					if m.cursor < m.scrollOffset {
						m.scrollOffset = m.cursor
					}
				}
			case "down", "j":
				if m.cursor < len(m.diffLines)-1 {
					m.cursor++
					// Auto-scroll when cursor reaches bottom
					visibleLines := 10 // Approximate visible lines
					if m.cursor >= m.scrollOffset+visibleLines {
						m.scrollOffset = m.cursor - visibleLines + 1
					}
				}
			case "enter":
				m.confirmed = true
			case "q", "esc":
				return m, func() tea.Msg {
					return PreviewFinishedMsg{Success: false}
				}
			}
		}
	}
	return m, nil
}

func (m PreviewModel) writeFile() error {
	// Create the output file
	file, err := os.Create(m.outputPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	// Write the generated entries
	return parser.Write(file, m.generatedEntries)
}

func (m PreviewModel) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("Preview .env.example generation")

	if m.confirmed && m.success {
		successMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true).
			Render(fmt.Sprintf("✓ Successfully wrote %s", m.outputPath))

		continueMsg := lipgloss.NewStyle().
			Faint(true).
			Render("Press Enter to continue")

		return "\n" + title + "\n\n" + successMsg + "\n\n" + continueMsg + "\n"
	}

	if m.confirmed {
		confirmMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Render(fmt.Sprintf("Write changes to %s?", m.outputPath))

		help := lipgloss.NewStyle().
			Faint(true).
			Render("Y/Enter: yes • N: no • Q/Esc: cancel")

		return "\n" + title + "\n\n" + confirmMsg + "\n\n" + help + "\n"
	}

	// Show diff
	var diff strings.Builder

	// Calculate visible range
	visibleLines := 10
	start := m.scrollOffset
	end := start + visibleLines
	if end > len(m.diffLines) {
		end = len(m.diffLines)
	}

	for i := start; i < end; i++ {
		line := m.diffLines[i]
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		// Apply colors based on line type
		style := lipgloss.NewStyle()
		if strings.HasPrefix(line, "- ") {
			// Removed line - red
			style = style.Foreground(lipgloss.Color("#FF6B6B"))
		} else if strings.HasPrefix(line, "+ ") {
			// Added line - green
			style = style.Foreground(lipgloss.Color("#00FF00"))
		} else {
			// Unchanged line - green (kept)
			style = style.Foreground(lipgloss.Color("#00FF00"))
		}

		if i == m.cursor {
			style = style.Bold(true).Background(lipgloss.Color("#7D56F4"))
		}

		diff.WriteString(style.Render(cursor+" "+line) + "\n")
	}

	// Scroll indicator
	if len(m.diffLines) > visibleLines {
		scrollInfo := fmt.Sprintf("Line %d/%d", m.cursor+1, len(m.diffLines))
		diff.WriteString(lipgloss.NewStyle().Faint(true).Render(scrollInfo) + "\n")
	}

	help := lipgloss.NewStyle().
		Faint(true).
		Render("↑/k: up • ↓/j: down • Enter: confirm • Q/Esc: cancel")

	return "\n" + title + "\n\n" + diff.String() + "\n" + help + "\n"
}
