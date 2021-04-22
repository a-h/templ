package templ

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
			input: `{% call Other(p.Test) %}`,
			expected: CallTemplateExpression{
				Name: Expression{
					Value: "Other",
					Range: Range{
						From: Position{
							Index: 8,
							Line:  1,
							Col:   8,
						},
						To: Position{
							Index: 12,
							Line:  1,
							Col:   12,
						},
					},
				},
				Arguments: Expression{
					Value: `p.Test`,
					Range: Range{
						From: Position{
							Index: 14,
							Line:  1,
							Col:   14,
						},
						To: Position{
							Index: 20,
							Line:  1,
							Col:   20,
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
