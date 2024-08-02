package onlyscripts

import (
	_ "embed"
	"testing"

	"github.com/a-h/templ/generator/htmldiff"
)

//go:embed expected.html
var expected string

func Test(t *testing.T) {
	script := withParameters("hello", "world", 42069)

	diff, err := htmldiff.Diff(script, expected)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "" {
		t.Error(diff)
	}
}
