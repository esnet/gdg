package tui

import "charm.land/bubbles/v2/key"

// Keys holds all key bindings used by GDG TUI screens.
// It implements help.KeyMap so it can be passed directly to help.Model.View().
type Keys struct {
	Up       key.Binding
	Down     key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Enter    key.Binding
	Space    key.Binding
	Esc      key.Binding
	Quit     key.Binding
}

// DefaultKeys is the standard key map shared by all GDG screens.
var DefaultKeys = Keys{
	Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
	ShiftTab: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev field")),
	Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
	Space:    key.NewBinding(key.WithKeys(" ", "space"), key.WithHelp("space", "toggle")),
	Esc:      key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Quit:     key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
}

// ShortHelp implements help.KeyMap – shown in the compact one-line footer.
func (k Keys) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Enter, k.Esc, k.Quit}
}

// FullHelp implements help.KeyMap – shown when the user expands the footer.
func (k Keys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Tab, k.ShiftTab},
		{k.Enter, k.Space, k.Esc, k.Quit},
	}
}
