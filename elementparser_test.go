package templ

import (
	"testing"

	"github.com/a-h/lexical/input"
	"github.com/a-h/lexical/parse"
	"github.com/google/go-cmp/cmp"
)

func TestAttributeParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		parser   parse.Function
		expected interface{}
	}{
		{
			name:   "element: open",
			input:  `<a>`,
			parser: newElementOpenTagParser().Parse,
			expected: elementOpenTag{
				Name:       "a",
				Attributes: []Attribute{},
			},
		},
		{
			name:   "element: open with attributes",
			input:  `<div id="123" style="padding: 10px">`,
			parser: newElementOpenTagParser().Parse,
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
			name:   "constant attribute",
			input:  ` href="test"`,
			parser: newConstantAttributeParser().Parse,
			expected: ConstantAttribute{
				Name:  "href",
				Value: "test",
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

func TestElementParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected Element
	}{
		{
			name:  "element: self-closing with single constant attribute",
			input: `<a href="test"/>`,
			expected: Element{
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
			name:  "element: self-closing with single expression attribute",
			input: `<a href={%= "test" %}/>`,
			expected: Element{
				Name: "a",
				Attributes: []Attribute{
					ExpressionAttribute{
						Name: "href",
						Value: StringExpression{
							Expression: Expression{
								Value: `"test"`,
								Range: Range{
									From: Position{
										Index: 12,
										Line:  1,
										Col:   12,
									},
									To: Position{

										Index: 18,
										Line:  1,
										Col:   18,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "element: self-closing with multiple constant attributes",
			input: `<a href="test" style="text-underline: auto"/>`,
			expected: Element{
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
			name:  "element: self-closing with multiple constant and expr attributes",
			input: `<a href="test" title={%= localisation.Get("a_title") %} style="text-underline: auto"/>`,
			expected: Element{
				Name: "a",
				Attributes: []Attribute{
					ConstantAttribute{
						Name:  "href",
						Value: "test",
					},
					ExpressionAttribute{
						Name: "title",
						Value: StringExpression{
							Expression: Expression{
								Value: `localisation.Get("a_title")`,
								Range: Range{
									From: Position{
										Index: 25,
										Line:  1,
										Col:   25,
									},
									To: Position{

										Index: 52,
										Line:  1,
										Col:   52,
									},
								},
							},
						},
					},
					ConstantAttribute{
						Name:  "style",
						Value: "text-underline: auto",
					},
				},
			},
		},
		{
			name:  "element: self closing with no attributes",
			input: `<hr/>`,
			expected: Element{
				Name:       "hr",
				Attributes: []Attribute{},
			},
		},
		{
			name:  "element: self closing with attribute",
			input: `<hr style="padding: 10px" />`,
			expected: Element{
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
			name:  "element: open and close",
			input: `<a></a>`,
			expected: Element{
				Name:       "a",
				Attributes: []Attribute{},
			},
		},
		{
			name:  "element: with self-closing child element",
			input: `<a><b/></a>`,
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
			name:  "element: with non-self-closing child element",
			input: `<a><b></b></a>`,
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
			name:  "element: containing space",
			input: `<a> <b> </b> </a>`,
			expected: Element{
				Name:       "a",
				Attributes: []Attribute{},
				Children: []Node{
					Whitespace{Value: " "},
					Element{
						Name:       "b",
						Attributes: []Attribute{},
						Children: []Node{
							Whitespace{Value: " "},
						},
					},
					Whitespace{Value: " "},
				},
			},
		},
		{
			name:  "element: with multiple child elements",
			input: `<a><b></b><c><d/></c></a>`,
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
			name:  "element: empty",
			input: `<div></div>`,
			expected: Element{
				Name:       "div",
				Attributes: []Attribute{},
			},
		},
		{
			name:  "element: containing string expression",
			input: `<div>{%= "test" %}</div>`,
			expected: Element{
				Name:       "div",
				Attributes: []Attribute{},
				Children: []Node{
					StringExpression{
						Expression: Expression{
							Value: `"test"`,
							Range: Range{
								From: Position{
									Index: 9,
									Line:  1,
									Col:   9,
								},
								To: Position{
									Index: 15,
									Line:  1,
									Col:   15,
								},
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
			result := newElementParser().Parse(input)
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

func TestElementParserErrors(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:  "element: mismatched end tag",
			input: `<a></b>`,
			expected: newParseError("element: mismatched end tag, expected '</a>', got '</b>'",
				Position{
					Index: 3,
					Line:  1,
					Col:   3,
				},
				Position{
					Index: 7,
					Line:  1,
					Col:   7,
				}),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := input.NewFromString(tt.input)
			result := newElementParser().Parse(input)
			if diff := cmp.Diff(tt.expected, result.Error); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
