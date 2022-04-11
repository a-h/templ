package parser

import (
	"testing"

	"github.com/a-h/lexical/input"
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
							Line:  1,
							Col:   4,
						},
						To: Position{

							Index: 28,
							Line:  1,
							Col:   28,
						},
					},
				},
				Children: []Node{
					Whitespace{Value: "\t\t\t\t\t"},
					Element{
						Name:       "div",
						Attributes: []Attribute{},
						Children: []Node{
							StringExpression{
								Expression: Expression{
									Value: `item`,
									Range: Range{
										From: Position{
											Index: 43,
											Line:  2,
											Col:   12,
										},
										To: Position{

											Index: 47,
											Line:  2,
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
							Line:  1,
							Col:   4,
						},
						To: Position{

							Index: 28,
							Line:  1,
							Col:   28,
						},
					},
				},
				Children: []Node{
					Whitespace{Value: "\t\t\t\t\t"},
					Element{
						Name:       "div",
						Attributes: []Attribute{},
						Children: []Node{
							StringExpression{
								Expression: Expression{
									Value: `item`,
									Range: Range{
										From: Position{
											Index: 42,
											Line:  2,
											Col:   12,
										},
										To: Position{

											Index: 46,
											Line:  2,
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
			input := input.NewFromString(tt.input)
			result := forExpression.Parse(input)
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


