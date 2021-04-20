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
			parser: newElementOpenTagParser(NewSourceRangeToItemLookup()).Parse,
			expected: elementOpenTag{
				Name:       "a",
				Attributes: []Attribute{},
			},
		},
		{
			name:   "element: open with attributes",
			input:  `<div id="123" style="padding: 10px">`,
			parser: newElementOpenTagParser(NewSourceRangeToItemLookup()).Parse,
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
			parser: newConstantAttributeParser(NewSourceRangeToItemLookup()).Parse,
			expected: ConstantAttribute{
				Name:  "href",
				Value: "test",
			},
		},
		{
			name:   "stringexpression: constant",
			input:  `{%= "test" %}`,
			parser: newStringExpressionParser(NewSourceRangeToItemLookup()).Parse,
			expected: StringExpression{
				Expression: `"test"`,
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

func TestElementParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected interface{}
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
							Expression: `"test"`,
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
							`localisation.Get("a_title")`,
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
			name:  "element: containing string expression",
			input: `<div>{%= "test" %}</div>`,
			expected: Element{
				Name:       "div",
				Attributes: []Attribute{},
				Children: []Node{
					StringExpression{
						Expression: `"test"`,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := input.NewFromString(tt.input)
			sril := NewSourceRangeToItemLookup()
			parser := newElementParser(sril)
			result := parser.Parse(input)
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

func TestElementParserLocations(t *testing.T) {
	input := input.NewFromString(`<div id="123">{%= "test" %}</div>`)
	sril := NewSourceRangeToItemLookup()
	parser := newElementParser(sril)

	result := parser.Parse(input)
	if result.Error != nil {
		t.Fatalf("paser error: %v", result.Error)
	}
	if !result.Success {
		t.Fatalf("failed to parse at %d", input.Index())
	}
	element := result.Item.(Element)

	t.Run("lookup child string expression by index", func(t *testing.T) {
		actualItemRange, ok := parser.SourceRangeToItemLookup.LookupByIndex(18)
		if !ok {
			t.Errorf("expected ok, got %v from %+v", ok, parser.SourceRangeToItemLookup)
			return
		}
		a := element.Children[0].(StringExpression)
		actual := actualItemRange.Item.(StringExpression)
		if a != actual {
			t.Errorf("expected %v, got %v", element, actual)
		}
	})
}
