package fuzzing

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"golang.org/x/net/html"

	v8 "rogchap.com/v8go"
)

var iso = v8.NewIsolate()

var testcases = []string{
	"hello",
	"<script>console.log('hello')</script>",
	"text data\n can contain all sorts of \"characters\" & symbols",
	"123",
	"</script>",
}

func FuzzComponentString(f *testing.F) {
	for _, tc := range testcases {
		f.Add([]byte(tc))
	}
	f.Fuzz(func(t *testing.T, v []byte) {
		// Render template.
		buf := new(strings.Builder)

		values := []any{
			string(v),
			[]string{string(v)},
			map[string]string{"value": string(v)},
		}
		for _, value := range values {
			buf.Reset()
			if err := String(value).Render(context.Background(), buf); err != nil {
				t.Skip(err)
			}
			runTest(t, buf.String())
		}
	})
}

func FuzzComponentAny(f *testing.F) {
	for _, tc := range testcases {
		jsonValue, err := json.Marshal(tc)
		if err != nil {
			panic(err)
		}
		f.Add(jsonValue)
	}
	f.Fuzz(func(t *testing.T, v []byte) {
		// Render template.
		buf := new(strings.Builder)

		values := []any{
			string(v),
			[]string{string(v)},
			map[string]string{"value": string(v)},
		}
		for _, value := range values {
			buf.Reset()
			if err := Any(value).Render(context.Background(), buf); err != nil {
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

func runTest(t *testing.T, templateOutput string) {
	// Parse HTML.
	n, err := html.Parse(strings.NewReader(templateOutput))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}
	sn := getFirstScript(n)
	if sn == nil {
		t.Fatalf("no script tag found")
	}

	// Extract JavaScript.
	script := sn.FirstChild.Data

	// Run JavaScript.
	v8ctx := v8.NewContext(iso)
	if _, err = v8ctx.RunScript(script, "component.js"); err != nil {
		t.Fatalf("failed to parse script: %v", err)
	}
	if _, err = v8ctx.RunScript("const result = logValue()", "component.js"); err != nil {
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
