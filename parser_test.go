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
			name:   "import: named",
			input:  `{% import name "github.com/a-h/something" %}`,
			parser: importParser,
			expected: Import{
				Expression: `name "github.com/a-h/something"`,
			},
		},
		{
			name:   "import: default",
			input:  `{% import "github.com/a-h/something" %}`,
			parser: importParser,
			expected: Import{
				Expression: `"github.com/a-h/something"`,
			},
		},
		{
			name:   "templateFileWhitespace: various spaces",
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
			name: "template: no parameters",
			input: `{% templ Name() %}
{% endtmpl %}`,
			parser: newTemplateParser(NewSourceRangeToItemLookup()).Parse,
			expected: Template{
				Name:                "Name",
				ParameterExpression: "",
				Children:            []Node{},
			},
		},
		{
			name: "template: single parameter",
			input: `{% templ Name(p Parameter) %}
{% endtmpl %}`,
			parser: newTemplateParser(NewSourceRangeToItemLookup()).Parse,
			expected: Template{
				Name:                "Name",
				ParameterExpression: "p Parameter",
				Children:            []Node{},
			},
		},
		{
			name: "template: containing element",
			input: `{% templ Name(p Parameter) %}
<span>{%= "span content" %}</span>
{% endtmpl %}`,
			parser: newTemplateParser(NewSourceRangeToItemLookup()).Parse,
			expected: Template{
				Name:                "Name",
				ParameterExpression: "p Parameter",
				Children: []Node{
					Element{
						Name:       "span",
						Attributes: []Attribute{},
						Children: []Node{
							StringExpression{
								Expression: `"span content"`,
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
			parser: newTemplateParser(NewSourceRangeToItemLookup()).Parse,
			expected: Template{
				Name:                "Name",
				ParameterExpression: "p Parameter",
				Children: []Node{
					Element{
						Name:       "div",
						Attributes: []Attribute{},
						Children: []Node{
							Whitespace{Value: "\n  "},
							StringExpression{Expression: `"div content"`},
							Whitespace{Value: "\n  "},
							Element{
								Name:       "span",
								Attributes: []Attribute{},
								Children: []Node{
									Whitespace{Value: "\n\t"},
									StringExpression{Expression: `"span content"`},
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
			name: "if: simple expression",
			input: `{% if p.Test %}
<span>
  {%= "span content" %}
</span>
{% endif %}
`,
			parser: newIfExpressionParser(NewSourceRangeToItemLookup()).Parse,
			expected: IfExpression{
				Expression: `p.Test`,
				Then: []Node{
					Element{
						Name:       "span",
						Attributes: []Attribute{},
						Children: []Node{
							Whitespace{Value: "\n  "},
							StringExpression{Expression: `"span content"`},
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
			parser: newIfExpressionParser(NewSourceRangeToItemLookup()).Parse,
			expected: IfExpression{
				Expression: `p.A`,
				Then: []Node{
					Whitespace{Value: "\t"},
					StringExpression{Expression: `"A"`},
					Whitespace{Value: "\n"},
				},
				Else: []Node{
					Whitespace{Value: "\n\t"},
					StringExpression{Expression: `"B"`},
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
			parser: newIfExpressionParser(NewSourceRangeToItemLookup()).Parse,
			expected: IfExpression{
				Expression: `p.A`,
				Then: []Node{
					Whitespace{Value: "\t\t\t\t\t"},
					IfExpression{
						Expression: `p.B`,
						Then: []Node{
							Whitespace{Value: "\t\t\t\t\t\t"},
							Element{
								Name:       "div",
								Attributes: []Attribute{},
								Children: []Node{
									StringExpression{Expression: `"B"`},
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
			name: "template: containing if element",
			input: `{% templ Name(p Parameter) %}
	{% if p.Test %}
		<span>
			{%= "span content" %}
		</span>
	{% endif %}
{% endtmpl %}`,
			parser: newTemplateParser(NewSourceRangeToItemLookup()).Parse,
			expected: Template{
				Name:                "Name",
				ParameterExpression: "p Parameter",
				Children: []Node{
					Whitespace{Value: "\t"},
					IfExpression{
						Expression: `p.Test`,
						Then: []Node{
							Whitespace{Value: "\t\t"},
							Element{
								Name:       "span",
								Attributes: []Attribute{},
								Children: []Node{
									Whitespace{"\n\t\t\t"},
									StringExpression{Expression: `"span content"`},
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
			name: "for: simple",
			input: `{% for _, item := range p.Items %}
					<div>{%= item %}</div>
				{% endfor %}`,
			parser: newForExpressionParser(NewSourceRangeToItemLookup()).Parse,
			expected: ForExpression{
				Expression: `_, item := range p.Items`,
				Children: []Node{
					Whitespace{Value: "\t\t\t\t\t"},
					Element{
						Name:       "div",
						Attributes: []Attribute{},
						Children: []Node{
							StringExpression{Expression: `item`},
						},
					},
					Whitespace{Value: "\n\t\t\t\t"},
				},
			},
		},
		{
			name:   "call: simple",
			input:  `{% call Other(p.Test) %}`,
			parser: callTemplateExpressionParser{}.Parse,
			expected: CallTemplateExpression{
				Name:               "Other",
				ArgumentExpression: `p.Test`,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := input.NewFromString(tt.input)
			result := tt.parser(input)
			if result.Error != nil {
				t.Fatalf("paser error: %v", result.Error)
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
