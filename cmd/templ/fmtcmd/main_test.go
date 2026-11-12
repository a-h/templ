package fmtcmd

import (
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
)

//go:embed testdata.txtar
var testDataTxTar []byte

type testProject struct {
	dir       string
	cleanup   func() error
	testFiles map[string]testFile
}

type testFile struct {
	name            string
	input, expected string
}

func setupProjectDir() (tp testProject, err error) {
	tp.dir, err = os.MkdirTemp("", "fmtcmd_test_*")
	if err != nil {
		return tp, fmt.Errorf("failed to make test dir: %w", err)
	}
	tp.testFiles = make(map[string]testFile)
	testData := txtar.Parse(testDataTxTar)
	for i := 0; i < len(testData.Files); i += 2 {
		file := testData.Files[i]
		err = os.WriteFile(filepath.Join(tp.dir, file.Name), file.Data, 0660)
		if err != nil {
			return tp, fmt.Errorf("failed to write file: %w", err)
		}
		tp.testFiles[file.Name] = testFile{
			name:     filepath.Join(tp.dir, file.Name),
			input:    string(file.Data),
			expected: string(testData.Files[i+1].Data),
		}
	}
	tp.cleanup = func() error {
		if err := os.RemoveAll(tp.dir); err != nil {
			return fmt.Errorf("failed to remove test directory %q: %w", tp.dir, err)
		}
		return nil
	}
	return tp, nil
}

func TestFormat(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	t.Run("can format a single file from stdin to stdout", func(t *testing.T) {
		tp, err := setupProjectDir()
		if err != nil {
			t.Fatalf("failed to setup project dir: %v", err)
		}
		defer func() {
			if err := tp.cleanup(); err != nil {
				t.Errorf("cleanup error: %v", err)
			}
		}()
		stdin := strings.NewReader(tp.testFiles["a.templ"].input)
		stdout := new(strings.Builder)
		if err = Run(log, stdin, stdout, Arguments{
			ToStdout: true,
		}); err != nil {
			t.Fatalf("failed to run format command: %v", err)
		}
		if diff := cmp.Diff(tp.testFiles["a.templ"].expected, stdout.String()); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("can process a single file to stdout", func(t *testing.T) {
		tp, err := setupProjectDir()
		if err != nil {
			t.Fatalf("failed to setup project dir: %v", err)
		}
		defer func() {
			if err := tp.cleanup(); err != nil {
				t.Errorf("cleanup error: %v", err)
			}
		}()
		stdout := new(strings.Builder)
		if err = Run(log, nil, stdout, Arguments{
			ToStdout: true,
			Files: []string{
				tp.testFiles["a.templ"].name,
			},
			FailIfChanged: false,
		}); err != nil {
			t.Fatalf("failed to run format command: %v", err)
		}
		if diff := cmp.Diff(tp.testFiles["a.templ"].expected, stdout.String()); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("can process a single file in place", func(t *testing.T) {
		tp, err := setupProjectDir()
		if err != nil {
			t.Fatalf("failed to setup project dir: %v", err)
		}
		defer func() {
			if err := tp.cleanup(); err != nil {
				t.Errorf("cleanup error: %v", err)
			}
		}()
		if err = Run(log, nil, nil, Arguments{
			Files: []string{
				tp.testFiles["a.templ"].name,
			},
			FailIfChanged: false,
		}); err != nil {
			t.Fatalf("failed to run format command: %v", err)
		}
		data, err := os.ReadFile(tp.testFiles["a.templ"].name)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if diff := cmp.Diff(tp.testFiles["a.templ"].expected, string(data)); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("fails when fail flag used and change occurs", func(t *testing.T) {
		tp, err := setupProjectDir()
		if err != nil {
			t.Fatalf("failed to setup project dir: %v", err)
		}
		defer func() {
			if err := tp.cleanup(); err != nil {
				t.Errorf("cleanup error: %v", err)
			}
		}()
		if err = Run(log, nil, nil, Arguments{
			Files: []string{
				tp.testFiles["a.templ"].name,
			},
			FailIfChanged: true,
		}); err == nil {
			t.Fatal("command should have exited with an error and did not")
		}
		data, err := os.ReadFile(tp.testFiles["a.templ"].name)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if diff := cmp.Diff(tp.testFiles["a.templ"].expected, string(data)); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("passes when fail flag used and no change occurs", func(t *testing.T) {
		tp, err := setupProjectDir()
		if err != nil {
			t.Fatalf("failed to setup project dir: %v", err)
		}
		defer func() {
			if err := tp.cleanup(); err != nil {
				t.Errorf("cleanup error: %v", err)
			}
		}()
		if err = Run(log, nil, nil, Arguments{
			Files: []string{
				tp.testFiles["c.templ"].name,
			},
			FailIfChanged: true,
		}); err != nil {
			t.Fatalf("failed to run format command: %v", err)
		}
		data, err := os.ReadFile(tp.testFiles["c.templ"].name)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if diff := cmp.Diff(tp.testFiles["c.templ"].expected, string(data)); diff != "" {
			t.Error(diff)
		}
	})
}

func TestFormatRespectsIgnoreFile(t *testing.T) {
	for _, tt := range []struct {
		name     string
		absolute bool
		file     bool
	}{
		{name: "relative directory"},
		{name: "absolute directory", absolute: true},
		{name: "relative file", file: true},
		{name: "absolute file", absolute: true, file: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			base := t.TempDir()
			t.Chdir(base)
			if err := os.MkdirAll(filepath.Join("views", "fixtures"), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join("views", ".templignore_fmt"), []byte("ignored.templ\nfixtures\n"), 0644); err != nil {
				t.Fatal(err)
			}
			const input = "package test\n\ntempl example() {\n  <div>Example</div>\n}\n"
			const formatted = "package test\n\ntempl example() {\n\t<div>Example</div>\n}\n"
			files := []string{"ignored.templ", filepath.Join("fixtures", "example.templ"), "included.templ"}
			for _, name := range files {
				if err := os.WriteFile(filepath.Join("views", name), []byte(input), 0644); err != nil {
					t.Fatal(err)
				}
			}
			path := "views"
			if tt.file {
				path = filepath.Join(path, "ignored.templ")
			}
			if tt.absolute {
				path = filepath.Join(base, path)
			}
			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			if err := Run(log, nil, io.Discard, Arguments{Files: []string{path}, PrettierCommand: "cat"}); err != nil {
				t.Fatal(err)
			}
			for _, name := range files {
				got, err := os.ReadFile(filepath.Join("views", name))
				if err != nil {
					t.Fatal(err)
				}
				want := input
				if name == "included.templ" && !tt.file {
					want = formatted
				}
				if diff := cmp.Diff(want, string(got)); diff != "" {
					t.Errorf("%s (-want +got):\n%s", name, diff)
				}
			}
		})
	}
}
