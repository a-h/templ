package testgotemplates

import (
	"context"
	_ "embed"
	"os"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/a-h/templ/generator/htmldiff"
	"github.com/a-h/templ/internal/prettier"
)

//go:embed expected.html
var expected string

func TestExample(t *testing.T) {
	if !prettier.IsAvailable(prettier.DefaultCommand) {
		t.Skip("prettier is not available, skipping test")
	}

	// Create the templ component.
	templComponent := greeting()
	html, err := templ.ToGoHTML(context.Background(), templComponent)
	if err != nil {
		t.Fatalf("failed to convert to html: %v", err)
	}

	// Use it within the text/html template.
	b := new(strings.Builder)
	err = example.Execute(b, html)
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	// Compare the output with the expected.
	actual, diff, err := htmldiff.DiffStrings(expected, b.String())
	if err != nil {
		t.Fatalf("failed to diff strings: %v", err)
	}
	if diff != "" {
		if err := os.WriteFile("actual.html", []byte(actual), 0644); err != nil {
			t.Errorf("failed to write actual.html: %v", err)
		}
		t.Error(diff)
	}
}
