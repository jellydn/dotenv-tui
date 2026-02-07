// Package tui provides Bubble Tea components for the terminal UI.
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jellydn/dotenv-tui/internal/generator"
	"github.com/jellydn/dotenv-tui/internal/parser"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PreviewModel is the Bubble Tea model for previewing .env.example diffs.
type PreviewModel struct {
	originalEntries  []parser.Entry
	generatedEntries []parser.Entry
	diffLines        []string
	cursor           int
	scrollOffset     int
	outputPath       string
	confirmed        bool
	success          bool
	fileIndex        int
	totalFiles       int
	filePath         string
}

// PreviewFinishedMsg signals the preview has completed with success status.
type PreviewFinishedMsg struct {
	Success bool
}

// NewPreviewModel creates a preview for the generated .env.example output.
func NewPreviewModel(filePath string, fileIndex, totalFiles int) tea.Cmd {
	return func() tea.Msg {
		file, err := os.Open(filePath)
		if err != nil {
			return previewInitMsg{
				originalEntries:  []parser.Entry{},
				generatedEntries: []parser.Entry{},
				diffLines:        []string{fmt.Sprintf("Error reading file: %v", err)},
				outputPath:       filepath.Join(filepath.Dir(filePath), ".env.example"),
				fileIndex:        fileIndex,
				totalFiles:       totalFiles,
				filePath:         filePath,
			}
		}
		defer func() { _ = file.Close() }()

		originalEntries, err := parser.Parse(file)
		if err != nil {
			return previewInitMsg{
				originalEntries:  []parser.Entry{},
				generatedEntries: []parser.Entry{},
				diffLines:        []string{fmt.Sprintf("Error parsing file: %v", err)},
				outputPath:       filepath.Join(filepath.Dir(filePath), ".env.example"),
				fileIndex:        fileIndex,
				totalFiles:       totalFiles,
				filePath:         filePath,
			}
		}

		generatedEntries := generator.GenerateExample(originalEntries)

		var diffLines []string
		for i, orig := range originalEntries {
			if i < len(generatedEntries) {
				origLine := parser.EntryToString(orig)
				genLine := parser.EntryToString(generatedEntries[i])

				if origLine == genLine {
					diffLines = append(diffLines, fmt.Sprintf("  %s", origLine))
				} else {
					diffLines = append(diffLines, fmt.Sprintf("  %s [masked]", genLine))
				}
			}
		}

		outputPath := filepath.Join(filepath.Dir(filePath), ".env.example")

		return previewInitMsg{
			originalEntries:  originalEntries,
			generatedEntries: generatedEntries,
			diffLines:        diffLines,
			outputPath:       outputPath,
			fileIndex:        fileIndex,
			totalFiles:       totalFiles,
			filePath:         filePath,
		}
	}
}

type previewInitMsg struct {
	originalEntries  []parser.Entry
	generatedEntries []parser.Entry
	diffLines        []string
	outputPath       string
	fileIndex        int
	totalFiles       int
	filePath         string
}

// Init initializes the preview model.
func (m PreviewModel) Init() tea.Cmd {
	return nil
}

const visibleDiffLines = 10

// adjustScroll ensures the cursor remains visible by adjusting scrollOffset
// when the cursor moves outside the currently visible area.
func (m *PreviewModel) adjustScroll() {
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	} else if m.cursor >= m.scrollOffset+visibleDiffLines {
		m.scrollOffset = m.cursor - visibleDiffLines + 1
	}
}

// Update handles messages and updates the preview model.
func (m PreviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case previewInitMsg:
		m.originalEntries = msg.originalEntries
		m.generatedEntries = msg.generatedEntries
		m.diffLines = msg.diffLines
		m.outputPath = msg.outputPath
		m.fileIndex = msg.fileIndex
		m.totalFiles = msg.totalFiles
		m.filePath = msg.filePath
		return m, nil

	case tea.KeyMsg:
		switch {
		case m.confirmed && m.success:
			switch msg.String() {
			case "enter", "q", "esc":
				return m, func() tea.Msg {
					return PreviewFinishedMsg{Success: true}
				}
			}
		case m.confirmed:
			switch msg.String() {
			case "y", "Y", "enter":
				if err := m.writeFile(); err != nil {
					m.success = false
					return m, nil
				}
				m.success = true
				return m, nil
			case "n", "N", "q", "esc":
				return m, func() tea.Msg {
					return PreviewFinishedMsg{Success: false}
				}
			}
		default:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					m.adjustScroll()
				}
			case "down", "j":
				if m.cursor < len(m.diffLines)-1 {
					m.cursor++
					m.adjustScroll()
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

// writeFile writes the generated entries to the output .env.example file.
func (m PreviewModel) writeFile() error {
	file, err := os.OpenFile(m.outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	return parser.Write(file, m.generatedEntries)
}

// View renders the diff preview UI.
func (m PreviewModel) View() string {
	positionText := fmt.Sprintf("[%d/%d] %s", m.fileIndex+1, m.totalFiles, m.filePath)
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("Preview .env.example generation")

	position := lipgloss.NewStyle().
		Faint(true).
		Render(positionText)

	if m.confirmed && m.success {
		successMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true).
			Render(fmt.Sprintf("✓ Successfully wrote %s", m.outputPath))

		continueMsg := lipgloss.NewStyle().
			Faint(true).
			Render("Press Enter to continue")

		return "\n" + title + "\n" + position + "\n\n" + successMsg + "\n\n" + continueMsg + "\n"
	}

	if m.confirmed {
		confirmMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Render(fmt.Sprintf("Write changes to %s?", m.outputPath))

		help := lipgloss.NewStyle().
			Faint(true).
			Render("Y/Enter: yes • N: no • Q/Esc: cancel")

		return "\n" + title + "\n" + position + "\n\n" + confirmMsg + "\n\n" + help + "\n"
	}

	var diff strings.Builder

	start := m.scrollOffset
	end := start + visibleDiffLines
	if end > len(m.diffLines) {
		end = len(m.diffLines)
	}

	for i := start; i < end; i++ {
		line := m.diffLines[i]
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		// Apply colors: green for unchanged, yellow for masked
		style := lipgloss.NewStyle()
		if strings.Contains(line, "[masked]") {
			style = style.Foreground(lipgloss.Color("#FFFF00")) // Yellow for masked
		} else {
			style = style.Foreground(lipgloss.Color("#00FF00")) // Green for unchanged
		}

		if i == m.cursor {
			style = style.Bold(true).Background(lipgloss.Color("#7D56F4"))
		}

		diff.WriteString(style.Render(cursor+" "+line) + "\n")
	}

	if len(m.diffLines) > visibleDiffLines {
		scrollInfo := fmt.Sprintf("Line %d/%d", m.cursor+1, len(m.diffLines))
		diff.WriteString(lipgloss.NewStyle().Faint(true).Render(scrollInfo) + "\n")
	}

	help := lipgloss.NewStyle().
		Faint(true).
		Render("↑/k: up • ↓/j: down • Enter: confirm • Q/Esc: cancel")

	return "\n" + title + "\n" + position + "\n\n" + diff.String() + "\n" + help + "\n"
}
