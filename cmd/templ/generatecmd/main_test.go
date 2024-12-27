package generatecmd

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path"
	"regexp"
	"testing"

	"github.com/a-h/templ/cmd/templ/testproject"
)

func TestGenerate(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	t.Run("can generate a file in place", func(t *testing.T) {
		// templ generate -f templates.templ
		dir, err := testproject.Create("github.com/a-h/templ/cmd/templ/testproject")
		if err != nil {
			t.Fatalf("failed to create test project: %v", err)
		}
		defer os.RemoveAll(dir)

		// Delete the templates_templ.go file to ensure it is generated.
		err = os.Remove(path.Join(dir, "templates_templ.go"))
		if err != nil {
			t.Fatalf("failed to remove templates_templ.go: %v", err)
		}

		// Run the generate command.
		err = Run(context.Background(), log, Arguments{
			FileName: path.Join(dir, "templates.templ"),
		})
		if err != nil {
			t.Fatalf("failed to run generate command: %v", err)
		}

		// Check the templates_templ.go file was created.
		_, err = os.Stat(path.Join(dir, "templates_templ.go"))
		if err != nil {
			t.Fatalf("templates_templ.go was not created: %v", err)
		}
	})
}

func TestDefaultWatchPattern(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		matches bool
	}{
		{
			name:    "empty file names do not match",
			input:   "",
			matches: false,
		},
		{
			name:    "*_templ.txt matches, Windows",
			input:   `C:\Users\adrian\github.com\a-h\templ\cmd\templ\testproject\strings_templ.txt`,
			matches: true,
		},
		{
			name:    "*_templ.txt matches, Unix",
			input:   "/Users/adrian/github.com/a-h/templ/cmd/templ/testproject/strings_templ.txt",
			matches: true,
		},
		{
			name:    "*.templ files match, Windows",
			input:   `C:\Users\adrian\github.com\a-h\templ\cmd\templ\testproject\templates.templ`,
			matches: true,
		},
		{
			name:    "*.templ files match, Unix",
			input:   "/Users/adrian/github.com/a-h/templ/cmd/templ/testproject/templates.templ",
			matches: true,
		},
		{
			name:    "*_templ.go files match, Windows",
			input:   `C:\Users\adrian\github.com\a-h\templ\cmd\templ\testproject\templates_templ.go`,
			matches: true,
		},
		{
			name:    "*_templ.go files match, Unix",
			input:   "/Users/adrian/github.com/a-h/templ/cmd/templ/testproject/templates_templ.go",
			matches: true,
		},
		{
			name:    "*.go files match, Windows",
			input:   `C:\Users\adrian\github.com\a-h\templ\cmd\templ\testproject\templates.go`,
			matches: true,
		},
		{
			name:    "*.go files match, Unix",
			input:   "/Users/adrian/github.com/a-h/templ/cmd/templ/testproject/templates.go",
			matches: true,
		},
		{
			name:    "*.css files do not match",
			input:   "/Users/adrian/github.com/a-h/templ/cmd/templ/testproject/templates.css",
			matches: false,
		},
	}
	wpRegexp, err := regexp.Compile(defaultWatchPattern)
	if err != nil {
		t.Fatalf("failed to compile default watch pattern: %v", err)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test := test
			t.Parallel()
			if wpRegexp.MatchString(test.input) != test.matches {
				t.Fatalf("expected match of %q to be %v", test.input, test.matches)
			}
		})
	}
}
