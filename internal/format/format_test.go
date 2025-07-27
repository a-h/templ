package format

import (
	"bytes"
	"path/filepath"
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
			a, err := txtar.ParseFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if len(a.Files) != 2 {
				t.Fatalf("expected 2 files, got %d", len(a.Files))
			}
			actual, _, err := Templ(clean(a.Files[0].Data), "")
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(string(a.Files[1].Data), string(actual)); diff != "" {
				t.Errorf("Expected:\n%s\nActual:\n%s\n", showWhitespace(string(a.Files[1].Data)), showWhitespace(string(actual)))
			}
		})
	}
}

func clean(b []byte) []byte {
	b = bytes.ReplaceAll(b, []byte("$\n"), []byte("\n"))
	b = bytes.TrimSuffix(b, []byte("\n"))
	return b
}
