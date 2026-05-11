package ignorefile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected Patterns
	}{
		{
			name:     "empty file returns no patterns",
			content:  "",
			expected: nil,
		},
		{
			name:     "comments and blank lines are skipped",
			content:  "# comment\n\n# another comment\n",
			expected: nil,
		},
		{
			name:     "non-comment lines are returned as patterns",
			content:  "# Ignore test fixtures.\ngenerator/test-*\n\n# Ignore vendor.\nvendor\n",
			expected: Patterns{"generator/test-*", "vendor"},
		},
		{
			name:     "leading and trailing whitespace is trimmed",
			content:  "  generator/test-*  \n",
			expected: Patterns{"generator/test-*"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), ".templignore")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}
			got, err := Parse(path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.expected) {
				t.Fatalf("got %d patterns, expected %d", len(got), len(tt.expected))
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("pattern %d: got %q, expected %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestParseMissingFile(t *testing.T) {
	got, err := Parse(filepath.Join(t.TempDir(), "nonexistent"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil patterns, got %v", got)
	}
}

func TestMatches(t *testing.T) {
	tests := []struct {
		name     string
		patterns Patterns
		path     string
		expected bool
	}{
		{
			name:     "nil patterns never match",
			patterns: nil,
			path:     "anything",
			expected: false,
		},
		{
			name:     "literal pattern matches the same path",
			patterns: Patterns{"vendor"},
			path:     "vendor",
			expected: true,
		},
		{
			name:     "glob pattern matches a directory",
			patterns: Patterns{"generator/test-*"},
			path:     "generator/test-foo",
			expected: true,
		},
		{
			name:     "glob pattern matches a file inside a matched directory",
			patterns: Patterns{"generator/test-*"},
			path:     "generator/test-foo/bar.templ",
			expected: true,
		},
		{
			name:     "glob pattern does not match unrelated paths",
			patterns: Patterns{"generator/test-*"},
			path:     "generator/real-code/bar.templ",
			expected: false,
		},
		{
			name:     "star matches files in the same directory",
			patterns: Patterns{"*.bak"},
			path:     "foo.bak",
			expected: true,
		},
		{
			name:     "star does not match files in subdirectories",
			patterns: Patterns{"*.bak"},
			path:     "dir/foo.bak",
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.patterns.Matches(tt.path)
			if got != tt.expected {
				t.Errorf("got %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestShouldSkipFunc(t *testing.T) {
	t.Run("missing file returns a function that never skips", func(t *testing.T) {
		skip, err := ShouldSkipFunc(t.TempDir(), ".templignore_fmt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if skip("anything") {
			t.Error("expected skip to return false")
		}
	})
	t.Run("existing file returns a function that matches patterns", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, ".templignore_fmt"), []byte("generator/test-*\n"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
		skip, err := ShouldSkipFunc(dir, ".templignore_fmt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !skip("generator/test-foo") {
			t.Error("expected skip to return true for matching path")
		}
		if skip("generator/real-code") {
			t.Error("expected skip to return false for non-matching path")
		}
	})
}
