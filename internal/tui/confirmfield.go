package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ConfirmField is a yes/no two-option select bound to a *bool.
// ↑/← or y sets Yes; ↓/→ or n sets No.
type ConfirmField struct {
	title  string
	desc   string
	cursor int // 0 = Yes, 1 = No
	ptr    *bool
}

// NewConfirmField creates a ConfirmField bound to ptr.
// The initial selection reflects *ptr: true → Yes (cursor 0), false → No (cursor 1).
func NewConfirmField(title, desc string, ptr *bool) *ConfirmField {
	f := &ConfirmField{title: title, desc: desc, ptr: ptr}
	if ptr != nil && !*ptr {
		f.cursor = 1
	}
	return f
}

// Value returns the current boolean selection.
func (f *ConfirmField) Value() bool { return f.cursor == 0 }

// ── Field interface ───────────────────────────────────────────────────────────

func (f *ConfirmField) Focus() tea.Cmd  { return nil }
func (f *ConfirmField) Blur()           {}
func (f *ConfirmField) Focusable() bool { return true }
func (f *ConfirmField) Validate() error { return nil }

func (f *ConfirmField) Update(msg tea.Msg) (Field, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return f, nil
	}
	switch key.String() {
	case "up", "k", "left", "h", "y", "Y":
		f.cursor = 0
	case "down", "j", "right", "l", "n", "N":
		f.cursor = 1
	}
	if f.ptr != nil {
		*f.ptr = (f.cursor == 0)
	}
	return f, nil
}

func (f *ConfirmField) View(focused bool, width int) string {
	var sb strings.Builder

	sb.WriteString(TitleStyle.Render(f.title))
	sb.WriteByte('\n')

	if f.desc != "" {
		for _, line := range strings.Split(f.desc, "\n") {
			sb.WriteString(DescStyle.Render("  " + line))
			sb.WriteByte('\n')
		}
	}

	labels := []string{"Yes", "No"}
	for i, label := range labels {
		var prefix string
		var labelStyle lipgloss.Style
		if i == f.cursor {
			if focused {
				prefix = FocusStyle.Render(GlyphCursor + " ")
			} else {
				prefix = "  "
			}
			labelStyle = lipgloss.NewStyle().Foreground(ColorBright)
		} else {
			prefix = "  "
			labelStyle = BlurStyle
		}
		sb.WriteString(prefix)
		sb.WriteString(labelStyle.Render(label))
		sb.WriteByte('\n')
	}

	return sb.String()
}
