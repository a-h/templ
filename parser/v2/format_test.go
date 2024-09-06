package parser

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
)

func TestFormatting(t *testing.T) {
	files, _ := filepath.Glob("formattestdata/*.txt")
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
			tem, err := ParseString(clean(a.Files[0].Data))
			if err != nil {
				t.Fatal(err)
			}
			var actual bytes.Buffer
			if err := tem.Write(&actual); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(string(a.Files[1].Data), actual.String()); diff != "" {
				t.Fatalf("%s:\n%s", file, diff)
			}
		})
	}
}

func clean(b []byte) string {
	b = bytes.ReplaceAll(b, []byte("$\n"), []byte("\n"))
	b = bytes.TrimSuffix(b, []byte("\n"))
	return string(b)
}
