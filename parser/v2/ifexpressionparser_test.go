package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestIfExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected IfExpression
	}{
		{
			name: "if: simple expression",
			input: `if p.Test {
<span>
  { "span content" }
</span>
}
`,
			expected: IfExpression{
				Expression: Expression{
					Value: `p.Test`,
					Range: Range{
						From: Position{
							Index: 3,
							Line:  0,
							Col:   3,
						},
						To: Position{
							Index: 9,
							Line:  0,
							Col:   9,
						},
					},
				},
				Then: []Node{
					Element{
						Name: "span",
						NameRange: Range{
							From: Position{Index: 13, Line: 1, Col: 1},
							To:   Position{Index: 17, Line: 1, Col: 5},
						},

						Children: []Node{
							Whitespace{Value: "\n  "},
							StringExpression{
								Expression: Expression{
									Value: `"span content"`,
									Range: Range{
										From: Position{
											Index: 23,
											Line:  2,
											Col:   4,
										},
										To: Position{
											Index: 37,
											Line:  2,
											Col:   18,
										},
									},
								},
								TrailingSpace: SpaceVertical,
							},
						},
						IndentChildren: true,
						TrailingSpace:  SpaceVertical,
					},
				},
			},
		},
		{
			name: "if: else",
			input: `if p.A {
	{ "A" }
} else {
	{ "B" }
}`,
			expected: IfExpression{
				Expression: Expression{
					Value: `p.A`,
					Range: Range{
						From: Position{
							Index: 3,
							Line:  0,
							Col:   3,
						},
						To: Position{
							Index: 6,
							Line:  0,
							Col:   6,
						},
					},
				},
				Then: []Node{
					Whitespace{Value: "\t"},
					StringExpression{
						Expression: Expression{
							Value: `"A"`,
							Range: Range{
								From: Position{
									Index: 12,
									Line:  1,
									Col:   3,
								},
								To: Position{
									Index: 15,
									Line:  1,
									Col:   6,
								},
							},
						},
						TrailingSpace: SpaceVertical,
					},
				},
				Else: []Node{
					StringExpression{
						Expression: Expression{
							Value: `"B"`,
							Range: Range{
								From: Position{
									Index: 30,
									Line:  3,
									Col:   3,
								},
								To: Position{
									Index: 33,
									Line:  3,
									Col:   6,
								},
							},
						},
						TrailingSpace: SpaceVertical,
					},
				},
			},
		},
		{
			name: "if: expressions can have a space after the opening brace",
			input: `if p.Test { 
  text
}
`,
			expected: IfExpression{
				Expression: Expression{
					Value: `p.Test`,
					Range: Range{
						From: Position{
							Index: 3,
							Line:  0,
							Col:   3,
						},
						To: Position{
							Index: 9,
							Line:  0,
							Col:   9,
						},
					},
				},
				Then: []Node{
					Whitespace{Value: "  "},
					Text{
						Value: "text",
						Range: Range{
							From: Position{Index: 15, Line: 1, Col: 2},
							To:   Position{Index: 19, Line: 1, Col: 6},
						},
						TrailingSpace: SpaceVertical,
					},
				},
			},
		},
		{
			name: "if: simple expression, without spaces",
			input: `if p.Test {
<span>
  { "span content" }
</span>
}
`,
			expected: IfExpression{
				Expression: Expression{
					Value: `p.Test`,
					Range: Range{
						From: Position{
							Index: 3,
							Line:  0,
							Col:   3,
						},
						To: Position{
							Index: 9,
							Line:  0,
							Col:   9,
						},
					},
				},
				Then: []Node{
					Element{
						Name: "span",
						NameRange: Range{
							From: Position{Index: 13, Line: 1, Col: 1},
							To:   Position{Index: 17, Line: 1, Col: 5},
						},

						Children: []Node{
							Whitespace{Value: "\n  "},
							StringExpression{
								Expression: Expression{
									Value: `"span content"`,
									Range: Range{
										From: Position{
											Index: 23,
											Line:  2,
											Col:   4,
										},
										To: Position{
											Index: 37,
											Line:  2,
											Col:   18,
										},
									},
								},
								TrailingSpace: SpaceVertical,
							},
						},
						IndentChildren: true,
						TrailingSpace:  SpaceVertical,
					},
				},
			},
		},
		{
			name: "if: else, without spaces",
			input: `if p.A{
	{ "A" }
} else {
	{ "B" }
}`,
			expected: IfExpression{
				Expression: Expression{
					Value: `p.A`,
					Range: Range{
						From: Position{
							Index: 3,
							Line:  0,
							Col:   3,
						},
						To: Position{
							Index: 6,
							Line:  0,
							Col:   6,
						},
					},
				},
				Then: []Node{
					Whitespace{Value: "\t"},
					StringExpression{
						Expression: Expression{
							Value: `"A"`,
							Range: Range{
								From: Position{
									Index: 11,
									Line:  1,
									Col:   3,
								},
								To: Position{
									Index: 14,
									Line:  1,
									Col:   6,
								},
							},
						},
						TrailingSpace: SpaceVertical,
					},
				},
				Else: []Node{
					StringExpression{
						Expression: Expression{
							Value: `"B"`,
							Range: Range{
								From: Position{
									Index: 29,
									Line:  3,
									Col:   3,
								},
								To: Position{
									Index: 32,
									Line:  3,
									Col:   6,
								},
							},
						},
						TrailingSpace: SpaceVertical,
					},
				},
			},
		},
		{
			name: "if: nested",
			input: `if p.A {
					if p.B {
						<div>{ "B" }</div>
					}
				}`,
			expected: IfExpression{
				Expression: Expression{
					Value: `p.A`,
					Range: Range{
						From: Position{
							Index: 3,
							Line:  0,
							Col:   3,
						},
						To: Position{
							Index: 6,
							Line:  0,
							Col:   6,
						},
					},
				},
				Then: []Node{
					Whitespace{Value: "\t\t\t\t\t"},
					IfExpression{
						Expression: Expression{
							Value: `p.B`,
							Range: Range{
								From: Position{
									Index: 17,
									Line:  1,
									Col:   8,
								},
								To: Position{
									Index: 20,
									Line:  1,
									Col:   11,
								},
							},
						},
						Then: []Node{
							Whitespace{Value: "\t\t\t\t\t\t"},
							Element{
								Name: "div",
								NameRange: Range{
									From: Position{Index: 30, Line: 2, Col: 7},
									To:   Position{Index: 33, Line: 2, Col: 10},
								},

								Children: []Node{
									StringExpression{
										Expression: Expression{
											Value: `"B"`,
											Range: Range{
												From: Position{
													Index: 36,
													Line:  2,
													Col:   13,
												},
												To: Position{
													Index: 39,
													Line:  2,
													Col:   16,
												},
											},
										},
									},
								},
								TrailingSpace: SpaceVertical,
							},
						},
					},
					Whitespace{Value: "\n\t\t\t\t"},
				},
			},
		},
		{
			name: "if: else if",
			input: `if p.A {
	{ "A" }
} else if p.B {
	{ "B" }
}`,
			expected: IfExpression{
				Expression: Expression{
					Value: `p.A`,
					Range: Range{
						From: Position{Index: 3, Line: 0, Col: 3},
						To:   Position{Index: 6, Line: 0, Col: 6},
					},
				},
				Then: []Node{
					Whitespace{Value: "\t"},
					StringExpression{
						Expression: Expression{
							Value: `"A"`,
							Range: Range{
								From: Position{Index: 12, Line: 1, Col: 3},
								To:   Position{Index: 15, Line: 1, Col: 6},
							},
						},
						TrailingSpace: SpaceVertical,
					},
				},
				ElseIfs: []ElseIfExpression{
					{
						Expression: Expression{
							Value: `p.B`,
							Range: Range{
								From: Position{Index: 28, Line: 2, Col: 10},
								To:   Position{Index: 31, Line: 2, Col: 13},
							},
						},
						Then: []Node{
							Whitespace{Value: "\t"},
							StringExpression{
								Expression: Expression{
									Value: `"B"`,
									Range: Range{
										From: Position{Index: 37, Line: 3, Col: 3},
										To:   Position{Index: 40, Line: 3, Col: 6},
									},
								},
								TrailingSpace: SpaceVertical,
							},
						},
					},
				},
			},
		},
		{
			name: "if: else if, else if",
			input: `if p.A {
	{ "A" }
} else if p.B {
	{ "B" }
} else if p.C {
	{ "C" }
}`,
			expected: IfExpression{
				Expression: Expression{
					Value: `p.A`,
					Range: Range{
						From: Position{Index: 3, Line: 0, Col: 3},
						To:   Position{Index: 6, Line: 0, Col: 6},
					},
				},
				Then: []Node{
					Whitespace{Value: "\t"},
					StringExpression{
						Expression: Expression{
							Value: `"A"`,
							Range: Range{
								From: Position{Index: 12, Line: 1, Col: 3},
								To:   Position{Index: 15, Line: 1, Col: 6},
							},
						},
						TrailingSpace: SpaceVertical,
					},
				},
				ElseIfs: []ElseIfExpression{
					{
						Expression: Expression{
							Value: `p.B`,
							Range: Range{
								From: Position{Index: 28, Line: 2, Col: 10},
								To:   Position{Index: 31, Line: 2, Col: 13},
							},
						},
						Then: []Node{
							Whitespace{Value: "\t"},
							StringExpression{
								Expression: Expression{
									Value: `"B"`,
									Range: Range{
										From: Position{Index: 37, Line: 3, Col: 3},
										To:   Position{Index: 40, Line: 3, Col: 6},
									},
								},
								TrailingSpace: SpaceVertical,
							},
						},
					},
					{
						Expression: Expression{
							Value: `p.C`,
							Range: Range{
								From: Position{Index: 53, Line: 4, Col: 10},
								To:   Position{Index: 56, Line: 4, Col: 13},
							},
						},
						Then: []Node{
							Whitespace{Value: "\t"},
							StringExpression{
								Expression: Expression{
									Value: `"C"`,
									Range: Range{
										From: Position{Index: 62, Line: 5, Col: 3},
										To:   Position{Index: 65, Line: 5, Col: 6},
									},
								},
								TrailingSpace: SpaceVertical,
							},
						},
					},
				},
			},
		},
		{
			name: "if: else if, else if, else",
			input: `if p.A {
	{ "A" }
} else if p.B {
	{ "B" }
} else if p.C {
	{ "C" }
} else {
	{ "D" }
}`,
			expected: IfExpression{
				Expression: Expression{
					Value: `p.A`,
					Range: Range{
						From: Position{Index: 3, Line: 0, Col: 3},
						To:   Position{Index: 6, Line: 0, Col: 6},
					},
				},
				Then: []Node{
					Whitespace{Value: "\t"},
					StringExpression{
						Expression: Expression{
							Value: `"A"`,
							Range: Range{
								From: Position{Index: 12, Line: 1, Col: 3},
								To:   Position{Index: 15, Line: 1, Col: 6},
							},
						},
						TrailingSpace: SpaceVertical,
					},
				},
				ElseIfs: []ElseIfExpression{
					{
						Expression: Expression{
							Value: `p.B`,
							Range: Range{
								From: Position{Index: 28, Line: 2, Col: 10},
								To:   Position{Index: 31, Line: 2, Col: 13},
							},
						},
						Then: []Node{
							Whitespace{Value: "\t"},
							StringExpression{
								Expression: Expression{
									Value: `"B"`,
									Range: Range{
										From: Position{Index: 37, Line: 3, Col: 3},
										To:   Position{Index: 40, Line: 3, Col: 6},
									},
								},
								TrailingSpace: SpaceVertical,
							},
						},
					},
					{
						Expression: Expression{
							Value: `p.C`,
							Range: Range{
								From: Position{Index: 53, Line: 4, Col: 10},
								To:   Position{Index: 56, Line: 4, Col: 13},
							},
						},
						Then: []Node{
							Whitespace{Value: "\t"},
							StringExpression{
								Expression: Expression{
									Value: `"C"`,
									Range: Range{
										From: Position{Index: 62, Line: 5, Col: 3},
										To:   Position{Index: 65, Line: 5, Col: 6},
									},
								},
								TrailingSpace: SpaceVertical,
							},
						},
					},
				},
				Else: []Node{
					StringExpression{
						Expression: Expression{
							Value: `"D"`,
							Range: Range{
								From: Position{Index: 80, Line: 7, Col: 3},
								To:   Position{Index: 83, Line: 7, Col: 6},
							},
						},
						TrailingSpace: SpaceVertical,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			actual, ok, err := ifExpression.Parse(input)
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

func TestIncompleteIf(t *testing.T) {
	t.Run("no opening brace", func(t *testing.T) {
		input := parse.NewInput(`if a tree falls in the woods`)
		_, _, err := ifExpression.Parse(input)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		pe, isParseError := err.(parse.ParseError)
		if !isParseError {
			t.Fatalf("expected a parse error, got %T", err)
		}
		if pe.Msg != "if: unterminated (missing closing '{\\n') - https://templ.guide/syntax-and-usage/statements#incomplete-statements" {
			t.Fatalf("unexpected error: %v", err)
		}
		if pe.Pos.Line != 0 {
			t.Fatalf("unexpected line: %d", pe.Pos.Line)
		}
	})
	t.Run("capitalised If", func(t *testing.T) {
		input := parse.NewInput(`If a tree falls in the woods`)
		_, ok, err := ifExpression.Parse(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatal("expected a non match")
		}
	})
}
