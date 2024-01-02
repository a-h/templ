package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestTemplElementExpressionParser(t *testing.T) {
	tests := []struct {
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
			name:  "templelement: simple with underscore",
			input: `@Other_Component(p.Test)` + "\n",
			expected: TemplElementExpression{
				Expression: Expression{
					Value: "Other_Component(p.Test)",
					Range: Range{
						From: Position{
							Index: 1,
							Line:  0,
							Col:   1,
						},
						To: Position{
							Index: 24,
							Line:  0,
							Col:   24,
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
					Whitespace{Value: "\n\t"},
					Text{Value: "some words",
						TrailingSpace: SpaceVertical,
					},
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
					Whitespace{Value: "\n\t\t\t"},
					Element{Name: "a", Attributes: []Attribute{
						ConstantAttribute{
							Name:  "href",
							Value: "someurl",
						},
					},
						TrailingSpace: SpaceVertical,
					},
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
					Whitespace{Value: "\n\t\t\t\t"},
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
		{
			name: "templelement: can parse the initial expression and leave the text",
			input: `@Icon("home", Inline) Home</a>
}`,
			expected: TemplElementExpression{
				Expression: Expression{
					Value: `Icon("home", Inline)`,
					Range: Range{
						From: Position{
							Index: 1,
							Line:  0,
							Col:   1,
						},
						To: Position{
							Index: 21,
							Line:  0,
							Col:   21,
						},
					},
				},
			},
		},
		{
			name:  "templelement: supports the use of templ elements in other packages",
			input: `@templates.Icon("home", Inline)`,
			expected: TemplElementExpression{
				Expression: Expression{
					Value: `templates.Icon("home", Inline)`,
					Range: Range{
						From: Position{
							Index: 1,
							Line:  0,
							Col:   1,
						},
						To: Position{
							Index: 31,
							Line:  0,
							Col:   31,
						},
					},
				},
			},
		},
		{
			name:  "templelement: supports the use of params which contain braces and params",
			input: `@templates.New(test{}, other())`,
			expected: TemplElementExpression{
				Expression: Expression{
					Value: `templates.New(test{}, other())`,
					Range: Range{
						From: Position{
							Index: 1,
							Line:  0,
							Col:   1,
						},
						To: Position{
							Index: 31,
							Line:  0,
							Col:   31,
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
			actual, ok, err := templElementExpression.Parse(input)
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
