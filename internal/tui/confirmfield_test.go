package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// ── test helpers ─────────────────────────────────────────────────────────────

// mockKeyMsg is a test double for the tea.KeyMsg interface.
type mockKeyMsg struct {
	key tea.Key
}

func (m mockKeyMsg) String() string { return m.key.String() }
func (m mockKeyMsg) Key() tea.Key   { return m.key }

// runeKey builds a mockKeyMsg for a printable character, e.g. 'y', 'n', 'k'.
func runeKey(r rune) mockKeyMsg {
	return mockKeyMsg{key: tea.Key{Code: r, Text: string(r)}}
}

// specialKey builds a mockKeyMsg for a named key constant, e.g. tea.KeyUp.
func specialKey(code rune) mockKeyMsg {
	return mockKeyMsg{key: tea.Key{Code: code}}
}

// ── NewConfirmField ───────────────────────────────────────────────────────────

func TestNewConfirmField_TruePtr(t *testing.T) {
	v := true
	f := NewConfirmField("Title", "Desc", &v)

	if f.cursor != 0 {
		t.Errorf("expected cursor 0 (Yes) for *ptr=true, got %d", f.cursor)
	}
	if !f.Value() {
		t.Error("expected Value() == true")
	}
}

func TestNewConfirmField_FalsePtr(t *testing.T) {
	v := false
	f := NewConfirmField("Title", "Desc", &v)

	if f.cursor != 1 {
		t.Errorf("expected cursor 1 (No) for *ptr=false, got %d", f.cursor)
	}
	if f.Value() {
		t.Error("expected Value() == false")
	}
}

func TestNewConfirmField_NilPtr(t *testing.T) {
	f := NewConfirmField("Title", "Desc", nil)

	if f.cursor != 0 {
		t.Errorf("expected cursor 0 (Yes) for nil ptr, got %d", f.cursor)
	}
	if !f.Value() {
		t.Error("expected Value() == true for nil ptr default")
	}
}

func TestNewConfirmField_StoresTitleAndDesc(t *testing.T) {
	v := true
	f := NewConfirmField("My Title", "My Desc", &v)

	if f.title != "My Title" {
		t.Errorf("expected title %q, got %q", "My Title", f.title)
	}
	if f.desc != "My Desc" {
		t.Errorf("expected desc %q, got %q", "My Desc", f.desc)
	}
}

// ── Value ─────────────────────────────────────────────────────────────────────

func TestValue_ReflectsCursor(t *testing.T) {
	v := true
	f := NewConfirmField("", "", &v)

	f.cursor = 0
	if !f.Value() {
		t.Error("cursor=0 should give Value()==true")
	}

	f.cursor = 1
	if f.Value() {
		t.Error("cursor=1 should give Value()==false")
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestUpdate_KeyBindings(t *testing.T) {
	tests := []struct {
		name       string
		msg        mockKeyMsg
		startValue bool
		wantCursor int
		wantValue  bool
	}{
		// Yes-selecting keys (cursor → 0)
		{name: "up", msg: specialKey(tea.KeyUp), startValue: false, wantCursor: 0, wantValue: true},
		{name: "left", msg: specialKey(tea.KeyLeft), startValue: false, wantCursor: 0, wantValue: true},
		{name: "k", msg: runeKey('k'), startValue: false, wantCursor: 0, wantValue: true},
		{name: "h", msg: runeKey('h'), startValue: false, wantCursor: 0, wantValue: true},
		{name: "y", msg: runeKey('y'), startValue: false, wantCursor: 0, wantValue: true},
		{name: "Y", msg: runeKey('Y'), startValue: false, wantCursor: 0, wantValue: true},

		// No-selecting keys (cursor → 1)
		{name: "down", msg: specialKey(tea.KeyDown), startValue: true, wantCursor: 1, wantValue: false},
		{name: "right", msg: specialKey(tea.KeyRight), startValue: true, wantCursor: 1, wantValue: false},
		{name: "j", msg: runeKey('j'), startValue: true, wantCursor: 1, wantValue: false},
		{name: "l", msg: runeKey('l'), startValue: true, wantCursor: 1, wantValue: false},
		{name: "n", msg: runeKey('n'), startValue: true, wantCursor: 1, wantValue: false},
		{name: "N", msg: runeKey('N'), startValue: true, wantCursor: 1, wantValue: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.startValue
			f := NewConfirmField("", "", &v)

			result, cmd := f.Update(tt.msg)

			if cmd != nil {
				t.Error("expected nil cmd")
			}
			cf, ok := result.(*ConfirmField)
			if !ok {
				t.Fatalf("Update() did not return *ConfirmField, got %T", result)
			}
			if cf.cursor != tt.wantCursor {
				t.Errorf("cursor: got %d, want %d", cf.cursor, tt.wantCursor)
			}
			if cf.Value() != tt.wantValue {
				t.Errorf("Value(): got %v, want %v", cf.Value(), tt.wantValue)
			}
			if v != tt.wantValue {
				t.Errorf("bound ptr: got %v, want %v", v, tt.wantValue)
			}
		})
	}
}

func TestUpdate_UnmappedKey_NoChange(t *testing.T) {
	v := true
	f := NewConfirmField("", "", &v)

	_, _ = f.Update(runeKey('z'))

	if !f.Value() {
		t.Error("unmapped key should not change selection")
	}
	if !v {
		t.Error("unmapped key should not update bound ptr")
	}
}

func TestUpdate_NonKeyMsg_NoChange(t *testing.T) {
	v := false
	f := NewConfirmField("", "", &v)

	result, cmd := f.Update("not a key message")

	if cmd != nil {
		t.Error("expected nil cmd for non-KeyMsg")
	}
	cf, ok := result.(*ConfirmField)
	if !ok {
		t.Fatalf("Update() did not return *ConfirmField, got %T", result)
	}
	if cf.cursor != 1 {
		t.Errorf("cursor should be unchanged: got %d, want 1", cf.cursor)
	}
}

func TestUpdate_NilPtr_DoesNotPanic(t *testing.T) {
	f := NewConfirmField("", "", nil)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Update() panicked with nil ptr: %v", r)
		}
	}()

	_, _ = f.Update(runeKey('n'))
}

// ── Field interface surface ───────────────────────────────────────────────────

func TestFocus_ReturnsNil(t *testing.T) {
	f := NewConfirmField("", "", nil)
	if cmd := f.Focus(); cmd != nil {
		t.Error("Focus() should return nil")
	}
}

func TestFocusable_ReturnsTrue(t *testing.T) {
	f := NewConfirmField("", "", nil)
	if !f.Focusable() {
		t.Error("Focusable() should return true")
	}
}

func TestValidate_ReturnsNil(t *testing.T) {
	f := NewConfirmField("", "", nil)
	if err := f.Validate(); err != nil {
		t.Errorf("Validate() should return nil, got %v", err)
	}
}

func TestBlur_DoesNotPanic(t *testing.T) {
	f := NewConfirmField("", "", nil)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Blur() panicked: %v", r)
		}
	}()
	f.Blur()
}

// ── View (smoke tests) ────────────────────────────────────────────────────────

func TestView_ContainsTitleAndLabels(t *testing.T) {
	v := true
	f := NewConfirmField("Are you sure?", "This cannot be undone.", &v)
	out := f.View(true, 80)

	for _, want := range []string{"Are you sure?", "This cannot be undone.", "Yes", "No"} {
		if !containsStripped(out, want) {
			t.Errorf("View() output missing %q", want)
		}
	}
}

func TestView_MultilineDesc(t *testing.T) {
	v := true
	f := NewConfirmField("Title", "Line one\nLine two", &v)
	out := f.View(false, 80)

	if !containsStripped(out, "Line one") {
		t.Error("View() missing first desc line")
	}
	if !containsStripped(out, "Line two") {
		t.Error("View() missing second desc line")
	}
}

func TestView_EmptyDesc(t *testing.T) {
	v := true
	f := NewConfirmField("Title", "", &v)
	out := f.View(false, 80)

	for _, want := range []string{"Title", "Yes", "No"} {
		if !containsStripped(out, want) {
			t.Errorf("View() with empty desc missing %q", want)
		}
	}
}

// ── ANSI-stripping helpers ────────────────────────────────────────────────────

// containsStripped strips ANSI escape codes from s before checking for substr.
// This keeps View tests stable across terminal environments where lipgloss may
// or may not emit colour codes.
func containsStripped(s, substr string) bool {
	return strings.Contains(stripANSI(s), substr)
}

func stripANSI(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++ // consume 'm'
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}
