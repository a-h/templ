package testrawelements

import (
	_ "embed"
	"testing"

	"github.com/a-h/templ/generator/htmldiff"
)

//go:embed expected.html
var expected string

//go:embed expected_with_minification.html
var expectedWithMinification string

func Test(t *testing.T) {
	component := Example()
	diff, err := htmldiff.Diff(component, expected)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "" {
		t.Error(diff)
	}
}

func TestWithJSMinification(t *testing.T) {
	component := ExampleWithMinification()
	diff, err := htmldiff.Diff(component, expectedWithMinification)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "" {
		t.Error(diff)
	}
}
