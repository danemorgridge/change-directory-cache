// Package picker provides an interactive, arrow-key directory selector. It
// renders to stderr so that stdout stays clean for the chosen path (which the
// shell wrapper captures to perform the cd).
package picker

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/danemorgridge/change-directory-cache/internal/cache"
)

var (
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	pathStyle     = lipgloss.NewStyle().Faint(true)
	helpStyle     = lipgloss.NewStyle().Faint(true)
)

type model struct {
	entries  []cache.Entry
	cursor   int
	chosen   int
	canceled bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c", "esc", "q":
			m.canceled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
		case "home", "g":
			m.cursor = 0
		case "end", "G":
			m.cursor = len(m.entries) - 1
		case "enter":
			m.chosen = m.cursor
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.canceled || m.chosen >= 0 {
		return "" // clear the UI once we're done
	}
	var b strings.Builder
	b.WriteString("Select a directory:\n\n")
	for i, e := range m.entries {
		if i == m.cursor {
			b.WriteString(cursorStyle.Render("> ") +
				selectedStyle.Render(e.Name) + "  " + pathStyle.Render(e.Path) + "\n")
		} else {
			b.WriteString("  " + e.Name + "  " + pathStyle.Render(e.Path) + "\n")
		}
	}
	b.WriteString("\n" + helpStyle.Render("↑/↓ move · enter select · esc cancel"))
	return b.String()
}

// Pick shows the interactive picker on stderr and returns the chosen entry.
// ok is false if the user cancels.
func Pick(entries []cache.Entry) (entry cache.Entry, ok bool, err error) {
	m := model{entries: entries, chosen: -1}
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	res, err := p.Run()
	if err != nil {
		return cache.Entry{}, false, err
	}
	final := res.(model)
	if final.canceled || final.chosen < 0 {
		return cache.Entry{}, false, nil
	}
	return final.entries[final.chosen], true, nil
}
