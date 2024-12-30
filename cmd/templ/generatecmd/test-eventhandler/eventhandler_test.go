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
	"testing"

	"github.com/a-h/templ/cmd/templ/generatecmd"
	"github.com/a-h/templ/generator"
	"github.com/fsnotify/fsnotify"
	"github.com/google/go-cmp/cmp"
)

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
		// The raw files cannot end in .templ because they will cause the generator to fail. Instead,
		// we create a tmp file that ends in .templ only for the duration of the test.
		rawFile, err := os.Open(test.rawFileName)
		if err != nil {
			t.Errorf("%s: Failed to open file %s: %v", test.name, test.rawFileName, err)
			break
		}
		file, err := os.CreateTemp("", fmt.Sprintf("*%s.templ", test.rawFileName))
		if err != nil {
			t.Errorf("%s: Failed to create a tmp file at %s: %v", test.name, file.Name(), err)
			break
		}
		defer os.Remove(file.Name())
		if _, err = io.Copy(file, rawFile); err != nil {
			t.Errorf("%s: Failed to copy contents from raw file %s to tmp %s: %v", test.name, test.rawFileName, file.Name(), err)
		}

		event := fsnotify.Event{Name: file.Name(), Op: fsnotify.Write}
		_, err = fseh.HandleEvent(context.Background(), event)
		if err == nil {
			t.Errorf("%s: no error was thrown", test.name)
			break
		}
		list, ok := err.(scanner.ErrorList)
		for !ok {
			err = errors.Unwrap(err)
			if err == nil {
				t.Errorf("%s: reached end of error wrapping before finding an ErrorList", test.name)
				break
			} else {
				list, ok = err.(scanner.ErrorList)
			}
		}
		if !ok {
			break
		}

		if len(list) != len(test.errorPositions) {
			t.Errorf("%s: expected %d errors but got %d", test.name, len(test.errorPositions), len(list))
			break
		}
		for i, err := range list {
			test.errorPositions[i].Filename = file.Name()
			diff := cmp.Diff(test.errorPositions[i], err.Pos)
			if diff != "" {
				t.Error(diff)
				t.Error("expected:")
				t.Error(test.errorPositions[i])
				t.Error("actual:")
				t.Error(err.Pos)
			}
		}
	}
}
