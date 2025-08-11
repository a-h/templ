package format

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
)

func TestFormatting(t *testing.T) {
	files, _ := filepath.Glob("testdata/*.txt")
	if len(files) == 0 {
		t.Errorf("no test files found")
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			t.Parallel()
			a, err := txtar.ParseFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if len(a.Files) != 2 {
				t.Fatalf("expected 2 files, got %d", len(a.Files))
			}
			actual, _, err := Templ(a.Files[0].Data, "", Config{PrettierRequired: true})
			if err != nil {
				t.Fatalf("failed to format input: %v", err)
			}
			expected := string(a.Files[1].Data)
			if diff := cmp.Diff(expected, string(actual)); diff != "" {
				t.Errorf("Expected:\n%s\nActual:\n%s\n", showWhitespace(expected), showWhitespace(string(actual)))

				expectedLines := strings.Split(expected, "\n")
				actualLines := strings.Split(string(actual), "\n")
				if len(expectedLines) != len(actualLines) {
					t.Errorf("Expected %d lines, got %d lines", len(expectedLines), len(actualLines))
				}
			}
		})
	}
}

func showWhitespace(s string) string {
	s = strings.ReplaceAll(s, "\n", "⏎\n")
	s = strings.ReplaceAll(s, "\t", "→")
	s = strings.ReplaceAll(s, " ", "·")
	return s
}
