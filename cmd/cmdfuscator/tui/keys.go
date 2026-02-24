package tui

import "github.com/charmbracelet/bubbles/key"

// keyMap defines all key bindings used by the TUI.
type keyMap struct {
	NextPanel  key.Binding
	PrevPanel  key.Binding
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
	Toggle     key.Binding
	Apply      key.Binding
	Copy       key.Binding
	Reset      key.Binding
	Search     key.Binding
	Escape     key.Binding
	Quit       key.Binding
}

// keys is the global keyMap used throughout the TUI.
var keys = keyMap{
	NextPanel: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("Tab", "next panel"),
	),
	PrevPanel: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("Shift+Tab", "prev panel"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "prev OS filter"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "next OS filter"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Toggle: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("Space", "toggle modifier"),
	),
	Apply: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("Enter", "apply obfuscation"),
	),
	Copy: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy output"),
	),
	Reset: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reset output"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "cancel"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
