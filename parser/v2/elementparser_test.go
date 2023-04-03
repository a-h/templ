package parser

import (
	"strings"
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

type attributeTest[T any] struct {
	name     string
	input    string
	parser   parse.Parser[T]
	expected T
}

func TestAttributeParser(t *testing.T) {
	var tests = []attributeTest[any]{
		{
			name:   "element: open",
			input:  `<a>`,
			parser: StripType(elementOpenTagParser),
			expected: elementOpenTag{
				Name: "a",
			},
		},
		{
			name:   "element: hyphen in name",
			input:  `<turbo-frame>`,
			parser: StripType(elementOpenTagParser),
			expected: elementOpenTag{
				Name: "turbo-frame",
			},
		},
		{
			name:   "element: open with attributes",
			input:  `<div id="123" style="padding: 10px">`,
			parser: StripType(elementOpenTagParser),
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
			name:   "boolean expression attribute",
			input:  ` noshade?={ true }"`,
			parser: StripType(boolExpressionAttributeParser),
			expected: BoolExpressionAttribute{
				Name: "noshade",
				Expression: Expression{
					Value: "true",
					Range: Range{
						From: Position{
							Index: 12,
							Line:  0,
							Col:   12,
						},
						To: Position{
							Index: 16,
							Line:  0,
							Col:   16,
						},
					},
				},
			},
		},
		{
			name:   "boolean expression attribute without spaces",
			input:  ` noshade?={true}"`,
			parser: StripType(boolExpressionAttributeParser),
			expected: BoolExpressionAttribute{
				Name: "noshade",
				Expression: Expression{
					Value: "true",
					Range: Range{
						From: Position{
							Index: 11,
							Line:  0,
							Col:   11,
						},
						To: Position{
							Index: 15,
							Line:  0,
							Col:   15,
						},
					},
				},
			},
		},
		{
			name:   "attribute parsing handles boolean expression attributes",
			input:  ` noshade?={ true }`,
			parser: StripType(attributeParser),
			expected: BoolExpressionAttribute{
				Name: "noshade",
				Expression: Expression{
					Value: "true",
					Range: Range{
						From: Position{
							Index: 12,
							Line:  0,
							Col:   12,
						},
						To: Position{
							Index: 16,
							Line:  0,
							Col:   16,
						},
					},
				},
			},
		},
		{
			name:   "constant attribute",
			input:  ` href="test"`,
			parser: StripType(constantAttributeParser),
			expected: ConstantAttribute{
				Name:  "href",
				Value: "test",
			},
		},
		{
			name:   "attribute name with hyphens",
			input:  ` data-turbo-permanent="value"`,
			parser: StripType(constantAttributeParser),
			expected: ConstantAttribute{
				Name:  "data-turbo-permanent",
				Value: "value",
			},
		},
		{
			name:   "empty attribute",
			input:  ` data=""`,
			parser: StripType(constantAttributeParser),
			expected: ConstantAttribute{
				Name:  "data",
				Value: "",
			},
		},
		{
			name:   "attribute containing escaped text",
			input:  ` href="&lt;&quot;&gt;"`,
			parser: StripType(constantAttributeParser),
			expected: ConstantAttribute{
				Name:  "href",
				Value: `<">`,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			result, ok, err := tt.parser.Parse(input)
			if err != nil {
				t.Error(err)
			}
			if !ok {
				t.Errorf("failed to parse at %d", input.Index())
			}
			if diff := cmp.Diff(tt.expected, result); diff != "" {
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
			name:  "element: self-closing with single bool expression attribute",
			input: `<hr noshade?={ true }/>`,
			expected: Element{
				Name: "hr",
				Attributes: []Attribute{
					BoolExpressionAttribute{
						Name: "noshade",
						Expression: Expression{
							Value: `true`,
							Range: Range{
								From: Position{
									Index: 15,
									Line:  0,
									Col:   15,
								},
								To: Position{

									Index: 19,
									Line:  0,
									Col:   19,
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "element: self-closing with single expression attribute",
			input: `<a href={ "test" }/>`,
			expected: Element{
				Name: "a",
				Attributes: []Attribute{
					ExpressionAttribute{
						Name: "href",
						Expression: Expression{
							Value: `"test"`,
							Range: Range{
								From: Position{
									Index: 10,
									Line:  0,
									Col:   10,
								},
								To: Position{

									Index: 16,
									Line:  0,
									Col:   16,
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
			name:  "element: self-closing with multiple boolean attributes",
			input: `<hr optionA optionB?={ true } optionC="other"/>`,
			expected: Element{
				Name: "hr",
				Attributes: []Attribute{
					BoolConstantAttribute{
						Name: "optionA",
					},
					BoolExpressionAttribute{
						Name: "optionB",
						Expression: Expression{
							Value: `true`,
							Range: Range{
								From: Position{
									Index: 23,
									Line:  0,
									Col:   23,
								},
								To: Position{

									Index: 27,
									Line:  0,
									Col:   27,
								},
							},
						},
					},
					ConstantAttribute{
						Name:  "optionC",
						Value: "other",
					},
				},
			},
		},
		{
			name:  "element: self-closing with multiple constant and expr attributes",
			input: `<a href="test" title={ localisation.Get("a_title") } style="text-underline: auto"/>`,
			expected: Element{
				Name: "a",
				Attributes: []Attribute{
					ConstantAttribute{
						Name:  "href",
						Value: "test",
					},
					ExpressionAttribute{
						Name: "title",
						Expression: Expression{
							Value: `localisation.Get("a_title")`,
							Range: Range{
								From: Position{
									Index: 23,
									Line:  0,
									Col:   23,
								},
								To: Position{

									Index: 50,
									Line:  0,
									Col:   50,
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
				Name: "hr",
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
				Name: "a",
			},
		},
		{
			name:  "element: open and close with text",
			input: `<a>The text</a>`,
			expected: Element{
				Name: "a",
				Children: []Node{
					Text{
						Value: "The text",
					},
				},
			},
		},
		{
			name:  "element: with self-closing child element",
			input: `<a><b/></a>`,
			expected: Element{
				Name: "a",
				Children: []Node{
					Element{
						Name: "b",
					},
				},
			},
		},
		{
			name:  "element: with non-self-closing child element",
			input: `<a><b></b></a>`,
			expected: Element{
				Name: "a",
				Children: []Node{
					Element{
						Name: "b",
					},
				},
			},
		},
		{
			name:  "element: containing space",
			input: `<a> <b> </b> </a>`,
			expected: Element{
				Name: "a",
				Children: []Node{
					Whitespace{Value: " "},
					Element{
						Name: "b",
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
				Name: "a",
				Children: []Node{
					Element{
						Name: "b",
					},
					Element{
						Name: "c",
						Children: []Node{
							Element{
								Name: "d",
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
				Name: "div",
			},
		},
		{
			name:  "element: containing string expression",
			input: `<div>{ "test" }</div>`,
			expected: Element{
				Name: "div",
				Children: []Node{
					StringExpression{
						Expression: Expression{
							Value: `"test"`,
							Range: Range{
								From: Position{
									Index: 7,
									Line:  0,
									Col:   7,
								},
								To: Position{
									Index: 13,
									Line:  0,
									Col:   13,
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
			input := parse.NewInput(tt.input)
			result, ok, err := element.Parse(input)
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}
			if !ok {
				t.Fatalf("failed to parse at %d", input.Index())
			}
			if diff := cmp.Diff(tt.expected, result); diff != "" {
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
			expected: parse.Error("<a>: mismatched end tag, expected '</a>', got '</b>'",
				parse.Position{
					Index: 3,
					Line:  0,
					Col:   3,
				}),
		},
		{
			name:  "element: style must only contain text",
			input: `<style><button /></style>`,
			expected: parse.Error("<style>: invalid node contents: script and style attributes must only contain text",
				parse.Position{
					Index: 0,
					Line:  0,
					Col:   0,
				}),
		},
		{
			name:  "element: script must only contain text",
			input: `<script><button /></script>`,
			expected: parse.Error("<script>: invalid node contents: script and style attributes must only contain text",
				parse.Position{
					Index: 0,
					Line:  0,
					Col:   0,
				}),
		},
		{
			name:  "element: attempted use of expression for style attribute (open/close)",
			input: `<a style={ value }></a>`,
			expected: parse.Error(`<a>: invalid style attribute: style attributes cannot be a templ expression`,
				parse.Position{
					Index: 0,
					Line:  0,
					Col:   0,
				}),
		},
		{
			name:  "element: attempted use of expression for style attribute (self-closing)",
			input: `<a style={ value }/>`,
			expected: parse.Error(`<a>: invalid style attribute: style attributes cannot be a templ expression`,
				parse.Position{
					Index: 0,
					Line:  0,
					Col:   0,
				}),
		},
		{
			name:  "element: script tags cannot contain non-text nodes",
			input: `<script>{ "value" }</script>`,
			expected: parse.Error("<script>: invalid node contents: script and style attributes must only contain text",
				parse.Position{
					Index: 0,
					Line:  0,
					Col:   0,
				}),
		},
		{
			name:  "element: style tags cannot contain non-text nodes",
			input: `<style>{ "value" }</style>`,
			expected: parse.Error("<style>: invalid node contents: script and style attributes must only contain text",
				parse.Position{
					Index: 0,
					Line:  0,
					Col:   0,
				}),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			_, _, err := element.Parse(input)
			if diff := cmp.Diff(tt.expected, err); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestBigElement(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("<div>")
	for i := 0; i < 4096*4; i++ {
		sb.WriteString("a")
	}
	sb.WriteString("</div>")
	_, ok, err := element.Parse(parse.NewInput(sb.String()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Errorf("unexpected failure to parse")
	}
}
