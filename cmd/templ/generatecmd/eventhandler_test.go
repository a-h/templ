package generatecmd

import (
	"context"
	"go/scanner"
	"go/token"
	"io"
	"log/slog"
	"testing"

	"github.com/a-h/templ/generator"
)

func TestEventHandler(t *testing.T) {
	tests := []struct {
		name           string
		fileName       string
		errorPositions []token.Position
	}{
		{
			name:     "single error outputs location in srcFile",
			fileName: "single_error.templ",
			errorPositions: []token.Position{
				{Filename: "single_error.templ", Offset: 244, Line: 3, Column: 20},
			},
		},
		{
			name:     "multiple errors all output locations in srcFile",
			fileName: "multiple_errors.templ",
			errorPositions: []token.Position{
				{Filename: "multiple_errors.templ", Offset: 240, Line: 3, Column: 15},
				{Filename: "multiple_errors.templ", Offset: 299, Line: 7, Column: 22},
			},
		},
	}

	slog := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	fseh := NewFSEventHandler(slog, ".", false, []generator.GenerateOpt{}, false, false, true)
	for _, test := range tests {
		_, _, _, err := fseh.generate(context.Background(), test.fileName)
		if err == nil {
			t.Errorf("No error was thrown for file %s", test.fileName)
		}

		list, ok := err.(scanner.ErrorList)
		if !ok {
			t.Errorf("Error is not of type scanner.ErrorList")
		}

		if len(list) != len(test.errorPositions) {
			t.Errorf("Expected %d errors but got %d", len(test.errorPositions), len(list))
		}

		for i, err := range list {
			if err.Pos != test.errorPositions[i] {
				t.Errorf("Got error %s at pos %v, expected this error at pos %v", err.Msg, err.Pos, test.errorPositions[i])
			}
		}

	}
}
