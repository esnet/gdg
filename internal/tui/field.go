package tui

import tea "charm.land/bubbletea/v2"

// Field is the interface implemented by every interactive element in a Screen.
//
// All concrete types (*TextField, *SelectField, etc.) use pointer receivers so
// they satisfy the interface while being mutated in-place inside a Screen's
// field slice – no extra copying overhead per frame.
type Field interface {
	// Update handles keyboard / mouse input when this field is focused.
	// Returns the (possibly mutated) field and any follow-up command.
	Update(msg tea.Msg) (Field, tea.Cmd)

	// View renders the field to a string.  focused controls cursor / highlight
	// visibility.  width is the available horizontal space in terminal columns.
	View(focused bool, width int) string

	// Focus activates the field (shows cursor, starts blink animation, etc.)
	// and returns any initialisation command required.
	Focus() tea.Cmd

	// Blur deactivates the field (hides cursor, dims colours, etc.).
	Blur()

	// Focusable returns false for purely decorative / read-only fields (e.g.
	// NoteField) that the Screen should skip when cycling focus.
	Focusable() bool

	// Validate returns a non-nil error when the field's current value is
	// invalid.  Called by the Screen before advancing focus or submitting.
	Validate() error
}
