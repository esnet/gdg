package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// Screen manages a list of Fields for one wizard step.
//
// Responsibilities:
//   - Tab / Shift-Tab cycle focus through focusable fields (wrapping).
//   - Enter validates the current field and advances focus, or submits the
//     screen when the last focusable field is confirmed.
//   - Esc sets Cancelled = true (the parent model decides what "back" means).
//   - Ctrl+C is intentionally NOT handled here; the parent model intercepts it.
//
// After Update returns, callers check:
//
//	screen.Submitted — all fields valid, phase may advance
//	screen.Cancelled — user pressed Esc, phase should go back
//
// Field values are bound to external pointers at construction time, so the
// parent can read results directly from its state struct without extracting
// them from the Screen.
type Screen struct {
	fields    []Field
	focused   int
	width     int
	Submitted bool
	Cancelled bool
	ErrMsg    string // screen-level error (e.g. from multi-field validation)
}

// NewScreen creates a Screen with the given width and fields.
//
// The first focusable field is identified immediately so that key events are
// routed correctly even before Init() is called. This matters because bubbletea
// v2's Init() returns only a tea.Cmd — model mutations inside Init() are
// discarded — so we cannot rely on Init() to set the focus index.
func NewScreen(width int, fields ...Field) Screen {
	s := Screen{fields: fields, width: width, focused: -1}
	for i, f := range s.fields {
		if f.Focusable() {
			s.focused = i
			break
		}
	}
	return s
}

// Init fires the Focus() command on the already-focused field so that
// textinput components start with their cursor blinking immediately.
// The focus *index* is set in NewScreen rather than here because bubbletea v2
// discards model-level mutations made inside Init().
func (s Screen) Init() (Screen, tea.Cmd) {
	if s.focused >= 0 && s.focused < len(s.fields) {
		return s, s.fields[s.focused].Focus()
	}
	return s, nil
}

// SetWidth updates the available horizontal space (called on WindowSizeMsg).
func (s Screen) SetWidth(w int) Screen {
	s.width = w
	return s
}

// Update handles input for the active screen.
func (s Screen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	key, isKey := msg.(tea.KeyMsg)
	if isKey {
		switch key.String() {
		case "esc":
			s.Cancelled = true
			return s, nil

		case "tab":
			return s.cycleFocus(+1)

		case "shift+tab":
			return s.cycleFocus(-1)

		case "enter":
			return s.handleEnter()
		}
	}

	// Delegate all other input to the focused field.
	if s.focused >= 0 && s.focused < len(s.fields) {
		updated, cmd := s.fields[s.focused].Update(msg)
		s.fields[s.focused] = updated
		return s, cmd
	}
	return s, nil
}

// View renders all fields stacked vertically.
func (s Screen) View() string {
	var sb strings.Builder
	for i, f := range s.fields {
		sb.WriteString(f.View(i == s.focused, s.width))
		sb.WriteByte('\n')
	}
	if s.ErrMsg != "" {
		sb.WriteString(ErrorStyle.Render("  " + GlyphCross + " " + s.ErrMsg))
		sb.WriteByte('\n')
	}
	return sb.String()
}

// ── internal helpers ─────────────────────────────────────────────────────────
// cycleFocus shifts focus by dir (+1 forward, -1 backward), wrapping around.
func (s Screen) cycleFocus(dir int) (Screen, tea.Cmd) {
	n := len(s.fields)
	if n == 0 {
		return s, nil
	}
	start := s.focused
	if start < 0 {
		start = 0
	}
	for i := 1; i <= n; i++ {
		next := ((start+i*dir)%n + n) % n
		if s.fields[next].Focusable() {
			if s.focused >= 0 && s.focused < n {
				s.fields[s.focused].Blur()
			}
			s.focused = next
			s.ErrMsg = ""
			return s, s.fields[next].Focus()
		}
	}
	return s, nil
}

// handleEnter validates the focused field, then either advances focus or submits.
func (s Screen) handleEnter() (Screen, tea.Cmd) {
	s.ErrMsg = ""

	// Validate the current field (skip non-focusable).
	if s.focused >= 0 && s.focused < len(s.fields) && s.fields[s.focused].Focusable() {
		if err := s.fields[s.focused].Validate(); err != nil {
			s.ErrMsg = err.Error()
			// Surface the error inside field types that support inline display.
			switch f := s.fields[s.focused].(type) {
			case *TextField:
				f.SetError(err.Error())
			case *MultiSelectField:
				f.errMsg = err.Error()
			}
			return s, nil
		}
	}

	// Advance to the next focusable field.
	for i := s.focused + 1; i < len(s.fields); i++ {
		if s.fields[i].Focusable() {
			if s.focused >= 0 && s.focused < len(s.fields) {
				s.fields[s.focused].Blur()
			}
			s.focused = i
			return s, s.fields[i].Focus()
		}
	}

	// No further focusable fields — validate everything and submit.
	for _, f := range s.fields {
		if f.Focusable() {
			if err := f.Validate(); err != nil {
				s.ErrMsg = err.Error()
				return s, nil
			}
		}
	}
	s.Submitted = true
	return s, nil
}
