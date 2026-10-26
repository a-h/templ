package jstokenizer

import "testing"

func TestSlashClassifierDecidesWhetherASlashStartsARegexp(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []string
		expected bool
	}{
		{
			name:     "a slash at the start of the source starts a regexp",
			tokens:   nil,
			expected: true,
		},
		{
			name:     "a slash after an assignment starts a regexp",
			tokens:   []string{"x", " ", "="},
			expected: true,
		},
		{
			name:     "a slash after an identifier is division",
			tokens:   []string{"x"},
			expected: false,
		},
		{
			name:     "a slash after whitespace after an identifier is division",
			tokens:   []string{"x", " ", "\n"},
			expected: false,
		},
		{
			name:     "a slash after a number is division",
			tokens:   []string{"1.5"},
			expected: false,
		},
		{
			name:     "a slash after a number that starts with a decimal point is division",
			tokens:   []string{".5"},
			expected: false,
		},
		{
			name:     "a slash after a keyword starts a regexp",
			tokens:   []string{"return", " "},
			expected: true,
		},
		{
			name:     "a slash after a property named like a keyword is division",
			tokens:   []string{"b", ".", "return"},
			expected: false,
		},
		{
			name:     "a slash after an escape sequence in an identifier is division",
			tokens:   []string{`a`},
			expected: false,
		},
		{
			name:     "a slash after a closing bracket is division",
			tokens:   []string{"b", "[", "0", "]"},
			expected: false,
		},
		{
			name:     "a slash after parentheses is division",
			tokens:   []string{"(", "a", ")"},
			expected: false,
		},
		{
			name:     "a slash after the condition of an if statement starts a regexp",
			tokens:   []string{"if", " ", "(", "a", ")"},
			expected: true,
		},
		{
			name:     "a slash after parentheses within a condition is division",
			tokens:   []string{"if", "(", "(", "a", ")"},
			expected: false,
		},
		{
			name:     "a slash after the head of a for await statement starts a regexp",
			tokens:   []string{"for", " ", "await", " ", "(", "a", ")"},
			expected: true,
		},
		{
			name:     "a slash after a call to a method named if is division",
			tokens:   []string{"b", ".", "if", "(", "c", ")"},
			expected: false,
		},
		{
			name:     "a slash after a postfix increment is division",
			tokens:   []string{"i", "+", "+"},
			expected: false,
		},
		{
			name:     "a slash after a prefix increment starts a regexp",
			tokens:   []string{"=", "+", "+"},
			expected: true,
		},
		{
			name:     "a slash after an increment on the line after an identifier starts a regexp",
			tokens:   []string{"i", "\n", "+", "+"},
			expected: true,
		},
		{
			name:     "a slash after addition starts a regexp",
			tokens:   []string{"i", "+"},
			expected: true,
		},
		{
			name:     "a slash after a closing brace starts a regexp",
			tokens:   []string{"{", "}"},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			classifier := newSlashClassifier()
			for _, token := range tt.tokens {
				classifier.Record(token)
			}
			if classifier.RegexpAllowed != tt.expected {
				t.Errorf("expected RegexpAllowed to be %v, got %v", tt.expected, classifier.RegexpAllowed)
			}
		})
	}
}

func TestSlashClassifierReportsAmbiguousSlashes(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []string
		expected bool
	}{
		{
			name:     "a slash after a closing brace is ambiguous",
			tokens:   []string{"{", "}"},
			expected: true,
		},
		{
			name:     "a slash after whitespace after a closing brace is ambiguous",
			tokens:   []string{"}", " ", "\n"},
			expected: true,
		},
		{
			name:     "a slash after yield is ambiguous",
			tokens:   []string{"yield"},
			expected: true,
		},
		{
			name:     "a slash after a property named yield is not ambiguous",
			tokens:   []string{"a", ".", "yield"},
			expected: false,
		},
		{
			name:     "a slash after an identifier on the same line is not ambiguous",
			tokens:   []string{"a", " "},
			expected: false,
		},
		{
			name:     "a slash after an identifier and a line feed is ambiguous",
			tokens:   []string{"a", "\n"},
			expected: true,
		},
		{
			name:     "a slash after an identifier and a line separator is ambiguous",
			tokens:   []string{"a", "\u2028"},
			expected: true,
		},
		{
			name:     "a slash after a property name and a line feed is not ambiguous",
			tokens:   []string{"a", ".", "b", "\n"},
			expected: false,
		},
		{
			name:     "a slash after a number and a line feed is not ambiguous",
			tokens:   []string{"1", "\n"},
			expected: false,
		},
		{
			name:     "a slash after a keyword and a line feed is not ambiguous",
			tokens:   []string{"return", "\n"},
			expected: false,
		},
		{
			name:     "a slash after an operator is not ambiguous",
			tokens:   []string{"}", "="},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			classifier := newSlashClassifier()
			for _, token := range tt.tokens {
				classifier.Record(token)
			}
			if classifier.IsAmbiguous() != tt.expected {
				t.Errorf("expected IsAmbiguous to return %v, got %v", tt.expected, classifier.IsAmbiguous())
			}
		})
	}
}
