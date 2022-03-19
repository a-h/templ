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
			input: `{% for _, item := range p.Items %}
					<div>{%= item %}</div>
				{% endfor %}`,
			expected: ForExpression{
				Expression: Expression{
					Value: `_, item := range p.Items`,
					Range: Range{
						From: Position{
							Index: 7,
							Line:  1,
							Col:   7,
						},
						To: Position{

							Index: 31,
							Line:  1,
							Col:   31,
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
											Index: 49,
											Line:  2,
											Col:   14,
										},
										To: Position{

											Index: 53,
											Line:  2,
											Col:   18,
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
			input: `{%for _, item := range p.Items%}
					<div>{%= item %}</div>
				{% endfor %}`,
			expected: ForExpression{
				Expression: Expression{
					Value: `_, item := range p.Items`,
					Range: Range{
						From: Position{
							Index: 6,
							Line:  1,
							Col:   6,
						},
						To: Position{

							Index: 30,
							Line:  1,
							Col:   30,
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
											Index: 47,
											Line:  2,
											Col:   14,
										},
										To: Position{

											Index: 51,
											Line:  2,
											Col:   18,
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
			result := newForExpressionParser().Parse(input)
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
