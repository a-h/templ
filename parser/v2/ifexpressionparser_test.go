package parser

import (
	"testing"

	"github.com/a-h/lexical/input"
	"github.com/google/go-cmp/cmp"
)

func TestIfExpression(t *testing.T) {
	var tests = []struct {
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
						Name:       "span",
						Attributes: []Attribute{},
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
							},
							Whitespace{Value: "\n"},
						},
					},
					Whitespace{Value: "\n"},
				},
				Else: []Node{},
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
					},
					Whitespace{Value: "\n"},
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
					},
					Whitespace{Value: "\n"},
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
						Name:       "span",
						Attributes: []Attribute{},
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
							},
							Whitespace{Value: "\n"},
						},
					},
					Whitespace{Value: "\n"},
				},
				Else: []Node{},
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
					},
					Whitespace{Value: "\n"},
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
					},
					Whitespace{Value: "\n"},
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
								Name:       "div",
								Attributes: []Attribute{},
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
							},
							Whitespace{Value: "\n\t\t\t\t\t"},
						},
						Else: []Node{},
					},
					Whitespace{Value: "\n\t\t\t\t"},
				},
				Else: []Node{},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := input.NewFromString(tt.input)
			result := ifExpression.Parse(input)
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
