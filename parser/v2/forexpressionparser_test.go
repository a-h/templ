package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestForExpressionParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name: "for: simple",
			input: `for _, item := range p.Items {
					<div>{ item }</div>
				}`,
			expected: ForExpression{
				Expression: Expression{
					Value: `_, item := range p.Items`,
					Range: Range{
						From: Position{
							Index: 4,
							Line:  0,
							Col:   4,
						},
						To: Position{

							Index: 28,
							Line:  0,
							Col:   28,
						},
					},
				},
				Children: []Node{
					Whitespace{Value: "\t\t\t\t\t"},
					Element{
						Name: "div",
						Children: []Node{
							StringExpression{
								Expression: Expression{
									Value: `item`,
									Range: Range{
										From: Position{
											Index: 43,
											Line:  1,
											Col:   12,
										},
										To: Position{

											Index: 47,
											Line:  1,
											Col:   16,
										},
									},
								},
							},
						},
					},
					Whitespace{Value: "\n\t\t\t\t"},
				},
			},
		},
		{
			name: "for: simple, without spaces",
			input: `for _, item := range p.Items{
					<div>{ item }</div>
				}`,
			expected: ForExpression{
				Expression: Expression{
					Value: `_, item := range p.Items`,
					Range: Range{
						From: Position{
							Index: 4,
							Line:  0,
							Col:   4,
						},
						To: Position{

							Index: 28,
							Line:  0,
							Col:   28,
						},
					},
				},
				Children: []Node{
					Whitespace{Value: "\t\t\t\t\t"},
					Element{
						Name: "div",
						Children: []Node{
							StringExpression{
								Expression: Expression{
									Value: `item`,
									Range: Range{
										From: Position{
											Index: 42,
											Line:  1,
											Col:   12,
										},
										To: Position{

											Index: 46,
											Line:  1,
											Col:   16,
										},
									},
								},
							},
						},
					},
					Whitespace{Value: "\n\t\t\t\t"},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			actual, ok, err := forExpression.Parse(input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !ok {
				t.Fatalf("unexpected failure for input %q", tt.input)
			}
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Error(diff)
			}
		})
	}
}
