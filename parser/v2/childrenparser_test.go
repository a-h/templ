package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestChildrenExpressionParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected ChildrenExpression
	}{
		{
			name:     "standard",
			input:    `{ children...}`,
			expected: ChildrenExpression{},
		},
		{
			name:     "condensed",
			input:    `{children...}`,
			expected: ChildrenExpression{},
		},
		{
			name:     "extra spaces",
			input:    `{  children...  }`,
			expected: ChildrenExpression{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			result, ok, err := childrenExpression.Parse(input)
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}
			if !ok {
				t.Errorf("failed to parse at %d", input.Index())
			}
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestChildrenExpressionParserAllocsOK(t *testing.T) {
	RunParserAllocTest(t, childrenExpression, true, 2, `{ children... }`)
}

func TestChildrenExpressionParserAllocsSkip(t *testing.T) {
	RunParserAllocTest(t, childrenExpression, false, 2, ``)
}
