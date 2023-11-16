package teststringerrs

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"testing"

	"github.com/a-h/templ/generator/htmldiff"
)

//go:embed expected.html
var expected string

func Test(t *testing.T) {
	component := render(false)

	diff, err := htmldiff.Diff(component, expected)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "" {
		t.Error(diff)
	}

	renderErr := render(true).Render(context.Background(), &bytes.Buffer{})
	if !errors.Is(renderErr, errSomethingBad) {
		t.Errorf("expected error: %v, but got %v", errSomethingBad, renderErr)
	}
}
