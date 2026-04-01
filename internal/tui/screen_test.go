package tui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// ── mock field ────────────────────────────────────────────────────────────────

// mockField is a minimal Field implementation for Screen tests.
// It records Focus/Blur calls and can be configured to fail Validate.
type mockField struct {
	focusable    bool
	focused      bool
	blurred      bool
	focusCount   int
	blurCount    int
	validateErr  error
	updateCalled bool
	viewCalled   bool
}

func (m *mockField) Focus() tea.Cmd {
	m.focused = true
	m.focusCount++
	return nil
}
func (m *mockField) Blur() {
	m.blurred = true
	m.blurCount++
}
func (m *mockField) Focusable() bool                   { return m.focusable }
func (m *mockField) Validate() error                   { return m.validateErr }
func (m *mockField) Update(_ tea.Msg) (Field, tea.Cmd) { m.updateCalled = true; return m, nil }
func (m *mockField) View(_ bool, _ int) string         { m.viewCalled = true; return "" }

func focusable() *mockField   { return &mockField{focusable: true} }
func unfocusable() *mockField { return &mockField{focusable: false} }
func failing() *mockField     { return &mockField{focusable: true, validateErr: errors.New("invalid")} }

// ── key helpers ───────────────────────────────────────────────────────────────

func escKey() mockKeyMsg      { return specialKey(tea.KeyEsc) }
func tabKey() mockKeyMsg      { return specialKey(tea.KeyTab) }
func enterKey() mockKeyMsg    { return specialKey(tea.KeyEnter) }
func shiftTabKey() mockKeyMsg { return mockKeyMsg{key: tea.Key{Code: tea.KeyTab, Mod: tea.ModShift}} }

// ── NewScreen ─────────────────────────────────────────────────────────────────

func TestNewScreen_FocusesFirstFocusableField(t *testing.T) {
	note := unfocusable()
	field := focusable()
	s := NewScreen(80, note, field)

	if s.focused != 1 {
		t.Errorf("focused: got %d, want 1", s.focused)
	}
}

func TestNewScreen_NoFocusableFields_FocusedIsNegOne(t *testing.T) {
	s := NewScreen(80, unfocusable(), unfocusable())

	if s.focused != -1 {
		t.Errorf("focused: got %d, want -1", s.focused)
	}
}

func TestNewScreen_EmptyFields_FocusedIsNegOne(t *testing.T) {
	s := NewScreen(80)

	if s.focused != -1 {
		t.Errorf("focused: got %d, want -1", s.focused)
	}
}

func TestNewScreen_StoresWidth(t *testing.T) {
	s := NewScreen(120, focusable())

	if s.width != 120 {
		t.Errorf("width: got %d, want 120", s.width)
	}
}

func TestNewScreen_FirstFocusableSkipsLeadingNotes(t *testing.T) {
	n1, n2 := unfocusable(), unfocusable()
	f := focusable()
	s := NewScreen(80, n1, n2, f)

	if s.focused != 2 {
		t.Errorf("focused: got %d, want 2", s.focused)
	}
}

// ── Init ──────────────────────────────────────────────────────────────────────

func TestInit_CallsFocusOnFocusedField(t *testing.T) {
	f := focusable()
	s := NewScreen(80, f)
	s.Init()

	if !f.focused {
		t.Error("Init() should call Focus() on the initially focused field")
	}
}

func TestInit_NoFocusableFields_DoesNotPanic(t *testing.T) {
	s := NewScreen(80, unfocusable())

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Init() panicked with no focusable fields: %v", r)
		}
	}()
	s.Init()
}

func TestInit_EmptyScreen_ReturnsNilCmd(t *testing.T) {
	s := NewScreen(80)
	_, cmd := s.Init()

	if cmd != nil {
		t.Error("Init() with no fields should return nil cmd")
	}
}

// ── SetWidth ──────────────────────────────────────────────────────────────────

func TestSetWidth_UpdatesWidth(t *testing.T) {
	s := NewScreen(80, focusable())
	s = s.SetWidth(120)

	if s.width != 120 {
		t.Errorf("width: got %d, want 120", s.width)
	}
}

func TestSetWidth_ReturnsNewScreen(t *testing.T) {
	s := NewScreen(80, focusable())
	s2 := s.SetWidth(100)

	// Value receiver — s2 is a copy, original unchanged.
	if s.width != 80 {
		t.Errorf("original width should be unchanged: got %d", s.width)
	}
	if s2.width != 100 {
		t.Errorf("new width: got %d, want 100", s2.width)
	}
}

// ── Update: Esc ───────────────────────────────────────────────────────────────

func TestUpdate_Esc_SetsCancelled(t *testing.T) {
	s := NewScreen(80, focusable())
	s, _ = s.Update(escKey())

	if !s.Cancelled {
		t.Error("Esc should set Cancelled=true")
	}
}

func TestUpdate_Esc_DoesNotSetSubmitted(t *testing.T) {
	s := NewScreen(80, focusable())
	s, _ = s.Update(escKey())

	if s.Submitted {
		t.Error("Esc should not set Submitted")
	}
}

// ── Update: Tab ───────────────────────────────────────────────────────────────

func TestUpdate_Tab_AdvancesFocus(t *testing.T) {
	f1, f2 := focusable(), focusable()
	s := NewScreen(80, f1, f2)
	s, _ = s.Update(tabKey())

	if s.focused != 1 {
		t.Errorf("focused after Tab: got %d, want 1", s.focused)
	}
}

func TestUpdate_Tab_BlursPreviousField(t *testing.T) {
	f1, f2 := focusable(), focusable()
	s := NewScreen(80, f1, f2)
	s, _ = s.Update(tabKey())

	if !f1.blurred {
		t.Error("Tab should blur the previously focused field")
	}
}

func TestUpdate_Tab_FocusesNextField(t *testing.T) {
	f1, f2 := focusable(), focusable()
	s := NewScreen(80, f1, f2)
	s, _ = s.Update(tabKey())

	if !f2.focused {
		t.Error("Tab should focus the next field")
	}
}

func TestUpdate_Tab_WrapsAroundToFirst(t *testing.T) {
	f1, f2 := focusable(), focusable()
	s := NewScreen(80, f1, f2)
	s.focused = 1
	s, _ = s.Update(tabKey())

	if s.focused != 0 {
		t.Errorf("Tab at last field should wrap to 0, got %d", s.focused)
	}
}

func TestUpdate_Tab_SkipsUnfocusableFields(t *testing.T) {
	f1, note, f2 := focusable(), unfocusable(), focusable()
	s := NewScreen(80, f1, note, f2)
	s, _ = s.Update(tabKey())

	if s.focused != 2 {
		t.Errorf("Tab should skip unfocusable fields, focused=%d want 2", s.focused)
	}
}

func TestUpdate_Tab_ClearsErrMsg(t *testing.T) {
	f1, f2 := focusable(), focusable()
	s := NewScreen(80, f1, f2)
	s.ErrMsg = "some error"
	s, _ = s.Update(tabKey())

	if s.ErrMsg != "" {
		t.Errorf("Tab should clear ErrMsg, got %q", s.ErrMsg)
	}
}

// ── Update: ShiftTab ──────────────────────────────────────────────────────────

func TestUpdate_ShiftTab_MovesFocusBackward(t *testing.T) {
	f1, f2 := focusable(), focusable()
	s := NewScreen(80, f1, f2)
	s.focused = 1
	s, _ = s.Update(shiftTabKey())

	if s.focused != 0 {
		t.Errorf("ShiftTab should move focus to 0, got %d", s.focused)
	}
}

func TestUpdate_ShiftTab_WrapsAroundToLast(t *testing.T) {
	f1, f2 := focusable(), focusable()
	s := NewScreen(80, f1, f2)
	// focused starts at 0
	s, _ = s.Update(shiftTabKey())

	if s.focused != 1 {
		t.Errorf("ShiftTab at first field should wrap to last, got %d", s.focused)
	}
}

func TestUpdate_ShiftTab_SkipsUnfocusableFields(t *testing.T) {
	f1, note, f2 := focusable(), unfocusable(), focusable()
	s := NewScreen(80, f1, note, f2)
	s.focused = 2
	s, _ = s.Update(shiftTabKey())

	if s.focused != 0 {
		t.Errorf("ShiftTab should skip unfocusable, focused=%d want 0", s.focused)
	}
}

// ── Update: Enter ─────────────────────────────────────────────────────────────

func TestUpdate_Enter_ValidationFailure_SetsErrMsg(t *testing.T) {
	f := failing()
	s := NewScreen(80, f)
	s, _ = s.Update(enterKey())

	if s.ErrMsg == "" {
		t.Error("Enter with invalid field should set ErrMsg")
	}
	if s.Submitted {
		t.Error("Enter with invalid field should not submit")
	}
}

func TestUpdate_Enter_AdvancesToNextFocusableField(t *testing.T) {
	f1, f2 := focusable(), focusable()
	s := NewScreen(80, f1, f2)
	s, _ = s.Update(enterKey())

	if s.focused != 1 {
		t.Errorf("Enter should advance focus to 1, got %d", s.focused)
	}
	if s.Submitted {
		t.Error("Enter with remaining fields should not submit")
	}
}

func TestUpdate_Enter_SkipsUnfocusableOnAdvance(t *testing.T) {
	f1, note, f2 := focusable(), unfocusable(), focusable()
	s := NewScreen(80, f1, note, f2)
	s, _ = s.Update(enterKey())

	if s.focused != 2 {
		t.Errorf("Enter should skip unfocusable, focused=%d want 2", s.focused)
	}
}

func TestUpdate_Enter_LastField_AllValid_Submits(t *testing.T) {
	f1, f2 := focusable(), focusable()
	s := NewScreen(80, f1, f2)
	s.focused = 1
	s, _ = s.Update(enterKey())

	if !s.Submitted {
		t.Error("Enter on last valid field should submit")
	}
}

func TestUpdate_Enter_LastField_OtherFieldInvalid_DoesNotSubmit(t *testing.T) {
	bad, good := failing(), focusable()
	s := NewScreen(80, bad, good)
	s.focused = 1
	s, _ = s.Update(enterKey())

	if s.Submitted {
		t.Error("should not submit when another field is invalid")
	}
	if s.ErrMsg == "" {
		t.Error("should set ErrMsg from failing field")
	}
}

func TestUpdate_Enter_ClearsErrMsgBeforeValidating(t *testing.T) {
	f := focusable()
	s := NewScreen(80, f)
	s.ErrMsg = "stale"
	s.focused = 0
	s, _ = s.Update(enterKey())

	// Field is valid and it's the only field, so it submits — ErrMsg cleared.
	if s.ErrMsg != "" {
		t.Errorf("Enter should clear stale ErrMsg, got %q", s.ErrMsg)
	}
}

func TestUpdate_Enter_TextField_SetsInlineError(t *testing.T) {
	ptr := ""
	tf := NewTextField("", "", &ptr)
	tf.WithValidate(func(s string) error {
		if s == "" {
			return errors.New("required")
		}
		return nil
	})
	s := NewScreen(80, tf)
	s, _ = s.Update(enterKey())

	if tf.errMsg != "required" {
		t.Errorf("TextField inline error: got %q, want %q", tf.errMsg, "required")
	}
}

func TestUpdate_Enter_MultiSelectField_SetsInlineError(t *testing.T) {
	ptr := []string{}
	msf := NewMultiSelectField("", "", testOpts, &ptr)
	msf.WithValidate(func(vals []string) error {
		if len(vals) == 0 {
			return errors.New("pick one")
		}
		return nil
	})
	s := NewScreen(80, msf)
	s, _ = s.Update(enterKey())

	if msf.errMsg != "pick one" {
		t.Errorf("MultiSelectField inline error: got %q, want %q", msf.errMsg, "pick one")
	}
}

// ── Update: non-key messages delegated to focused field ───────────────────────

func TestUpdate_NonKeyMsg_DelegatesToFocusedField(t *testing.T) {
	f := focusable()
	s := NewScreen(80, f)
	s, _ = s.Update("not a key")

	if !f.updateCalled {
		t.Error("non-key message should be delegated to focused field")
	}
}

func TestUpdate_NonKeyMsg_NoFocusedField_DoesNotPanic(t *testing.T) {
	s := NewScreen(80, unfocusable())

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Update() panicked with no focused field: %v", r)
		}
	}()
	s.Update("anything")
}

// ── View ──────────────────────────────────────────────────────────────────────

func TestView_CallsViewOnAllFields(t *testing.T) {
	f1, f2 := focusable(), focusable()
	s := NewScreen(80, f1, f2)
	s.View()

	if !f1.viewCalled {
		t.Error("View() should call View on f1")
	}
	if !f2.viewCalled {
		t.Error("View() should call View on f2")
	}
}

func TestView_FocusedFieldReceivesFocusedTrue(t *testing.T) {
	// Use a real NoteField as unfocusable and a spy that records focused param.
	spy := &viewSpyField{focusable: true}
	s := NewScreen(80, spy)
	s.View()

	if !spy.lastFocused {
		t.Error("View() should pass focused=true to the focused field")
	}
}

func TestView_UnfocusedFieldReceivesFocusedFalse(t *testing.T) {
	spy1 := &viewSpyField{focusable: true}
	spy2 := &viewSpyField{focusable: true}
	s := NewScreen(80, spy1, spy2)
	// focused=0, so spy2 should receive focused=false
	s.View()

	if spy2.lastFocused {
		t.Error("View() should pass focused=false to non-focused fields")
	}
}

func TestView_RendersErrMsg(t *testing.T) {
	s := NewScreen(80, focusable())
	s.ErrMsg = "something went wrong"
	out := s.View()

	if !containsStripped(out, "something went wrong") {
		t.Error("View() should render ErrMsg when set")
	}
}

func TestView_EmptyErrMsg_NotRendered(t *testing.T) {
	s := NewScreen(80, focusable())
	out := s.View()

	// GlyphCross only appears with an error — rough proxy for error line absence.
	if containsStripped(out, GlyphCross) {
		t.Error("View() should not render error line when ErrMsg is empty")
	}
}

// viewSpyField records what focused value it was last rendered with.
type viewSpyField struct {
	focusable   bool
	lastFocused bool
}

func (v *viewSpyField) Focus() tea.Cmd                    { return nil }
func (v *viewSpyField) Blur()                             {}
func (v *viewSpyField) Focusable() bool                   { return v.focusable }
func (v *viewSpyField) Validate() error                   { return nil }
func (v *viewSpyField) Update(_ tea.Msg) (Field, tea.Cmd) { return v, nil }
func (v *viewSpyField) View(focused bool, _ int) string {
	v.lastFocused = focused
	return ""
}
