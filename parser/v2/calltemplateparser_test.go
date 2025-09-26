package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestCallTemplateExpressionParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *CallTemplateExpression
	}{
		{
			name:  "call: simple",
			input: `{! Other(p.Test) }`,
			expected: &CallTemplateExpression{
				Expression: Expression{
					Value: "Other(p.Test)",
					Range: Range{
						From: Position{
							Index: 3,
							Line:  0,
							Col:   3,
						},
						To: Position{
							Index: 16,
							Line:  0,
							Col:   16,
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 18, Line: 0, Col: 18},
				},
			},
		},
		{
			name:  "call: simple, missing start space",
			input: `{!Other(p.Test) }`,
			expected: &CallTemplateExpression{
				Expression: Expression{
					Value: "Other(p.Test)",
					Range: Range{
						From: Position{
							Index: 2,
							Line:  0,
							Col:   2,
						},
						To: Position{
							Index: 15,
							Line:  0,
							Col:   15,
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 17, Line: 0, Col: 17},
				},
			},
		},
		{
			name:  "call: simple, missing start and end space",
			input: `{!Other(p.Test)}`,
			expected: &CallTemplateExpression{
				Expression: Expression{
					Value: "Other(p.Test)",
					Range: Range{
						From: Position{
							Index: 2,
							Line:  0,
							Col:   2,
						},
						To: Position{
							Index: 15,
							Line:  0,
							Col:   15,
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 16, Line: 0, Col: 16},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			result, ok, err := callTemplateExpression.Parse(input)
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

func TestCallTemplateParserAllocsSkip(t *testing.T) {
	RunParserAllocTest(t, callTemplateExpression, false, 0, ``)
}
