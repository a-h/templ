package templ

import (
	"testing"

	"github.com/a-h/lexical/input"
	"github.com/a-h/lexical/parse"
	"github.com/google/go-cmp/cmp"
)

func TestParsers(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		parser   parse.Function
		expected interface{}
	}{
		{
			name:   "whitespace: various spaces",
			input:  "  \t ",
			parser: whitespaceParser,
			expected: Whitespace{
				Value: "  \t ",
			},
		},
		{
			name:   "whitespace: spaces and newline",
			input:  " \n",
			parser: whitespaceParser,
			expected: Whitespace{
				Value: " \n",
			},
		},
		{
			name:   "whitespace: newline",
			input:  "\n",
			parser: whitespaceParser,
			expected: Whitespace{
				Value: "\n",
			},
		},
		{
			name: "if: simple expression",
			input: `{% if p.Test %}
<span>
  {%= "span content" %}
</span>
{% endif %}
`,
			parser: newIfExpressionParser().Parse,
			expected: IfExpression{
				Expression: Expression{
					Value: `p.Test`,
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
								},
							},
							Whitespace{Value: "\n"},
						},
					},
					Whitespace{
						Value: "\n",
					},
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
			parser: newIfExpressionParser().Parse,
			expected: IfExpression{
				Expression: Expression{
					Value: `p.A`,
				},
				Then: []Node{
					Whitespace{Value: "\t"},
					StringExpression{
						Expression: Expression{
							Value: `"A"`,
						},
					},
					Whitespace{Value: "\n"},
				},
				Else: []Node{
					Whitespace{Value: "\n\t"},
					StringExpression{
						Expression: Expression{
							Value: `"B"`,
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
			parser: newIfExpressionParser().Parse,
			expected: IfExpression{
				Expression: Expression{
					Value: `p.A`,
				},
				Then: []Node{
					Whitespace{Value: "\t\t\t\t\t"},
					IfExpression{
						Expression: Expression{
							Value: `p.B`,
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
		{
			name: "for: simple",
			input: `{% for _, item := range p.Items %}
					<div>{%= item %}</div>
				{% endfor %}`,
			parser: newForExpressionParser().Parse,
			expected: ForExpression{
				Expression: Expression{
					Value: `_, item := range p.Items`,
				},
				Children: []Node{
					Whitespace{Value: "\t\t\t\t\t"},
					Element{
						Name:       "div",
						Attributes: []Attribute{},
						Children: []Node{
							StringExpression{
								Expression: Expression{
									Value: `item`,
								},
							},
						},
					},
					Whitespace{Value: "\n\t\t\t\t"},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := input.NewFromString(tt.input)
			result := tt.parser(input)
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
