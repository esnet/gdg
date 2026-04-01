package tui

import tea "charm.land/bubbletea/v2"

// RunConfirm runs a minimal full-screen program asking a single yes/no question.
// Returns true when the user selects Yes, false for No, Esc, or Ctrl+C.
// Intended for simple interactive confirmations outside of a larger TUI wizard.
func RunConfirm(title, description string) bool {
	var result bool
	scr := NewScreen(60, NewConfirmField(title, description, &result))
	m := confirmRunner{screen: scr}
	prog := tea.NewProgram(m)
	final, err := prog.Run()
	if err != nil {
		return false
	}
	r := final.(confirmRunner)
	return !r.cancelled && result
}

// ── confirmRunner – minimal tea.Model for a single confirm prompt ─────────────

type confirmRunner struct {
	screen    Screen
	width     int
	cancelled bool
}

func (m confirmRunner) Init() tea.Cmd {
	var cmd tea.Cmd
	_, cmd = m.screen.Init()
	return cmd
}

func (m confirmRunner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.screen = m.screen.SetWidth(m.width)
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.cancelled = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.screen, cmd = m.screen.Update(msg)
	if m.screen.Submitted || m.screen.Cancelled {
		m.cancelled = m.screen.Cancelled
		return m, tea.Quit
	}
	return m, cmd
}

func (m confirmRunner) View() tea.View {
	v := tea.NewView("")
	v.Content = "\n" + m.screen.View()
	return v
}
