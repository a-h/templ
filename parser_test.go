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
			name:   "package: standard",
			input:  `{% package templ %}`,
			parser: packageParser,
			expected: PackageExpression{
				Expression: "templ",
			},
		},
		{
			name:   "import: named",
			input:  `{% import name "github.com/a-h/something" %}`,
			parser: importParser,
			expected: ImportExpression{
				Expression: `name "github.com/a-h/something"`,
			},
		},
		{
			name:   "import: default",
			input:  `{% import "github.com/a-h/something" %}`,
			parser: importParser,
			expected: ImportExpression{
				Expression: `"github.com/a-h/something"`,
			},
		},
		{
			name:   "templateFileWhitespace: various spaces",
			input:  "  \t ",
			parser: templateFileWhitespaceParser,
			expected: templateFileWhitespace{
				Text: "  \t ",
			},
		},
		{
			name:   "templateFileWhitespace: spaces and newline",
			input:  " \n",
			parser: templateFileWhitespaceParser,
			expected: templateFileWhitespace{
				Text: " \n",
			},
		},
		{
			name:   "templateFileWhitespace: newline",
			input:  "\n",
			parser: templateFileWhitespaceParser,
			expected: templateFileWhitespace{
				Text: "\n",
			},
		},
		{
			name: "template: no parameters",
			input: `{% templ Name() %}
			{% endtmpl %}`,
			parser: templateParser,
			expected: Template{
				Name:                "Name",
				ParameterExpression: "",
			},
		},
		{
			name: "template: single parameter",
			input: `{% templ Name(p Parameter) %}
			{% endtmpl %}`,
			parser: templateParser,
			expected: Template{
				Name:                "Name",
				ParameterExpression: "p Parameter",
			},
		},
		{
			name:   "constant attribute",
			input:  ` href="test"`,
			parser: constAttrParser,
			expected: ConstantAttribute{
				Name:  "href",
				Value: "test",
			},
		},
		{
			name:   "element self-closing with single constant attribute",
			input:  `<a href="test"/>`,
			parser: elementSelfClosingParser,
			expected: elementSelfClosing{
				Name: "a",
				Attributes: []Attribute{
					ConstantAttribute{
						Name:  "href",
						Value: "test",
					},
				},
			},
		},
		{
			name:   "element self-closing with multiple attributes",
			input:  `<a href="test" style="text-underline: auto"/>`,
			parser: elementSelfClosingParser,
			expected: elementSelfClosing{
				Name: "a",
				Attributes: []Attribute{
					ConstantAttribute{
						Name:  "href",
						Value: "test",
					},
					ConstantAttribute{
						Name:  "style",
						Value: "text-underline: auto",
					},
				},
			},
		},
		{
			name:   "element: self closing with no attributes",
			input:  `<hr/>`,
			parser: elementSelfClosingParser,
			expected: elementSelfClosing{
				Name:       "hr",
				Attributes: []Attribute{},
			},
		},
		{
			name:   "element: self closing with attribute",
			input:  `<hr style="padding: 10px" />`,
			parser: elementSelfClosingParser,
			expected: elementSelfClosing{
				Name: "hr",
				Attributes: []Attribute{
					ConstantAttribute{
						Name:  "style",
						Value: "padding: 10px",
					},
				},
			},
		},
		{
			name:   "element: open",
			input:  `<a>`,
			parser: elementOpenTagParser,
			expected: elementOpenTag{
				Name:       "a",
				Attributes: []Attribute{},
			},
		},
		{
			name:   "element: open with attributes",
			input:  `<div id="123" style="padding: 10px">`,
			parser: elementOpenTagParser,
			expected: elementOpenTag{
				Name: "div",
				Attributes: []Attribute{
					ConstantAttribute{
						Name:  "id",
						Value: "123",
					},
					ConstantAttribute{
						Name:  "style",
						Value: "padding: 10px",
					},
				},
			},
		},
		{
			name:   "element: open and close",
			input:  `<a></a>`,
			parser: elementParser{}.Parse,
			expected: Element{
				Name:       "a",
				Attributes: []Attribute{},
			},
		},
		{
			name:   "element: with self-closing child element",
			input:  `<a><b/></a>`,
			parser: elementParser{}.Parse,
			expected: Element{
				Name:       "a",
				Attributes: []Attribute{},
				Children: []Node{
					Element{
						Name:       "b",
						Attributes: []Attribute{},
					},
				},
			},
		},
		{
			name:   "element: with non-self-closing child element",
			input:  `<a><b></b></a>`,
			parser: elementParser{}.Parse,
			expected: Element{
				Name:       "a",
				Attributes: []Attribute{},
				Children: []Node{
					Element{
						Name:       "b",
						Attributes: []Attribute{},
					},
				},
			},
		},
		{
			name:   "element: with multiple child elements",
			input:  `<a><b></b><c><d/></c></a>`,
			parser: elementParser{}.Parse,
			expected: Element{
				Name:       "a",
				Attributes: []Attribute{},
				Children: []Node{
					Element{
						Name:       "b",
						Attributes: []Attribute{},
					},
					Element{
						Name:       "c",
						Attributes: []Attribute{},
						Children: []Node{
							Element{
								Name:       "d",
								Attributes: []Attribute{},
							},
						},
					},
				},
			},
		},
		{
			name:   "nodestringexpression: constant",
			input:  `{%= "test" %}`,
			parser: nodeStringExpressionParser,
			expected: NodeStringExpression{
				Expression: `"test"`,
			},
		},
		{
			name:   "element: containing string expression",
			input:  `<div>{%= "test" %}</div>`,
			parser: elementParser{}.Parse,
			expected: Element{
				Name:       "a",
				Attributes: []Attribute{},
				Children: []Node{
					NodeStringExpression{
						Expression: `"test"`,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result := tt.parser(input.NewFromString(tt.input))
			if result.Error != nil {
				t.Fatalf("paser error: %v", result.Error)
			}
			if !result.Success {
				t.Fatalf("failed to parse!")
			}
			if diff := cmp.Diff(tt.expected, result.Item); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
