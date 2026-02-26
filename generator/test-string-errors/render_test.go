package teststringerrs

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"testing"

	"github.com/a-h/templ"
	"github.com/a-h/templ/generator/htmldiff"
	"github.com/a-h/templ/internal/prettier"
)

//go:embed expected.html
var expected string

func Test(t *testing.T) {
	if !prettier.IsAvailable(prettier.DefaultCommand) {
		t.Skip("prettier is not available, skipping test")
	}

	t.Run("can render without error", func(t *testing.T) {
		component := TestComponent(nil)

		_, _, err := htmldiff.Diff(component, expected)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("string expressions can return errors", func(t *testing.T) {
		errSomethingBad := errors.New("bad error")

		err := TestComponent(errSomethingBad).Render(context.Background(), &bytes.Buffer{})
		if err == nil {
			t.Fatalf("expected error, but got nil")
		}

		t.Run("the errors are templ errors", func(t *testing.T) {
			var templateErr templ.Error
			if !errors.As(err, &templateErr) {
				t.Fatalf("expected error to be templ.Error, but got %T", err)
			}
			if templateErr.FileName != `generator/test-string-errors/template.templ` {
				t.Errorf("expected error in `generator/test-string-errors/template.templ`, but got %v", templateErr.FileName)
			}
			if templateErr.Line != 18 {
				t.Errorf("expected error on line 18, but got %v", templateErr.Line)
			}
			if templateErr.Col != 26 {
				t.Errorf("expected error on column 26, but got %v", templateErr.Col)
			}
		})

		t.Run("the underlying error can be unwrapped", func(t *testing.T) {
			if !errors.Is(err, errSomethingBad) {
				t.Errorf("expected error: %v, but got %v", errSomethingBad, err)
			}
		})

	})
}
