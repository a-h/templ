package testeventhandler

import (
	"context"
	"errors"
	"fmt"
	"go/scanner"
	"go/token"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/go-cmp/cmp"

	"github.com/a-h/templ/cmd/templ/generatecmd"
	"github.com/a-h/templ/generator"
)

func TestGoFileWritten(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	templContent := []byte(`package testgofilewritten

templ hello() {
	<div>Hello</div>
}
`)
	file, err := os.CreateTemp("", "*.templ")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	templFileName := file.Name()
	goFileName := strings.TrimSuffix(templFileName, ".templ") + "_templ.go"
	defer func() {
		_ = file.Close()
		if err := os.Remove(templFileName); err != nil {
			t.Logf("Warning: failed to remove temp file %s: %v", templFileName, err)
		}
		if err := os.Remove(goFileName); err != nil {
			t.Logf("Warning: failed to remove generated file %s: %v", goFileName, err)
		}
	}()

	if _, err = file.Write(templContent); err != nil {
		t.Fatalf("failed to write templ content: %v", err)
	}
	if err = file.Sync(); err != nil {
		t.Fatalf("failed to sync file: %v", err)
	}

	dir := filepath.Dir(templFileName)
	fseh := generatecmd.NewFSEventHandler(log, dir, false, []generator.GenerateOpt{}, false, false, generatecmd.FileWriter, false)

	t.Run("first generation writes the Go file", func(t *testing.T) {
		result, err := fseh.HandleEvent(context.Background(), fsnotify.Event{Name: templFileName, Op: fsnotify.Create})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.GoFileWritten {
			t.Fatal("expected GoFileWritten to be true on first generation")
		}
		if _, err := os.Stat(goFileName); err != nil {
			t.Fatalf("expected generated Go file to exist: %v", err)
		}
	})

	t.Run("second generation with unchanged content does not write", func(t *testing.T) {
		now := time.Now().Add(time.Second)
		if err := os.Chtimes(templFileName, now, now); err != nil {
			t.Fatalf("failed to update file times: %v", err)
		}
		result, err := fseh.HandleEvent(context.Background(), fsnotify.Event{Name: templFileName, Op: fsnotify.Write})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.GoFileWritten {
			t.Fatal("expected GoFileWritten to be false when content is unchanged")
		}
	})

	t.Run("fresh handler skips write when Go file on disk is up to date", func(t *testing.T) {
		goFileInfoBefore, err := os.Stat(goFileName)
		if err != nil {
			t.Fatalf("failed to stat generated Go file: %v", err)
		}

		now := time.Now().Add(2 * time.Second)
		if err := os.Chtimes(templFileName, now, now); err != nil {
			t.Fatalf("failed to update file times: %v", err)
		}

		freshHandler := generatecmd.NewFSEventHandler(log, dir, false, []generator.GenerateOpt{}, false, false, generatecmd.FileWriter, false)
		result, err := freshHandler.HandleEvent(context.Background(), fsnotify.Event{Name: templFileName, Op: fsnotify.Create})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.GoFileWritten {
			t.Fatal("expected GoFileWritten to be false when Go file on disk is already up to date")
		}

		goFileInfoAfter, err := os.Stat(goFileName)
		if err != nil {
			t.Fatalf("failed to stat generated Go file: %v", err)
		}
		if goFileInfoAfter.ModTime() != goFileInfoBefore.ModTime() {
			t.Fatal("expected Go file not to be rewritten")
		}
	})

	t.Run("fresh handler writes when templ content has changed", func(t *testing.T) {
		changedContent := []byte(`package testgofilewritten

templ hello() {
	<div>Hello, World</div>
}
`)
		if err := os.WriteFile(templFileName, changedContent, 0o644); err != nil {
			t.Fatalf("failed to write changed templ content: %v", err)
		}

		freshHandler := generatecmd.NewFSEventHandler(log, dir, false, []generator.GenerateOpt{}, false, false, generatecmd.FileWriter, false)
		result, err := freshHandler.HandleEvent(context.Background(), fsnotify.Event{Name: templFileName, Op: fsnotify.Create})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.GoFileWritten {
			t.Fatal("expected GoFileWritten to be true when templ content has changed")
		}
	})
}

// extractErrorList unwraps errors until it finds a scanner.ErrorList
func extractErrorList(err error) (scanner.ErrorList, bool) {
	if err == nil {
		return nil, false
	}

	if list, ok := err.(scanner.ErrorList); ok {
		return list, true
	}

	return extractErrorList(errors.Unwrap(err))
}

func TestErrorLocationMapping(t *testing.T) {
	tests := []struct {
		name           string
		rawFileName    string
		errorPositions []token.Position
	}{
		{
			name:        "single error outputs location in srcFile",
			rawFileName: "single_error.templ.error",
			errorPositions: []token.Position{
				{Offset: 46, Line: 3, Column: 20},
			},
		},
		{
			name:        "multiple errors all output locations in srcFile",
			rawFileName: "multiple_errors.templ.error",
			errorPositions: []token.Position{
				{Offset: 41, Line: 3, Column: 15},
				{Offset: 101, Line: 7, Column: 22},
				{Offset: 126, Line: 10, Column: 1},
			},
		},
	}

	slog := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	var fw generatecmd.FileWriterFunc
	fseh := generatecmd.NewFSEventHandler(slog, ".", false, []generator.GenerateOpt{}, false, false, fw, false)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// The raw files cannot end in .templ because they will cause the generator to fail. Instead,
			// we create a tmp file that ends in .templ only for the duration of the test.
			rawFile, err := os.Open(test.rawFileName)
			if err != nil {
				t.Fatalf("Failed to open file %s: %v", test.rawFileName, err)
			}
			defer func() {
				if err = rawFile.Close(); err != nil {
					t.Fatalf("Failed to close raw file %s: %v", test.rawFileName, err)
				}
			}()

			file, err := os.CreateTemp("", fmt.Sprintf("*%s.templ", test.rawFileName))
			if err != nil {
				t.Fatalf("Failed to create a tmp file at %s: %v", file.Name(), err)
			}
			tempFileName := file.Name()
			defer func() {
				_ = file.Close()
				if err := os.Remove(tempFileName); err != nil {
					t.Logf("Warning: Failed to remove tmp file %s: %v", tempFileName, err)
				}
			}()

			if _, err = io.Copy(file, rawFile); err != nil {
				t.Fatalf("Failed to copy contents from raw file %s to tmp %s: %v", test.rawFileName, tempFileName, err)
			}

			// Ensure file is synced to disk and file pointer is at the beginning
			if err = file.Sync(); err != nil {
				t.Fatalf("Failed to sync file: %v", err)
			}

			event := fsnotify.Event{Name: tempFileName, Op: fsnotify.Write}
			_, err = fseh.HandleEvent(context.Background(), event)
			if err == nil {
				t.Fatal("Expected an error but none was thrown")
			}

			list, ok := extractErrorList(err)
			if !ok {
				t.Fatal("Failed to extract ErrorList from error")
			}

			if len(list) != len(test.errorPositions) {
				t.Fatalf("Expected %d errors but got %d", len(test.errorPositions), len(list))
			}

			for i, err := range list {
				expected := test.errorPositions[i]
				expected.Filename = tempFileName

				if diff := cmp.Diff(expected, err.Pos); diff != "" {
					t.Errorf("Error position mismatch (-expected +actual):\n%s", diff)
				}
			}
		})
	}
}
