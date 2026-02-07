// dotenv-tui is a terminal UI tool for managing .env files.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/jellydn/dotenv-tui/internal/generator"
	"github.com/jellydn/dotenv-tui/internal/parser"
	"github.com/jellydn/dotenv-tui/internal/scanner"
	"github.com/jellydn/dotenv-tui/internal/tui"
	"github.com/jellydn/dotenv-tui/internal/upgrade"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var Version = ""

// getVersion returns the version of the application.
// It checks the Version variable first, then falls back to build info.
func getVersion() string {
	if Version != "" {
		return Version
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}

	return "dev"
}

type model struct {
	currentScreen screen
	menu          tui.MenuModel
	picker        tui.PickerModel
	preview       tui.PreviewModel
	form          tui.FormModel
	fileList      []string
	fileIndex     int
	pickerMode    tui.MenuChoice
	windowHeight  int
}

type screen int

const (
	menuScreen screen = iota
	pickerScreen
	previewScreen
	formScreen
	doneScreen
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
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.windowHeight = wsm.Height
	}

	switch m.currentScreen {
	case menuScreen:
		return updateMenu(msg, m)
	case pickerScreen:
		return updatePicker(msg, m)
	case previewScreen:
		return updatePreview(msg, m)
	case formScreen:
		return updateForm(msg, m)
	case doneScreen:
		return updateDone(msg, m)
	}
	return m, nil
}

func updateMenu(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	menuModel, menuCmd := m.menu.Update(msg)
	m.menu = menuModel.(tui.MenuModel)
	cmd := menuCmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "enter" || keyMsg.String() == " " {
			m.currentScreen = pickerScreen
			m.picker.SetWindowHeight(m.windowHeight)
			return m, tui.NewPickerModel(m.menu.Choice(), ".")
		}
	}

	return m, cmd
}

func updatePicker(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	pickerModel, pickerCmd := m.picker.Update(msg)
	m.picker = pickerModel.(tui.PickerModel)
	cmd := pickerCmd

	switch msg := msg.(type) {
	case tui.PickerFinishedMsg:
		if len(msg.Selected) > 0 {
			m.fileList = msg.Selected
			m.fileIndex = 0
			m.pickerMode = msg.Mode

			if msg.Mode == tui.GenerateExample {
				m.currentScreen = previewScreen
				return m, tui.NewPreviewModel(msg.Selected[0], 0, len(msg.Selected))
			}
			if msg.Mode == tui.GenerateEnv {
				m.currentScreen = formScreen
				return m, tui.NewFormModel(msg.Selected[0], 0, len(msg.Selected))
			}
		}
		m.currentScreen = menuScreen
		m.menu = tui.NewMenuModel()
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "esc" {
			return returnToMenu(m), nil
		}
	}

	return m, cmd
}

func updatePreview(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	previewModel, previewCmd := m.preview.Update(msg)
	m.preview = previewModel.(tui.PreviewModel)
	cmd := previewCmd

	switch msg := msg.(type) {
	case tui.PreviewFinishedMsg:
		m.currentScreen = doneScreen
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "esc" {
			return returnToMenu(m), nil
		}
	}

	return m, cmd
}

func updateForm(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	formModel, formCmd := m.form.Update(msg)
	m.form = formModel.(tui.FormModel)
	cmd := formCmd

	switch msg := msg.(type) {
	case tui.FormFinishedMsg:
		m.currentScreen = doneScreen
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "esc" {
			return returnToMenu(m), nil
		}
	}

	return m, cmd
}

// navigateToFile transitions to the appropriate screen (preview or form)
// for the current file index in the file list.
func (m model) navigateToFile() (tea.Model, tea.Cmd) {
	if m.pickerMode == tui.GenerateExample {
		m.currentScreen = previewScreen
		return m, tui.NewPreviewModel(m.fileList[m.fileIndex], m.fileIndex, len(m.fileList))
	}
	m.currentScreen = formScreen
	return m, tui.NewFormModel(m.fileList[m.fileIndex], m.fileIndex, len(m.fileList))
}

func returnToMenu(m model) tea.Model {
	m.currentScreen = menuScreen
	m.menu = tui.NewMenuModel()
	return m
}

// updateDone handles messages for the done/completion screen.
// It supports Tab/Shift+Tab navigation between files.
func updateDone(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			if len(m.fileList) > 1 {
				m.fileIndex = (m.fileIndex + 1) % len(m.fileList)
				return m.navigateToFile()
			}
		case "shift+tab":
			if len(m.fileList) > 1 {
				m.fileIndex = (m.fileIndex - 1 + len(m.fileList)) % len(m.fileList)
				return m.navigateToFile()
			}
		case "q", "esc":
			return returnToMenu(m), nil
		}
	}
	return m, nil
}

func (m model) viewDone() string {
	var title string
	if m.pickerMode == tui.GenerateExample {
		title = ".env.example Generation Complete"
	} else {
		title = ".env Generation Complete"
	}

	currentFile := ""
	if len(m.fileList) > 0 {
		currentFile = m.fileList[m.fileIndex]
	}

	status := fmt.Sprintf("Processed: %s [%d/%d]", currentFile, m.fileIndex+1, len(m.fileList))
	help := "q: back to menu"
	if len(m.fileList) > 1 {
		help = "Tab: next file • Shift+Tab: previous file • q: back to menu"
	}

	return fmt.Sprintf(
		"\n%s\n\n%s\n\n%s\n",
		lipgloss.NewStyle().Bold(true).Render(title),
		status,
		lipgloss.NewStyle().Faint(true).Render(help),
	)
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
	case doneScreen:
		return m.viewDone()
	default:
		return ""
	}
}

func main() {
	var (
		generateExample = flag.String("generate-example", "", "Generate .env.example from specified .env file")
		generateEnv     = flag.String("generate-env", "", "Generate .env from specified .env.example file")
		showHelp        = flag.Bool("help", false, "Show help information")
		showVersion     = flag.Bool("version", false, "Show version information")
		scanFlag        = flag.Bool("scan", false, "Scan directory for .env files")
		yoloFlag        = flag.Bool("yolo", false, "Auto-generate .env from all .env.example files")
		forceFlag       = flag.Bool("force", false, "Force overwrite existing files")
		upgradeFlag     = flag.Bool("upgrade", false, "Upgrade to the latest version")
	)

	flag.Parse()

	if *showVersion {
		fmt.Printf("dotenv-tui version %s\n", getVersion())
		return
	}

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

	if *yoloFlag {
		if err := generateAllEnvFiles(*forceFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *upgradeFlag {
		if err := upgrade.Upgrade(getVersion()); err != nil {
			fmt.Fprintf(os.Stderr, "Error upgrading: %v\n", err)
			os.Exit(1)
		}
		return
	}

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
    --yolo                       Auto-generate .env from all .env.example files
    --force                      Force overwrite existing files
    --upgrade                    Upgrade to the latest version
    --version                    Show version information
    --help                       Show this help message

EXAMPLES:
    dotenv-tui                                    # Launch interactive TUI
    dotenv-tui --generate-example .env            # Generate .env.example from .env
    dotenv-tui --generate-env .env.example       # Generate .env from .env.example
    dotenv-tui --scan                             # Scan current directory for .env files
    dotenv-tui --scan ./myproject                 # Scan specific directory
    dotenv-tui --upgrade                          # Upgrade to the latest version
`)
}

type entryProcessor func([]parser.Entry) []parser.Entry

func generateFile(inputPath string, force bool, outputFilename string, processEntries entryProcessor, parseErrMsg string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer func() { _ = file.Close() }()

	entries, err := parser.Parse(file)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", parseErrMsg, err)
	}

	processedEntries := processEntries(entries)

	outputPath := filepath.Join(filepath.Dir(inputPath), outputFilename)

	if _, err := os.Stat(outputPath); err == nil && !force {
		return fmt.Errorf("%s already exists. Use --force to overwrite", outputPath)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	if err := parser.Write(outFile, processedEntries); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("Generated %s\n", outputPath)
	return nil
}

func generateExampleFile(inputPath string, force bool) error {
	return generateFile(inputPath, force, ".env.example", generator.GenerateExample, ".env file")
}

func generateEnvFile(inputPath string, force bool) error {
	return generateFile(inputPath, force, ".env", func(entries []parser.Entry) []parser.Entry {
		return entries
	}, ".env.example file")
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

func generateAllEnvFiles(force bool) error {
	exampleFiles, err := scanner.ScanExamples(".")
	if err != nil {
		return fmt.Errorf("failed to scan for .env.example files: %w", err)
	}

	if len(exampleFiles) == 0 {
		fmt.Fprintf(os.Stderr, "No .env.example files found\n")
		os.Exit(1)
	}

	fmt.Printf("Found %d .env.example file(s):\n", len(exampleFiles))
	for _, file := range exampleFiles {
		fmt.Printf("  %s\n", file)
	}

	var generated, skipped int
	for _, exampleFile := range exampleFiles {
		if err := processExampleFile(exampleFile, force, &generated, &skipped); err != nil {
			return err
		}
	}

	fmt.Printf("Done: %d generated, %d skipped\n", generated, skipped)
	return nil
}

func processExampleFile(exampleFile string, force bool, generated, skipped *int) error {
	outputPath := strings.TrimSuffix(exampleFile, ".example")

	file, err := os.Open(exampleFile)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", exampleFile, err)
	}
	defer func() { _ = file.Close() }()

	entries, err := parser.Parse(file)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", exampleFile, err)
	}

	if _, err := os.Stat(outputPath); err == nil && !force {
		fmt.Printf("%s already exists. Overwrite? [y/N] ", outputPath)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		if response != "y" && response != "Y" {
			fmt.Printf("Skipped %s\n", outputPath)
			*skipped++
			return nil
		}
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", outputPath, err)
	}
	defer func() { _ = outFile.Close() }()

	if err := parser.Write(outFile, entries); err != nil {
		return fmt.Errorf("failed to write %s: %w", outputPath, err)
	}

	fmt.Printf("Generated %s\n", outputPath)
	*generated++
	return nil
}
