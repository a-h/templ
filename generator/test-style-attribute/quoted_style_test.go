package teststyleattribute

import (
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
)

// TestQuotedStyleValue checks the value the browser actually hands to the CSS
// parser. Attribute values are HTML-decoded before they reach CSS, so any
// entity left in the value (&#39;, &amp;, ...) makes the declaration invalid
// and the browser drops it.
func TestQuotedStyleValue(t *testing.T) {
	const want = "font-family: 'Courier New', monospace;"

	tests := []struct {
		name      string
		component templ.Component
	}{
		{"safe css", QuotedSafeCSS()},
		{"string", QuotedString()},
		{"spread attributes", QuotedAttributes(templ.Attributes{
			"style": templ.SafeCSS("font-family: 'Courier New', monospace"),
		})},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var sb strings.Builder
			if err := test.component.Render(context.Background(), &sb); err != nil {
				t.Fatal(err)
			}
			got, ok := styleAttribute(t, sb.String())
			if !ok {
				t.Fatalf("no style attribute rendered in %q", sb.String())
			}
			if got != want {
				t.Errorf("CSS parser sees %q, want %q", got, want)
			}
		})
	}
}

// styleAttribute returns the decoded style attribute of the first element that
// has one, which is the same string a browser passes to the CSS parser.
func styleAttribute(t *testing.T, rendered string) (value string, ok bool) {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(rendered))
	if err != nil {
		t.Fatal(err)
	}
	for n := range doc.Descendants() {
		if n.Type != html.ElementNode {
			continue
		}
		for _, attr := range n.Attr {
			if attr.Key == "style" {
				return attr.Val, true
			}
		}
	}
	return "", false
}
