package tui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// These reuse the mockKeyMsg type defined in confirmfield_test.go.
// If running tests in isolation, move mockKeyMsg, runeKey, specialKey,
// stripANSI, and containsStripped into a shared testhelpers_test.go file.

func spaceKey() mockKeyMsg { return mockKeyMsg{key: tea.Key{Code: ' ', Text: " "}} }
func upKey() mockKeyMsg    { return specialKey(tea.KeyUp) }
func downKey() mockKeyMsg  { return specialKey(tea.KeyDown) }

var testOpts = []Option{
	{Label: "Alpha", Value: "alpha"},
	{Label: "Beta", Value: "beta"},
	{Label: "Gamma", Value: "gamma"},
}

// selectedValues returns a sorted-ish snapshot of the values currently
// selected in f, for easy comparison in tests.
func selectedValues(f *MultiSelectField) []string {
	// Walk opts in order so the result is deterministic.
	var out []string
	for i, o := range f.opts {
		if f.selected[i] {
			out = append(out, o.Value)
		}
	}
	return out
}

// ── NewMultiSelectField ───────────────────────────────────────────────────────

func TestNewMultiSelectField_EmptyPtr(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("Title", "Desc", testOpts, &ptr)

	if f.title != "Title" {
		t.Errorf("title: got %q, want %q", f.title, "Title")
	}
	if f.desc != "Desc" {
		t.Errorf("desc: got %q, want %q", f.desc, "Desc")
	}
	if len(f.selected) != 0 {
		t.Errorf("expected no pre-selected items, got %v", f.selected)
	}
	if f.cursor != 0 {
		t.Errorf("cursor: got %d, want 0", f.cursor)
	}
}

func TestNewMultiSelectField_PreSelectsFromPtr(t *testing.T) {
	ptr := []string{"alpha", "gamma"}
	f := NewMultiSelectField("", "", testOpts, &ptr)

	if !f.selected[0] {
		t.Error("expected opt[0] (alpha) to be pre-selected")
	}
	if f.selected[1] {
		t.Error("expected opt[1] (beta) to not be pre-selected")
	}
	if !f.selected[2] {
		t.Error("expected opt[2] (gamma) to be pre-selected")
	}
}

func TestNewMultiSelectField_NilPtr(t *testing.T) {
	f := NewMultiSelectField("", "", testOpts, nil)

	if f == nil {
		t.Fatal("NewMultiSelectField returned nil")
	}
	if len(f.selected) != 0 {
		t.Errorf("expected empty selection for nil ptr, got %v", f.selected)
	}
}

func TestNewMultiSelectField_UnknownValueInPtr(t *testing.T) {
	ptr := []string{"does-not-exist"}
	f := NewMultiSelectField("", "", testOpts, &ptr)

	if len(f.selected) != 0 {
		t.Errorf("unknown value should not select anything, got %v", f.selected)
	}
}

// ── WithReadOnly ──────────────────────────────────────────────────────────────

func TestWithReadOnly_SetsFlag(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr).WithReadOnly()

	if !f.readOnly {
		t.Error("expected readOnly to be true after WithReadOnly()")
	}
}

func TestWithReadOnly_ReturnsSameInstance(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	f2 := f.WithReadOnly()

	if f != f2 {
		t.Error("WithReadOnly() should return the same *MultiSelectField")
	}
}

// ── WithSelected ──────────────────────────────────────────────────────────────

func TestWithSelected_ReplacesSelection(t *testing.T) {
	ptr := []string{"alpha"}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	f.WithSelected([]string{"beta", "gamma"})

	if f.selected[0] {
		t.Error("alpha should have been deselected")
	}
	if !f.selected[1] {
		t.Error("beta should be selected")
	}
	if !f.selected[2] {
		t.Error("gamma should be selected")
	}
}

func TestWithSelected_SyncsPtr(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	f.WithSelected([]string{"beta"})

	if len(ptr) != 1 || ptr[0] != "beta" {
		t.Errorf("ptr after WithSelected: got %v, want [beta]", ptr)
	}
}

func TestWithSelected_EmptySlice_ClearsAll(t *testing.T) {
	ptr := []string{"alpha", "beta"}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	f.WithSelected([]string{})

	if len(f.selected) != 0 {
		t.Errorf("expected all deselected, got %v", f.selected)
	}
	if len(ptr) != 0 {
		t.Errorf("ptr should be empty, got %v", ptr)
	}
}

// ── WithItemSelected ──────────────────────────────────────────────────────────

func TestWithItemSelected_SelectsSingle(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	f.WithItemSelected("beta", true)

	if !f.selected[1] {
		t.Error("beta should be selected")
	}
	if f.selected[0] || f.selected[2] {
		t.Error("other items should remain unselected")
	}
	if len(ptr) != 1 || ptr[0] != "beta" {
		t.Errorf("ptr: got %v, want [beta]", ptr)
	}
}

func TestWithItemSelected_Deselects(t *testing.T) {
	ptr := []string{"alpha", "beta"}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	f.WithItemSelected("alpha", false)

	if f.selected[0] {
		t.Error("alpha should be deselected")
	}
	if !f.selected[1] {
		t.Error("beta should still be selected")
	}
}

func TestWithItemSelected_UnknownValue_NoChange(t *testing.T) {
	ptr := []string{"alpha"}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	f.WithItemSelected("unknown", true)

	vals := selectedValues(f)
	if len(vals) != 1 || vals[0] != "alpha" {
		t.Errorf("unknown value should not change selection, got %v", vals)
	}
}

// ── WithValidate ──────────────────────────────────────────────────────────────

func TestWithValidate_AttachesFn(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	called := false
	f.WithValidate(func(vals []string) error {
		called = true
		return nil
	})

	_ = f.Validate()

	if !called {
		t.Error("expected validate fn to be called")
	}
}

func TestWithValidate_ReturnsError(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	f.WithValidate(func(vals []string) error {
		if len(vals) == 0 {
			return errors.New("at least one required")
		}
		return nil
	})

	if err := f.Validate(); err == nil {
		t.Error("expected validation error for empty selection")
	}
}

func TestWithValidate_PassesCurrentPtrValues(t *testing.T) {
	ptr := []string{"alpha"}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	var got []string
	f.WithValidate(func(vals []string) error {
		got = vals
		return nil
	})

	_ = f.Validate()

	if len(got) != 1 || got[0] != "alpha" {
		t.Errorf("validate received %v, want [alpha]", got)
	}
}

// ── Validate ──────────────────────────────────────────────────────────────────

func TestValidate_ReadOnly_AlwaysNil(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr).WithReadOnly()
	f.WithValidate(func(_ []string) error {
		return errors.New("should not be called")
	})

	if err := f.Validate(); err != nil {
		t.Errorf("Validate() in read-only mode should return nil, got %v", err)
	}
}

func TestValidate_NoFn_ReturnsNil(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)

	if err := f.Validate(); err != nil {
		t.Errorf("Validate() with no fn should return nil, got %v", err)
	}
}

func TestValidate_NilPtr_DoesNotPanic(t *testing.T) {
	f := NewMultiSelectField("", "", testOpts, nil)
	f.WithValidate(func(vals []string) error { return nil })

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Validate() panicked with nil ptr: %v", r)
		}
	}()

	_ = f.Validate()
}

// ── Field interface surface ───────────────────────────────────────────────────

func TestMultiSelect_Focus_ReturnsNil(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	if cmd := f.Focus(); cmd != nil {
		t.Error("Focus() should return nil")
	}
}

func TestMultiSelect_Focusable_ReturnsTrue(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	if !f.Focusable() {
		t.Error("Focusable() should return true")
	}
}

func TestMultiSelect_Blur_DoesNotPanic(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Blur() panicked: %v", r)
		}
	}()
	f.Blur()
}

// ── Update: cursor movement ───────────────────────────────────────────────────

func TestUpdate_CursorDown(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)

	f.Update(downKey())

	if f.cursor != 1 {
		t.Errorf("cursor: got %d, want 1", f.cursor)
	}
}

func TestUpdate_CursorUp(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	f.cursor = 2

	f.Update(upKey())

	if f.cursor != 1 {
		t.Errorf("cursor: got %d, want 1", f.cursor)
	}
}

func TestUpdate_CursorDoesNotGoAboveZero(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)

	f.Update(upKey())

	if f.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", f.cursor)
	}
}

func TestUpdate_CursorDoesNotGoBeyondLastOption(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	f.cursor = len(testOpts) - 1

	f.Update(downKey())

	if f.cursor != len(testOpts)-1 {
		t.Errorf("cursor should stay at %d, got %d", len(testOpts)-1, f.cursor)
	}
}

func TestUpdate_ViKeys_UpDown(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)

	f.Update(runeKey('j'))
	if f.cursor != 1 {
		t.Errorf("j: cursor got %d, want 1", f.cursor)
	}

	f.Update(runeKey('k'))
	if f.cursor != 0 {
		t.Errorf("k: cursor got %d, want 0", f.cursor)
	}
}

// ── Update: space toggle ──────────────────────────────────────────────────────

func TestUpdate_Space_TogglesSelection(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)

	f.Update(spaceKey())

	if !f.selected[0] {
		t.Error("space should select item at cursor")
	}
	if len(ptr) != 1 || ptr[0] != "alpha" {
		t.Errorf("ptr after toggle on: got %v, want [alpha]", ptr)
	}

	f.Update(spaceKey())

	if f.selected[0] {
		t.Error("second space should deselect item at cursor")
	}
	if len(ptr) != 0 {
		t.Errorf("ptr after toggle off: got %v, want []", ptr)
	}
}

func TestUpdate_Space_ReadOnly_DoesNotToggle(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr).WithReadOnly()

	f.Update(spaceKey())

	if f.selected[0] {
		t.Error("space in read-only mode should not toggle selection")
	}
}

func TestUpdate_Space_ClearsErrMsg(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	f.errMsg = "some error"

	f.Update(spaceKey())

	if f.errMsg != "" {
		t.Errorf("space should clear errMsg, got %q", f.errMsg)
	}
}

func TestUpdate_Space_SyncsPtr(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	f.cursor = 1

	f.Update(spaceKey())

	if len(ptr) != 1 || ptr[0] != "beta" {
		t.Errorf("ptr after toggle: got %v, want [beta]", ptr)
	}
}

func TestUpdate_NonKeyMsg_NoChangeMultiSelect(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)

	result, cmd := f.Update("not a key")

	if cmd != nil {
		t.Error("expected nil cmd for non-KeyMsg")
	}
	cf := result.(*MultiSelectField)
	if cf.cursor != 0 || len(cf.selected) != 0 {
		t.Error("non-KeyMsg should cause no state change")
	}
}

// ── View (smoke tests) ────────────────────────────────────────────────────────

func TestMultiSelect_View_ContainsTitleAndOptions(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("Pick one", "Choose carefully", testOpts, &ptr)
	out := f.View(true, 80)

	for _, want := range []string{"Pick one", "Choose carefully", "Alpha", "Beta", "Gamma"} {
		if !containsStripped(out, want) {
			t.Errorf("View() missing %q", want)
		}
	}
}

func TestMultiSelect_View_HintShownWhenFocusedWritable(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	out := f.View(true, 80)

	if !containsStripped(out, "space: toggle") {
		t.Error("View() should show hint when focused and writable")
	}
}

func TestMultiSelect_View_HintHiddenWhenBlurred(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	out := f.View(false, 80)

	if containsStripped(out, "space: toggle") {
		t.Error("View() should not show hint when blurred")
	}
}

func TestMultiSelect_View_HintHiddenWhenReadOnly(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr).WithReadOnly()
	out := f.View(true, 80)

	if containsStripped(out, "space: toggle") {
		t.Error("View() should not show hint in read-only mode")
	}
}

func TestMultiSelect_View_MultilineDesc(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "Line one\nLine two", testOpts, &ptr)
	out := f.View(false, 80)

	if !containsStripped(out, "Line one") || !containsStripped(out, "Line two") {
		t.Error("View() should render all desc lines")
	}
}

func TestMultiSelect_View_ErrorMsg(t *testing.T) {
	ptr := []string{}
	f := NewMultiSelectField("", "", testOpts, &ptr)
	f.errMsg = "at least one required"
	out := f.View(true, 80)

	if !containsStripped(out, "at least one required") {
		t.Error("View() should render errMsg when set")
	}
}

// stripANSI and containsStripped are defined in confirmfield_test.go.
// If running this file in isolation, move them to a shared testhelpers_test.go.

func containsStrippedStr(s, substr string) bool {
	return strings.Contains(stripANSI(s), substr)
}
