package testcssexpression

import (
	"testing"

	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

var expected = templ.ComponentCSSClass{
	ID:    "className_f179",
	Class: templ.SafeCSS(`.className_f179{background-color:#ffffff;color:#ff0000;}`),
}

func TestCSSExpression(t *testing.T) {
	if diff := cmp.Diff(expected, className()); diff != "" {
		t.Error(diff)
	}
}
