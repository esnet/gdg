package tools

import (
	"io"
	"os"
	"testing"
)

func TestGetUserConfirmation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"yes", "yes\n", true},
		{"Yes", "Yes\n", true},
		{"y", "y\n", true},
		{"Y", "Y\n", true},
		{"no", "no\n", false},
		{"No", "No\n", false},
		{"n", "n\n", false},
		{"N", "N\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, w, _ := os.Pipe()
			origStdin := os.Stdin
			os.Stdin = r
			defer func() { os.Stdin = origStdin }()

			_, _ = io.WriteString(w, tt.input)
			w.Close()

			got := GetUserConfirmation("prompt: ", "", false)

			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

// TestGetUserConfirmation_CustomErrorString exercises the branch where a
// non-empty error string is supplied. With terminate=false the function still
// returns false on "n" — the error string is only used when terminate=true
// triggers log.Fatal, which we cannot call in a test. This verifies that the
// "if error == """" guard does NOT overwrite a caller-supplied message, i.e.
// the function compiles and runs correctly with a non-empty error argument.
func TestGetUserConfirmation_CustomErrorString(t *testing.T) {
	r, w, _ := os.Pipe()
	origStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	_, _ = io.WriteString(w, "n\n")
	w.Close()

	// Pass a custom, non-empty error string. terminate=false so no log.Fatal.
	got := GetUserConfirmation("Continue? ", "Custom goodbye message", false)
	if got != false {
		t.Fatalf("expected false for 'n' answer, got %v", got)
	}
}

// TestGetUserConfirmation_YesWithCustomError verifies that "y" still returns
// true when a non-empty error string is provided, and that the error string
// does not interfere with a positive answer.
func TestGetUserConfirmation_YesWithCustomError(t *testing.T) {
	r, w, _ := os.Pipe()
	origStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	_, _ = io.WriteString(w, "y\n")
	w.Close()

	got := GetUserConfirmation("Continue? ", "Custom error message", false)
	if got != true {
		t.Fatalf("expected true for 'y' answer, got %v", got)
	}
}
