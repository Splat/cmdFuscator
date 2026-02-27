package tui

import "github.com/charmbracelet/lipgloss"

// ─── Colour palette ───────────────────────────────────────────────────────────

var (
	clrGreen    = lipgloss.Color("#00FF88")
	clrGold     = lipgloss.Color("#FFD700")
	clrGray     = lipgloss.Color("#888888")
	clrDimGray  = lipgloss.Color("#444444")
	clrWhite    = lipgloss.Color("#DDDDDD")
	clrBorder   = lipgloss.Color("#334455")
	clrFocusBdr = lipgloss.Color("#00FF88")
	clrRed      = lipgloss.Color("#FF4455")
	clrCyan     = lipgloss.Color("#00CCFF")
)

// ─── Panel borders ────────────────────────────────────────────────────────────

func panelStyle(focused bool) lipgloss.Style {
	bdr := clrBorder
	if focused {
		bdr = clrFocusBdr
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(bdr).
		Padding(0, 1)
}

// ─── Text styles ──────────────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrGreen)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(clrGray).
			Italic(true)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrGold)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrGreen)

	normalStyle = lipgloss.NewStyle().
			Foreground(clrWhite)

	dimStyle = lipgloss.NewStyle().
			Foreground(clrDimGray)

	checkedStyle = lipgloss.NewStyle().
			Foreground(clrGreen)

	uncheckedStyle = lipgloss.NewStyle().
			Foreground(clrDimGray)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			Foreground(clrGreen)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(clrGray)

	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrCyan)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(clrGray)

	copyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrGold)

	errorStyle = lipgloss.NewStyle().
			Foreground(clrRed)

	notImplStyle = lipgloss.NewStyle().
			Foreground(clrGold).
			Italic(true)

	rawEscapeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrCyan)
)

// statusBar renders the bottom help line.
func renderStatusBar(width int) string {
	keys := []struct{ key, action string }{
		{"Tab", "Focus"},
		{"↑↓", "Navigate"},
		{"←→", "OS Filter"},
		{"Space", "Toggle"},
		{"Enter", "Apply"},
		{"c", "Copy"},
		{"r", "Reset"},
		{"/", "Search"},
		{"q", "Quit"},
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts, keyStyle.Render(k.key)+" "+statusBarStyle.Render(k.action))
	}

	bar := ""
	for i, p := range parts {
		if i > 0 {
			bar += statusBarStyle.Render("  ·  ")
		}
		bar += p
	}

	return lipgloss.NewStyle().
		Width(width).
		Foreground(clrGray).
		Render(bar)
}
