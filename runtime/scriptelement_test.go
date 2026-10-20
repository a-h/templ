package runtime

import (
	"errors"
	"testing"
)

func TestScriptContent(t *testing.T) {
	tests := []struct {
		name                         string
		input                        any
		expectedInsideStringLiteral  string
		expectedOutsideStringLiteral string
	}{
		{
			name:                         "string",
			input:                        "hello",
			expectedInsideStringLiteral:  `hello`,
			expectedOutsideStringLiteral: `"hello"`,
		},
		{
			name:                         "string with single quotes",
			input:                        "hello 'world'",
			expectedInsideStringLiteral:  `hello \u0027world\u0027`,
			expectedOutsideStringLiteral: `"hello 'world'"`,
		},
		{
			name:                         "string with double quotes",
			input:                        "hello \"world\"",
			expectedInsideStringLiteral:  `hello \u0022world\u0022`,
			expectedOutsideStringLiteral: `"hello \"world\""`,
		},
		{
			name:                         "string with backticks",
			input:                        "hello `world`",
			expectedInsideStringLiteral:  `hello \u0060world\u0060`,
			expectedOutsideStringLiteral: "\"hello `world`\"",
		},
		{
			name:                         "int",
			input:                        1,
			expectedInsideStringLiteral:  `1`,
			expectedOutsideStringLiteral: `1`,
		},
		{
			name:                         "float",
			input:                        1.1,
			expectedInsideStringLiteral:  `1.1`,
			expectedOutsideStringLiteral: `1.1`,
		},
		{
			name:                         "bool",
			input:                        true,
			expectedInsideStringLiteral:  `true`,
			expectedOutsideStringLiteral: `true`,
		},
		{
			name:                         "array",
			input:                        []int{1, 2, 3},
			expectedInsideStringLiteral:  `[1,2,3]`,
			expectedOutsideStringLiteral: `[1,2,3]`,
		},
		{
			name:                         "object",
			input:                        struct{ Name string }{"Alice"},
			expectedInsideStringLiteral:  `{\u0022Name\u0022:\u0022Alice\u0022}`,
			expectedOutsideStringLiteral: `{"Name":"Alice"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Run("inside string literal", func(t *testing.T) {
				actualInsideStringLiteral, err := ScriptContentInsideStringLiteral(tt.input)
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if actualInsideStringLiteral != tt.expectedInsideStringLiteral {
					t.Errorf("expected:\n%s\ngot:\n%s", tt.expectedInsideStringLiteral, actualInsideStringLiteral)
				}
			})
			t.Run("outside string literal", func(t *testing.T) {
				actualOutsideStringLiteral, err := ScriptContentOutsideStringLiteral(tt.input)
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if actualOutsideStringLiteral != tt.expectedOutsideStringLiteral {
					t.Errorf("expected:\n%s\ngot:\n%s", tt.expectedOutsideStringLiteral, actualOutsideStringLiteral)
				}
			})
		})
	}
}

// TestScriptContentInsideTemplateLiteral covers GHSA-jq3r-rjqp-mwgj: a Go value
// interpolated into a backtick-quoted JavaScript template literal must have '$',
// '{' and '}' escaped, since an unescaped "${ ... }" sequence is evaluated as a
// JavaScript expression by the browser rather than treated as inert string data.
func TestScriptContentInsideTemplateLiteral(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "string with no special characters is untouched",
			input:    "hello",
			expected: `hello`,
		},
		{
			name:     "dollar brace substitution syntax is escaped",
			input:    "${alert(document.domain)}",
			expected: `\u0024\u007balert(document.domain)\u007d`,
		},
		{
			name:     "backtick is escaped so the literal cannot be closed early",
			input:    "hello `world`",
			expected: `hello \u0060world\u0060`,
		},
		{
			name:     "closing brace alone is escaped",
			input:    "}",
			expected: `\u007d`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual, err := ScriptContentInsideTemplateLiteral(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if actual != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, actual)
			}
		})
	}
}

// TestScriptContentInsideRegexp covers a Go value interpolated into the pattern of a
// JavaScript regular expression literal. JSON encoding does not escape '/', so a
// value could end the regular expression, and the rest of the value would execute.
// The value must also match itself, rather than change the pattern.
func TestScriptContentInsideRegexp(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "string with no special characters is untouched",
			input:    "hello",
			expected: `hello`,
		},
		{
			name:     "slashes are escaped so the regexp cannot be ended early",
			input:    "/;alert(1);//",
			expected: `\/;alert\(1\);\/\/`,
		},
		{
			name:     "a backslash is escaped so it cannot escape the closing slash",
			input:    `\`,
			expected: `\\`,
		},
		{
			name:     "metacharacters are escaped so they match themselves",
			input:    "a.b*c+d?e^f$g|h",
			expected: `a\.b\*c\+d\?e\^f\$g\|h`,
		},
		{
			name:     "brackets, braces, and parentheses are escaped so they cannot open a group, class, or quantifier",
			input:    "[a]{1}(b)",
			expected: `\[a\]\{1\}\(b\)`,
		},
		{
			name:     "a hyphen is hex escaped so that it is valid in a regexp with the u flag",
			input:    "a-z",
			expected: `a\x2dz`,
		},
		{
			name:     "a line terminator is escaped so that it cannot end the regexp",
			input:    "a\nb",
			expected: `a\nb`,
		},
		{
			name:     "an HTML end tag is escaped so that it cannot end the script element",
			input:    "</script>",
			expected: `\u003c\/script\u003e`,
		},
		{
			name:     "an HTML start tag is escaped so that it cannot start double escaped script data",
			input:    "<script>",
			expected: `\u003cscript\u003e`,
		},
		{
			name:     "quotes and ampersands are hex escaped",
			input:    `"'&`,
			expected: `\u0022\u0027\u0026`,
		},
		{
			name:     "a number is JSON encoded and then escaped",
			input:    1.5,
			expected: `1\.5`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual, err := ScriptContentInsideRegexp(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if actual != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, actual)
			}
		})
	}
}

func TestScriptContentErrors(t *testing.T) {
	t.Run("inside string literal", func(t *testing.T) {
		_, err := ScriptContentInsideStringLiteral("s", errors.New("error"))
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "error" {
			t.Errorf("expected error: %s", err)
		}
	})
	t.Run("outside string literal", func(t *testing.T) {
		_, err := ScriptContentOutsideStringLiteral("s", errors.New("error"))
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "error" {
			t.Errorf("expected error: %s", err)
		}
	})
	t.Run("unmarshal error", func(t *testing.T) {
		_, err := ScriptContentInsideStringLiteral(func() string {
			return "hello"
		})
		if err == nil {
			t.Fatalf("expected unmarshal error, but got %v", err)
		}
	})
}
