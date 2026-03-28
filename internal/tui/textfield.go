package tui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// TextField wraps a bubbles textinput.Model, binding its live value to an
// external *string pointer so callers can read the result without an extra
// step after the screen is submitted.
type TextField struct {
	title    string
	desc     string
	ti       textinput.Model
	ptr      *string
	validate func(string) error
	errMsg   string
}

// NewTextField creates a TextField with title, description, and a bound pointer.
// If ptr already holds a non-empty string the input is pre-filled.
func NewTextField(title, desc string, ptr *string) *TextField {
	ti := textinput.New()
	ti.Prompt = ""
	if ptr != nil && *ptr != "" {
		ti.SetValue(*ptr)
	}
	return &TextField{title: title, desc: desc, ti: ti, ptr: ptr}
}

// WithMask enables password masking (dots instead of characters).
func (f *TextField) WithMask() *TextField {
	f.ti.EchoMode = textinput.EchoPassword
	f.ti.EchoCharacter = '•'
	return f
}

// WithValidate attaches a validation function called on Enter / submit.
func (f *TextField) WithValidate(fn func(string) error) *TextField {
	f.validate = fn
	return f
}

// WithPlaceholder sets greyed-out placeholder text shown when the field is empty.
func (f *TextField) WithPlaceholder(p string) *TextField {
	f.ti.Placeholder = p
	return f
}

// Value returns the current text in the input.
func (f *TextField) Value() string { return f.ti.Value() }

// SetError displays an inline error message (used by Screen on validation failure).
func (f *TextField) SetError(msg string) { f.errMsg = msg }

// ── Field interface ───────────────────────────────────────────────────────────

func (f *TextField) Focus() tea.Cmd {
	f.errMsg = ""
	return f.ti.Focus()
}

func (f *TextField) Blur() { f.ti.Blur() }

func (f *TextField) Focusable() bool { return true }

func (f *TextField) Validate() error {
	if f.validate != nil {
		return f.validate(f.ti.Value())
	}
	return nil
}

func (f *TextField) Update(msg tea.Msg) (Field, tea.Cmd) {
	var cmd tea.Cmd
	f.ti, cmd = f.ti.Update(msg)
	if f.ptr != nil {
		*f.ptr = f.ti.Value()
	}
	return f, cmd
}

func (f *TextField) View(focused bool, width int) string {
	inputWidth := width - 6
	if inputWidth < 10 {
		inputWidth = 10
	}
	f.ti.SetWidth(inputWidth)

	var sb strings.Builder

	// Title
	sb.WriteString(TitleStyle.Render(f.title))
	sb.WriteByte('\n')

	// Description (multi-line, indented)
	if f.desc != "" {
		for _, line := range strings.Split(f.desc, "\n") {
			sb.WriteString(DescStyle.Render("  " + line))
			sb.WriteByte('\n')
		}
	}

	// Input box.
	//
	// Width(inputWidth+4): lipgloss v2 uses a border-box model where Width
	// includes the border (2 chars) but NOT padding. With Padding(0,1) that
	// consumes another 2 chars, leaving exactly inputWidth for the text area —
	// matching the SetWidth(inputWidth) call above. Using inputWidth+2 (v1
	// assumption) left only inputWidth-2 for wrapping, causing the textinput
	// content to overflow and produce a spurious second line inside the box.
	//
	// Indent: lipgloss MarginLeft in v2 does not reliably distribute the margin
	// to every line of a multi-line rendered block.  We split on \n and prepend
	// "  " to each line manually so the top-border, content, and bottom-border
	// all shift left by the same 2 columns.
	borderColor := ColorDim
	if focused {
		borderColor = ColorFocus
	}
	tiView := f.ti.View()
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(inputWidth).
		Render(tiView)
	for _, line := range strings.Split(box, "\n") {
		sb.WriteString("  ")
		sb.WriteString(line)
		sb.WriteByte('\n')
	}

	// Inline validation error
	if f.errMsg != "" {
		sb.WriteString(ErrorStyle.Render("  " + GlyphCross + " " + f.errMsg))
		sb.WriteByte('\n')
	}

	return sb.String()
}
