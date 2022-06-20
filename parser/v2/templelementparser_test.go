package parser

import (
	"testing"

	"github.com/a-h/lexical/input"
	"github.com/google/go-cmp/cmp"
)

func TestTemplElementExpressionParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected TemplElementExpression
	}{
		{
			name:  "templelement: simple",
			input: `@Other(p.Test)` + "\n",
			expected: TemplElementExpression{
				Expression: Expression{
					Value: "Other(p.Test)",
					Range: Range{
						From: Position{
							Index: 1,
							Line:  0,
							Col:   1,
						},
						To: Position{
							Index: 14,
							Line:  0,
							Col:   14,
						},
					},
				},
			},
		},
		{
			name: "templelement: simple, block with text",
			input: `@Other(p.Test) {
	some words
}`,
			expected: TemplElementExpression{
				Expression: Expression{
					Value: "Other(p.Test)",
					Range: Range{
						From: Position{
							Index: 1,
							Line:  0,
							Col:   1,
						},
						To: Position{
							Index: 14,
							Line:  0,
							Col:   14,
						},
					},
				},
				Children: []Node{
					Whitespace{Value: "\t"},
					Text{Value: "some words"},
					Whitespace{Value: "\n"},
				},
			},
		},
		{
			name: "templelement: simple, block with anchor",
			input: `@Other(p.Test){
			<a href="someurl" />
		}`,
			expected: TemplElementExpression{
				Expression: Expression{
					Value: "Other(p.Test)",
					Range: Range{
						From: Position{
							Index: 1,
							Line:  0,
							Col:   1,
						},
						To: Position{
							Index: 14,
							Line:  0,
							Col:   14,
						},
					},
				},
				Children: []Node{
					Whitespace{Value: "\t\t\t"},
					Element{Name: "a", Attributes: []Attribute{
						ConstantAttribute{"href", "someurl"},
					}},
					Whitespace{Value: "\n\t\t"},
				},
			},
		},
		{
			name: "templelement: simple, block with templelement as child",
			input: `@Other(p.Test) {
				@other2
			}`,
			expected: TemplElementExpression{
				Expression: Expression{
					Value: "Other(p.Test)",
					Range: Range{
						From: Position{
							Index: 1,
							Line:  0,
							Col:   1,
						},
						To: Position{
							Index: 14,
							Line:  0,
							Col:   14,
						},
					},
				},
				Children: []Node{
					Whitespace{Value: "\t\t\t\t"},
					TemplElementExpression{
						Expression: Expression{
							Value: "other2",
							Range: Range{
								From: Position{22, 1, 5},
								To:   Position{28, 1, 11},
							},
						},
					},
					Whitespace{Value: "\n\t\t\t"},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := input.NewFromString(tt.input)
			result := templElementExpression.Parse(input)
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
