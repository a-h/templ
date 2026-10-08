package fuzzing

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/a-h/templ"
	"github.com/a-h/templ/internal/htmlattr"
	"golang.org/x/net/html"
)

// attributeTestcases seeds the fuzz corpus with known-tricky attribute names
// and values: javascript: URLs, including the forms that browsers normalize,
// attribute names that would add attributes, and the names of attributes that
// require sanitization (GHSA-94c2-xg24-pwhv).
var attributeTestcases = []struct {
	name  string
	value string
}{
	{"title", "hello"},
	{"href", "javascript:alert(document.domain)"},
	{"HREF", "JaVaScRiPt:alert(document.domain)"},
	{"src", " javascript:alert(document.domain)"},
	{"formaction", "java\tscript:alert(document.domain)"},
	{"action", "\x01javascript:alert(document.domain)"},
	{"data", "vbscript:msgbox(1)"},
	{"to", "https://example.com;javascript:alert(document.domain)"},
	{"values", "0;javascript:alert(document.domain)"},
	{"srcdoc", "<script>alert(document.domain)</script>"},
	{"onclick", "alert(document.domain)"},
	{"OnClick", "alert(document.domain)"},
	{"hx-on:click", "alert(document.domain)"},
	{"hx-on-click", "alert(document.domain)"},
	{"data-hx-on:click", "alert(document.domain)"},
	{"x onmouseover=alert(document.domain) y", "value"},
	{`x" onmouseover="alert(document.domain)`, "value"},
	{"x>", "<script>alert(document.domain)</script>"},
	{"x/", "value"},
	{"x&y", "value"},
	{"", "value"},
	{"title", `" onmouseover="alert(document.domain)`},
}

var attributeElementNames = []string{"div", "iframe", "form", "set"}

var attributeNameCases = []struct {
	name      string
	component func(elementName, name, value string) templ.Component
}{
	{"attribute key expression", DynamicAttribute},
	{"spread attribute", SpreadAttribute},
}

// FuzzAttributeName asserts that, for attribute key expressions and spread
// attributes on elements where attributes have different meanings, a fuzzed
// attribute name can't add attributes to the element, and the attribute value
// is inert, as parsed by an HTML5 parser. Any additional attribute, or any
// executable value, is the shape of GHSA-94c2-xg24-pwhv.
func FuzzAttributeName(f *testing.F) {
	for _, tc := range attributeTestcases {
		f.Add(tc.name, tc.value)
	}
	f.Fuzz(func(t *testing.T, name, value string) {
		skipUnnormalizedInput(t, name)
		skipUnnormalizedInput(t, value)
		buf := new(strings.Builder)
		for _, tc := range attributeNameCases {
			for _, elementName := range attributeElementNames {
				buf.Reset()
				if err := tc.component(elementName, name, value).Render(context.Background(), buf); err != nil {
					t.Fatalf("%s on %s: failed to render: %v", tc.name, elementName, err)
				}
				n := findElement(parseHTML(t, buf.String()), elementName)
				if n == nil {
					t.Fatalf("%s on %s: element not found in %q", tc.name, elementName, buf.String())
				}
				if len(n.Attr) != 1 {
					t.Fatalf("%s on %s: name %q added attributes: %q", tc.name, elementName, name, buf.String())
				}
				expectedName := templ.FailedSanitizationAttributeName
				if htmlattr.IsValidName(name) {
					expectedName = asciiLower(name)
				}
				if n.Attr[0].Key != expectedName {
					t.Fatalf("%s on %s: expected attribute name %q, got %q in %q", tc.name, elementName, expectedName, n.Attr[0].Key, buf.String())
				}
				assertAttributeValueIsInert(t, elementName, n.Attr[0].Key, n.Attr[0].Val, value)
			}
		}
	})
}

// FuzzAttributeValue asserts that a fuzzed value written to each attribute
// with a constant name that requires sanitization is inert, as parsed by an
// HTML5 parser (GHSA-94c2-xg24-pwhv).
func FuzzAttributeValue(f *testing.F) {
	for _, tc := range attributeTestcases {
		f.Add(tc.value)
	}
	f.Fuzz(func(t *testing.T, value string) {
		skipUnnormalizedInput(t, value)
		buf := new(strings.Builder)
		if err := ConstantNameAttributes(value).Render(context.Background(), buf); err != nil {
			t.Fatalf("failed to render: %v", err)
		}
		var checked int
		for n := range parseHTML(t, buf.String()).Descendants() {
			if n.Type != html.ElementNode {
				continue
			}
			for _, attr := range n.Attr {
				if attr.Key == "attributename" {
					continue
				}
				assertAttributeValueIsInert(t, n.Data, attr.Key, attr.Val, value)
				checked++
			}
		}
		if checked != 10 {
			t.Fatalf("expected 10 attributes to be checked, checked %d in %q", checked, buf.String())
		}
	})
}

// skipUnnormalizedInput skips inputs that an HTML5 parser normalizes, so that
// they can't be expected to round trip byte-for-byte. The parser replaces NUL
// with U+FFFD, and CR and CRLF with LF, which is a property of HTML parsing
// rather than of the escaping being exercised here.
func skipUnnormalizedInput(t *testing.T, s string) {
	t.Helper()
	if !utf8.ValidString(s) {
		t.Skip("input is not valid UTF-8")
	}
	if strings.ContainsAny(s, "\x00\r") {
		t.Skip("input contains characters that the HTML parser normalizes")
	}
}

// assertAttributeValueIsInert asserts that the value of the attribute, as
// parsed by an HTML5 parser, can't execute JavaScript, and that values that
// don't require sanitization round trip unchanged.
func assertAttributeValueIsInert(t *testing.T, elementName, name, parsed, input string) {
	t.Helper()
	assertBrowserCannotExecute(t, elementName, name, parsed)
	switch htmlattr.Classify(elementName, name) {
	case htmlattr.ContextEventHandler:
		if parsed != templ.FailedSanitizationJS {
			t.Fatalf("<%s %s>: expected event handler to be replaced, got %q", elementName, name, parsed)
		}
	case htmlattr.ContextURL:
		if isJavaScriptURL(parsed) {
			t.Fatalf("<%s %s>: javascript URL was not sanitized: %q", elementName, name, parsed)
		}
		if parsed != input && parsed != string(templ.FailedSanitizationURL) {
			t.Fatalf("<%s %s>: expected %q or %q, got %q", elementName, name, input, templ.FailedSanitizationURL, parsed)
		}
	case htmlattr.ContextAnimationValue:
		for item := range strings.SplitSeq(parsed, ";") {
			if isJavaScriptURL(item) {
				t.Fatalf("<%s %s>: javascript URL was not sanitized: %q", elementName, name, parsed)
			}
		}
	case htmlattr.ContextSrcdoc:
		for n := range parseHTML(t, parsed).Descendants() {
			if n.Type == html.TextNode {
				continue
			}
			if n.Type == html.ElementNode && (n.Data == "html" || n.Data == "head" || n.Data == "body") {
				continue
			}
			t.Fatalf("<%s %s>: srcdoc contains markup: input %q, srcdoc %q", elementName, name, input, parsed)
		}
	default:
		if parsed != input {
			t.Fatalf("<%s %s>: round trip failed: input %q, got back %q", elementName, name, input, parsed)
		}
	}
}

// assertBrowserCannotExecute asserts that a browser, or htmx, can't execute the
// value of the attribute as JavaScript. It's written independently of
// htmlattr.Classify, as a statement of how browsers treat attribute values, so
// that a missing classification in htmlattr fails the fuzz tests, rather than
// changing what they expect.
func assertBrowserCannotExecute(t *testing.T, elementName, name, parsed string) {
	t.Helper()
	switch {
	case strings.HasPrefix(name, "on"), strings.HasPrefix(name, "hx-on"), strings.HasPrefix(name, "data-hx-on"):
		if parsed != templ.FailedSanitizationJS {
			t.Fatalf("<%s %s>: event handler can execute %q", elementName, name, parsed)
		}
	case name == "href", name == "xlink:href",
		name == "formaction" && (elementName == "button" || elementName == "input"),
		name == "action" && elementName == "form",
		name == "data" && elementName == "object",
		name == "src" && (elementName == "iframe" || elementName == "frame" || elementName == "embed"):
		if isJavaScriptURL(parsed) {
			t.Fatalf("<%s %s>: browser can navigate to %q", elementName, name, parsed)
		}
	case (elementName == "set" || elementName == "animate") && (name == "to" || name == "from" || name == "values" || name == "by"):
		for item := range strings.SplitSeq(parsed, ";") {
			if isJavaScriptURL(item) {
				t.Fatalf("<%s %s>: animation can set href to %q", elementName, name, parsed)
			}
		}
	case name == "srcdoc" && elementName == "iframe":
		for n := range parseHTML(t, parsed).Descendants() {
			if n.Type == html.ElementNode && n.Data != "html" && n.Data != "head" && n.Data != "body" {
				t.Fatalf("<%s %s>: srcdoc contains the %s element: %q", elementName, name, n.Data, parsed)
			}
		}
	}
}

// isJavaScriptURL returns true if a browser would treat the URL as a
// javascript: or vbscript: URL. The WHATWG URL parser strips leading and
// trailing C0 control characters and spaces, and removes all tabs and
// newlines, before parsing the scheme.
func isJavaScriptURL(s string) bool {
	s = strings.TrimFunc(s, isC0ControlOrSpace)
	s = strings.NewReplacer("\t", "", "\n", "", "\r", "").Replace(s)
	scheme, _, ok := strings.Cut(s, ":")
	if !ok {
		return false
	}
	scheme = asciiLower(scheme)
	return scheme == "javascript" || scheme == "vbscript"
}

func isC0ControlOrSpace(r rune) bool {
	return r <= 0x20
}

// asciiLower lowercases ASCII letters only, in the same way as the HTML5
// tokenizer lowercases attribute names, and the URL parser lowercases schemes.
func asciiLower(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if 'A' <= r && r <= 'Z' {
			r += 'a' - 'A'
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

func parseHTML(t *testing.T, s string) *html.Node {
	t.Helper()
	n, err := html.Parse(strings.NewReader(s))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}
	return n
}

func findElement(root *html.Node, elementName string) *html.Node {
	for n := range root.Descendants() {
		if n.Type == html.ElementNode && n.Data == elementName {
			return n
		}
	}
	return nil
}
