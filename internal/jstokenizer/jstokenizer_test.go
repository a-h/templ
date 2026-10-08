package jstokenizer

import (
	"testing"

	"github.com/a-h/parse"
)

const value = "{{ x }}"

func TestTokenizerDeterminesTheContextOfValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Context
	}{
		{
			name:     "a value in code is in code context",
			input:    `var a = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a value in a double quoted string is in string context",
			input:    `var a = "{{ x }}";`,
			expected: ContextString,
		},
		{
			name:     "a value in a single quoted string is in string context",
			input:    `var a = '{{ x }}';`,
			expected: ContextString,
		},
		{
			name:     "an escaped quote does not end a string",
			input:    `var a = "\"{{ x }}";`,
			expected: ContextString,
		},
		{
			name:     "a different quote does not end a string",
			input:    `var a = "'{{ x }}";`,
			expected: ContextString,
		},
		{
			name:     "dollar brace in a double quoted string is string data",
			input:    `var a = "${ {{ x }} }";`,
			expected: ContextString,
		},
		{
			name:     "a value after a string has ended is in code context",
			input:    `var a = "a" + {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a value in a template literal is in template literal context",
			input:    "var a = `{{ x }}`;",
			expected: ContextTemplateLiteral,
		},
		{
			name:     "a value in a substitution is in code context",
			input:    "var a = `${ {{ x }} }`;",
			expected: ContextCode,
		},
		{
			name:     "a value after a substitution has closed is in template literal context",
			input:    "var a = `${ b }{{ x }}`;",
			expected: ContextTemplateLiteral,
		},
		{
			name:     "a value in a string inside a substitution is in string context",
			input:    "var a = `${ \"{{ x }}\" }`;",
			expected: ContextString,
		},
		{
			name:     "a value in a template literal nested in a substitution is in template literal context",
			input:    "var a = `${ `{{ x }}` }`;",
			expected: ContextTemplateLiteral,
		},
		{
			name:     "a value in a substitution nested in a template literal in a substitution is in code context",
			input:    "var a = `${ `${ {{ x }} }` }`;",
			expected: ContextCode,
		},
		{
			name:     "a value after a nested template literal in a substitution is in code context",
			input:    "var a = `${ `b` + {{ x }} }`;",
			expected: ContextCode,
		},
		{
			name:     "an object literal in a substitution does not end the substitution",
			input:    "var a = `${ f({ b: { c: 1 } }) + {{ x }} }`;",
			expected: ContextCode,
		},
		{
			name:     "an object literal in a substitution is closed before the substitution",
			input:    "var a = `${ f({ b: 1 }) }{{ x }}`;",
			expected: ContextTemplateLiteral,
		},
		{
			name:     "an opening brace in a string in a substitution is not counted",
			input:    "var a = `${\"{\"}{{ x }}`;",
			expected: ContextTemplateLiteral,
		},
		{
			name:     "a closing brace in a string in a substitution is not counted",
			input:    "var a = `${ \"}\" + {{ x }} }`;",
			expected: ContextCode,
		},
		{
			name:     "an opening brace in a comment in a substitution is not counted",
			input:    "var a = `${ 1 /* { */ }{{ x }}`;",
			expected: ContextTemplateLiteral,
		},
		{
			name:     "an opening brace in a regexp in a substitution is not counted",
			input:    "var a = `${ /{/.source }{{ x }}`;",
			expected: ContextTemplateLiteral,
		},
		{
			name:     "a quote in a single line comment does not start a string",
			input:    "// \"\nvar a = {{ x }};",
			expected: ContextCode,
		},
		{
			name:     "a quote in a multi-line comment does not start a string",
			input:    `/* ' */ var a = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "comment markers in a string are string data",
			input:    `var a = "//" + "{{ x }}";`,
			expected: ContextString,
		},
		{
			name:     "a slash after an assignment starts a regexp, so a quote in it does not start a string",
			input:    `var r = /"/; var a = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after an opening parenthesis starts a regexp",
			input:    `f(/'/, {{ x }});`,
			expected: ContextCode,
		},
		{
			name:     "a slash after a keyword starts a regexp",
			input:    `function f() { return /"/; } var a = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after a block starts a regexp",
			input:    `if (b) {} /"/.test(s); var a = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after an identifier is division",
			input:    `var a = b / 2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after an identifier that starts with a keyword is division",
			input:    `var a = returned / 2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after a number is division",
			input:    `var a = 4 / 2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after a closing parenthesis is division",
			input:    `var a = (b) / 2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after a closing bracket is division",
			input:    `var a = b[0] / 2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after a string is division",
			input:    `var a = "b" / 2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after a value is division",
			input:    `var a = {{ x }} / 2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after a regexp is division",
			input:    `var a = /b/ / 2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pi := parse.NewInput(tt.input)
			tokenizer := New()
			var actual Context
			var found bool
			for {
				peeked, ok := pi.Peek(len(value))
				if ok && peeked == value {
					pi.Take(len(value))
					actual = tokenizer.GetContext()
					found = true
					tokenizer.RecordValue()
					continue
				}
				_, ok, err := tokenizer.ReadComment(pi)
				if err != nil {
					t.Fatalf("unexpected error reading comment: %v", err)
				}
				if ok {
					continue
				}
				_, ok, err = tokenizer.Read(pi)
				if err != nil {
					t.Fatalf("unexpected error reading token: %v", err)
				}
				if !ok {
					break
				}
			}
			if !found {
				t.Fatalf("no value found")
			}
			if actual != tt.expected {
				t.Errorf("expected context %v, got %v", tt.expected, actual)
			}
		})
	}
}

func TestRegexpLiteralParser(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expected   string
		expectedOK bool
	}{
		{
			name:       "no content is considered to be a comment",
			input:      `//`,
			expectedOK: false,
		},
		{
			name:       "must not be multiline",
			input:      "/div>\n</div>",
			expectedOK: false,
		},
		{
			name:       "match a single char",
			input:      `/a/`,
			expected:   `/a/`,
			expectedOK: true,
		},
		{
			name:       "match a simple regex",
			input:      `/a|b/`,
			expected:   `/a|b/`,
			expectedOK: true,
		},
		{
			name:       "match a complex regex",
			input:      `/a(b|c)*d{2,4}/`,
			expected:   `/a(b|c)*d{2,4}/`,
			expectedOK: true,
		},
		{
			name:       "match a regex with flags",
			input:      `/a/i`,
			expected:   `/a/i`,
			expectedOK: true,
		},
		{
			name:       "match a regex with multiple flags",
			input:      `/a/gmi`,
			expected:   `/a/gmi`,
			expectedOK: true,
		},
		{
			name:       "escaped slashes",
			input:      `/a\/b\/c/`,
			expected:   `/a\/b\/c/`,
			expectedOK: true,
		},
		{
			name:       "no match: missing closing slash",
			input:      `/a|b`,
			expected:   "",
			expectedOK: false,
		},
		{
			name:       "no match: missing opening slash",
			input:      `a|b/`,
			expected:   "",
			expectedOK: false,
		},
		{
			name:       "must not contain interpolated go expressions",
			input:      `/a{{ x }}/`,
			expected:   "",
			expectedOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := parse.NewInput(tt.input)
			result, ok, err := regexpLiteral.Parse(input)
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}
			if ok != tt.expectedOK {
				t.Fatalf("expected ok to be %v, got %v", tt.expectedOK, ok)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
