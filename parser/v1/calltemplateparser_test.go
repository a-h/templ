package parser

import (
	"testing"

	"github.com/a-h/lexical/input"
	"github.com/google/go-cmp/cmp"
)

func TestCallTemplateExpressionParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected CallTemplateExpression
	}{
		{
			name:  "call: simple",
			input: `{%! Other(p.Test) %}`,
			expected: CallTemplateExpression{
				Expression: Expression{
					Value: "Other(p.Test)",
					Range: Range{
						From: Position{
							Index: 4,
							Line:  1,
							Col:   4,
						},
						To: Position{
							Index: 17,
							Line:  1,
							Col:   17,
						},
					},
				},
			},
		},
		{
			name:  "call: simple, missing start space",
			input: `{%!Other(p.Test) %}`,
			expected: CallTemplateExpression{
				Expression: Expression{
					Value: "Other(p.Test)",
					Range: Range{
						From: Position{
							Index: 3,
							Line:  1,
							Col:   3,
						},
						To: Position{
							Index: 16,
							Line:  1,
							Col:   16,
						},
					},
				},
			},
		},
		{
			name:  "call: simple, missing start and end space",
			input: `{%!Other(p.Test)%}`,
			expected: CallTemplateExpression{
				Expression: Expression{
					Value: "Other(p.Test)",
					Range: Range{
						From: Position{
							Index: 3,
							Line:  1,
							Col:   3,
						},
						To: Position{
							Index: 16,
							Line:  1,
							Col:   16,
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
			result := newCallTemplateExpressionParser().Parse(input)
			if result.Error != nil {
				t.Fatalf("parser error: %v", result.Error)
			}
			if !result.Success {
				t.Errorf("failed to parse at %d", input.Index())
			}
			if diff := cmp.Diff(tt.expected, result.Item); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
