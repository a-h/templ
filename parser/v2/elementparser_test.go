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
	tests := []attributeTest[any]{
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
			name:   "element: open with hyperscript attribute",
			input:  `<div _="show = true">`,
			parser: StripType(elementOpenTagParser),
			expected: elementOpenTag{
				Name: "div",
				Attributes: []Attribute{
					ConstantAttribute{
						Name:  "_",
						Value: "show = true",
					},
				},
			},
		},
		{
			name:   "element: open with complex attributes",
			input:  `<div @click="show = true" :class="{'foo': true}">`,
			parser: StripType(elementOpenTagParser),
			expected: elementOpenTag{
				Name: "div",
				Attributes: []Attribute{
					ConstantAttribute{
						Name:  "@click",
						Value: "show = true",
					},
					ConstantAttribute{
						Name:  ":class",
						Value: "{'foo': true}",
					},
				},
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
			name: "conditional expression attribute - single",
			input: `
		if p.important {
			class="important"
		}
"`,
			parser: StripType(conditionalAttribute),
			expected: ConditionalAttribute{
				Expression: Expression{
					Value: "p.important",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  1,
							Col:   5,
						},
						To: Position{
							Index: 17,
							Line:  1,
							Col:   16,
						},
					},
				},
				Then: []Attribute{
					ConstantAttribute{
						Name:  "class",
						Value: "important",
					},
				},
			},
		},
		{
			name: "conditional expression attribute - multiple",
			input: `
if test { 
	class="itIsTrue"
	noshade
	name={ "other" }
}
"`,
			parser: StripType(conditionalAttribute),
			expected: ConditionalAttribute{
				Expression: Expression{
					Value: "test",
					Range: Range{
						From: Position{
							Index: 4,
							Line:  1,
							Col:   3,
						},
						To: Position{
							Index: 8,
							Line:  1,
							Col:   7,
						},
					},
				},
				Then: []Attribute{
					ConstantAttribute{
						Name:  "class",
						Value: "itIsTrue",
					},
					BoolConstantAttribute{
						Name: "noshade",
					},
					ExpressionAttribute{
						Name: "name",
						Expression: Expression{
							Value: `"other"`,
							Range: Range{
								From: Position{
									Index: 47,
									Line:  4,
									Col:   8,
								},
								To: Position{
									Index: 54,
									Line:  4,
									Col:   15,
								},
							},
						},
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
			parser: StripType[Attribute](attributeParser{}),
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
			name:   "spread attributes",
			input:  ` { spread... }"`,
			parser: StripType(spreadAttributesParser),
			expected: SpreadAttributes{
				Expression{
					Value: "spread",
					Range: Range{
						From: Position{
							Index: 3,
							Line:  0,
							Col:   3,
						},
						To: Position{
							Index: 9,
							Line:  0,
							Col:   9,
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
			name:   "single quote not required constant attribute",
			input:  ` href='no double quote in value'`,
			parser: StripType(constantAttributeParser),
			expected: ConstantAttribute{
				Name:        "href",
				Value:       `no double quote in value`,
				SingleQuote: false,
			},
		},
		{
			name:   "single quote required constant attribute",
			input:  ` href='"test"'`,
			parser: StripType(constantAttributeParser),
			expected: ConstantAttribute{
				Name:        "href",
				Value:       `"test"`,
				SingleQuote: true,
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
			name: "multiline attribute",
			input: ` data-script="on click
                do something
             end"
`,
			parser: StripType(constantAttributeParser),
			expected: ConstantAttribute{
				Name:  "data-script",
				Value: "on click\n                do something\n             end",
			},
		},
		{
			name:   "bool constant attribute",
			input:  `<div data>`,
			parser: StripType(elementOpenTagParser),
			expected: elementOpenTag{
				Name: "div",
				Attributes: []Attribute{
					BoolConstantAttribute{
						Name: "data",
					},
				},
			},
		},
		{
			name:   "bool constant attributes can end with a Unix newline",
			input:  "<input\n\t\trequired\n\t/>",
			parser: StripType[Node](element),
			expected: Element{
				Name:        "input",
				IndentAttrs: true,
				Attributes: []Attribute{
					BoolConstantAttribute{
						Name: "required",
					},
				},
			},
		},
		{
			name:   "bool constant attributes can end with a Windows newline",
			input:  "<input\r\n\t\trequired\r\n\t/>",
			parser: StripType[Node](element),
			expected: Element{
				Name:        "input",
				IndentAttrs: true,
				Attributes: []Attribute{
					BoolConstantAttribute{
						Name: "required",
					},
				},
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
		{
			name:   "HTMX wildcard attribute names are supported",
			input:  ` hx-target-*="#errors"`,
			parser: StripType(constantAttributeParser),
			expected: ConstantAttribute{
				Name:  "hx-target-*",
				Value: `#errors`,
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
				t.Errorf("failed to parse at %v", input.Position())
			}
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestElementParser(t *testing.T) {
	tests := []struct {
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
			name:  "element: self-closing with multiple spreads attributes",
			input: `<a { firstSpread... } { children... }/>`,
			expected: Element{
				Name: "a",
				Attributes: []Attribute{
					SpreadAttributes{
						Expression: Expression{
							Value: "firstSpread",
							Range: Range{
								From: Position{
									Index: 5,
									Line:  0,
									Col:   5,
								},
								To: Position{
									Index: 16,
									Line:  0,
									Col:   16,
								},
							},
						},
					},
					SpreadAttributes{
						Expression: Expression{
							Value: "children",
							Range: Range{
								From: Position{
									Index: 24,
									Line:  0,
									Col:   24,
								},
								To: Position{
									Index: 32,
									Line:  0,
									Col:   32,
								},
							},
						},
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
			name: "element: self-closing with multiple constant, conditional and expr attributes",
			input: `<div style="width: 100;"
		if p.important {
			class="important"
		}
>Test</div>
}

`,
			expected: Element{
				Name: "div",
				Attributes: []Attribute{
					ConstantAttribute{
						Name:  "style",
						Value: "width: 100;",
					},
					ConditionalAttribute{
						Expression: Expression{
							Value: `p.important`,
							Range: Range{
								From: Position{
									Index: 30,
									Line:  1,
									Col:   5,
								},
								To: Position{
									Index: 41,
									Line:  1,
									Col:   16,
								},
							},
						},
						Then: []Attribute{
							ConstantAttribute{
								Name:  "class",
								Value: "important",
							},
						},
					},
				},
				IndentAttrs: true,
				Children: []Node{
					Text{
						Value: "Test",
					},
				},
				TrailingSpace: SpaceVertical,
			},
		},
		{
			name:  "element: self-closing with no attributes",
			input: `<hr/>`,
			expected: Element{
				Name: "hr",
			},
		},
		{
			name:  "element: self-closing with attribute",
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
			name: "element: self-closing with conditional attribute",
			input: `<hr style="padding: 10px" 
			if true {
				class="itIsTrue"
			}
/>`,
			expected: Element{
				Name: "hr",
				Attributes: []Attribute{
					ConstantAttribute{
						Name:  "style",
						Value: "padding: 10px",
					},
					ConditionalAttribute{
						Expression: Expression{
							Value: "true",
							Range: Range{
								From: Position{
									Index: 33,
									Line:  1,
									Col:   6,
								},
								To: Position{
									Index: 37,
									Line:  1,
									Col:   10,
								},
							},
						},
						Then: []Attribute{
							ConstantAttribute{
								Name:  "class",
								Value: "itIsTrue",
							},
						},
					},
				},
				IndentAttrs: true,
			},
		},
		{
			name: "element: self-closing with conditional attribute with else block",
			input: `<hr style="padding: 10px" 
			if true {
				class="itIsTrue"
			} else {
				class="itIsNotTrue"
			}
/>`,
			expected: Element{
				Name: "hr",
				Attributes: []Attribute{
					ConstantAttribute{
						Name:  "style",
						Value: "padding: 10px",
					},
					ConditionalAttribute{
						Expression: Expression{
							Value: "true",
							Range: Range{
								From: Position{
									Index: 33,
									Line:  1,
									Col:   6,
								},
								To: Position{
									Index: 37,
									Line:  1,
									Col:   10,
								},
							},
						},
						Then: []Attribute{
							ConstantAttribute{
								Name:  "class",
								Value: "itIsTrue",
							},
						},
						Else: []Attribute{
							ConstantAttribute{
								Name:  "class",
								Value: "itIsNotTrue",
							},
						},
					},
				},
				IndentAttrs: true,
			},
		},
		{
			name: "element: open and close with conditional attribute",
			input: `<p style="padding: 10px" 
			if true {
				class="itIsTrue"
			}
>Test</p>`,
			expected: Element{
				Name: "p",
				Attributes: []Attribute{
					ConstantAttribute{
						Name:  "style",
						Value: "padding: 10px",
					},
					ConditionalAttribute{
						Expression: Expression{
							Value: "true",
							Range: Range{
								From: Position{
									Index: 32,
									Line:  1,
									Col:   6,
								},
								To: Position{
									Index: 36,
									Line:  1,
									Col:   10,
								},
							},
						},
						Then: []Attribute{
							ConstantAttribute{
								Name:  "class",
								Value: "itIsTrue",
							},
						},
					},
				},
				IndentAttrs: true,
				Children: []Node{
					Text{Value: "Test"},
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
						TrailingSpace: SpaceHorizontal,
					},
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
		{
			name:  "element: inputs can contain class attributes",
			input: `<input  type="email" id="email" name="email" class={ "a", "b", "c",  templ.KV("c", false)}	placeholder="your@email.com" autocomplete="off"/>`,
			expected: Element{
				Name: "input",
				Attributes: []Attribute{
					ConstantAttribute{
						Name:  "type",
						Value: "email",
					},
					ConstantAttribute{
						Name:  "id",
						Value: "email",
					},
					ConstantAttribute{
						Name:  "name",
						Value: "email",
					},
					ExpressionAttribute{
						Name: "class",
						Expression: Expression{
							Value: `"a", "b", "c",  templ.KV("c", false)`,
							Range: Range{
								From: Position{
									Index: 53,
									Line:  0,
									Col:   53,
								},
								To: Position{
									Index: 89,
									Line:  0,
									Col:   89,
								},
							},
						},
					},
					ConstantAttribute{
						Name:  "placeholder",
						Value: "your@email.com",
					},
					ConstantAttribute{
						Name:  "autocomplete",
						Value: "off",
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
	tests := []struct {
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
		{
			name:  "element: names cannot be greater than 128 characters",
			input: `<aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa></aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa>`,
			expected: parse.Error("element names must be < 128 characters long",
				parse.Position{
					Index: 130,
					Line:  0,
					Col:   130,
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
	sb := new(strings.Builder)
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
