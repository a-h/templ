package imports

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/a-h/templ/cmd/templ/testproject"
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

func TestImport(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
		return
	}

	tests := []struct {
		name       string
		src        string
		assertions func(t *testing.T, updated string)
	}{
		{
			name: "un-named imports are removed",
			src: `package main

import "fmt"
import "github.com/a-h/templ/cmd/templ/testproject/css-classes"

templ Page(count int) {
	{ fmt.Sprintf("%d", count) }
	{ cssclasses.Header }
}
`,
			assertions: func(t *testing.T, updated string) {
				if count := strings.Count(updated, "github.com/a-h/templ/cmd/templ/testproject/css-classes"); count != 0 {
					t.Errorf("expected un-named import to be removed, but got %d instance of it", count)
				}
			},
		},
		{
			name: "named imports are retained",
			src: `package main

import "fmt"
import  cssclasses "github.com/a-h/templ/cmd/templ/testproject/css-classes"

templ Page(count int) {
	{ fmt.Sprintf("%d", count) }
	{ cssclasses.Header }
}
`,
			assertions: func(t *testing.T, updated string) {
				if count := strings.Count(updated, "cssclasses \"github.com/a-h/templ/cmd/templ/testproject/css-classes\""); count != 1 {
					t.Errorf("expected named import to be retained, got %d instances of it", count)
				}
				if count := strings.Count(updated, "github.com/a-h/templ/cmd/templ/testproject/css-classes"); count != 1 {
					t.Errorf("expected one import, got %d", count)
				}
			},
		},
	}

	for _, test := range tests {
		// Create test project.
		dir, err := testproject.Create("github.com/a-h/templ/cmd/templ/testproject")
		if err != nil {
			t.Fatalf("failed to create test project: %v", err)
		}
		defer os.RemoveAll(dir)

		// Load the templates.templ file.
		filePath := path.Join(dir, "templates.templ")
		err = os.WriteFile(filePath, []byte(test.src), 0660)
		if err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		// Parse the new file.
		template, err := parser.Parse(filePath)
		if err != nil {
			t.Fatalf("failed to parse %v", err)
		}
		template.Filepath = filePath
		tf, err := Process(template)
		if err != nil {
			t.Fatalf("failed to process file: %v", err)
		}

		// Write it back out after processing.
		buf := new(strings.Builder)
		if err := tf.Write(buf); err != nil {
			t.Fatalf("failed to write template file: %v", err)
		}

		// Assert.
		test.assertions(t, buf.String())
	}
}
