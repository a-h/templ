package parser

import (
	"io"
	"testing"

	"github.com/a-h/lexical/input"
)

// # List of situations where a templ file could contain braces.

// Inside a HTML attribute.
// <a style="font-family: { arial }">That does not make sense, but still...</a>

// Inside a script tag.
// <script>var value = { test: 123 };</script>

// Inside a templ definition expression.
// { templ Name(data map[string]interface{}) }

// Inside a templ script.
// { script Name(data map[string]interface{}) }
//   { something }
// { endscript }

// Inside a call to a template, passing some data.
// {! localisations(map[string]interface{} { "key": 123 }) }

// Inside a string.
// {! localisations("\"value{'data'}") }

// Inside a tick string.
// {! localisations(`value{'data'}`) }

// Parser logic...
// Read until ( ` | " | { | } | EOL/EOF )
//  If " handle any escaped quotes or ticks until the end of the string.
//  If ` read until the closing tick.
//  If { increment the brace count up
//  If } increment the brace count down
//    If brace count == 0, break
//  If EOL, break
//  If EOF, break
// If brace count != 0 throw an error

func TestRuneLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "rune literal with escaped newline",
			input:    `'\n' `,
			expected: `'\n'`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := rune_lit(input.NewFromString(tt.input))
			if pr.Error != nil && pr.Error != io.EOF {
				t.Fatalf("unexpected error: %v", pr.Error)
			}
			if !pr.Success {
				t.Fatalf("unexpected failure for input %q: %v", tt.input, pr)
			}
			actual, ok := pr.Item.(string)
			if !ok {
				t.Fatalf("unexpectedly not a string, got %t (%v) instead", pr.Item, pr.Item)
			}
			if string(actual) != tt.expected {
				t.Errorf("expected vs got:\n%q\n%q", tt.expected, actual)
			}
		})
	}
}

func TestStringLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "string literal with escaped newline",
			input:    `"\n" `,
			expected: `"\n"`,
		},
		{
			name:     "raw literal with \n",
			input:    "`\\n` ",
			expected: "`\\n`",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := string_lit(input.NewFromString(tt.input))
			if pr.Error != nil && pr.Error != io.EOF {
				t.Fatalf("unexpected error: %v", pr.Error)
			}
			if !pr.Success {
				t.Fatalf("unexpected failure for input %q: %v", tt.input, pr)
			}
			actual, ok := pr.Item.(string)
			if !ok {
				t.Fatalf("unexpectedly not a string, got %t (%v) instead", pr.Item, pr.Item)
			}
			if string(actual) != tt.expected {
				t.Errorf("expected vs got:\n%q\n%q", tt.expected, actual)
			}
		})
	}
}

func TestExpressions(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		prefix          string
		startBraceCount int
		expected        string
	}{
		{
			name:            "templ: no parameters",
			input:           "{ templ TemplName() }\n",
			prefix:          "{ templ ",
			startBraceCount: 1,
			expected:        "TemplName()",
		},
		{
			name:            "templ: string parameter",
			input:           `{ templ TemplName(a string) }`,
			prefix:          "{ templ ",
			startBraceCount: 1,
			expected:        `TemplName(a string)`,
		},
		{
			name:            "templ: map parameter",
			input:           `{ templ TemplName(data map[string]interface{}) }`,
			prefix:          "{ templ ",
			startBraceCount: 1,
			expected:        `TemplName(data map[string]interface{})`,
		},
		{
			name:            "call: string parameter",
			input:           `{! Header("test") }`,
			prefix:          "{! ",
			startBraceCount: 1,
			expected:        `Header("test")`,
		},
		{
			name:            "call: string parameter with escaped values and mismatched braces",
			input:           `{! Header("\"}}") }`,
			prefix:          "{! ",
			startBraceCount: 1,
			expected:        `Header("\"}}")`,
		},
		{
			name:            "call: string parameter, with rune literals",
			input:           `{! Header('\"') }`,
			prefix:          "{! ",
			startBraceCount: 1,
			expected:        `Header('\"')`,
		},
		{
			name:            "call: map literal",
			input:           `{! Header(map[string]interface{}{ "test": 123 }) }`,
			prefix:          "{! ",
			startBraceCount: 1,
			expected:        `Header(map[string]interface{}{ "test": 123 })`,
		},
		{
			name:            "call: rune and map literal",
			input:           `{! Header('\"', map[string]interface{}{ "test": 123 }) }`,
			prefix:          "{! ",
			startBraceCount: 1,
			expected:        `Header('\"', map[string]interface{}{ "test": 123 })`,
		},
		{
			name:            "if: function call",
			input:           `{ if findOut("}") }`,
			prefix:          "{ if ",
			startBraceCount: 1,
			expected:        `findOut("}")`,
		},
		{
			name:            "if: function call, tricky string/rune params",
			input:           `{ if findOut("}", '}', '\'') }`,
			prefix:          "{ if ",
			startBraceCount: 1,
			expected:        `findOut("}", '}', '\'')`,
		},
		{
			name:            "if: function call, function param",
			input:           `{ if findOut(func() bool { return true }) }`,
			prefix:          "{ if ",
			startBraceCount: 1,
			expected:        `findOut(func() bool { return true })`,
		},
		{
			name: "attribute value: simple string",
			// Used to be {%= "data" %}, but can be simplified, since the position
			// of the node in the document defines how it can be used.
			// As an attribute value, it must be a Go expression that returns a string.
			input:           `{ "data" }`,
			prefix:          "{ ",
			startBraceCount: 1,
			expected:        `"data"`,
		},
		{
			name:            "javascript expression",
			input:           "var x = 123;",
			prefix:          "",
			startBraceCount: 0,
			expected:        "var x = 123;",
		},
		{
			name:            "javascript expression",
			input:           `var x = "}";`,
			prefix:          "",
			startBraceCount: 0,
			expected:        `var x = "}";`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := &expressionParser{
				startBraceCount: tt.startBraceCount,
			}
			expr := tt.input[len(tt.prefix):]
			pr := ep.Parse(input.NewFromString(expr))
			if pr.Error != nil && pr.Error != io.EOF {
				t.Fatalf("unexpected error: %v", pr.Error)
			}
			if !pr.Success {
				t.Fatalf("unexpected failure for input %q: %v", expr, pr)
			}
			actual, ok := pr.Item.(string)
			if !ok {
				t.Fatalf("unexpectedly not an expression, got %t (%v) instead", pr.Item, pr.Item)
			}
			if string(actual) != tt.expected {
				t.Errorf("expected vs got:\n%q\n%q", tt.expected, actual)
			}
		})
	}
}
