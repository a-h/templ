package modcheck

import (
	"testing"

	"golang.org/x/mod/modfile"
)

func TestPatchGoVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "go 1.20",
			expected: "1.20",
		},
		{
			input:    "go 1.20.123",
			expected: "1.20.123",
		},
		{
			input:    "go 1.20.1",
			expected: "1.20.1",
		},
		{
			input:    "go 1.20rc1",
			expected: "1.20rc1",
		},
		{
			input:    "go 1.15",
			expected: "1.15",
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			input := "module github.com/a-h/templ\n\n" + string(test.input) + "\n" + "toolchain go1.27.9\n"
			mf, err := modfile.Parse("go.mod", []byte(input), nil)
			if err != nil {
				t.Fatalf("failed to parse go.mod: %v", err)
			}
			if test.expected != mf.Go.Version {
				t.Errorf("expected %q, got %q", test.expected, mf.Go.Version)
			}
		})
	}
}
