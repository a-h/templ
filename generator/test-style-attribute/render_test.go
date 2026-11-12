package teststyleattribute

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/a-h/templ/generator/htmldiff"
)

//go:embed expected.html
var expected string

func TestSafeCSSPropertyKeyValue(t *testing.T) {
	t.Parallel()

	style := templ.KV("color", templ.SafeCSSProperty("red"))
	tests := []struct {
		name  string
		style any
	}{
		{name: "key value", style: style},
		{name: "slice", style: []templ.KeyValue[string, templ.SafeCSSProperty]{style}},
		{name: "function", style: func() templ.KeyValue[string, templ.SafeCSSProperty] { return style }},
	}
	const want = `<button style="color:red;">Click me</button> <button style="background-color: red;">Click me</button> <button STYLE="color:red;">Click me</button>`
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var output strings.Builder
			if err := Button(tt.style, "Click me").Render(t.Context(), &output); err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}

func Test(t *testing.T) {
	var stringCSS = "background-color:blue;color:red"
	var safeCSS = templ.SafeCSS("background-color:blue;color:red;")
	var mapStringString = map[string]string{
		"color":            "red",
		"background-color": "blue",
	}
	var mapStringSafeCSSProperty = map[string]templ.SafeCSSProperty{
		"color":            templ.SafeCSSProperty("red"),
		"background-color": templ.SafeCSSProperty("blue"),
	}
	var kvStringStringSlice = []templ.KeyValue[string, string]{
		templ.KV("background-color", "blue"),
		templ.KV("color", "red"),
	}
	var kvStringBoolSlice = []templ.KeyValue[string, bool]{
		templ.KV("background-color:blue", true),
		templ.KV("color:red", true),
		templ.KV("color:blue", false),
	}
	var kvSafeCSSBoolSlice = []templ.KeyValue[templ.SafeCSS, bool]{
		templ.KV(templ.SafeCSS("background-color:blue"), true),
		templ.KV(templ.SafeCSS("color:red"), true),
		templ.KV(templ.SafeCSS("color:blue"), false),
	}

	tests := []any{
		stringCSS,
		safeCSS,
		mapStringString,
		mapStringSafeCSSProperty,
		kvStringStringSlice,
		kvStringBoolSlice,
		kvSafeCSSBoolSlice,
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%T", test), func(t *testing.T) {
			component := Button(test, "Click me")

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
		})
	}
}
