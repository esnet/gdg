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
