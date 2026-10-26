package jstokenizer

import (
	"errors"
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

const value = "{{ x }}"

func TestTokenizerDeterminesTheContextOfValues(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Context
		expectedErr error
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
			name:     "a hashbang at the start of the source comments out the rest of the line",
			input:    "#! \"\nvar a = {{ x }};",
			expected: ContextCode,
		},
		{
			name:     "a slash after a keyword and a line separator starts a regexp",
			input:    "function f() { return /\"/; } var a = {{ x }};",
			expected: ContextCode,
		},
		{
			name:     "a line separator ends a single line comment",
			input:    "// a var b = `{{ x }}`;",
			expected: ContextTemplateLiteral,
		},
		{
			name:     "a paragraph separator ends a single line comment",
			input:    "// a var b = `{{ x }}`;",
			expected: ContextTemplateLiteral,
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
			name:        "a value after a slash after a closing brace is ambiguous when a block and an object literal give different contexts",
			input:       `if (b) {} /"/.test(s); var a = {{ x }};`,
			expectedErr: ErrAmbiguousContext,
		},
		{
			name:     "a value after a slash after a closing brace is in code context once a block and an object literal give the same state",
			input:    "if (b) {} /x/.test(s);\nvar a = {{ x }};",
			expected: ContextCode,
		},
		{
			name:        "a value after a slash after an object literal on the same line is ambiguous",
			input:       `var a = {} / 2; var b = {{ x }};`,
			expectedErr: ErrAmbiguousContext,
		},
		{
			name:     "a value on the line after a slash after an object literal is in code context",
			input:    "var a = {} / 2;\nvar b = {{ x }};",
			expected: ContextCode,
		},
		{
			name:        "a value after a slash after yield is ambiguous, because yield is an identifier outside a generator",
			input:       `var a = yield / {{ x }};`,
			expectedErr: ErrAmbiguousContext,
		},
		{
			name:        "a value after a slash after await is ambiguous, because await is an identifier outside an async function",
			input:       `var a = await /"/; var b = {{ x }};`,
			expectedErr: ErrAmbiguousContext,
		},
		{
			name:        "a value after a slash on the line after a declaration without an initializer is ambiguous",
			input:       "var a\n/\"/.test(s); var b = {{ x }};",
			expectedErr: ErrAmbiguousContext,
		},
		{
			name:        "a comment that contains a line terminator makes a slash after an identifier ambiguous",
			input:       "var a /*\n*/ /\"/.test(s); var b = {{ x }};",
			expectedErr: ErrAmbiguousContext,
		},
		{
			name:     "a slash after an identifier on the same line is division",
			input:    `var a = b /\n2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after a prefix increment starts a regexp",
			input:    `++/"/.lastIndex; var a = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after an increment on the line after an identifier starts a regexp",
			input:    "a\n++/\"/.lastIndex; var b = {{ x }};",
			expected: ContextCode,
		},
		{
			name:        "a value in a comment in one interpretation of an ambiguous slash is ambiguous",
			input:       "var a = {}/ // {{ x }}\n;",
			expectedErr: ErrAmbiguousContext,
		},
		{
			name:     "a slash after an identifier is division",
			input:    `var a = b / 2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after an identifier that ends with a combining mark is division",
			input:    "var a = A\u05B1/\"/g, b = \"/ 1; var c = {{ x }};",
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
			name:     "a slash after a number with a trailing decimal point is division",
			input:    `var a = 1. / 2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after a property of a number is division",
			input:    `var a = 1..x / 2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after a number that starts with a decimal point is division",
			input:    `var a = .5/2; var c = "/"; var d = {{ x }};`,
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
		{
			name:     "a slash after the condition of an if statement starts a regexp",
			input:    `if (ok) /"/.test(s); var a = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after the condition of a while statement starts a regexp",
			input:    `while (ok) /"/.test(s); var a = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after the head of a for statement starts a regexp",
			input:    `for (;;) /"/.test(s); var a = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after the object of a with statement starts a regexp",
			input:    `with (o) /"/.test(s); var a = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after the head of a for await statement starts a regexp",
			input:    `for await (const a of b) /"/.test(a); var c = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after a condition that contains parentheses starts a regexp",
			input:    `if (f(a)) /"/.test(s); var a = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after parentheses within a condition is division",
			input:    `if ((a) / 2) { var c = "/"; } var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after a call to a method named if is division",
			input:    `var a = b.if(c) / 2; var d = "/"; var e = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after a property named return is division",
			input:    `var a = b.return / 2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after an optional chained property named return is division",
			input:    `var a = b?.return / 2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after a postfix increment is division",
			input:    `var a = i++ / 2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after a postfix decrement is division",
			input:    `var a = i-- / 2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after addition of a postfix increment starts a regexp",
			input:    `var a = i+++/"/.source; var b = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "an HTML open comment comments out the rest of the line",
			input:    "<!-- \"\nvar a = {{ x }};",
			expected: ContextCode,
		},
		{
			name:     "an HTML open comment after code comments out the rest of the line",
			input:    "var a = b <!-- \"\nvar c = {{ x }};",
			expected: ContextCode,
		},
		{
			name:     "an HTML close comment at the start of a line comments out the rest of the line",
			input:    "var a = b;\n--> \"\nvar c = {{ x }};",
			expected: ContextCode,
		},
		{
			name:     "an HTML close comment after whitespace at the start of a line comments out the rest of the line",
			input:    "var a = b;\n  --> \"\nvar c = {{ x }};",
			expected: ContextCode,
		},
		{
			name:     "a postfix decrement followed by greater than within a line is not a comment",
			input:    `var a = b --> 0; var c = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a value in a regexp is in regexp context",
			input:    `var r = /a{{ x }}/;`,
			expected: ContextRegexp,
		},
		{
			name:     "a quote in a regexp before a value does not start a string",
			input:    `var r = /"{{ x }}/;`,
			expected: ContextRegexp,
		},
		{
			name:     "a slash in a character class does not end a regexp",
			input:    `var r = /[/]{{ x }}/;`,
			expected: ContextRegexp,
		},
		{
			name:     "an escaped slash does not end a regexp",
			input:    `var r = /\/{{ x }}/;`,
			expected: ContextRegexp,
		},
		{
			name:     "a value at the start of a regexp is in regexp start context",
			input:    `var r = /{{ x }}/;`,
			expected: ContextRegexpStart,
		},
		{
			name:     "a value after a value at the start of a regexp is in regexp context",
			input:    `var r = /{{ x }}{{ x }}/;`,
			expected: ContextRegexp,
		},
		{
			name:     "a value at the start of a character class is in regexp context",
			input:    `var r = /[{{ x }}]/;`,
			expected: ContextRegexp,
		},
		{
			name:     "a value after a regexp that contains a value is in code context",
			input:    `var r = /{{ x }}/g; var a = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after a regexp that contains a value is division",
			input:    `var r = /{{ x }}/ / 2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a slash after the flags of a regexp is division",
			input:    `var r = /a/gi / 2; var c = "/"; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a quote after an escaped slash in a regexp does not start a string",
			input:    `var r = /a\/"/; var d = {{ x }};`,
			expected: ContextCode,
		},
		{
			name:     "a line feed ends an unterminated string",
			input:    "var a = \"b\nvar c = {{ x }};",
			expected: ContextCode,
		},
		{
			name:     "a line separator does not end a string",
			input:    "var a = \"b\u2028{{ x }}\";",
			expected: ContextString,
		},
		{
			name:     "a value on the line after an ambiguous slash is in code context once an unterminated string ends",
			input:    "if (ok) {} /\"/.test(s);\nconst message = {{ x }};",
			expected: ContextCode,
		},
		{
			name:     "a line terminator ends an unterminated regexp",
			input:    "var r = /a\nvar c = {{ x }};",
			expected: ContextCode,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pi := parse.NewInput(tt.input)
			tokenizer := New()
			var actual Context
			var actualErr error
			var found bool
			for actualErr == nil {
				peeked, ok := pi.Peek(len(value))
				if ok && peeked == value {
					start := pi.Index()
					pi.Take(len(value))
					actual, actualErr = tokenizer.RecordValue(pi, start)
					found = true
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
			if !errors.Is(actualErr, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, actualErr)
			}
			if tt.expectedErr == nil && actual != tt.expected {
				t.Errorf("expected context %v, got %v", tt.expected, actual)
			}
		})
	}
}

func TestTokenizerReadsTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "a number with a decimal point is one token",
			input:    "1_000.5e3",
			expected: []string{"1_000.5e3"},
		},
		{
			name:     "a number that starts with a decimal point is one token",
			input:    ".0.b",
			expected: []string{".0", ".", "b"},
		},
		{
			name:     "a property of an identifier that starts with an underscore is not a number",
			input:    "_.b",
			expected: []string{"_", ".", "b"},
		},
		{
			name:     "an identifier can contain a combining mark",
			input:    "A\u05B1/2",
			expected: []string{"A\u05B1", "/", "2"},
		},
		{
			name:     "an identifier can contain a zero width joiner",
			input:    "a\u200db/2",
			expected: []string{"a\u200db", "/", "2"},
		},
		{
			name:     "a line separator is one token",
			input:    "a b",
			expected: []string{"a", " ", "b"},
		},
		{
			name:     "an identifier can contain non-ASCII letters",
			input:    "café/2",
			expected: []string{"café", "/", "2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pi := parse.NewInput(tt.input)
			tokenizer := New()
			var actual []string
			for {
				token, ok, err := tokenizer.Read(pi)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !ok {
					break
				}
				actual = append(actual, token)
			}
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Error(diff)
			}
		})
	}
}
