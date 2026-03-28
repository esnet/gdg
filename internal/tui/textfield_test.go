package tui

import (
	"errors"
	"testing"

	"charm.land/bubbles/v2/textinput"
)

// ── NewTextField ──────────────────────────────────────────────────────────────

func TestNewTextField_StoresTitleAndDesc(t *testing.T) {
	ptr := ""
	f := NewTextField("My Title", "My Desc", &ptr)

	if f.title != "My Title" {
		t.Errorf("title: got %q, want %q", f.title, "My Title")
	}
	if f.desc != "My Desc" {
		t.Errorf("desc: got %q, want %q", f.desc, "My Desc")
	}
}

func TestNewTextField_PreFillsFromPtr(t *testing.T) {
	ptr := "prefilled"
	f := NewTextField("", "", &ptr)

	if f.Value() != "prefilled" {
		t.Errorf("Value(): got %q, want %q", f.Value(), "prefilled")
	}
}

func TestNewTextField_EmptyPtr_NoPreFill(t *testing.T) {
	ptr := ""
	f := NewTextField("", "", &ptr)

	if f.Value() != "" {
		t.Errorf("Value(): got %q, want empty string", f.Value())
	}
}

func TestNewTextField_NilPtr_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewTextField panicked with nil ptr: %v", r)
		}
	}()
	f := NewTextField("", "", nil)
	if f == nil {
		t.Error("NewTextField returned nil")
	}
}

func TestNewTextField_PromptIsEmpty(t *testing.T) {
	ptr := ""
	f := NewTextField("", "", &ptr)

	if f.ti.Prompt != "" {
		t.Errorf("Prompt: got %q, want empty string", f.ti.Prompt)
	}
}

// ── WithMask ──────────────────────────────────────────────────────────────────

func TestWithMask_SetsEchoMode(t *testing.T) {
	ptr := ""
	f := NewTextField("", "", &ptr).WithMask()

	if f.ti.EchoMode != textinput.EchoPassword {
		t.Errorf("EchoMode: got %v, want EchoPassword", f.ti.EchoMode)
	}
	if f.ti.EchoCharacter != '•' {
		t.Errorf("EchoCharacter: got %q, want '•'", f.ti.EchoCharacter)
	}
}

func TestWithMask_ReturnsSameInstance(t *testing.T) {
	ptr := ""
	f := NewTextField("", "", &ptr)
	f2 := f.WithMask()

	if f != f2 {
		t.Error("WithMask() should return the same *TextField")
	}
}

// ── WithPlaceholder ───────────────────────────────────────────────────────────

func TestWithPlaceholder_SetsPlaceholder(t *testing.T) {
	ptr := ""
	f := NewTextField("", "", &ptr).WithPlaceholder("enter text here")

	if f.ti.Placeholder != "enter text here" {
		t.Errorf("Placeholder: got %q, want %q", f.ti.Placeholder, "enter text here")
	}
}

func TestWithPlaceholder_ReturnsSameInstance(t *testing.T) {
	ptr := ""
	f := NewTextField("", "", &ptr)
	f2 := f.WithPlaceholder("hint")

	if f != f2 {
		t.Error("WithPlaceholder() should return the same *TextField")
	}
}

// ── WithValidate ──────────────────────────────────────────────────────────────

func TestWithValidate_ReturnsSameInstance(t *testing.T) {
	ptr := ""
	f := NewTextField("", "", &ptr)
	f2 := f.WithValidate(func(s string) error { return nil })

	if f != f2 {
		t.Error("WithValidate() should return the same *TextField")
	}
}

// ── Value ─────────────────────────────────────────────────────────────────────

func TestValue_ReturnsCurrentInput(t *testing.T) {
	ptr := "hello"
	f := NewTextField("", "", &ptr)

	if f.Value() != "hello" {
		t.Errorf("Value(): got %q, want %q", f.Value(), "hello")
	}
}

// ── SetError ──────────────────────────────────────────────────────────────────

func TestSetError_StoresMessage(t *testing.T) {
	ptr := ""
	f := NewTextField("", "", &ptr)
	f.SetError("something went wrong")

	if f.errMsg != "something went wrong" {
		t.Errorf("errMsg: got %q, want %q", f.errMsg, "something went wrong")
	}
}

// ── Field interface ───────────────────────────────────────────────────────────

func TestTextField_Focusable_ReturnsTrue(t *testing.T) {
	ptr := ""
	f := NewTextField("", "", &ptr)

	if !f.Focusable() {
		t.Error("Focusable() should return true")
	}
}

func TestTextField_Focus_ClearsErrMsg(t *testing.T) {
	ptr := ""
	f := NewTextField("", "", &ptr)
	f.errMsg = "stale error"

	f.Focus()

	if f.errMsg != "" {
		t.Errorf("Focus() should clear errMsg, got %q", f.errMsg)
	}
}

func TestTextField_Focus_DoesNotPanic(t *testing.T) {
	ptr := ""
	f := NewTextField("", "", &ptr)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Focus() panicked: %v", r)
		}
	}()
	f.Focus()
}

func TestTextField_Blur_DoesNotPanic(t *testing.T) {
	ptr := ""
	f := NewTextField("", "", &ptr)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Blur() panicked: %v", r)
		}
	}()
	f.Blur()
}

// ── Validate ──────────────────────────────────────────────────────────────────

func TestTextField_Validate_NoFn_ReturnsNil(t *testing.T) {
	ptr := ""
	f := NewTextField("", "", &ptr)

	if err := f.Validate(); err != nil {
		t.Errorf("Validate() with no fn should return nil, got %v", err)
	}
}

func TestTextField_Validate_PassesCurrentValue(t *testing.T) {
	ptr := "myvalue"
	f := NewTextField("", "", &ptr)
	var got string
	f.WithValidate(func(s string) error {
		got = s
		return nil
	})

	_ = f.Validate()

	if got != "myvalue" {
		t.Errorf("Validate() passed %q, want %q", got, "myvalue")
	}
}

func TestTextField_Validate_ReturnsError(t *testing.T) {
	ptr := ""
	f := NewTextField("", "", &ptr)
	f.WithValidate(func(s string) error {
		if s == "" {
			return errors.New("required")
		}
		return nil
	})

	if err := f.Validate(); err == nil {
		t.Error("Validate() should return error for empty input")
	}
}

func TestTextField_Validate_ReturnsNilOnValid(t *testing.T) {
	ptr := "filled"
	f := NewTextField("", "", &ptr)
	f.WithValidate(func(s string) error {
		if s == "" {
			return errors.New("required")
		}
		return nil
	})

	if err := f.Validate(); err != nil {
		t.Errorf("Validate() should return nil for valid input, got %v", err)
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestTextField_Update_NonKeyMsg_ReturnsField(t *testing.T) {
	ptr := ""
	f := NewTextField("", "", &ptr)

	result, _ := f.Update("not a key")

	if result != f {
		t.Error("Update() should return the same *TextField")
	}
}

func TestTextField_Update_SyncsPtrOnChange(t *testing.T) {
	ptr := "initial"
	f := NewTextField("", "", &ptr)

	// Focus the field first so it accepts input.
	f.Focus()

	// Send a rune key to append a character.
	f.Update(runeKey('!'))

	// The ptr should now reflect whatever the textinput holds.
	if *f.ptr != f.Value() {
		t.Errorf("ptr %q out of sync with Value() %q", *f.ptr, f.Value())
	}
}

func TestTextField_Update_NilPtr_DoesNotPanic(t *testing.T) {
	f := NewTextField("", "", nil)
	f.Focus()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Update() panicked with nil ptr: %v", r)
		}
	}()

	f.Update(runeKey('a'))
}

// ── View (smoke tests) ────────────────────────────────────────────────────────

func TestTextField_View_ContainsTitleAndDesc(t *testing.T) {
	ptr := ""
	f := NewTextField("Username", "Enter your username.", &ptr)
	out := f.View(false, 80)

	for _, want := range []string{"Username", "Enter your username."} {
		if !containsStripped(out, want) {
			t.Errorf("View() missing %q", want)
		}
	}
}

func TestTextField_View_MultilineDesc(t *testing.T) {
	ptr := ""
	f := NewTextField("", "Line one\nLine two", &ptr)
	out := f.View(false, 80)

	for _, want := range []string{"Line one", "Line two"} {
		if !containsStripped(out, want) {
			t.Errorf("View() missing %q", want)
		}
	}
}

func TestTextField_View_EmptyDesc_NoExtraContent(t *testing.T) {
	ptr := ""
	f := NewTextField("Title", "", &ptr)
	out := f.View(false, 80)

	if !containsStripped(out, "Title") {
		t.Error("View() missing title")
	}
}

func TestTextField_View_ShowsErrMsg(t *testing.T) {
	ptr := ""
	f := NewTextField("", "", &ptr)
	f.SetError("field is required")
	out := f.View(false, 80)

	if !containsStripped(out, "field is required") {
		t.Error("View() should render errMsg when set")
	}
}

func TestTextField_View_NarrowWidth_ClampsToMinimum(t *testing.T) {
	ptr := ""
	f := NewTextField("Title", "Desc", &ptr)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("View() panicked with zero width: %v", r)
		}
	}()

	out := f.View(false, 0)
	if !containsStripped(out, "Title") {
		t.Error("View() with zero width should still render title")
	}
}

func TestTextField_View_PrefilledValueVisible(t *testing.T) {
	ptr := "existingvalue"
	f := NewTextField("", "", &ptr)
	out := f.View(false, 80)

	if !containsStripped(out, "existingvalue") {
		t.Error("View() should render prefilled value")
	}
}
