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
			name:  "tempelement: simple",
			input: `<! Other(p.Test) />`,
			expected: TemplElementExpression{
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
			},
		},
		{
			name:  "tempelement: simple, missing start space",
			input: `<!Other(p.Test) />`,
			expected: TemplElementExpression{
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
			},
		},
		{
			name:  "tempelement: simple, missing start and end space",
			input: `<!Other(p.Test)/>`,
			expected: TemplElementExpression{
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
			},
		},
		{
			name: "tempelement: simple, block with text",
			input: `<!Other(p.Test)>
			some words
			</>`,
			expected: TemplElementExpression{
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
				Children: []Node{
					Whitespace{Value: "\n\t\t\t"},
					Text{Value: "some words"},
					Whitespace{Value: "\n\t\t\t"},
				},
			},
		},
		{
			name: "tempelement: simple, block with anchor",
			input: `<!Other(p.Test)>
			<a href="someurl" />
			</>`,
			expected: TemplElementExpression{
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
				Children: []Node{
					Whitespace{Value: "\n\t\t\t"},
					Element{Name: "a", Attributes: []Attribute{
						ConstantAttribute{"href", "someurl"},
					}},
					Whitespace{Value: "\n\t\t\t"},
				},
			},
		},
		{
			name: "tempelement: simple, block with templelement as child",
			input: `<!Other(p.Test)>
				<!other2 />
			</>`,
			expected: TemplElementExpression{
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
				Children: []Node{
					Whitespace{Value: "\n\t\t\t\t"},
					TemplElementExpression{
						Expression: Expression{
							Value: "other2",
							Range: Range{
								From: Position{23, 1, 6},
								To:   Position{29, 1, 12},
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
