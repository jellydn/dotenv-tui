package tui

import "github.com/charmbracelet/lipgloss"

func Logo() string {
	teal := lipgloss.Color("#0F766E")
	green := lipgloss.Color("#34D399")
	blue := lipgloss.Color("#93C5FD")

	frame := lipgloss.NewStyle().
		Foreground(teal)

	envLabel := lipgloss.NewStyle().
		Foreground(green).
		Bold(true)

	lockShackle := lipgloss.NewStyle().
		Foreground(blue)

	lockBody := lipgloss.NewStyle().
		Foreground(green)

	lines := []string{
		frame.Render("  ╭──────────────────╮"),
		frame.Render("  │") + "  " + envLabel.Render(".env") + "            " + frame.Render("│"),
		frame.Render("  │") + "                  " + frame.Render("│"),
		frame.Render("  │") + "      " + lockShackle.Render("┌──────┐") + "    " + frame.Render("│"),
		frame.Render("  │") + "      " + lockShackle.Render("│") + "      " + lockShackle.Render("│") + "    " + frame.Render("│"),
		frame.Render("  │") + "    " + lockShackle.Render("┌─┘") + "      " + lockShackle.Render("└─┐") + "  " + frame.Render("│"),
		frame.Render("  │") + "    " + lockBody.Render("│") + "    " + lockBody.Render("◆") + "     " + lockBody.Render("│") + "  " + frame.Render("│"),
		frame.Render("  │") + "    " + lockBody.Render("│") + "    " + lockBody.Render("│") + "     " + lockBody.Render("│") + "  " + frame.Render("│"),
		frame.Render("  │") + "    " + lockBody.Render("└──────────┘") + "  " + frame.Render("│"),
		frame.Render("  ╰──────────────────╯"),
	}

	var logo string
	for _, line := range lines {
		logo += line + "\n"
	}

	return logo
}

func Wordmark() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#E5E7EB")).
		Render("dotenv-tui")

	tagline := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#93A3B8")).
		Italic(true).
		Render("secure .env workflows in your terminal")

	return title + "\n" + tagline
}
