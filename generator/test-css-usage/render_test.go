package testcssusage

import (
	"context"
	_ "embed"
	"io"
	"strings"
	"testing"

	"github.com/a-h/htmlformat"
	"github.com/a-h/templ/generator/htmldiff"
	"github.com/google/go-cmp/cmp"
)

//go:embed expected.html
var expected string

func TestHTML(t *testing.T) {
	component := ThreeButtons()

	diff, err := htmldiff.Diff(component, expected)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "" {
	if err != nil {
		t.Error(diff)
	}
}
