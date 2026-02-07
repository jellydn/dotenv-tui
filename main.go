package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jellydn/dotenv-tui/internal/generator"
	"github.com/jellydn/dotenv-tui/internal/parser"
	"github.com/jellydn/dotenv-tui/internal/scanner"
	"github.com/jellydn/dotenv-tui/internal/tui"

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
	var (
		generateExample = flag.String("generate-example", "", "Generate .env.example from specified .env file")
		generateEnv     = flag.String("generate-env", "", "Generate .env from specified .env.example file")
		showHelp        = flag.Bool("help", false, "Show help information")
		scanFlag        = flag.Bool("scan", false, "Scan directory for .env files")
		forceFlag       = flag.Bool("force", false, "Force overwrite existing files")
	)

	flag.Parse()

	if *showHelp {
		showUsage()
		return
	}

	if *generateExample != "" {
		if err := generateExampleFile(*generateExample, *forceFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating .env.example: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *generateEnv != "" {
		if err := generateEnvFile(*generateEnv, *forceFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating .env: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *scanFlag {
		// Check if next argument is a directory path
		args := flag.Args()
		scanPath := "."
		if len(args) > 0 {
			scanPath = args[0]
		}
		if err := scanAndList(scanPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// No flags provided, launch TUI
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Printf(`dotenv-tui - A terminal UI tool for managing .env files

USAGE:
    dotenv-tui [FLAGS]

FLAGS:
    --generate-example <path>    Generate .env.example from specified .env file
    --generate-env <path>        Generate .env from specified .env.example file
    --scan [directory]           List discovered .env files (default: current directory)
    --force                      Force overwrite existing files
    --help                       Show this help message

EXAMPLES:
    dotenv-tui                                    # Launch interactive TUI
    dotenv-tui --generate-example .env            # Generate .env.example from .env
    dotenv-tui --generate-env .env.example       # Generate .env from .env.example
    dotenv-tui --scan                             # Scan current directory for .env files
    dotenv-tui --scan ./myproject                 # Scan specific directory
`)
}

func generateExampleFile(inputPath string, force bool) error {
	// Read the input .env file
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer func() { _ = file.Close() }()

	entries, err := parser.Parse(file)
	if err != nil {
		return fmt.Errorf("failed to parse .env file: %w", err)
	}

	// Generate example entries
	exampleEntries := generator.GenerateExample(entries)

	// Determine output path (always output to .env.example in the same directory)
	outputPath := filepath.Join(filepath.Dir(inputPath), ".env.example")

	// Check if file exists and handle overwrite
	if _, err := os.Stat(outputPath); err == nil && !force {
		return fmt.Errorf("%s already exists. Use --force to overwrite", outputPath)
	}

	// Write to output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	if err := parser.Write(outFile, exampleEntries); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("Generated %s\n", outputPath)
	return nil
}

func generateEnvFile(inputPath string, force bool) error {
	// Read the input .env.example file
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer func() { _ = file.Close() }()

	entries, err := parser.Parse(file)
	if err != nil {
		return fmt.Errorf("failed to parse .env.example file: %w", err)
	}

	// Determine output path
	outputPath := filepath.Join(filepath.Dir(inputPath), ".env")

	// Check if file exists and handle overwrite
	if _, err := os.Stat(outputPath); err == nil && !force {
		return fmt.Errorf("%s already exists. Use --force to overwrite", outputPath)
	}

	// Write to output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	if err := parser.Write(outFile, entries); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("Generated %s\n", outputPath)
	return nil
}

func scanAndList(dir string) error {
	if dir == "" {
		dir = "."
	}

	files, err := scanner.Scan(dir)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No .env files found")
		return nil
	}

	fmt.Printf("Found %d .env file(s):\n", len(files))
	for _, file := range files {
		fmt.Printf("  %s\n", file)
	}

	return nil
}
