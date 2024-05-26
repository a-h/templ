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
	cleanup   func()
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
	tp.cleanup = func() {
		os.RemoveAll(tp.dir)
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
		defer tp.cleanup()
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
		defer tp.cleanup()
		stdout := new(strings.Builder)
		if err = Run(log, nil, stdout, Arguments{
			ToStdout: true,
			Files: []string{
				tp.testFiles["a.templ"].name,
			},
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
		defer tp.cleanup()
		if err = Run(log, nil, nil, Arguments{
			Files: []string{
				tp.testFiles["a.templ"].name,
			},
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
}
