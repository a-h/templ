package testgotemplates

import (
	"context"
	_ "embed"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/a-h/templ/generator/htmldiff"
)

//go:embed expected.html
var expected string

func TestExample(t *testing.T) {
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
	diff, err := htmldiff.DiffStrings(expected, b.String())
	if err != nil {
		t.Fatalf("failed to diff strings: %v", err)
	}
	if diff != "" {
		t.Error(diff)
	}
}
