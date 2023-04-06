package cmd

import (
	"github.com/muesli/reflow/indent"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	firstChoice   int
	firstChosen   bool
	firstOptions  []string
	secondChoice  int
	secondChosen  bool
	secondOptions []string
	quitting      bool
}

func initialModel() model {
	return model{
		firstChoice:   0,
		firstChosen:   false,
		firstOptions:  []string{"Context", "Dashboard", "Datasources", "Developer Tooling", "Folders", "Library", "Organizations", "User", "Version"},
		secondChoice:  0,
		secondChosen:  false,
		secondOptions: []string{},
		quitting:      false,
	}
}

// Main Init Function
func (m model) Init() tea.Cmd {
	return nil
}

// Main Update Function (calls sub-update function handlers)
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	// Hand off to appropriate update handler if the above don't occur
	if !m.firstChosen {
		return updateFirstChoice(msg, m)
	} else if !m.secondChosen {
		return updateSecondChoice(msg, m)
	} else {
		return processInformation(msg, m)
	}
}

// Main View Function (calls sub-view function handlers)
func (m model) View() string {
	var s string
	if m.quitting {
		return "\n  See you later!\n\n"
	} else if !m.firstChosen {
		return viewFirstChoice(m)
	} else if !m.secondChosen {
		return viewSecondChoice(m)
	} else {
		return viewInformation(m)
	}
	return indent.String("\n"+s+"\n\n", 2)
}

// Update sub-handlers

func updateFirstChoice(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			m.firstChoice = m.firstChoice + 1
			if m.firstChoice >= len(m.firstOptions) {
				m.firstChoice = len(m.firstOptions) - 1
			}
		case "up":
			m.firstChoice = m.firstChoice - 1
			if m.firstChoice < 0 {
				m.firstChoice = 0
			}
		case "enter":
			m.firstChosen = true
			return m, nil
		}
	}
	return m, nil
}

func viewFirstChoice(m model) string {

}

// Second set of menus: handlers below
