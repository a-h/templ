package parser

import (
	"testing"

	"github.com/a-h/lexical/input"
	"github.com/a-h/lexical/parse"
	"github.com/google/go-cmp/cmp"
)

func TestWhitespace(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		parser   parse.Function
		expected interface{}
	}{
		{
			name:   "whitespace: various spaces",
			input:  "  \t ",
			parser: whitespaceParser,
			expected: Whitespace{
				Value: "  \t ",
			},
		},
		{
			name:   "whitespace: spaces and newline",
			input:  " \n",
			parser: whitespaceParser,
			expected: Whitespace{
				Value: " \n",
			},
		},
		{
			name:   "whitespace: newline",
			input:  "\n",
			parser: whitespaceParser,
			expected: Whitespace{
				Value: "\n",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := input.NewFromString(tt.input)
			result := tt.parser(input)
			if result.Error != nil {
				t.Fatalf("parser error: %v", result.Error)
			}
			if !result.Success {
				t.Fatalf("failed to parse at %d", input.Index())
			}
			if diff := cmp.Diff(tt.expected, result.Item); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
