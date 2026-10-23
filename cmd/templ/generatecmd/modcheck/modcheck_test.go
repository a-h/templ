package modcheck

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/mod/modfile"
)

func TestWalkUp(t *testing.T) {
	t.Run("finds go.mod in a parent directory", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n"), 0600); err != nil {
			t.Fatal(err)
		}
		subdir := filepath.Join(dir, "child")
		if err := os.Mkdir(subdir, 0700); err != nil {
			t.Fatal(err)
		}
		got, err := WalkUp(subdir)
		if err != nil {
			t.Fatal(err)
		}
		if got != dir {
			t.Fatalf("expected module directory %q, got %q", dir, got)
		}
	})
	t.Run("returns an error when go.mod is missing", func(t *testing.T) {
		_, err := WalkUp(t.TempDir())
		if err == nil || err.Error() != "could not find go.mod file" {
			t.Fatalf("expected could not find go.mod file error, got %v", err)
		}
	})
}

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
