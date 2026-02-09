// Package tui provides Bubble Tea components for the terminal UI.
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jellydn/dotenv-tui/internal/backup"
	"github.com/jellydn/dotenv-tui/internal/generator"
	"github.com/jellydn/dotenv-tui/internal/parser"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type filePreview struct {
	filePath         string
	outputPath       string
	generatedEntries []parser.Entry
	diffLines        []string
	errMsg           string
}

// PreviewModel is the Bubble Tea model for previewing .env.example diffs.
type PreviewModel struct {
	files        []filePreview
	currentFile  int
	cursor       int
	scrollOffset int
	written      bool
	writeResults []writeResult
	windowHeight int
	enableBackup bool
}

type writeResult struct {
	OutputPath string
	Success    bool
	Error      string
}

// PreviewFinishedMsg signals the preview has completed.
type PreviewFinishedMsg struct {
	Results []writeResult
}

type previewInitMsg struct {
	files        []filePreview
	enableBackup bool
}

// NewPreviewModel creates a preview for multiple files at once.
func NewPreviewModel(filePaths []string, enableBackup bool) tea.Cmd {
	return func() tea.Msg {
		var files []filePreview
		for _, fp := range filePaths {
			files = append(files, loadFilePreview(fp))
		}
		return previewInitMsg{files: files, enableBackup: enableBackup}
	}
}

func loadFilePreview(filePath string) filePreview {
	outputPath := filepath.Join(filepath.Dir(filePath), ".env.example")

	file, err := os.Open(filePath)
	if err != nil {
		return filePreview{
			filePath:   filePath,
			outputPath: outputPath,
			diffLines:  []string{fmt.Sprintf("Error reading file: %v", err)},
			errMsg:     err.Error(),
		}
	}
	defer func() { _ = file.Close() }()

	originalEntries, err := parser.Parse(file)
	if err != nil {
		return filePreview{
			filePath:   filePath,
			outputPath: outputPath,
			diffLines:  []string{fmt.Sprintf("Error parsing file: %v", err)},
			errMsg:     err.Error(),
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

	return filePreview{
		filePath:         filePath,
		outputPath:       outputPath,
		generatedEntries: generatedEntries,
		diffLines:        diffLines,
	}
}

// Init initializes the preview model.
func (m PreviewModel) Init() tea.Cmd {
	return nil
}

// SetWindowHeight sets the terminal height for scroll calculations.
func (m *PreviewModel) SetWindowHeight(h int) {
	m.windowHeight = h
}

const previewOverheadLines = 8 // title + position + 2 newlines + scroll info + help + 2 newlines

func (m PreviewModel) visibleLines() int {
	if m.windowHeight <= previewOverheadLines {
		return 10 // fallback to default if window is too small
	}
	return m.windowHeight - previewOverheadLines
}

func (m *PreviewModel) adjustScroll() {
	visible := m.visibleLines()
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	} else if m.cursor >= m.scrollOffset+visible {
		m.scrollOffset = m.cursor - visible + 1
	}
}

func (m *PreviewModel) switchFile(dir int) {
	n := len(m.files)
	if n <= 1 {
		return
	}
	m.currentFile = (m.currentFile + dir + n) % n
	m.cursor = 0
	m.scrollOffset = 0
}

// Update handles messages and updates the preview model.
func (m PreviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case previewInitMsg:
		m.files = msg.files
		m.currentFile = 0
		m.cursor = 0
		m.scrollOffset = 0
		m.written = false
		m.writeResults = nil
		m.enableBackup = msg.enableBackup
		return m, nil

	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height
		m.adjustScroll()
		return m, nil

	case tea.KeyMsg:
		if len(m.files) == 0 {
			if msg.String() == "q" || msg.String() == "esc" {
				return m, func() tea.Msg { return PreviewFinishedMsg{} }
			}
			return m, nil
		}

		if m.written {
			switch msg.String() {
			case "enter", "q", "esc":
				return m, func() tea.Msg {
					return PreviewFinishedMsg{Results: m.writeResults}
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "tab":
			m.switchFile(1)
		case "shift+tab":
			m.switchFile(-1)
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.adjustScroll()
			}
		case "down", "j":
			f := m.files[m.currentFile]
			if m.cursor < len(f.diffLines)-1 {
				m.cursor++
				m.adjustScroll()
			}
		case "enter":
			m.writeResults = m.writeAllFiles()
			m.written = true
		case "q", "esc":
			return m, func() tea.Msg {
				return PreviewFinishedMsg{}
			}
		}
	}
	return m, nil
}

func (m PreviewModel) writeAllFiles() []writeResult {
	var results []writeResult
	for _, f := range m.files {
		if f.errMsg != "" {
			results = append(results, writeResult{
				OutputPath: f.outputPath,
				Success:    false,
				Error:      f.errMsg,
			})
			continue
		}
		err := m.writePreviewFile(f.outputPath, f.generatedEntries)
		if err != nil {
			results = append(results, writeResult{
				OutputPath: f.outputPath,
				Success:    false,
				Error:      err.Error(),
			})
		} else {
			results = append(results, writeResult{
				OutputPath: f.outputPath,
				Success:    true,
			})
		}
	}
	return results
}

func (m PreviewModel) writePreviewFile(outputPath string, entries []parser.Entry) error {
	if m.enableBackup {
		if _, err := os.Stat(outputPath); err == nil {
			if _, err := backup.CreateBackup(outputPath); err != nil {
				return fmt.Errorf("failed to create backup: %w", err)
			}
		}
	}

	file, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	return parser.Write(file, entries)
}

// View renders the diff preview UI.
func (m PreviewModel) View() string {
	if len(m.files) == 0 {
		return "\nNo files to preview\n"
	}

	if m.written {
		return m.viewWriteResults()
	}

	f := m.files[m.currentFile]

	positionText := fmt.Sprintf("[%d/%d] %s", m.currentFile+1, len(m.files), f.filePath)
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("Preview .env.example generation")

	position := lipgloss.NewStyle().
		Faint(true).
		Render(positionText)

	var diff strings.Builder

	visible := m.visibleLines()
	start := m.scrollOffset
	end := start + visible
	if end > len(f.diffLines) {
		end = len(f.diffLines)
	}

	for i := start; i < end; i++ {
		line := f.diffLines[i]
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		style := lipgloss.NewStyle()
		if strings.Contains(line, "[masked]") {
			style = style.Foreground(lipgloss.Color("#FFFF00"))
		} else {
			style = style.Foreground(lipgloss.Color("#00FF00"))
		}

		if i == m.cursor {
			style = style.Bold(true).Background(lipgloss.Color("#7D56F4"))
		}

		diff.WriteString(style.Render(cursor+" "+line) + "\n")
	}

	if len(f.diffLines) > visible {
		scrollInfo := fmt.Sprintf("Line %d/%d", m.cursor+1, len(f.diffLines))
		diff.WriteString(lipgloss.NewStyle().Faint(true).Render(scrollInfo) + "\n")
	}

	helpParts := []string{"↑/k: up", "↓/j: down"}
	if len(m.files) > 1 {
		helpParts = append(helpParts, "Tab: next file", "Shift+Tab: prev file")
	}
	helpParts = append(helpParts, "Enter: write all", "q/Esc: cancel")

	help := lipgloss.NewStyle().
		Faint(true).
		Render(strings.Join(helpParts, " • "))

	return "\n" + title + "\n" + position + "\n\n" + diff.String() + "\n" + help + "\n"
}

func (m PreviewModel) viewWriteResults() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render(".env.example Generation Complete")

	var lines strings.Builder
	successCount := 0
	for _, r := range m.writeResults {
		if r.Success {
			successCount++
			line := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Render(fmt.Sprintf("  ✓ %s", r.OutputPath))
			lines.WriteString(line + "\n")
		} else {
			line := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF5F56")).
				Render(fmt.Sprintf("  ✗ %s: %s", r.OutputPath, r.Error))
			lines.WriteString(line + "\n")
		}
	}

	summary := fmt.Sprintf("Wrote %d/%d files", successCount, len(m.writeResults))

	help := lipgloss.NewStyle().
		Faint(true).
		Render("Press Enter or q to return to menu")

	return "\n" + title + "\n\n" + lines.String() + "\n" + summary + "\n\n" + help + "\n"
}
