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
