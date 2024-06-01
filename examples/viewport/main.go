package main

// An example program demonstrating the viewport component

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/bubbleo/viewport"
)

// RenderableString is a simple type that wraps a string and implements the Renderable interface
type RenderableString struct {
	content string
}

func (r RenderableString) Render() string {
	return r.content
}

type model struct {
	lines    []RenderableString
	ready    bool
	viewport viewport.Model[RenderableString]
	header   []string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return m, tea.Quit
		}
		if k := msg.String(); k == "w" {
			m.viewport.SetWrapText(!m.viewport.GetWrapText())
		}
		if k := msg.String(); k == "s" {
			m.viewport.SetSelectionEnabled(!m.viewport.GetSelectionEnabled())
		}

	case tea.WindowSizeMsg:
		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New[RenderableString](msg.Width-2, msg.Height-5-2)
			m.viewport.SetContent(m.lines)
			m.viewport.SetSelectionEnabled(false)
			m.viewport.SetStringToHighlight("surf")
			m.viewport.SetWrapText(true)
			m.ready = true
		} else {
			m.viewport.SetWidth(msg.Width - 2)
			m.viewport.SetHeight(msg.Height - 5 - 2)
		}
	}

	// Handle keyboard events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	var header = strings.Join(getHeader(
		m.viewport.GetWrapText(),
		m.viewport.GetSelectionEnabled(),
		[]key.Binding{
			m.viewport.KeyMap.PageDown,
			m.viewport.KeyMap.PageUp,
			m.viewport.KeyMap.HalfPageUp,
			m.viewport.KeyMap.HalfPageDown,
			m.viewport.KeyMap.Up,
			m.viewport.KeyMap.Down,
			m.viewport.KeyMap.Left,
			m.viewport.KeyMap.Right,
			m.viewport.KeyMap.Top,
			m.viewport.KeyMap.Bottom,
		},
	), "\n")
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(m.viewport.View()),
	)
}

func getHeader(wrapped, selectionEnabled bool, bindings []key.Binding) []string {
	var header []string
	header = append(header, lipgloss.NewStyle().Bold(true).Render("A Supercharged Viewport"))
	header = append(header, "- Wrapping enabled: "+fmt.Sprint(wrapped)+" (w to toggle)")
	header = append(header, "- Selection enabled: "+fmt.Sprint(selectionEnabled)+" (s to toggle)")
	header = append(header, "- Text to highlight: 'surf'")
	header = append(header, getShortHelp(bindings))
	return header
}

func getShortHelp(bindings []key.Binding) string {
	var output string
	for _, km := range bindings {
		output += km.Help().Key + " " + km.Help().Desc + "  "
	}
	output = strings.TrimSpace(output)
	return output
}

func main() {
	// Load some text for our viewport
	content, err := os.ReadFile("example.txt")
	if err != nil {
		fmt.Println("could not load file:", err)
		os.Exit(1)
	}

	lines := strings.Split(string(content), "\n")
	renderableLines := make([]RenderableString, len(lines))
	for i, line := range lines {
		renderableLines[i] = RenderableString{content: line}
	}

	p := tea.NewProgram(
		model{lines: renderableLines},
		tea.WithAltScreen(), // use the full size of the terminal in its "alternate screen buffer"
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
