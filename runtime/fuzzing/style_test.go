package fuzzing

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
	templruntime "github.com/a-h/templ/runtime"
	"github.com/a-h/templ/safehtml"
	"github.com/google/go-cmp/cmp"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
)

// styleTestcases seeds the style fuzz corpus with values that contain HTML
// character references, quotes that end an attribute value, markup, and CSS
// that ends a declaration, block, string, or comment early.
var styleTestcases = []string{
	"color:red",
	"font-family: 'Courier New', monospace",
	`font-family: "Courier New", monospace`,
	`background-image: url("a.png?x=1&y=2")`,
	"background-image: url(a.png?x=1&y=2)",
	"content: '&amp;'",
	"content: '&#39;'",
	"content: '&#x27'",
	"content: '&lt;/style&gt;'",
	"&",
	"&#",
	"&amp;lt;",
	`color: red" onmouseover="alert(document.domain)`,
	`color: red' onmouseover='alert(document.domain)`,
	"color: red></div><script>alert(document.domain)</script>",
	"</style><script>alert(document.domain)</script>",
	"<!-- -->",
	"red; position: fixed",
	"red} div {position: fixed",
	"red /* comment */",
	"red */ position: fixed; /*",
	`"a"; position: fixed; x: "b"`,
	`"a\"; position: fixed; x: \"b"`,
	`url("a"); position: fixed; x: url("b")`,
	`url(a) , url("b"); position: fixed; x: url("c")`,
	`url("/");position:fixed;x:url("/")`,
	`url('/');position:fixed;x:url('/')`,
	`url(/);position:fixed;x:url(/)`,
	`url(")`,
	`"`,
	`"a\"`,
	"url(javascript:alert(document.domain))",
	"expression(alert(document.domain))",
	"red !important",
	"--x: {;}",
	"@import url(https://example.com/)",
	"\\",
	`\"`,
	"\\000022",
	`"unterminated`,
	"'unterminated",
	" ",
	"\t;\n",
	";;",
	" ",
	"",
}

// sanitizedCSSProperties are properties whose values the sanitizer checks with
// different rules.
var sanitizedCSSProperties = []string{"color", "display", "font-family", "background-image"}

// FuzzStyleValue asserts that a fuzzed value of each type that a style
// attribute accepts, written through each path that templ uses to render a
// style attribute, can't add attributes to the element, and that the HTML5
// parser decodes the attribute to the CSS that templ produced. Any other value
// means that the CSS was not HTML-escaped exactly once. Spread attributes and
// attribute key expressions accept fewer types than constant names.
func FuzzStyleValue(f *testing.F) {
	for _, value := range styleTestcases {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, value string) {
		skipUnnormalizedInput(t, value)
		constantNameValues := []any{
			value,
			templ.SafeCSS(value),
			templ.KV(value, true),
			templ.KV(templ.SafeCSS(value), true),
			map[string]string{"color": value},
			map[string]string{value: "red"},
			map[string]templ.SafeCSSProperty{"color": templ.SafeCSSProperty(value)},
			map[string]templ.SafeCSSProperty{value: "red"},
			templ.KV("color", value),
			templ.KV(value, "red"),
			[]any{value, templ.SafeCSS(value)},
		}
		for _, v := range constantNameValues {
			assertStyleIsSanitizedCSS(t, "constant name", ConstantNameStyle(v), v)
		}
		for _, v := range []any{value, templ.SafeCSS(value), templ.KV(value, true)} {
			assertStyleIsSanitizedCSS(t, "spread attribute", SpreadStyle(v), v)
		}
		assertStyleIsSanitizedCSS(t, "attribute key expression", DynamicNameStyle(value), value)
		assertStyleIsSanitizedCSS(t, "attribute key expression", DynamicNameStyle(templ.SafeCSS(value)), templ.SafeCSS(value))
	})
}

// assertStyleIsSanitizedCSS asserts that the HTML5 parser decodes the style
// attribute that the component renders for value to the CSS that templ
// produces for value.
func assertStyleIsSanitizedCSS(t *testing.T, path string, component templ.Component, value any) {
	t.Helper()
	expected, err := templruntime.SanitizeStyleAttributeValues(value)
	if err != nil {
		t.Fatalf("%s with %T: failed to sanitize: %v", path, value, err)
	}
	actual := renderStyleAttribute(t, component)
	if actual != expected {
		t.Fatalf("%s with %T: expected the CSS parser to receive %q, got %q", path, value, expected, actual)
	}
}

// FuzzSafeCSSStyle asserts that the CSS parser receives a fuzzed templ.SafeCSS
// value exactly as written, with a semicolon appended if it's missing, through
// each path that templ uses to render a style attribute.
func FuzzSafeCSSStyle(f *testing.F) {
	for _, value := range styleTestcases {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, value string) {
		skipUnnormalizedInput(t, value)
		expected := value
		if strings.TrimSpace(value) == "" {
			expected = ""
		} else if !strings.HasSuffix(value, ";") {
			expected += ";"
		}
		components := []struct {
			path      string
			component templ.Component
		}{
			{"constant name", ConstantNameStyle(templ.SafeCSS(value))},
			{"constant name with templ.KeyValue", ConstantNameStyle(templ.KV(templ.SafeCSS(value), true))},
			{"spread attribute", SpreadStyle(templ.SafeCSS(value))},
			{"attribute key expression", DynamicNameStyle(templ.SafeCSS(value))},
		}
		for _, c := range components {
			actual := renderStyleAttribute(t, c.component)
			if actual != expected {
				t.Fatalf("%s: expected the CSS parser to receive %q, got %q", c.path, expected, actual)
			}
		}
	})
}

// FuzzStyleDeclaration asserts that a fuzzed CSS property name or value, in
// each type that the sanitizer checks as a property and value pair, is parsed
// by a CSS parser as exactly one declaration, so that it can't end the
// declaration and add another.
func FuzzStyleDeclaration(f *testing.F) {
	for _, value := range styleTestcases {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, value string) {
		skipUnnormalizedInput(t, value)
		for _, property := range sanitizedCSSProperties {
			for _, v := range []any{map[string]string{property: value}, templ.KV(property, value)} {
				assertDeclarationOfValue(t, renderStyleAttribute(t, ConstantNameStyle(v)), property, value)
			}
		}
		for _, v := range []any{map[string]string{value: "red"}, map[string]templ.SafeCSSProperty{value: "red"}, templ.KV(value, "red")} {
			style := renderStyleAttribute(t, ConstantNameStyle(v))
			declarations, ok := parseDeclarations(style)
			if !ok || len(declarations) != 1 {
				t.Fatalf("%T: property %q: expected one declaration, got %#v (valid: %v) from %q", v, value, declarations, ok, style)
			}
			if declarations[0].Property != asciiLower(value) && declarations[0].Property != asciiLower(safehtml.InnocuousPropertyName) {
				t.Fatalf("%T: expected property %q or %q, got %q", v, asciiLower(value), safehtml.InnocuousPropertyName, declarations[0].Property)
			}
		}
	})
}

// assertDeclarationOfValue asserts that style is parsed as exactly one
// declaration of property, and that its value is either the innocuous value, or
// parsed in the same way as value on its own.
func assertDeclarationOfValue(t *testing.T, style, property, value string) {
	t.Helper()
	declarations, ok := parseDeclarations(style)
	if !ok || len(declarations) != 1 || declarations[0].Property != property {
		t.Fatalf("%s: value %q: expected one declaration, got %#v (valid: %v) from %q", property, value, declarations, ok, style)
	}
	if declarations[0].Value == safehtml.InnocuousPropertyValue {
		return
	}
	expected, ok := parseDeclarations(property + ":" + value)
	if !ok {
		t.Fatalf("%s: value %q is not valid CSS, but was not replaced: %q", property, value, style)
	}
	if diff := cmp.Diff(expected, declarations); diff != "" {
		t.Fatalf("%s: value %q: declarations differ from the value on its own:\n%s", property, value, diff)
	}
}

// renderStyleAttribute renders the component, parses the output with an HTML5
// parser, and returns the style attribute of its only element, as decoded by
// the parser. It fails the test if the element has any other attribute.
func renderStyleAttribute(t *testing.T, component templ.Component) (style string) {
	t.Helper()
	buf := new(strings.Builder)
	if err := component.Render(context.Background(), buf); err != nil {
		t.Fatalf("failed to render: %v", err)
	}
	n := findElement(parseHTML(t, buf.String()), "div")
	if n == nil {
		t.Fatalf("element not found in %q", buf.String())
	}
	if len(n.Attr) != 1 || n.Attr[0].Key != "style" {
		t.Fatalf("expected only a style attribute, got %q", buf.String())
	}
	return n.Attr[0].Val
}

// declaration is a CSS declaration, as parsed by a CSS parser.
type declaration struct {
	Property string
	Value    string
}

// parseDeclarations parses s as the contents of a style attribute. ok is false
// if the CSS parser reports an error, because a browser drops a declaration
// that it can't parse, and recovers from the error by skipping to the next
// semicolon.
func parseDeclarations(s string) (declarations []declaration, ok bool) {
	p := css.NewParser(parse.NewInputString(s), true)
	for {
		gt, _, data := p.Next()
		switch gt {
		case css.ErrorGrammar:
			return declarations, errors.Is(p.Err(), io.EOF)
		case css.DeclarationGrammar, css.CustomPropertyGrammar:
			var value strings.Builder
			for _, token := range p.Values() {
				value.Write(token.Data)
			}
			declarations = append(declarations, declaration{Property: string(data), Value: value.String()})
		case css.CommentGrammar:
		default:
			return declarations, false
		}
	}
}
