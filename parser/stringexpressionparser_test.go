package parser

import (
	"testing"

	"github.com/a-h/lexical/input"
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
			input: `{%= "this" %}`,
			expected: StringExpression{
				Expression: Expression{
					Value: `"this"`,
					Range: Range{
						From: Position{
							Index: 4,
							Line:  1,
							Col:   4,
						},
						To: Position{

							Index: 10,
							Line:  1,
							Col:   10,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := input.NewFromString(tt.input)
			result := newStringExpressionParser().Parse(input)
			if result.Error != nil {
				t.Fatalf("parser error: %v", result.Error)
			}
			if !result.Success {
				t.Fatalf("failed to parse at %d", input.Index())
			}
			if diff := cmp.Diff(tt.expected, result.Item); diff != "" {
				t.Errorf(diff)
			}

			// Check the index.
			se := result.Item.(StringExpression)
			cut := tt.input[se.Expression.Range.From.Index:se.Expression.Range.To.Index]
			if tt.expected.Expression.Value != cut {
				t.Errorf("range, expected %q, got %q", tt.expected.Expression.Value, cut)
			}
		})
	}
}
