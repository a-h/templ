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
			name: "templelement: simple multiline call",
			input: `@Other_Component(
				p.Test,
				"something" + "else",
			)` + "\n",
			expected: TemplElementExpression{
				Expression: Expression{
					Value: `Other_Component(
				p.Test,
				"something" + "else",
			)`,
					Range: Range{
						From: Position{
							Index: 1,
							Line:  0,
							Col:   1,
						},
						To: Position{
							Index: 60,
							Line:  3,
							Col:   4,
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
					Text{
						Value: "some words",
						Range: Range{
							From: Position{Index: 18, Line: 1, Col: 1},
							To:   Position{Index: 28, Line: 1, Col: 11},
						},
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
					Element{Name: "a",
						NameRange: Range{
							From: Position{Index: 20, Line: 1, Col: 4},
							To:   Position{Index: 21, Line: 1, Col: 5},
						},
						Attributes: []Attribute{
							ConstantAttribute{
								Name:  "href",
								Value: "someurl",
								NameRange: Range{
									From: Position{Index: 22, Line: 1, Col: 6},
									To:   Position{Index: 26, Line: 1, Col: 10},
								},
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
		{
			name:  "templelement: supports a slice of functions",
			input: `@templates[0]()`,
			expected: TemplElementExpression{
				Expression: Expression{
					Value: `templates[0]()`,
					Range: Range{
						From: Position{
							Index: 1,
							Line:  0,
							Col:   1,
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
			name:  "templelement: supports a map of functions",
			input: `@templates["key"]()`,
			expected: TemplElementExpression{
				Expression: Expression{
					Value: `templates["key"]()`,
					Range: Range{
						From: Position{
							Index: 1,
							Line:  0,
							Col:   1,
						},
						To: Position{
							Index: 19,
							Line:  0,
							Col:   19,
						},
					},
				},
			},
		},
		{
			name:  "templelement: supports a slice of structs/interfaces",
			input: `@templates[0].CreateTemplate()`,
			expected: TemplElementExpression{
				Expression: Expression{
					Value: `templates[0].CreateTemplate()`,
					Range: Range{
						From: Position{
							Index: 1,
							Line:  0,
							Col:   1,
						},
						To: Position{
							Index: 30,
							Line:  0,
							Col:   30,
						},
					},
				},
			},
		},
		{
			name:  "templelement: supports a slice of structs/interfaces",
			input: `@templates[0].CreateTemplate()`,
			expected: TemplElementExpression{
				Expression: Expression{
					Value: `templates[0].CreateTemplate()`,
					Range: Range{
						From: Position{
							Index: 1,
							Line:  0,
							Col:   1,
						},
						To: Position{
							Index: 30,
							Line:  0,
							Col:   30,
						},
					},
				},
			},
		},
		{
			name:  "templelement: bare variables are read until the end of the token",
			input: `@template</div>`,
			expected: TemplElementExpression{
				Expression: Expression{
					Value: `template`,
					Range: Range{
						From: Position{
							Index: 1,
							Line:  0,
							Col:   1,
						},
						To: Position{
							Index: 9,
							Line:  0,
							Col:   9,
						},
					},
				},
			},
		},
		{
			name:  "templelement: struct literal method calls are supported",
			input: `@layout.DefaultLayout{}.Compile()<div>`,
			expected: TemplElementExpression{
				Expression: Expression{
					Value: `layout.DefaultLayout{}.Compile()`,
					Range: Range{
						From: Position{1, 0, 1},
						To:   Position{33, 0, 33},
					},
				},
			},
		},
		{
			name: "templelement: struct literal method calls are supported, with child elements",
			input: `@layout.DefaultLayout{}.Compile() {
  <div>hello</div>
}`,
			expected: TemplElementExpression{
				Expression: Expression{
					Value: `layout.DefaultLayout{}.Compile()`,
					Range: Range{
						From: Position{1, 0, 1},
						To:   Position{33, 0, 33},
					},
				},
				Children: []Node{
					Whitespace{Value: "\n  "},
					Element{
						Name: "div",
						NameRange: Range{
							From: Position{Index: 39, Line: 1, Col: 3},
							To:   Position{Index: 42, Line: 1, Col: 6},
						},
						Children: []Node{
							Text{
								Value: "hello",
								Range: Range{
									From: Position{Index: 43, Line: 1, Col: 7},
									To:   Position{Index: 48, Line: 1, Col: 12},
								},
							},
						},
						TrailingSpace: SpaceVertical,
					},
				},
			},
		},
		{
			name: "templelement: arguments can receive a slice of complex types",
			input: `@tabs([]*TabData{
  {Name: "A"},
  {Name: "B"},
})`,
			expected: TemplElementExpression{
				Expression: Expression{
					Value: `tabs([]*TabData{
  {Name: "A"},
  {Name: "B"},
})`,
					Range: Range{
						From: Position{1, 0, 1},
						To:   Position{50, 3, 2},
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
