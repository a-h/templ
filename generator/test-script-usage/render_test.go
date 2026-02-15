package testscriptusage

import (
	_ "embed"
	"os"
	"testing"

	"github.com/a-h/templ/generator/htmldiff"
	"github.com/a-h/templ/internal/prettier"
)

//go:embed expected.html
var expected string

func Test(t *testing.T) {
	if !prettier.IsAvailable(prettier.DefaultCommand) {
		t.Skip("prettier is not available, skipping test")
	}

	component := ThreeButtons()

	actual, diff, err := htmldiff.Diff(component, expected)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "" {
		if err := os.WriteFile("actual.html", []byte(actual), 0644); err != nil {
			t.Errorf("failed to write actual.html: %v", err)
		}
		t.Error(diff)
	}
}
