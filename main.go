// dotenv-tui is a terminal UI tool for managing .env files.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jellydn/dotenv-tui/internal/cli"
	"github.com/jellydn/dotenv-tui/internal/tui"
	"github.com/jellydn/dotenv-tui/internal/upgrade"
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
		if err := cli.GenerateExampleFile(*generateExample, *forceFlag, cli.RealFileSystem{}, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating .env.example: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *generateEnv != "" {
		if err := cli.GenerateEnvFile(*generateEnv, *forceFlag, cli.RealFileSystem{}, os.Stdout); err != nil {
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
		if err := cli.ScanAndList(scanPath, cli.RealDirScanner{}, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *yoloFlag {
		if err := cli.GenerateAllEnvFiles(*forceFlag, cli.RealFileSystem{}, cli.RealDirScanner{}, os.Stdin, os.Stdout); err != nil {
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
    dotenv-tui --generate-env .env.example        # Generate .env from .env.example
    dotenv-tui --scan                             # Scan current directory for .env files
    dotenv-tui --scan ./myproject                 # Scan specific directory
    dotenv-tui --yolo                             # Auto-generate .env from all .env.example files
    dotenv-tui --yolo --force                     # Force overwrite existing .env files
    dotenv-tui --upgrade                          # Upgrade to the latest version
 `)
}
