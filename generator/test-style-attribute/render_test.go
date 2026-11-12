package teststyleattribute

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/a-h/templ/generator/htmldiff"
	"github.com/google/go-cmp/cmp"
)

//go:embed expected.html
var expected string

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

func TestQuotedStyleValuesAreEscapedOnce(t *testing.T) {
	tests := []struct {
		name     string
		style    any
		expected string
	}{
		{
			name:     "string quotes are HTML-escaped once",
			style:    "font-family: 'Courier New', monospace",
			expected: `<button style="font-family: &#39;Courier New&#39;, monospace;">Click me</button>`,
		},
		{
			name:     "SafeCSS quotes and ampersands are HTML-escaped once",
			style:    templ.SafeCSS(`background-image: url("a.png?x=1&y=2")`),
			expected: `<button style="background-image: url(&#34;a.png?x=1&amp;y=2&#34;);">Click me</button>`,
		},
		{
			name:     "map[string]templ.SafeCSSProperty quotes are HTML-escaped once",
			style:    map[string]templ.SafeCSSProperty{"font-family": "'Courier New', monospace"},
			expected: `<button style="font-family:&#39;Courier New&#39;, monospace;">Click me</button>`,
		},
		{
			name:     "KeyValue[templ.SafeCSS, bool] quotes are HTML-escaped once",
			style:    templ.KV(templ.SafeCSS("font-family: 'Courier New'"), true),
			expected: `<button style="font-family: &#39;Courier New&#39;;">Click me</button>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			if err := QuotedButton(tt.style).Render(context.Background(), &sb); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.expected, sb.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}
