package parser

import (
	"testing"

	"github.com/a-h/lexical/input"
	"github.com/google/go-cmp/cmp"
)

func TestTemplateParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected HTMLTemplate
	}{
		{
			name: "template: no parameters",
			input: `{% templ Name() %}
{% endtempl %}`,
			expected: HTMLTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
						To: Position{
							Index: 12,
							Line:  1,
							Col:   12,
						},
					},
				},
				Parameters: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 14,
							Line:  1,
							Col:   14,
						},
						To: Position{
							Index: 14,
							Line:  1,
							Col:   14,
						},
					},
				},
				Children: []Node{},
			},
		},
		{
			name: "template: no spaces",
			input: `{%templ Name()%}
{% endtempl %}`,
			expected: HTMLTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 8,
							Line:  1,
							Col:   8,
						},
						To: Position{
							Index: 11,
							Line:  1,
							Col:   11,
						},
					},
				},
				Parameters: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 13,
							Line:  1,
							Col:   13,
						},
						To: Position{
							Index: 13,
							Line:  1,
							Col:   13,
						},
					},
				},
				Children: []Node{},
			},
		},
		{
			name: "template: single parameter",
			input: `{% templ Name(p Parameter) %}
{% endtempl %}`,
			expected: HTMLTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
						To: Position{
							Index: 12,
							Line:  1,
							Col:   12,
						},
					},
				},
				Parameters: Expression{
					Value: "p Parameter",
					Range: Range{
						From: Position{
							Index: 14,
							Line:  1,
							Col:   14,
						},
						To: Position{
							Index: 25,
							Line:  1,
							Col:   25,
						},
					},
				},
				Children: []Node{},
			},
		},
		{
			name: "template: containing element",
			input: `{% templ Name(p Parameter) %}
<span>{%= "span content" %}</span>
{% endtempl %}`,
			expected: HTMLTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
						To: Position{
							Index: 12,
							Line:  1,
							Col:   12,
						},
					},
				},
				Parameters: Expression{
					Value: "p Parameter",
					Range: Range{
						From: Position{
							Index: 14,
							Line:  1,
							Col:   14,
						},
						To: Position{
							Index: 25,
							Line:  1,
							Col:   25,
						},
					},
				},
				Children: []Node{
					Element{
						Name:       "span",
						Attributes: []Attribute{},
						Children: []Node{
							StringExpression{
								Expression: Expression{
									Value: `"span content"`,
									Range: Range{
										From: Position{
											Index: 40,
											Line:  2,
											Col:   10,
										},
										To: Position{
											Index: 54,
											Line:  2,
											Col:   24,
										},
									},
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
		{
			name: "template: containing nested elements",
			input: `{% templ Name(p Parameter) %}
<div>
  {%= "div content" %}
  <span>
	{%= "span content" %}
  </span>
</div>
{% endtempl %}`,
			expected: HTMLTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
						To: Position{
							Index: 12,
							Line:  1,
							Col:   12,
						},
					},
				},
				Parameters: Expression{
					Value: "p Parameter",
					Range: Range{
						From: Position{
							Index: 14,
							Line:  1,
							Col:   14,
						},
						To: Position{
							Index: 25,
							Line:  1,
							Col:   25,
						},
					},
				},
				Children: []Node{
					Element{
						Name:       "div",
						Attributes: []Attribute{},
						Children: []Node{
							Whitespace{Value: "\n  "},
							StringExpression{
								Expression: Expression{
									Value: `"div content"`,
									Range: Range{
										From: Position{
											Index: 42,
											Line:  3,
											Col:   6,
										},
										To: Position{
											Index: 55,
											Line:  3,
											Col:   19,
										},
									},
								},
							},
							Whitespace{Value: "\n  "},
							Element{
								Name:       "span",
								Attributes: []Attribute{},
								Children: []Node{
									Whitespace{Value: "\n\t"},
									StringExpression{
										Expression: Expression{
											Value: `"span content"`,
											Range: Range{
												From: Position{
													Index: 73,
													Line:  5,
													Col:   5,
												},
												To: Position{
													Index: 87,
													Line:  5,
													Col:   19,
												},
											},
										},
									},
									Whitespace{Value: "\n  "},
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
			name: "template: containing if element",
			input: `{% templ Name(p Parameter) %}
	{% if p.Test %}
		<span>
			{%= "span content" %}
		</span>
	{% endif %}
{% endtempl %}`,
			expected: HTMLTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
						To: Position{
							Index: 12,
							Line:  1,
							Col:   12,
						},
					},
				},
				Parameters: Expression{
					Value: "p Parameter",
					Range: Range{
						From: Position{
							Index: 14,
							Line:  1,
							Col:   14,
						},
						To: Position{
							Index: 25,
							Line:  1,
							Col:   25,
						},
					},
				},
				Children: []Node{
					Whitespace{Value: "\t"},
					IfExpression{
						Expression: Expression{
							Value: `p.Test`,
							Range: Range{
								From: Position{
									Index: 37,
									Line:  2,
									Col:   7,
								},
								To: Position{
									Index: 43,
									Line:  2,
									Col:   13,
								},
							},
						},
						Then: []Node{
							Whitespace{Value: "\t\t"},
							Element{
								Name:       "span",
								Attributes: []Attribute{},
								Children: []Node{
									Whitespace{"\n\t\t\t"},
									StringExpression{
										Expression: Expression{
											Value: `"span content"`,
											Range: Range{
												From: Position{
													Index: 63,
													Line:  4,
													Col:   7,
												},
												To: Position{
													Index: 77,
													Line:  4,
													Col:   21,
												},
											},
										},
									},
									Whitespace{"\n\t\t"},
								},
							},
							Whitespace{
								Value: "\n\t",
							},
						},
						Else: []Node{},
					},
					Whitespace{
						Value: "\n",
					},
				},
			},
		},
		{
			name: "template: inputs",
			input: `{% templ Name(p Parameter) %}
<input type="text" value="a" />
<input type="text" value="b" />
{% endtempl %}`,
			expected: HTMLTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
						To: Position{
							Index: 12,
							Line:  1,
							Col:   12,
						},
					},
				},
				Parameters: Expression{
					Value: "p Parameter",
					Range: Range{
						From: Position{
							Index: 14,
							Line:  1,
							Col:   14,
						},
						To: Position{
							Index: 25,
							Line:  1,
							Col:   25,
						},
					},
				},
				Children: []Node{
					Element{
						Name: "input",
						Attributes: []Attribute{
							ConstantAttribute{Name: "type", Value: "text"},
							ConstantAttribute{Name: "value", Value: "a"},
						},
					},
					Whitespace{Value: "\n"},
					Element{
						Name: "input",
						Attributes: []Attribute{
							ConstantAttribute{Name: "type", Value: "text"},
							ConstantAttribute{Name: "value", Value: "b"},
						},
					},
					Whitespace{Value: "\n"},
				},
			},
		},
		{
			name: "template: doctype",
			input: `{% templ Name() %}
<!DOCTYPE html>
{% endtempl %}`,
			expected: HTMLTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
						To: Position{
							Index: 12,
							Line:  1,
							Col:   12,
						},
					},
				},
				Parameters: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 14,
							Line:  1,
							Col:   14,
						},
						To: Position{
							Index: 14,
							Line:  1,
							Col:   14,
						},
					},
				},
				Children: []Node{
					DocType{
						Value: "html",
					},
					Whitespace{Value: "\n"},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := input.NewFromString(tt.input)
			result := newTemplateParser().Parse(input)
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
