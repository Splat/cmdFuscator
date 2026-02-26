// Package tui implements the cmdFuscator terminal user interface using
// Bubbletea (event loop), Lipgloss (layout/style), and Bubbles (widgets).
//
// Layout overview:
//
//	┌─ cmdFuscator ──────────────────────────────────────────────────────┐
//	│ subtitle                                                           │
//	├──────────────────┬─────────────────────────────────────────────────┤
//	│ OS tabs          │  Command Input                                  │
//	│ Search: [      ] │  ╭────────────────────────────────────────────╮ │
//	│                  │  │ certutil.exe -urlcache -split -f https://…  │ │
//	│ > certutil.exe   │  ╰────────────────────────────────────────────╯ │
//	│   cmd.exe        │                                                  │
//	│   curl.exe       │  Modifier Options          [Enter] Apply [r]Reset│
//	│   …              │  ╭────────────────────────────────────────────╮ │
//	│                  │  │ [✓] RandomCase    [✓] QuoteInsertion       │ │
//	│                  │  │ [ ] Shorthands    [ ] Sed                  │ │
//	│                  │  ╰────────────────────────────────────────────╯ │
//	│                  │                                                  │
//	│                  │  Output                           [c] Copy      │
//	│                  │  ╭────────────────────────────────────────────╮ │
//	│                  │  │ cErTuTiL.EXe -URLcAcHe …                  │ │
//	│                  │  ╰────────────────────────────────────────────╯ │
//	├──────────────────┴─────────────────────────────────────────────────┤
//	│ Tab Focus  ↑↓ Navigate  Space Toggle  Enter Apply  c Copy  q Quit │
//	└────────────────────────────────────────────────────────────────────┘
package tui

import (
	"fmt"
	"io/fs"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"cmdFuscator/engine"
	"cmdFuscator/loader"
	"cmdFuscator/models"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Panel focus ──────────────────────────────────────────────────────────────

type panel int

const (
	panelSidebar panel = iota
	panelInput
	panelOptions
	panelOutput
	panelCount
)

// ─── OS filter tabs ───────────────────────────────────────────────────────────

type osFilter int

const (
	osAll osFilter = iota
	osWindows
	osLinux
	osMacOS
)

var osLabels = map[osFilter]string{
	osAll:     "All",
	osWindows: "Win",
	osLinux:   "Lin",
	osMacOS:   "Mac",
}

var osPlatforms = map[osFilter]string{
	osWindows: "windows",
	osLinux:   "linux",
	osMacOS:   "macos",
}

// ─── Model ────────────────────────────────────────────────────────────────────

// Model is the root Bubbletea model for cmdFuscator.
type Model struct {
	// layout
	width  int
	height int

	// focus
	focused panel

	// sidebar
	osFilter    osFilter
	searchInput textinput.Model
	searching   bool
	allExes     []*models.ProfileFile // all loaded profiles
	filtered    []*models.ProfileFile // after OS filter + search
	exeCursor   int
	exeOffset   int // scroll offset for sidebar list

	// command input
	cmdInput textinput.Model

	// selected profile
	selected *models.ProfileFile

	// options panel – modifier toggles
	modifiers []engine.ModifierInfo
	modCursor int

	// output
	output     string
	outputView viewport.Model
	copyMsg    string

	// engine
	eng *engine.Engine

	// status / error
	statusMsg string
	lastErr   error
}

// New creates a Model and loads profiles from the provided fs.FS.
// Pass the embedded model FS from main.go.
func New(modelFS fs.FS) Model {
	// Command input widget
	ci := textinput.New()
	ci.Placeholder = "type a command…"
	ci.CharLimit = 512
	ci.Width = 60

	// Search input widget
	si := textinput.New()
	si.Placeholder = "search…"
	si.CharLimit = 64
	si.Width = 20

	// Output viewport
	ov := viewport.New(60, 5)

	m := Model{
		cmdInput:    ci,
		searchInput: si,
		outputView:  ov,
		eng:         engine.New(),
		focused:     panelSidebar,
	}

	// Load profiles from the embedded FS (sub-dir is "models" within the FS)
	sub, err := fs.Sub(modelFS, "models")
	if err != nil {
		m.statusMsg = fmt.Sprintf("load error: %v", err)
		return m
	}

	profiles, err := loader.LoadFS(sub)
	if err != nil {
		m.statusMsg = fmt.Sprintf("load error: %v", err)
		return m
	}

	// Sort alphabetically for a stable list
	sort.Slice(profiles, func(i, j int) bool {
		return strings.ToLower(profiles[i].Name) < strings.ToLower(profiles[j].Name)
	})

	m.allExes = profiles
	m.applyFilter()

	if len(m.filtered) > 0 {
		m.selectExe(0)
	}

	return m
}

// ─── Bubbletea interface ──────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalcSizes()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Propagate to focused widget
	return m.updateFocusedWidget(msg)
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading…"
	}

	sidebar := m.viewSidebar()
	main := m.viewMain()

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, main)
	statusBar := renderStatusBar(m.width)

	return lipgloss.JoinVertical(lipgloss.Left,
		m.viewHeader(),
		body,
		statusBar,
	)
}

// ─── Key handling ─────────────────────────────────────────────────────────────

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global: quit
	if keyStr := msg.String(); keyStr == "ctrl+c" {
		return m, tea.Quit
	}
	if !m.searching && msg.String() == "q" {
		return m, tea.Quit
	}

	// Searching mode captures all input for the search box
	if m.searching {
		return m.handleSearchKey(msg)
	}

	switch {
	case key.Matches(msg, keys.NextPanel):
		m.focused = (m.focused + 1) % panelCount
		m.syncFocusToWidget()

	case key.Matches(msg, keys.PrevPanel):
		m.focused = (m.focused + panelCount - 1) % panelCount
		m.syncFocusToWidget()

	case key.Matches(msg, keys.Search) && m.focused == panelSidebar:
		m.searching = true
		m.searchInput.Focus()

	case key.Matches(msg, keys.Left) && m.focused == panelSidebar:
		m.setOSFilter((m.osFilter + osFilter(len(osPlatforms))) % osFilter(len(osLabels)))

	case key.Matches(msg, keys.Right) && m.focused == panelSidebar:
		m.setOSFilter((m.osFilter + 1) % osFilter(len(osLabels)))

	case key.Matches(msg, keys.Up):
		m.handleUp()

	case key.Matches(msg, keys.Down):
		m.handleDown()

	case key.Matches(msg, keys.Toggle) && m.focused == panelOptions:
		m.toggleModifier()

	case key.Matches(msg, keys.Apply):
		m.applyObfuscation()

	case key.Matches(msg, keys.Copy):
		m.copyOutput()

	case key.Matches(msg, keys.Reset):
		m.output = ""
		m.outputView.SetContent("")
		m.outputView.GotoTop()
		m.copyMsg = ""
		m.lastErr = nil
		m.statusMsg = ""

	default:
		if m.focused == panelInput {
			var cmd tea.Cmd
			m.cmdInput, cmd = m.cmdInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.searching = false
		m.searchInput.Blur()
		m.applyFilter()
		if len(m.filtered) > 0 {
			m.exeCursor = 0
			m.exeOffset = 0
			m.selectExe(0)
		}
		return m, nil
	default:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.applyFilter()
		m.exeCursor = 0
		m.exeOffset = 0
		return m, cmd
	}
}

func (m *Model) handleUp() {
	switch m.focused {
	case panelSidebar:
		if m.exeCursor > 0 {
			m.exeCursor--
			if m.exeCursor < m.exeOffset {
				m.exeOffset--
			}
			m.selectExe(m.exeCursor)
		}
	case panelOptions:
		if m.modCursor > 0 {
			m.modCursor--
		}
	case panelOutput:
		m.outputView.LineUp(1)
	}
}

func (m *Model) handleDown() {
	switch m.focused {
	case panelSidebar:
		if m.exeCursor < len(m.filtered)-1 {
			m.exeCursor++
			visibleRows := m.sidebarListHeight()
			if m.exeCursor >= m.exeOffset+visibleRows {
				m.exeOffset++
			}
			m.selectExe(m.exeCursor)
		}
	case panelOptions:
		if m.modCursor < len(m.modifiers)-1 {
			m.modCursor++
		}
	case panelOutput:
		m.outputView.LineDown(1)
	}
}

func (m *Model) toggleModifier() {
	if m.modCursor >= 0 && m.modCursor < len(m.modifiers) {
		m.modifiers[m.modCursor].Enabled = !m.modifiers[m.modCursor].Enabled
	}
}

func (m *Model) applyObfuscation() {
	if m.selected == nil {
		m.statusMsg = "select an executable first"
		return
	}

	cmd := strings.TrimSpace(m.cmdInput.Value())
	if cmd == "" {
		m.statusMsg = "enter a command"
		return
	}

	enabled := make(map[string]bool)
	for _, mod := range m.modifiers {
		enabled[mod.Name] = mod.Enabled
	}

	result, err := m.eng.Obfuscate(cmd, m.selected, enabled)
	if err != nil {
		m.lastErr = err
		m.statusMsg = "error: " + err.Error()
		return
	}

	m.output = result.Output
	m.outputView.SetContent(result.Output)
	m.outputView.GotoTop()

	// Build status summary
	parts := []string{}
	if len(result.Applied) > 0 {
		parts = append(parts, "applied: "+strings.Join(result.Applied, ", "))
	}
	if len(result.Skipped) > 0 {
		parts = append(parts, notImplStyle.Render("not implemented: "+strings.Join(result.Skipped, ", ")))
	}
	if len(result.Errors) > 0 {
		for name, e := range result.Errors {
			parts = append(parts, errorStyle.Render(name+": "+e.Error()))
		}
	}
	m.statusMsg = strings.Join(parts, "  |  ")
	m.lastErr = nil
}

func (m *Model) copyOutput() {
	if m.output == "" {
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	default:
		m.copyMsg = "(copy not supported on this OS)"
		return
	}

	cmd.Stdin = strings.NewReader(m.output)
	if err := cmd.Run(); err != nil {
		m.copyMsg = errorStyle.Render("copy failed: " + err.Error())
		return
	}
	m.copyMsg = copyStyle.Render("COPIED!")
}

// ─── Profile selection ────────────────────────────────────────────────────────

func (m *Model) selectExe(idx int) {
	if idx < 0 || idx >= len(m.filtered) {
		return
	}
	m.selected = m.filtered[idx]

	// Populate command input with the template from the first profile
	if len(m.selected.Profiles) > 0 {
		m.cmdInput.SetValue(buildTemplateCommand(m.selected.Profiles[0]))
	}

	// Reset modifiers to defaults for this profile
	enabled := engine.DefaultEnabled(m.selected)
	m.modifiers = engine.ModifierSummary(enabled)
	m.modCursor = 0
	m.output = ""
	m.outputView.SetContent("")
	m.outputView.GotoTop()
	m.copyMsg = ""
	m.statusMsg = ""
}

// buildTemplateCommand converts the profile's command template into a string.
func buildTemplateCommand(p models.Profile) string {
	parts := make([]string, 0, len(p.Parameters.Command))
	for _, el := range p.Parameters.Command {
		parts = append(parts, el.StringValue())
	}
	return strings.Join(parts, " ")
}

// ─── Filtering ────────────────────────────────────────────────────────────────

func (m *Model) applyFilter() {
	query := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))
	platform := osPlatforms[m.osFilter] // empty string for osAll

	var out []*models.ProfileFile
	for _, pf := range m.allExes {
		// Platform filter
		if platform != "" {
			match := false
			for _, p := range pf.Profiles {
				if strings.EqualFold(p.Platform, platform) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		// Search filter
		if query != "" && !strings.Contains(strings.ToLower(pf.Name), query) {
			continue
		}
		out = append(out, pf)
	}
	m.filtered = out
}

func (m *Model) setOSFilter(f osFilter) {
	m.osFilter = f
	m.applyFilter()
	m.exeCursor = 0
	m.exeOffset = 0
	if len(m.filtered) > 0 {
		m.selectExe(0)
	}
}

// ─── Widget sync ──────────────────────────────────────────────────────────────

func (m *Model) syncFocusToWidget() {
	if m.focused == panelInput {
		m.cmdInput.Focus()
	} else {
		m.cmdInput.Blur()
	}
}

func (m *Model) updateFocusedWidget(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.focused {
	case panelInput:
		var cmd tea.Cmd
		m.cmdInput, cmd = m.cmdInput.Update(msg)
		return m, cmd
	case panelOutput:
		var cmd tea.Cmd
		m.outputView, cmd = m.outputView.Update(msg)
		return m, cmd
	}
	return m, nil
}

// ─── Layout calculations ──────────────────────────────────────────────────────

const (
	sidebarWidth = 26

	// panelBorderV is the number of lines a panelStyle border adds (top + bottom).
	panelBorderV = 2
	// panelBorderH is the number of columns a panelStyle border+padding adds
	// (left border + left pad + right pad + right border = 1+1+1+1 = 4).
	panelBorderH = 4
)

// bodyHeight is the number of terminal lines available between the header and
// the status bar.  The header is 1 line; the status bar is 1 line.
func (m *Model) bodyHeight() int {
	h := m.height - 2
	if h < 10 {
		return 10
	}
	return h
}

// mainWidth is the pixel-column width of the right-hand main area.
func (m *Model) mainWidth() int {
	w := m.width - sidebarWidth
	if w < 20 {
		return 20
	}
	return w
}

// panelContentWidth is the value to pass to panelStyle.Width() so that a
// sub-panel exactly fills the main area column.
func (m *Model) panelContentWidth() int {
	return m.mainWidth() - panelBorderH
}

// optModifierRows returns how many two-column rows the modifier grid occupies.
func (m *Model) optModifierRows() int {
	rows := (len(m.modifiers) + 1) / 2
	if rows < 1 {
		return 1
	}
	return rows
}

// outputViewHeight calculates the viewport height so the three panels plus
// gaps and status bar fill bodyHeight exactly.
//
// Accounting (lines):
//   cmdBox  = sectionLabel(1) + input(1) + panelBorderV(2)   = 4
//   gap                                                        = 1
//   optBox  = sectionLabel(1) + optRows  + panelBorderV(2)
//   gap                                                        = 1
//   outBox  = sectionLabel(1) + viewH    + panelBorderV(2)
//   gap                                                        = 1
//   status                                                     = 1
func (m *Model) outputViewHeight() int {
	fixed := 4 + 1 + (1+m.optModifierRows()+panelBorderV) + 1 + (1+panelBorderV) + 1 + 1
	h := m.bodyHeight() - fixed
	if h < 2 {
		return 2
	}
	return h
}

// sidebarListHeight returns how many executable-list rows fit inside the
// sidebar border.
//
// Inside the sidebar border: tabs(1) + search(1) + gap(1) = 3 fixed lines.
func (m *Model) sidebarListHeight() int {
	h := m.bodyHeight() - panelBorderV - 3
	if h < 1 {
		return 1
	}
	return h
}

func (m *Model) recalcSizes() {
	pw := m.panelContentWidth()
	if pw < 4 {
		pw = 4
	}
	m.cmdInput.Width = pw - 2
	m.outputView.Width = pw - 2
	m.outputView.Height = m.outputViewHeight()
}

// ─── Views ────────────────────────────────────────────────────────────────────

func (m Model) viewHeader() string {
	title := titleStyle.Render("cmdFuscator")
	sub := subtitleStyle.Render(" TUI port of ArgFuscator.net  •  security research tool")
	return title + sub
}

func (m Model) viewSidebar() string {
	focused := m.focused == panelSidebar

	// OS filter tabs
	tabs := ""
	for _, f := range []osFilter{osAll, osWindows, osLinux, osMacOS} {
		label := osLabels[f]
		if m.osFilter == f {
			tabs += activeTabStyle.Render("[" + label + "]")
		} else {
			tabs += inactiveTabStyle.Render("[" + label + "]")
		}
	}

	// Search input row
	searchRow := dimStyle.Render("/") + " " + m.searchInput.View()

	// Executable list — padded so the border fills the full body height
	listH := m.sidebarListHeight()
	listLines := make([]string, 0, listH)
	end := m.exeOffset + listH
	if end > len(m.filtered) {
		end = len(m.filtered)
	}
	maxNameLen := sidebarWidth - panelBorderH - 2 // 2 for "> " or "  " prefix
	for i := m.exeOffset; i < end; i++ {
		name := m.filtered[i].Name
		if len(name) > maxNameLen {
			name = name[:maxNameLen-1] + "…"
		}
		if i == m.exeCursor {
			listLines = append(listLines, selectedStyle.Render("> "+name))
		} else {
			listLines = append(listLines, normalStyle.Render("  "+name))
		}
	}
	if len(m.filtered) == 0 {
		listLines = append(listLines, dimStyle.Render("  (no results)"))
	}
	for len(listLines) < listH {
		listLines = append(listLines, "")
	}

	inner := lipgloss.JoinVertical(lipgloss.Left,
		tabs,
		searchRow,
		"",
		strings.Join(listLines, "\n"),
	)

	// Height(bodyHeight - panelBorderV) makes the border extend to the bottom
	// of the body, matching the stacked panels on the right.
	style := panelStyle(focused).
		Width(sidebarWidth - panelBorderH).
		Height(m.bodyHeight() - panelBorderV)
	return style.Render(inner)
}

func (m Model) viewMain() string {
	mw := m.mainWidth()
	pw := m.panelContentWidth()

	// ── Command input ─────────────────────────────────────────────────────
	// The section label lives INSIDE the panel border so the whole box
	// (label + input) lights up when this panel is focused.
	cmdFocused := m.focused == panelInput
	cmdInner := lipgloss.JoinVertical(lipgloss.Left,
		sectionStyle.Render("Command"),
		m.cmdInput.View(),
	)
	cmdBox := panelStyle(cmdFocused).Width(pw).Render(cmdInner)

	// ── Modifier options ──────────────────────────────────────────────────
	optFocused := m.focused == panelOptions
	optHeader := lipgloss.NewStyle().MaxWidth(pw).Render(
		sectionStyle.Render("Modifiers") + "  " + dimStyle.Render("[Enter] Apply  [r] Reset"),
	)
	optInner := lipgloss.JoinVertical(lipgloss.Left,
		optHeader,
		renderModifierGrid(m.modifiers, m.modCursor, pw),
	)
	optBox := panelStyle(optFocused).Width(pw).Render(optInner)

	// ── Output ────────────────────────────────────────────────────────────
	outFocused := m.focused == panelOutput
	var outViewStr string
	if m.output == "" {
		outViewStr = dimStyle.Render("(press Enter to apply obfuscation)")
	} else {
		outViewStr = m.outputView.View()
	}
	outInner := lipgloss.JoinVertical(lipgloss.Left,
		sectionStyle.Render("Output")+"  "+m.copyMsg,
		outViewStr,
	)
	outBox := panelStyle(outFocused).Width(pw).Render(outInner)

	// ── Status message ────────────────────────────────────────────────────
	statusStr := ""
	if m.statusMsg != "" {
		statusStr = m.statusMsg
	} else if m.selected != nil {
		statusStr = dimStyle.Render(fmt.Sprintf("%s  •  %d profile(s)", m.selected.Name, len(m.selected.Profiles)))
	}
	status := lipgloss.NewStyle().MaxWidth(mw).Render(statusStr)

	mainContent := lipgloss.JoinVertical(lipgloss.Left,
		cmdBox,
		"",
		optBox,
		"",
		outBox,
		"",
		status,
	)

	return lipgloss.NewStyle().Width(mw).Render(mainContent)
}

// renderModifierGrid lays out modifier checkboxes in two columns.
func renderModifierGrid(mods []engine.ModifierInfo, cursor, width int) string {
	if len(mods) == 0 {
		return dimStyle.Render("(no modifiers for this profile)")
	}

	colW := width / 2
	var lines []string

	for i := 0; i < len(mods); i += 2 {
		left := renderModifierItem(mods[i], i == cursor, colW)
		right := ""
		if i+1 < len(mods) {
			right = renderModifierItem(mods[i+1], i+1 == cursor, colW)
		}
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, left, right))
	}

	return strings.Join(lines, "\n")
}

func renderModifierItem(info engine.ModifierInfo, selected bool, width int) string {
	checkbox := uncheckedStyle.Render("[ ]")
	label := dimStyle.Render(info.Name)
	if info.Enabled {
		checkbox = checkedStyle.Render("[✓]")
		label = normalStyle.Render(info.Name)
	}
	item := checkbox + " " + label
	if selected {
		item = selectedStyle.Render("> ") + item
	} else {
		item = "  " + item
	}
	return lipgloss.NewStyle().Width(width).Render(item)
}
