package testcssexpression

import (
	"testing"

	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

var expected = templ.ComponentCSSClass{
	ID:    "className_34fc0328",
	Class: templ.SafeCSS(`.className_34fc0328{background-color:#ffffff;max-height:calc(100vh - 170px);color:#ff0000;}`),
}

func TestCSSExpression(t *testing.T) {
	if diff := cmp.Diff(expected, className()); diff != "" {
		t.Error(diff)
	}
}
