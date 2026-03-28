package tui

import (
	"testing"
)

// ── NewNoteField ──────────────────────────────────────────────────────────────

func TestNewNoteField_StoresTitleAndBody(t *testing.T) {
	f := NewNoteField("My Title", "My Body")

	if f.title != "My Title" {
		t.Errorf("title: got %q, want %q", f.title, "My Title")
	}
	if f.body != "My Body" {
		t.Errorf("body: got %q, want %q", f.body, "My Body")
	}
}

// ── Field interface ───────────────────────────────────────────────────────────

func TestNoteField_Focusable_ReturnsFalse(t *testing.T) {
	f := NewNoteField("", "")
	if f.Focusable() {
		t.Error("Focusable() must return false — Screen uses this to skip NoteField")
	}
}

func TestNoteField_Focus_ReturnsNil(t *testing.T) {
	f := NewNoteField("", "")
	if cmd := f.Focus(); cmd != nil {
		t.Error("Focus() should return nil")
	}
}

func TestNoteField_Validate_ReturnsNil(t *testing.T) {
	f := NewNoteField("", "")
	if err := f.Validate(); err != nil {
		t.Errorf("Validate() should return nil, got %v", err)
	}
}

func TestNoteField_Blur_DoesNotPanic(t *testing.T) {
	f := NewNoteField("", "")
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Blur() panicked: %v", r)
		}
	}()
	f.Blur()
}

func TestNoteField_Update_IsNoOp(t *testing.T) {
	f := NewNoteField("", "")

	result, cmd := f.Update("anything")

	if cmd != nil {
		t.Error("Update() should return nil cmd")
	}
	if result != f {
		t.Error("Update() should return the same *NoteField")
	}
}

// ── View ──────────────────────────────────────────────────────────────────────

func TestNoteField_View_ContainsTitleAndBody(t *testing.T) {
	f := NewNoteField("Important Notice", "Please read this carefully.")
	out := f.View(false, 80)

	for _, want := range []string{"Important Notice", "Please read this carefully."} {
		if !containsStripped(out, want) {
			t.Errorf("View() missing %q", want)
		}
	}
}

func TestNoteField_View_MultilineBody(t *testing.T) {
	f := NewNoteField("Title", "Line one\nLine two\nLine three")
	out := f.View(false, 80)

	for _, want := range []string{"Line one", "Line two", "Line three"} {
		if !containsStripped(out, want) {
			t.Errorf("View() missing %q", want)
		}
	}
}

func TestNoteField_View_NarrowWidth_ClampsToMinDivider(t *testing.T) {
	f := NewNoteField("Title", "Body")

	// width=0 would produce dividerWidth=-4, clamped to 4.
	// Should not panic and should still contain content.
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

func TestNoteField_View_FocusedParamIgnored(t *testing.T) {
	f := NewNoteField("Title", "Body")

	focused := f.View(true, 80)
	blurred := f.View(false, 80)

	if focused != blurred {
		t.Error("View() output should be identical regardless of focused param")
	}
}
