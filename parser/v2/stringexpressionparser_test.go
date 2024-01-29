package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestStringExpressionParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected StringExpression
	}{
		{
			name:  "basic expression",
			input: `{ "this" }`,
			expected: StringExpression{
				Expression: Expression{
					Value: `"this"`,
					Range: Range{
						From: Position{
							Index: 2,
							Line:  0,
							Col:   2,
						},
						To: Position{

							Index: 8,
							Line:  0,
							Col:   8,
						},
					},
				},
			},
		},
		{
			name:  "no spaces",
			input: `{"this"}`,
			expected: StringExpression{
				Expression: Expression{
					Value: `"this"`,
					Range: Range{
						From: Position{
							Index: 1,
							Line:  0,
							Col:   1,
						},
						To: Position{

							Index: 7,
							Line:  0,
							Col:   7,
						},
					},
				},
			},
		},
		{
			name: "multiple lines",
			input: `{ test{}.Call(a,
		b,
	  c) }`,
			expected: StringExpression{
				Expression: Expression{
					Value: "test{}.Call(a,\n\t\tb,\n\t  c)",
					Range: Range{
						From: Position{
							Index: 2,
							Line:  0,
							Col:   2,
						},
						To: Position{

							Index: 27,
							Line:  2,
							Col:   5,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			an, ok, err := stringExpression.Parse(input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !ok {
				t.Fatalf("unexpected failure for input %q", tt.input)
			}
			actual := an.(StringExpression)
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Error(diff)
			}

			// Check the index.
			cut := tt.input[actual.Expression.Range.From.Index:actual.Expression.Range.To.Index]
			if tt.expected.Expression.Value != cut {
				t.Errorf("range, expected %q, got %q", tt.expected.Expression.Value, cut)
			}
		})
	}
}
