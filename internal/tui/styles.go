// Package tui provides reusable terminal UI components built on top of
// Bubbletea v2 and Bubbles v2. All components use pointer receivers so they
// can be stored as the Field interface in a Screen without extra copying.
package tui

import "charm.land/lipgloss/v2"

// Palette — all colours used across GDG TUIs in one place.
var (
	ColorFocus   = lipgloss.Color("63")  // indigo  – focused borders / cursors
	ColorAccent  = lipgloss.Color("215") // amber   – titles / highlights
	ColorMuted   = lipgloss.Color("241") // grey    – descriptions / footer
	ColorError   = lipgloss.Color("196") // red     – validation errors
	ColorSuccess = lipgloss.Color("46")  // green   – success marks
	ColorText    = lipgloss.Color("252") // off-white – body text
	ColorBright  = lipgloss.Color("255") // white   – selected / important text
	ColorDim     = lipgloss.Color("240") // dark grey – unfocused options
)

// Shared text styles.
var (
	TitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(ColorBright)
	DescStyle    = lipgloss.NewStyle().Foreground(ColorMuted)
	ErrorStyle   = lipgloss.NewStyle().Foreground(ColorError)
	FocusStyle   = lipgloss.NewStyle().Foreground(ColorFocus)
	BlurStyle    = lipgloss.NewStyle().Foreground(ColorDim)
	AccentStyle  = lipgloss.NewStyle().Foreground(ColorAccent)
	SuccessStyle = lipgloss.NewStyle().Foreground(ColorSuccess)
)

// Glyph constants used across all field renders.
const (
	GlyphCursor = "❯"
	GlyphCheck  = "✓"
	GlyphCross  = "✗"
	GlyphBullet = "•"
)
