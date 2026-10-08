package fuzzing

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/a-h/templ"
	"golang.org/x/net/html"

	v8 "rogchap.com/v8go"
)

var iso = v8.NewIsolate()

// testcases seeds the fuzz corpus with known-tricky inputs: HTML/script
// structural tokens, and JavaScript template literal substitution syntax
// (GHSA-jq3r-rjqp-mwgj).
var testcases = []string{
	"hello",
	"<script>console.log('hello')</script>",
	"text data\n can contain all sorts of \"characters\" & symbols",
	"123",
	"</script>",
	"${alert(document.domain)}",
	"${ {} }",
	"$",
	"{",
	"}",
	"`",
}

var roundTripCases = []struct {
	name      string
	component func(string) templ.Component
}{
	{"outside string literal", func(v string) templ.Component { return StringOutsideStringLiteral(v) }},
	{"inside double quoted string", func(v string) templ.Component { return StringInsideDoubleQuotedString(v) }},
	{"inside single quoted string", func(v string) templ.Component { return StringInsideSingleQuotedString(v) }},
	{"inside template literal", func(v string) templ.Component { return StringInsideTemplateLiteral(v) }},
	{"inside template literal substitution", func(v string) templ.Component { return StringInsideTemplateLiteralSubstitution(v) }},
}

// FuzzScriptStringRoundTrip asserts that, for every JavaScript syntax
// position templ can interpolate a string into within a <script> element, a
// value read back out of the rendered script via getValue() is identical to
// the value that was rendered in. Any mismatch means that some part of the
// input was evaluated as JavaScript rather than reproduced as inert string
// data, which is precisely the shape of GHSA-jq3r-rjqp-mwgj.
func FuzzScriptStringRoundTrip(f *testing.F) {
	for _, tc := range testcases {
		f.Add([]byte(tc))
	}
	f.Fuzz(func(t *testing.T, v []byte) {
		value := string(v)
		if !utf8.ValidString(value) {
			// Invalid UTF-8 is lossily normalized to U+FFFD by encoding/json
			// when the value takes the JSON-marshalled path, which is a
			// property of encoding/json rather than of the escaping being
			// exercised here, so it isn't expected to round trip byte-for-byte.
			t.Skip("input is not valid UTF-8")
		}
		buf := new(strings.Builder)
		for _, tc := range roundTripCases {
			buf.Reset()
			if err := tc.component(value).Render(context.Background(), buf); err != nil {
				t.Skip(err)
			}
			actual := runGetValue(t, buf.String())
			if actual != value {
				t.Fatalf("%s: round trip failed: input %q, rendered %q, got back %q", tc.name, value, buf.String(), actual)
			}
		}
	})
}

// FuzzComponentAny exercises the non-string (JSON marshalled) branch of the
// template literal escaping path with string, slice and map shaped values.
// Non-string values don't have a simple round trip to assert on, so this only
// proves that the generated script still parses and runs without throwing.
func FuzzComponentAny(f *testing.F) {
	for _, tc := range testcases {
		jsonValue, err := json.Marshal(tc)
		if err != nil {
			panic(err)
		}
		f.Add(jsonValue)
	}
	f.Fuzz(func(t *testing.T, v []byte) {
		buf := new(strings.Builder)

		values := []any{
			string(v),
			[]string{string(v)},
			map[string]string{"value": string(v)},
		}
		for _, value := range values {
			buf.Reset()
			if err := AnyInsideTemplateLiteral(value).Render(context.Background(), buf); err != nil {
				t.Skip(err)
			}
			runTest(t, buf.String())
		}
	})
}

func getFirstScript(n *html.Node) *html.Node {
	if n.Data == "script" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if s := getFirstScript(c); s != nil {
			return s
		}
	}
	return nil
}

func extractScript(t *testing.T, templateOutput string) string {
	n, err := html.Parse(strings.NewReader(templateOutput))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}
	sn := getFirstScript(n)
	if sn == nil {
		t.Fatalf("no script tag found")
	}
	return sn.FirstChild.Data
}

// runGetValue runs script, which must define getValue(), calls it, and
// returns the result coerced to a Go string.
func runGetValue(t *testing.T, templateOutput string) string {
	script := extractScript(t, templateOutput)

	v8ctx := v8.NewContext(iso)
	defer v8ctx.Close()
	if _, err := v8ctx.RunScript(script, "component.js"); err != nil {
		t.Fatalf("failed to parse script: %v", err)
	}
	result, err := v8ctx.RunScript("getValue()", "component.js")
	if err != nil {
		t.Fatalf("failed to call getValue: %v", err)
	}
	return result.String()
}

func runTest(t *testing.T, templateOutput string) {
	script := extractScript(t, templateOutput)

	// Run JavaScript.
	v8ctx := v8.NewContext(iso)
	if _, err := v8ctx.RunScript(script, "component.js"); err != nil {
		t.Fatalf("failed to parse script: %v", err)
	}
	if _, err := v8ctx.RunScript("const result = logValue()", "component.js"); err != nil {
		t.Fatalf("failed to get value: %v", err)
	}
	actual, err := v8ctx.RunScript("result", "component.js")
	if err != nil {
		t.Fatalf("failed to get result: %v", err)
	}
	defer v8ctx.Close()

	// Assert.
	if !actual.IsString() {
		t.Fatalf("expected boolean, got %T", actual.Object().Value)
	}
	if actual.String() != "result_ok" {
		t.Fatalf("expected 'result_ok', got %v", actual.Boolean())
	}
}
