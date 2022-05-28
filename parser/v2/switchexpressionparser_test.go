package parser

import (
	"testing"

	"github.com/a-h/lexical/input"
	"github.com/google/go-cmp/cmp"
)

func TestSwitchExpressionParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected SwitchExpression
	}{
		{
			name: "switch: simple",
			input: `switch "stringy" {
}`,
			expected: SwitchExpression{
				Expression: Expression{
					Value: `"stringy"`,
					Range: Range{
						From: Position{
							Index: 7,
							Line:  0,
							Col:   7,
						},
						To: Position{
							Index: 16,
							Line:  0,
							Col:   16,
						},
					},
				},
				Cases: []CaseExpression{},
			},
		},
		{
			name: "switch: default only",
			input: `switch "stringy" {
default:
	<span>
	  { "span content" }
	</span>
}`,
			expected: SwitchExpression{
				Expression: Expression{
					Value: `"stringy"`,
					Range: Range{
						From: Position{
							Index: 7,
							Line:  0,
							Col:   7,
						},
						To: Position{
							Index: 16,
							Line:  0,
							Col:   16,
						},
					},
				},
				Cases: []CaseExpression{
					{
						Expression: Expression{
							Value: "default:\n",
							Range: Range{
								From: Position{
									Index: 19,
									Line:  1,
									Col:   0,
								},
								To: Position{
									Index: 28,
									Line:  2,
									Col:   0,
								},
							},
						},
						Children: []Node{
							Whitespace{Value: "\t"},
							Element{
								Name:       "span",
								Attributes: []Attribute{},
								Children: []Node{
									Whitespace{Value: "\n\t  "},
									StringExpression{
										Expression: Expression{
											Value: `"span content"`,
											Range: Range{
												From: Position{
													Index: 41,
													Line:  3,
													Col:   5,
												},
												To: Position{
													Index: 55,
													Line:  3,
													Col:   19,
												},
											},
										},
									},
									Whitespace{Value: "\n\t"},
								},
							},
							Whitespace{Value: "\n"},
						},
					},
				},
			},
		},
		{
			name: "switch: one case",
			input: `switch "stringy" {
	case "stringy":
<span>
  { "span content" }
</span>
}`,
			expected: SwitchExpression{
				Expression: Expression{
					Value: `"stringy"`,
					Range: Range{
						From: Position{
							Index: 7,
							Line:  0,
							Col:   7,
						},
						To: Position{
							Index: 16,
							Line:  0,
							Col:   16,
						},
					},
				},
				Cases: []CaseExpression{
					{
						Expression: Expression{
							Value: "case \"stringy\":\n",
							Range: Range{
								From: Position{
									Index: 20,
									Line:  1,
									Col:   1,
								},
								To: Position{
									Index: 36,
									Line:  2,
									Col:   0,
								},
							},
						},
						Children: []Node{
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
													Index: 47,
													Line:  3,
													Col:   4,
												},
												To: Position{
													Index: 61,
													Line:  3,
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
					},
				},
			},
		},
		{
			name: "switch: two cases",
			input: `switch "stringy" {
	case "a":
		{ "A" }
	case "b":
		{ "B" }
}`,
			expected: SwitchExpression{
				Expression: Expression{
					Value: `"stringy"`,
					Range: Range{
						From: Position{
							Index: 7,
							Line:  0,
							Col:   7,
						},
						To: Position{
							Index: 16,
							Line:  0,
							Col:   16,
						},
					},
				},
				Cases: []CaseExpression{
					{
						Expression: Expression{
							Value: "case \"a\":\n",
							Range: Range{
								From: Position{
									Index: 20,
									Line:  1,
									Col:   1,
								},
								To: Position{
									Index: 30,
									Line:  2,
									Col:   0,
								},
							},
						},
						Children: []Node{
							Whitespace{
								Value: "\t\t",
							},
							StringExpression{
								Expression: Expression{
									Value: `"A"`,
									Range: Range{
										From: Position{
											Index: 34,
											Line:  2,
											Col:   4,
										},
										To: Position{
											Index: 37,
											Line:  2,
											Col:   7,
										},
									},
								},
							},
						},
					},
					{
						Expression: Expression{
							Value: "case \"b\":\n",
							Range: Range{
								From: Position{
									Index: 41,
									Line:  3,
									Col:   1,
								},
								To: Position{
									Index: 51,
									Line:  4,
									Col:   0,
								},
							},
						},
						Children: []Node{
							Whitespace{
								Value: "\t\t",
							},
							StringExpression{
								Expression: Expression{
									Value: `"B"`,
									Range: Range{
										From: Position{
											Index: 55,
											Line:  4,
											Col:   4,
										},
										To: Position{
											Index: 58,
											Line:  4,
											Col:   7,
										},
									},
								},
							},
							Whitespace{
								Value: "\n",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := input.NewFromString(tt.input)
			result := switchExpression.Parse(input)
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
