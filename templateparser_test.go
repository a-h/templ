package templ

import (
	"testing"

	"github.com/a-h/lexical/input"
	"github.com/a-h/lexical/parse"
	"github.com/google/go-cmp/cmp"
)

func TestTemplateParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		parser   parse.Function
		expected interface{}
	}{
		{
			name: "template: no parameters",
			input: `{% templ Name() %}
{% endtmpl %}`,
			parser: newTemplateParser().Parse,
			expected: Template{
				Name: Expression{
					Value: "Name",
				},
				Parameters: Expression{
					Value: "",
				},
				Children: []Node{},
			},
		},
		{
			name: "template: single parameter",
			input: `{% templ Name(p Parameter) %}
{% endtmpl %}`,
			parser: newTemplateParser().Parse,
			expected: Template{
				Name: Expression{
					Value: "Name",
				},
				Parameters: Expression{
					Value: "p Parameter",
				},
				Children: []Node{},
			},
		},
		{
			name: "template: containing element",
			input: `{% templ Name(p Parameter) %}
<span>{%= "span content" %}</span>
{% endtmpl %}`,
			parser: newTemplateParser().Parse,
			expected: Template{
				Name: Expression{
					Value: "Name",
				},
				Parameters: Expression{
					Value: "p Parameter",
				},
				Children: []Node{
					Element{
						Name:       "span",
						Attributes: []Attribute{},
						Children: []Node{
							StringExpression{
								Expression: Expression{
									Value: `"span content"`,
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
{% endtmpl %}`,
			parser: newTemplateParser().Parse,
			expected: Template{
				Name: Expression{
					Value: "Name",
				},
				Parameters: Expression{
					Value: "p Parameter",
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
{% endtmpl %}`,
			parser: newTemplateParser().Parse,
			expected: Template{
				Name: Expression{
					Value: "Name",
				},
				Parameters: Expression{
					Value: "p Parameter",
				},
				Children: []Node{
					Whitespace{Value: "\t"},
					IfExpression{
						Expression: Expression{
							Value: `p.Test`,
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
