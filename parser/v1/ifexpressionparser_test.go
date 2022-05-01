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
			input: `{% if p.Test %}
<span>
  {%= "span content" %}
</span>
{% endif %}
`,
			expected: IfExpression{
				Expression: Expression{
					Value: `p.Test`,
					Range: Range{
						From: Position{
							Index: 6,
							Line:  1,
							Col:   6,
						},
						To: Position{
							Index: 12,
							Line:  1,
							Col:   12,
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
											Index: 29,
											Line:  3,
											Col:   6,
										},
										To: Position{
											Index: 43,
											Line:  3,
											Col:   20,
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
			input: `{% if p.A %}
	{%= "A" %}
{% else %}
	{%= "B" %}
{% endif %}`,
			expected: IfExpression{
				Expression: Expression{
					Value: `p.A`,
					Range: Range{
						From: Position{
							Index: 6,
							Line:  1,
							Col:   6,
						},
						To: Position{
							Index: 9,
							Line:  1,
							Col:   9,
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
									Index: 18,
									Line:  2,
									Col:   5,
								},
								To: Position{
									Index: 21,
									Line:  2,
									Col:   8,
								},
							},
						},
					},
					Whitespace{Value: "\n"},
				},
				Else: []Node{
					Whitespace{Value: "\n\t"},
					StringExpression{
						Expression: Expression{
							Value: `"B"`,
							Range: Range{
								From: Position{
									Index: 41,
									Line:  4,
									Col:   5,
								},
								To: Position{
									Index: 44,
									Line:  4,
									Col:   8,
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
			input: `{%if p.Test%}
<span>
  {%= "span content" %}
</span>
{%endif%}
`,
			expected: IfExpression{
				Expression: Expression{
					Value: `p.Test`,
					Range: Range{
						From: Position{
							Index: 5,
							Line:  1,
							Col:   5,
						},
						To: Position{
							Index: 11,
							Line:  1,
							Col:   11,
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
											Index: 27,
											Line:  3,
											Col:   6,
										},
										To: Position{
											Index: 41,
											Line:  3,
											Col:   20,
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
			input: `{%if p.A%}
	{%= "A" %}
{%else%}
	{%= "B" %}
{%endif%}`,
			expected: IfExpression{
				Expression: Expression{
					Value: `p.A`,
					Range: Range{
						From: Position{
							Index: 5,
							Line:  1,
							Col:   5,
						},
						To: Position{
							Index: 8,
							Line:  1,
							Col:   8,
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
									Index: 16,
									Line:  2,
									Col:   5,
								},
								To: Position{
									Index: 19,
									Line:  2,
									Col:   8,
								},
							},
						},
					},
					Whitespace{Value: "\n"},
				},
				Else: []Node{
					Whitespace{Value: "\n\t"},
					StringExpression{
						Expression: Expression{
							Value: `"B"`,
							Range: Range{
								From: Position{
									Index: 37,
									Line:  4,
									Col:   5,
								},
								To: Position{
									Index: 40,
									Line:  4,
									Col:   8,
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
			input: `{% if p.A %}
					{% if p.B %}
						<div>{%= "B" %}</div>
					{% endif %}
				{% endif %}`,
			expected: IfExpression{
				Expression: Expression{
					Value: `p.A`,
					Range: Range{
						From: Position{
							Index: 6,
							Line:  1,
							Col:   6,
						},
						To: Position{
							Index: 9,
							Line:  1,
							Col:   9,
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
									Index: 24,
									Line:  2,
									Col:   11,
								},
								To: Position{
									Index: 27,
									Line:  2,
									Col:   14,
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
													Index: 46,
													Line:  3,
													Col:   15,
												},
												To: Position{
													Index: 49,
													Line:  3,
													Col:   18,
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
			result := newIfExpressionParser().Parse(input)
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
