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
		expected interface{}
	}{
		{
			name: "switch: simple",
			input: `{% switch "stringy" %}
{% endswitch %}`,
			expected: SwitchExpression{
				Expression: Expression{
					Value: `"stringy"`,
					Range: Range{
						From: Position{
							Index: 10,
							Line:  1,
							Col:   10,
						},
						To: Position{
							Index: 19,
							Line:  1,
							Col:   19,
						},
					},
				},
				Default: nil,
				Cases:   []CaseExpression{},
			},
		},
		{
			name: "switch: default only",
			input: `{% switch "stringy" %}
{% default %}
<span>
  {%= "span content" %}
</span>
{% enddefault %}
{% endswitch %}`,
			expected: SwitchExpression{
				Expression: Expression{
					Value: `"stringy"`,
					Range: Range{
						From: Position{
							Index: 10,
							Line:  1,
							Col:   10,
						},
						To: Position{
							Index: 19,
							Line:  1,
							Col:   19,
						},
					},
				},
				Cases: []CaseExpression{},
				Default: []Node{
					Whitespace{Value: "\n"},
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
											Index: 50,
											Line:  4,
											Col:   6,
										},
										To: Position{
											Index: 64,
											Line:  4,
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
			},
		},
		{
			name: "switch: one case",
			input: `{% switch "stringy" %}
{% case "stringy" %}
<span>
  {%= "span content" %}
</span>
{% endcase %}
{% endswitch %}`,
			expected: SwitchExpression{
				Expression: Expression{
					Value: `"stringy"`,
					Range: Range{
						From: Position{
							Index: 10,
							Line:  1,
							Col:   10,
						},
						To: Position{
							Index: 19,
							Line:  1,
							Col:   19,
						},
					},
				},
				Cases: []CaseExpression{
					{
						Expression: Expression{
							Value: `"stringy"`,
							Range: Range{
								From: Position{
									Index: 31,
									Line:  2,
									Col:   8,
								},
								To: Position{
									Index: 40,
									Line:  2,
									Col:   17,
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
													Index: 57,
													Line:  4,
													Col:   6,
												},
												To: Position{
													Index: 71,
													Line:  4,
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
					},
				},
				Default: nil,
			},
		},
		{
			name: "switch: can be parsed without spaces",
			input: `{%switch "stringy"%}
{%case "stringy"%}
	stringy
{%endcase%}
{%default%}
	default
{%enddefault%}
{%endswitch%}`,
			expected: SwitchExpression{
				Expression: Expression{
					Value: `"stringy"`,
					Range: Range{
						From: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
						To: Position{
							Index: 18,
							Line:  1,
							Col:   18,
						},
					},
				},
				Cases: []CaseExpression{
					{
						Expression: Expression{
							Value: `"stringy"`,
							Range: Range{
								From: Position{
									Index: 28,
									Line:  2,
									Col:   7,
								},
								To: Position{
									Index: 37,
									Line:  2,
									Col:   16,
								},
							},
						},
						Children: []Node{Text{Value: "stringy\n"}},
					},
				},
				Default: []Node{Whitespace{Value: "\n\t"}, Text{Value: "default\n"}},
			},
		},
		{
			name: "switch: two cases",
			input: `{% switch "stringy" %}
{% case "stringy" %}
<span>
  {%= "span content" %}
</span>
{% endcase %}
{% case "other" %}
<span>
  {%= "other content" %}
</span>
{% endcase %}
{% endswitch %}`,
			expected: SwitchExpression{
				Expression: Expression{
					Value: `"stringy"`,
					Range: Range{
						From: Position{
							Index: 10,
							Line:  1,
							Col:   10,
						},
						To: Position{
							Index: 19,
							Line:  1,
							Col:   19,
						},
					},
				},
				Cases: []CaseExpression{
					{
						Expression: Expression{
							Value: `"stringy"`,
							Range: Range{
								From: Position{
									Index: 31,
									Line:  2,
									Col:   8,
								},
								To: Position{
									Index: 40,
									Line:  2,
									Col:   17,
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
													Index: 57,
													Line:  4,
													Col:   6,
												},
												To: Position{
													Index: 71,
													Line:  4,
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
					}, {
						Expression: Expression{
							Value: `"other"`,
							Range: Range{
								From: Position{
									Index: 105,
									Line:  7,
									Col:   8,
								},
								To: Position{
									Index: 112,
									Line:  7,
									Col:   15,
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
											Value: `"other content"`,
											Range: Range{
												From: Position{
													Index: 129,
													Line:  9,
													Col:   6,
												},
												To: Position{
													Index: 144,
													Line:  9,
													Col:   21,
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
				Default: nil,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := input.NewFromString(tt.input)
			result := newSwitchExpressionParser().Parse(input)
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
