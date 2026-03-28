package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// MultiSelectField renders a checklist of options.
//
// Key bindings (writable mode):
//
//	↑ / k   — move cursor up
//	↓ / j   — move cursor down
//	space   — toggle the focused item
//	enter   — confirm selection and advance / submit the screen
//
// In read-only mode the cursor still moves but Space is ignored.  Use
// WithReadOnly() to create a display-only list (e.g. showing which files will
// be affected before the user confirms).
//
// Selected values are synced live to *ptr (slice of option Values).
type MultiSelectField struct {
	title    string
	desc     string
	opts     []Option
	selected map[int]bool
	cursor   int
	ptr      *[]string
	validate func([]string) error
	errMsg   string
	readOnly bool
}

// NewMultiSelectField creates a writable MultiSelectField bound to ptr.
// Options whose Value appears in *ptr are pre-checked.
func NewMultiSelectField(title, desc string, opts []Option, ptr *[]string) *MultiSelectField {
	f := &MultiSelectField{
		title:    title,
		desc:     desc,
		opts:     opts,
		selected: make(map[int]bool),
		ptr:      ptr,
	}
	if ptr != nil {
		for _, v := range *ptr {
			for i, o := range opts {
				if o.Value == v {
					f.selected[i] = true
				}
			}
		}
	}
	return f
}

// WithReadOnly returns a copy of the field configured for display only.
// The cursor can still move for readability, but Space does not toggle items.
// Focusable() still returns true so the screen cycles through it normally.
func (f *MultiSelectField) WithReadOnly() *MultiSelectField {
	f.readOnly = true
	return f
}

// WithSelected pre-selects the options whose values are in the provided slice,
// replacing any prior selection. Useful for setting defaults after construction.
func (f *MultiSelectField) WithSelected(values []string) *MultiSelectField {
	f.selected = make(map[int]bool)
	for _, v := range values {
		for i, o := range f.opts {
			if o.Value == v {
				f.selected[i] = true
			}
		}
	}
	f.syncPtr()
	return f
}

// WithItemSelected pre-selects or deselects a single option by value.
func (f *MultiSelectField) WithItemSelected(value string, sel bool) *MultiSelectField {
	for i, o := range f.opts {
		if o.Value == value {
			f.selected[i] = sel
		}
	}
	f.syncPtr()
	return f
}

// WithValidate attaches a validation function (e.g. "at least one required").
// Validation is skipped in read-only mode.
func (f *MultiSelectField) WithValidate(fn func([]string) error) *MultiSelectField {
	f.validate = fn
	return f
}

func (f *MultiSelectField) syncPtr() {
	if f.ptr == nil {
		return
	}
	vals := make([]string, 0, len(f.selected))
	for i, o := range f.opts {
		if f.selected[i] {
			vals = append(vals, o.Value)
		}
	}
	*f.ptr = vals
}

// ── Field interface ───────────────────────────────────────────────────────────

func (f *MultiSelectField) Focus() tea.Cmd  { return nil }
func (f *MultiSelectField) Blur()           {}
func (f *MultiSelectField) Focusable() bool { return true }

func (f *MultiSelectField) Validate() error {
	if f.readOnly || f.validate == nil {
		return nil
	}
	var vals []string
	if f.ptr != nil {
		vals = *f.ptr
	}
	return f.validate(vals)
}

func (f *MultiSelectField) Update(msg tea.Msg) (Field, tea.Cmd) {
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
	case " ", "space":
		// bubbletea v2 may deliver the spacebar as the rune " " or as the
		// named key "space" depending on terminal/platform. Handle both.
		if !f.readOnly {
			f.selected[f.cursor] = !f.selected[f.cursor]
			f.syncPtr()
			f.errMsg = "" // clear stale validation error on change
		}
	}
	return f, nil
}

func (f *MultiSelectField) View(focused bool, width int) string {
	var sb strings.Builder

	// Title
	sb.WriteString(TitleStyle.Render(f.title))
	sb.WriteByte('\n')

	// Description
	if f.desc != "" {
		for _, line := range strings.Split(f.desc, "\n") {
			sb.WriteString(DescStyle.Render("  " + line))
			sb.WriteByte('\n')
		}
	}

	// Hint line (writable only)
	if !f.readOnly && focused {
		sb.WriteString(DescStyle.Render("  space: toggle  •  enter: confirm"))
		sb.WriteByte('\n')
	}

	// Options
	for i, opt := range f.opts {
		// Checkbox
		var check string
		if f.selected[i] {
			check = FocusStyle.Render("[" + GlyphCheck + "]")
		} else {
			check = BlurStyle.Render("[ ]")
		}

		// Cursor prefix + label colour
		var prefix string
		var labelStyle lipgloss.Style
		if i == f.cursor && focused {
			prefix = FocusStyle.Render(GlyphCursor + " ")
			labelStyle = lipgloss.NewStyle().Foreground(ColorBright)
		} else {
			prefix = "  "
			if f.selected[i] {
				labelStyle = lipgloss.NewStyle().Foreground(ColorText)
			} else {
				labelStyle = BlurStyle
			}
		}

		fmt.Fprintf(&sb, "%s%s %s\n", prefix, check, labelStyle.Render(opt.Label))
	}

	// Validation error
	if f.errMsg != "" {
		sb.WriteString(ErrorStyle.Render("  " + GlyphCross + " " + f.errMsg))
		sb.WriteByte('\n')
	}

	return sb.String()
}
