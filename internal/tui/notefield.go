package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// NoteField is a read-only informational panel.  It is not focusable; the
// Screen skips it when cycling focus.  Pressing Enter on a screen whose only
// fields are NoteFields submits the screen immediately (the user just needs to
// acknowledge the information and move on).
type NoteField struct {
	title string
	body  string
}

// NewNoteField creates an informational display field.
func NewNoteField(title, body string) *NoteField {
	return &NoteField{title: title, body: body}
}

// ── Field interface ───────────────────────────────────────────────────────────

func (f *NoteField) Update(_ tea.Msg) (Field, tea.Cmd) { return f, nil }
func (f *NoteField) Focus() tea.Cmd                    { return nil }
func (f *NoteField) Blur()                             {}
func (f *NoteField) Focusable() bool                   { return false }
func (f *NoteField) Validate() error                   { return nil }

func (f *NoteField) View(_ bool, width int) string {
	var sb strings.Builder

	sb.WriteString(TitleStyle.Render(f.title))
	sb.WriteByte('\n')

	dividerWidth := width - 4
	if dividerWidth < 4 {
		dividerWidth = 4
	}
	divider := lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render(strings.Repeat("─", dividerWidth))
	sb.WriteString("  ")
	sb.WriteString(divider)
	sb.WriteByte('\n')

	bodyStyle := lipgloss.NewStyle().Foreground(ColorText)
	for _, line := range strings.Split(f.body, "\n") {
		sb.WriteString(bodyStyle.Render("  " + line))
		sb.WriteByte('\n')
	}

	return sb.String()
}
