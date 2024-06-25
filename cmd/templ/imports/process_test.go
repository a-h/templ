package imports

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/a-h/templ/parser/v2"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
)

func TestFormatting(t *testing.T) {
	files, _ := filepath.Glob("testdata/*.txtar")
	if len(files) == 0 {
		t.Errorf("no test files found")
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			a, err := txtar.ParseFile(file)
			if err != nil {
				t.Fatalf("failed to parse txtar file: %v", err)
			}
			if len(a.Files) != 2 {
				t.Fatalf("expected 2 files, got %d", len(a.Files))
			}
			template, err := parser.ParseString(clean(a.Files[0].Data))
			if err != nil {
				t.Fatalf("failed to parse %v", err)
			}
			template.Filepath = a.Files[0].Name
			tf, err := Process(template)
			if err != nil {
				t.Fatalf("failed to process file: %v", err)
			}
			expected := string(a.Files[1].Data)
			actual := new(strings.Builder)
			if err := tf.Write(actual); err != nil {
				t.Fatalf("failed to write template file: %v", err)
			}
			if diff := cmp.Diff(expected, actual.String()); diff != "" {
				t.Errorf("%s:\n%s", file, diff)
				t.Errorf("expected:\n%s", showWhitespace(expected))
				t.Errorf("actual:\n%s", showWhitespace(actual.String()))
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

func clean(b []byte) string {
	b = bytes.ReplaceAll(b, []byte("$\n"), []byte("\n"))
	b = bytes.TrimSuffix(b, []byte("\n"))
	return string(b)
}
