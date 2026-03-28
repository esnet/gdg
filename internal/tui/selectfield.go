package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Option is a single choice shown in SelectField or MultiSelectField.
type Option struct {
	Label string
	Value string
}

// NewOption is a convenience constructor.
func NewOption(label, value string) Option { return Option{Label: label, Value: value} }

// SelectField renders a single-choice list.  ↑/↓ (or k/j) move the cursor;
// the selected value is synced live to *ptr.
type SelectField struct {
	title  string
	desc   string
	opts   []Option
	cursor int
	ptr    *string
	errMsg string
}

// NewSelectField creates a SelectField bound to ptr.
// The option whose Value matches *ptr is pre-selected; if *ptr is empty the
// first option is selected and *ptr is updated accordingly.
func NewSelectField(title, desc string, opts []Option, ptr *string) *SelectField {
	f := &SelectField{title: title, desc: desc, opts: opts, ptr: ptr}
	if ptr != nil && *ptr != "" {
		for i, o := range opts {
			if o.Value == *ptr {
				f.cursor = i
				return f
			}
		}
	}
	// Default: first option
	if len(opts) > 0 && ptr != nil {
		*ptr = opts[0].Value
	}
	return f
}

// Value returns the currently highlighted option's value.
func (f *SelectField) Value() string {
	if f.cursor < len(f.opts) {
		return f.opts[f.cursor].Value
	}
	return ""
}

// ── Field interface ───────────────────────────────────────────────────────────

func (f *SelectField) Focus() tea.Cmd   { return nil }
func (f *SelectField) Blur()            {}
func (f *SelectField) Focusable() bool  { return true }
func (f *SelectField) Validate() error  { return nil }

func (f *SelectField) Update(msg tea.Msg) (Field, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return f, nil
	}
	switch key.String() {
	case "up", "k":
		if f.cursor > 0 {
			f.cursor--
		}
	case "down", "j":
		if f.cursor < len(f.opts)-1 {
			f.cursor++
		}
	}
	if f.ptr != nil && f.cursor < len(f.opts) {
		*f.ptr = f.opts[f.cursor].Value
	}
	return f, nil
}

func (f *SelectField) View(focused bool, width int) string {
	var sb strings.Builder

	sb.WriteString(TitleStyle.Render(f.title))
	sb.WriteByte('\n')

	if f.desc != "" {
		for _, line := range strings.Split(f.desc, "\n") {
			sb.WriteString(DescStyle.Render("  " + line))
			sb.WriteByte('\n')
		}
	}

	for i, opt := range f.opts {
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
		sb.WriteString(labelStyle.Render(opt.Label))
		sb.WriteByte('\n')
	}

	if f.errMsg != "" {
		sb.WriteString(ErrorStyle.Render("  " + GlyphCross + " " + f.errMsg))
		sb.WriteByte('\n')
	}

	return sb.String()
}
