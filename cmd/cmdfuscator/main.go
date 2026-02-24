package main

import (
	"fmt"
	"os"

	"cmdFuscator/cmd/cmdfuscator/tui"
	"cmdFuscator/data"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	model := tui.New(data.ModelFS)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // full-screen TUI
		tea.WithMouseCellMotion(), // enable mouse scrolling in viewport
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "cmdFuscator: %v\n", err)
		os.Exit(1)
	}
}
