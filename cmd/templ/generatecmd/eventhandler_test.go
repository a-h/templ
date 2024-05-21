package generatecmd

import (
	"context"
	"errors"
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
			fileName: "single_error.templ.error",
			errorPositions: []token.Position{
				{Filename: "single_error.templ.error", Offset: 41, Line: 3, Column: 20},
			},
		},
		{
			name:     "multiple errors all output locations in srcFile",
			fileName: "multiple_errors.templ.error",
			errorPositions: []token.Position{
				{Filename: "multiple_errors.templ.error", Offset: 36, Line: 3, Column: 15},
				{Filename: "multiple_errors.templ.error", Offset: 96, Line: 7, Column: 22},
				{Filename: "multiple_errors.templ.error", Offset: 121, Line: 10, Column: 1},
			},
		},
	}

	slog := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	fseh := NewFSEventHandler(slog, ".", false, []generator.GenerateOpt{}, false, false, true)
	for _, test := range tests {
		_, _, _, err := fseh.generate(context.Background(), test.fileName)
		if err == nil {
			t.Errorf("%s: no error was thrown", test.name)
			break
		}

		tmp := err
		err = errors.Unwrap(err)
		if err == nil {
			t.Errorf("%s: thrown error could not be unwrapped %s", test.name, tmp.Error())
			break
		}
		list, ok := err.(scanner.ErrorList)
		if !ok {
			t.Errorf("%s: unwrapped error is not of type scanner.ErrorList %s", test.name, err.Error())
			break
		}

		if len(list) != len(test.errorPositions) {
			t.Errorf("%s: expected %d errors but got %d", test.name, len(test.errorPositions), len(list))
			break
		}

		for i, err := range list {
			if err.Pos != test.errorPositions[i] {
				t.Errorf("%s: got error %s at pos %v, expected this error at pos %v", test.name, err.Msg, err.Pos, test.errorPositions[i])
			}
		}
	}
}
