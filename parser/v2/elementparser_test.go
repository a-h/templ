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
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 2, Line: 0, Col: 2},
				},
			},
		},
		{
			name:   "element: hyphen in name",
			input:  `<turbo-frame>`,
			parser: StripType(elementOpenTagParser),
			expected: elementOpenTag{
				Name: "turbo-frame",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 12, Line: 0, Col: 12},
				},
			},
		},
		{
			name:   "element: open with hyperscript attribute",
			input:  `<div _="show = true">`,
			parser: StripType(elementOpenTagParser),
			expected: elementOpenTag{
				Name: "div",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 4, Line: 0, Col: 4},
				},
				Attributes: []Attribute{
					&ConstantAttribute{
						Value: "show = true",
						Key: ConstantAttributeKey{
							Name: "_",
							NameRange: Range{
								From: Position{Index: 5, Line: 0, Col: 5},
								To:   Position{Index: 6, Line: 0, Col: 6},
							},
						},
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
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 4, Line: 0, Col: 4},
				},
				Attributes: []Attribute{
					&ConstantAttribute{
						Value: "show = true",
						Key: ConstantAttributeKey{
							Name: "@click",
							NameRange: Range{
								From: Position{Index: 5, Line: 0, Col: 5},
								To:   Position{Index: 11, Line: 0, Col: 11},
							},
						},
					},
					&ConstantAttribute{
						Value: "{'foo': true}",
						Key: ConstantAttributeKey{
							Name: ":class",
							NameRange: Range{
								From: Position{Index: 26, Line: 0, Col: 26},
								To:   Position{Index: 32, Line: 0, Col: 32},
							},
						},
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
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 4, Line: 0, Col: 4},
				},
				Attributes: []Attribute{
					&ConstantAttribute{
						Value: "123",
						Key: ConstantAttributeKey{
							Name: "id",
							NameRange: Range{
								From: Position{Index: 5, Line: 0, Col: 5},
								To:   Position{Index: 7, Line: 0, Col: 7},
							},
						},
					},
					&ConstantAttribute{
						Value: "padding: 10px",
						Key: ConstantAttributeKey{
							Name: "style",
							NameRange: Range{
								From: Position{Index: 14, Line: 0, Col: 14},
								To:   Position{Index: 19, Line: 0, Col: 19},
							},
						},
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
			expected: &ConditionalAttribute{
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
					&ConstantAttribute{
						Value: "important",
						Key: ConstantAttributeKey{
							Name: "class",
							NameRange: Range{
								From: Position{Index: 23, Line: 2, Col: 3},
								To:   Position{Index: 28, Line: 2, Col: 8},
							},
						},
					},
				},
			},
		},
		{
			name: "conditional expression attribute - multiple",
			input: `
if test {` + " " + `
	class="itIsTrue"
	noshade
	name={ "other" }
}
"`,
			parser: StripType(conditionalAttribute),
			expected: &ConditionalAttribute{
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
					&ConstantAttribute{
						Value: "itIsTrue",
						Key: ConstantAttributeKey{
							Name: "class",
							NameRange: Range{
								From: Position{Index: 13, Line: 2, Col: 1},
								To:   Position{Index: 18, Line: 2, Col: 6},
							},
						},
					},
					&BoolConstantAttribute{
						Key: ConstantAttributeKey{
							Name: "noshade",
							NameRange: Range{
								From: Position{Index: 31, Line: 3, Col: 1},
								To:   Position{Index: 38, Line: 3, Col: 8},
							},
						},
					},
					&ExpressionAttribute{
						Key: ConstantAttributeKey{
							Name: "name",
							NameRange: Range{
								From: Position{Index: 40, Line: 4, Col: 1},
								To:   Position{Index: 44, Line: 4, Col: 5},
							},
						},
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
			expected: &BoolExpressionAttribute{
				Key: ConstantAttributeKey{
					Name: "noshade",
					NameRange: Range{
						From: Position{Index: 1, Line: 0, Col: 1},
						To:   Position{Index: 8, Line: 0, Col: 8},
					},
				},
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
			expected: &BoolExpressionAttribute{
				Key: ConstantAttributeKey{
					Name: "noshade",
					NameRange: Range{
						From: Position{Index: 1, Line: 0, Col: 1},
						To:   Position{Index: 8, Line: 0, Col: 8},
					},
				},
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
			parser: StripType(attributeParser{}),
			expected: &BoolExpressionAttribute{
				Key: ConstantAttributeKey{
					Name: "noshade",
					NameRange: Range{
						From: Position{Index: 1, Line: 0, Col: 1},
						To:   Position{Index: 8, Line: 0, Col: 8},
					},
				},
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
			name:   "boolean expression with excess spaces",
			input:  ` noshade?={ true   }"`,
			parser: StripType(boolExpressionAttributeParser),
			expected: &BoolExpressionAttribute{
				Key: ConstantAttributeKey{
					Name: "noshade",
					NameRange: Range{
						From: Position{Index: 1, Line: 0, Col: 1},
						To:   Position{Index: 8, Line: 0, Col: 8},
					},
				},
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
			expected: &SpreadAttributes{
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
			expected: &ConstantAttribute{
				Value: "test",
				Key: ConstantAttributeKey{
					Name: "href",
					NameRange: Range{
						From: Position{Index: 1, Line: 0, Col: 1},
						To:   Position{Index: 5, Line: 0, Col: 5},
					},
				},
			},
		},
		{
			name:   "single quote not required constant attribute",
			input:  ` href='no double quote in value'`,
			parser: StripType(constantAttributeParser),
			expected: &ConstantAttribute{
				Value:       `no double quote in value`,
				SingleQuote: false,
				Key: ConstantAttributeKey{
					Name: "href",
					NameRange: Range{
						From: Position{Index: 1, Line: 0, Col: 1},
						To:   Position{Index: 5, Line: 0, Col: 5},
					},
				},
			},
		},
		{
			name:   "single quote required constant attribute",
			input:  ` href='"test"'`,
			parser: StripType(constantAttributeParser),
			expected: &ConstantAttribute{
				Value:       `"test"`,
				SingleQuote: true,
				Key: ConstantAttributeKey{
					Name: "href",
					NameRange: Range{
						From: Position{Index: 1, Line: 0, Col: 1},
						To:   Position{Index: 5, Line: 0, Col: 5},
					},
				},
			},
		},
		{
			name:   "attribute name with hyphens",
			input:  ` data-turbo-permanent="value"`,
			parser: StripType(constantAttributeParser),
			expected: &ConstantAttribute{
				Value: "value",
				Key: ConstantAttributeKey{
					Name: "data-turbo-permanent",
					NameRange: Range{
						From: Position{Index: 1, Line: 0, Col: 1},
						To:   Position{Index: 21, Line: 0, Col: 21},
					},
				},
			},
		},
		{
			name:   "empty attribute",
			input:  ` data=""`,
			parser: StripType(constantAttributeParser),
			expected: &ConstantAttribute{
				Value: "",
				Key: ConstantAttributeKey{
					Name: "data",
					NameRange: Range{
						From: Position{Index: 1, Line: 0, Col: 1},
						To:   Position{Index: 5, Line: 0, Col: 5},
					},
				},
			},
		},
		{
			name: "multiline attribute",
			input: ` data-script="on click
                do something
             end"
`,
			parser: StripType(constantAttributeParser),
			expected: &ConstantAttribute{
				Value: "on click\n                do something\n             end",
				Key: ConstantAttributeKey{
					Name: "data-script",
					NameRange: Range{
						From: Position{Index: 1, Line: 0, Col: 1},
						To:   Position{Index: 12, Line: 0, Col: 12},
					},
				},
			},
		},
		{
			name:   "bool constant attribute",
			input:  `<div data>`,
			parser: StripType(elementOpenTagParser),
			expected: elementOpenTag{
				Name: "div",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 4, Line: 0, Col: 4},
				},
				Attributes: []Attribute{
					&BoolConstantAttribute{
						Key: ConstantAttributeKey{
							Name: "data",

							NameRange: Range{
								From: Position{Index: 5, Line: 0, Col: 5},
								To:   Position{Index: 9, Line: 0, Col: 9},
							},
						},
					},
				},
			},
		},
		{
			name:   "bool constant attribute with expression key",
			input:  `<div { "foo" }>`,
			parser: StripType(elementOpenTagParser),
			expected: elementOpenTag{
				Name: "div",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 4, Line: 0, Col: 4},
				},
				Attributes: []Attribute{
					&BoolConstantAttribute{
						Key: ExpressionAttributeKey{
							Expression: Expression{
								Value: `"foo"`,
								Range: Range{
									From: Position{
										Index: 7,
										Line:  0,
										Col:   7,
									},
									To: Position{
										Index: 12,
										Line:  0,
										Col:   12,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:   "bool constant attributes can end with a Unix newline",
			input:  "<input\n\t\trequired\n\t/>",
			parser: StripType(element),
			expected: &Element{
				Name:        "input",
				IndentAttrs: true,
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 6, Line: 0, Col: 6},
				},
				Attributes: []Attribute{
					&BoolConstantAttribute{
						Key: ConstantAttributeKey{
							Name: "required",
							NameRange: Range{
								From: Position{Index: 9, Line: 1, Col: 2},
								To:   Position{Index: 17, Line: 1, Col: 10},
							},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 21, Line: 2, Col: 3},
				},
			},
		},
		{
			name:   "bool constant attributes can end with a Windows newline",
			input:  "<input\r\n\t\trequired\r\n\t/>",
			parser: StripType(element),
			expected: &Element{
				Name:        "input",
				IndentAttrs: true,
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 6, Line: 0, Col: 6},
				},
				Attributes: []Attribute{
					&BoolConstantAttribute{
						Key: ConstantAttributeKey{
							Name: "required",
							NameRange: Range{
								From: Position{Index: 10, Line: 1, Col: 2},
								To:   Position{Index: 18, Line: 1, Col: 10},
							},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 23, Line: 2, Col: 3},
				},
			},
		},
		{
			name:   "constant attribute values are parsed as-is (single quoted)",
			input:  ` href='"'`,
			parser: StripType(constantAttributeParser),
			expected: &ConstantAttribute{
				Value:       `"`,
				SingleQuote: true,
				Key: ConstantAttributeKey{
					Name: "href",
					NameRange: Range{
						From: Position{Index: 1, Line: 0, Col: 1},
						To:   Position{Index: 5, Line: 0, Col: 5},
					},
				},
			},
		},
		{
			name:   "constant attribute values are parsed as-is (double quoted)",
			input:  ` href="&lt;'"`,
			parser: StripType(constantAttributeParser),
			expected: &ConstantAttribute{
				Value: `&lt;'`,
				Key: ConstantAttributeKey{
					Name: "href",
					NameRange: Range{
						From: Position{Index: 1, Line: 0, Col: 1},
						To:   Position{Index: 5, Line: 0, Col: 5},
					},
				},
			},
		},
		{
			name:   "HTMX wildcard attribute names are supported",
			input:  ` hx-target-*="#errors"`,
			parser: StripType(constantAttributeParser),
			expected: &ConstantAttribute{
				Value: `#errors`,
				Key: ConstantAttributeKey{
					Name: "hx-target-*",
					NameRange: Range{
						From: Position{Index: 1, Line: 0, Col: 1},
						To:   Position{Index: 12, Line: 0, Col: 12},
					},
				},
			},
		},
		{
			name:   "unquoted attributes are supported",
			input:  ` data=123`,
			parser: StripType(constantAttributeParser),
			expected: &ConstantAttribute{
				Value: "123",
				Key: ConstantAttributeKey{
					Name: "data",
					NameRange: Range{
						From: Position{Index: 1, Line: 0, Col: 1},
						To:   Position{Index: 5, Line: 0, Col: 5},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			result, matched, err := tt.parser.Parse(input)
			if err != nil {
				t.Error(err)
			}
			if !matched {
				t.Errorf("failed to parse at %v", input.Position())
			}
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestVoidElementCloserParser(t *testing.T) {
	t.Run("all void elements are parsed", func(t *testing.T) {
		for _, input := range voidElementCloseTags {
			_, matched, err := voidElementCloser.Parse(parse.NewInput(input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !matched {
				t.Fatalf("failed to parse %q", input)
			}
		}
	})
}

func TestElementParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Element
	}{
		{
			name:  "element: self-closing with single constant attribute",
			input: `<a href="test"/>`,
			expected: &Element{
				Name: "a",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 2, Line: 0, Col: 2},
				},
				Attributes: []Attribute{
					&ConstantAttribute{
						Value: "test",
						Key: ConstantAttributeKey{
							Name: "href",
							NameRange: Range{
								From: Position{Index: 3, Line: 0, Col: 3},
								To:   Position{Index: 7, Line: 0, Col: 7},
							},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 16, Line: 0, Col: 16},
				},
			},
		},
		{
			name:  "element: colon in name, empty",
			input: `<maps:map></maps:map>`,
			expected: &Element{
				Name: "maps:map",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 9, Line: 0, Col: 9},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 21, Line: 0, Col: 21},
				},
			},
		},
		{
			name:  "element: colon in name, with content",
			input: `<maps:map>Content</maps:map>`,
			expected: &Element{
				Name: "maps:map",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 9, Line: 0, Col: 9},
				},
				Children: []Node{
					&Text{
						Value: "Content",
						Range: Range{
							From: Position{Index: 10, Line: 0, Col: 10},
							To:   Position{Index: 17, Line: 0, Col: 17},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 28, Line: 0, Col: 28},
				},
			},
		},
		{
			name:  "element: void (input)",
			input: `<input>`,
			expected: &Element{
				Name: "input",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 6, Line: 0, Col: 6},
				},
				Children: nil,
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 7, Line: 0, Col: 7},
				},
			},
		},
		{
			name:  "element: void (br)",
			input: `<br>`,
			expected: &Element{
				Name: "br",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 3, Line: 0, Col: 3},
				},
				Children: nil,
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 4, Line: 0, Col: 4},
				},
			},
		},
		{
			name:  "element: void (hr)",
			input: `<hr noshade>`,
			expected: &Element{
				Name: "hr",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 3, Line: 0, Col: 3},
				},
				Attributes: []Attribute{
					&BoolConstantAttribute{
						Key: ConstantAttributeKey{
							Name: "noshade",
							NameRange: Range{
								From: Position{Index: 4, Line: 0, Col: 4},
								To:   Position{Index: 11, Line: 0, Col: 11},
							},
						},
					},
				},
				Children: nil,
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 12, Line: 0, Col: 12},
				},
			},
		},
		{
			name:  "element: void with content",
			input: `<input>Text</input>`,
			expected: &Element{
				Name: "input",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 6, Line: 0, Col: 6},
				},
				// <input> is a void element, so text is not a child of the input.
				// </input> is ignored.
				Children: nil,
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 7, Line: 0, Col: 7},
				},
			},
		},
		{
			name:  "element: self-closing with single bool expression attribute",
			input: `<hr noshade?={ true }/>`,
			expected: &Element{
				Name: "hr",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 3, Line: 0, Col: 3},
				},
				Attributes: []Attribute{
					&BoolExpressionAttribute{
						Key: ConstantAttributeKey{
							Name: "noshade",
							NameRange: Range{
								From: Position{Index: 4, Line: 0, Col: 4},
								To:   Position{Index: 11, Line: 0, Col: 11},
							},
						},
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
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 23, Line: 0, Col: 23},
				},
			},
		},
		{
			name:  "element: void nesting same is OK",
			input: `<div><br><br></br></div>`,
			expected: &Element{
				Name: "div",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 4, Line: 0, Col: 4},
				},
				Children: []Node{
					&Element{
						Name: "br", // The <br> one.
						NameRange: Range{
							From: Position{Index: 6, Line: 0, Col: 6},
							To:   Position{Index: 8, Line: 0, Col: 8},
						},
						Range: Range{
							From: Position{Index: 5, Line: 0, Col: 5},
							To:   Position{Index: 9, Line: 0, Col: 9},
						},
					},
					&Element{
						Name: "br", // The <br></br> one.
						NameRange: Range{
							From: Position{Index: 10, Line: 0, Col: 10},
							To:   Position{Index: 12, Line: 0, Col: 12},
						},
						Range: Range{
							From: Position{Index: 9, Line: 0, Col: 9},
							To:   Position{Index: 18, Line: 0, Col: 18},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 24, Line: 0, Col: 24},
				},
			},
		},
		{
			name:  "element: void nesting is ignored",
			input: `<br><hr></br>`,
			expected: &Element{
				Name: "br",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 3, Line: 0, Col: 3},
				},
				// <br> is a void element, so <hr> is not a child of the <br>.
				// </br> is ignored.
				Children: nil,
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 4, Line: 0, Col: 4},
				},
			},
		},
		{
			name:  "element: self-closing with single expression attribute",
			input: `<a href={ "test" }/>`,
			expected: &Element{
				Name: "a",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 2, Line: 0, Col: 2},
				},
				Attributes: []Attribute{
					&ExpressionAttribute{
						Key: ConstantAttributeKey{
							Name: "href",
							NameRange: Range{
								From: Position{Index: 3, Line: 0, Col: 3},
								To:   Position{Index: 7, Line: 0, Col: 7},
							},
						},
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
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 20, Line: 0, Col: 20},
				},
			},
		},
		{
			name:  "element: self-closing with multiple constant attributes",
			input: `<a href="test" style="text-underline: auto"/>`,
			expected: &Element{
				Name: "a",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 2, Line: 0, Col: 2},
				},
				Attributes: []Attribute{
					&ConstantAttribute{
						Value: "test",
						Key: ConstantAttributeKey{
							Name: "href",
							NameRange: Range{
								From: Position{Index: 3, Line: 0, Col: 3},
								To:   Position{Index: 7, Line: 0, Col: 7},
							},
						},
					},
					&ConstantAttribute{
						Value: "text-underline: auto",
						Key: ConstantAttributeKey{
							Name: "style",
							NameRange: Range{
								From: Position{Index: 15, Line: 0, Col: 15},
								To:   Position{Index: 20, Line: 0, Col: 20},
							},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 45, Line: 0, Col: 45},
				},
			},
		},
		{
			name:  "element: self-closing with multiple spreads attributes",
			input: `<a { firstSpread... } { children... }/>`,
			expected: &Element{
				Name: "a",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 2, Line: 0, Col: 2},
				},
				Attributes: []Attribute{
					&SpreadAttributes{
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
					&SpreadAttributes{
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
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 39, Line: 0, Col: 39},
				},
			},
		},
		{
			name:  "element: self-closing with multiple boolean attributes",
			input: `<hr optionA optionB?={ true } optionC="other"/>`,
			expected: &Element{
				Name: "hr",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 3, Line: 0, Col: 3},
				},
				Attributes: []Attribute{
					&BoolConstantAttribute{
						Key: ConstantAttributeKey{
							Name: "optionA",
							NameRange: Range{
								From: Position{Index: 4, Line: 0, Col: 4},
								To:   Position{Index: 11, Line: 0, Col: 11},
							},
						},
					},
					&BoolExpressionAttribute{
						Key: ConstantAttributeKey{
							Name: "optionB",
							NameRange: Range{
								From: Position{Index: 12, Line: 0, Col: 12},
								To:   Position{Index: 19, Line: 0, Col: 19},
							},
						},
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
					&ConstantAttribute{
						Value: "other",
						Key: ConstantAttributeKey{
							Name: "optionC",
							NameRange: Range{
								From: Position{Index: 30, Line: 0, Col: 30},
								To:   Position{Index: 37, Line: 0, Col: 37},
							},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 47, Line: 0, Col: 47},
				},
			},
		},
		{
			name:  "element: self-closing with multiple constant and expr attributes",
			input: `<a href="test" title={ localisation.Get("a_title") } style="text-underline: auto"/>`,
			expected: &Element{
				Name: "a",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 2, Line: 0, Col: 2},
				},
				Attributes: []Attribute{
					&ConstantAttribute{
						Value: "test",
						Key: ConstantAttributeKey{
							Name: "href",
							NameRange: Range{
								From: Position{Index: 3, Line: 0, Col: 3},
								To:   Position{Index: 7, Line: 0, Col: 7},
							},
						},
					},
					&ExpressionAttribute{
						Key: ConstantAttributeKey{
							Name: "title",
							NameRange: Range{
								From: Position{Index: 15, Line: 0, Col: 15},
								To:   Position{Index: 20, Line: 0, Col: 20},
							},
						},
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
					&ConstantAttribute{
						Value: "text-underline: auto",
						Key: ConstantAttributeKey{
							Name: "style",
							NameRange: Range{
								From: Position{Index: 53, Line: 0, Col: 53},
								To:   Position{Index: 58, Line: 0, Col: 58},
							},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 83, Line: 0, Col: 83},
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
			expected: &Element{
				Name: "div",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 4, Line: 0, Col: 4},
				},
				Attributes: []Attribute{
					&ConstantAttribute{
						Value: "width: 100;",
						Key: ConstantAttributeKey{
							Name: "style",
							NameRange: Range{
								From: Position{Index: 5, Line: 0, Col: 5},
								To:   Position{Index: 10, Line: 0, Col: 10},
							},
						},
					},
					&ConditionalAttribute{
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
							&ConstantAttribute{
								Value: "important",
								Key: ConstantAttributeKey{
									Name: "class",
									NameRange: Range{
										From: Position{Index: 47, Line: 2, Col: 3},
										To:   Position{Index: 52, Line: 2, Col: 8},
									},
								},
							},
						},
					},
				},
				IndentAttrs: true,
				Children: []Node{
					&Text{
						Value: "Test",
						Range: Range{
							From: Position{Index: 70, Line: 4, Col: 1},
							To:   Position{Index: 74, Line: 4, Col: 5},
						},
					},
				},
				TrailingSpace: SpaceVertical,
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 81, Line: 5, Col: 0},
				},
			},
		},
		{
			name:  "element: self-closing with no attributes",
			input: `<hr/>`,
			expected: &Element{
				Name: "hr",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 3, Line: 0, Col: 3},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 5, Line: 0, Col: 5},
				},
			},
		},
		{
			name:  "element: self-closing with attribute",
			input: `<hr style="padding: 10px" />`,
			expected: &Element{
				Name: "hr",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 3, Line: 0, Col: 3},
				},
				Attributes: []Attribute{
					&ConstantAttribute{
						Value: "padding: 10px",
						Key: ConstantAttributeKey{
							Name: "style",
							NameRange: Range{
								From: Position{Index: 4, Line: 0, Col: 4},
								To:   Position{Index: 9, Line: 0, Col: 9},
							},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 28, Line: 0, Col: 28},
				},
			},
		},
		{
			name: "element: self-closing with conditional attribute",
			input: `<hr style="padding: 10px"` + " " + `
			if true {
				class="itIsTrue"
			}
/>`,
			expected: &Element{
				Name: "hr",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 3, Line: 0, Col: 3},
				},
				Attributes: []Attribute{
					&ConstantAttribute{
						Value: "padding: 10px",
						Key: ConstantAttributeKey{
							Name: "style",
							NameRange: Range{
								From: Position{Index: 4, Line: 0, Col: 4},
								To:   Position{Index: 9, Line: 0, Col: 9},
							},
						},
					},
					&ConditionalAttribute{
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
							&ConstantAttribute{
								Value: "itIsTrue",
								Key: ConstantAttributeKey{
									Name: "class",
									NameRange: Range{
										From: Position{Index: 44, Line: 2, Col: 4},
										To:   Position{Index: 49, Line: 2, Col: 9},
									},
								},
							},
						},
					},
				},
				IndentAttrs: true,
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 68, Line: 4, Col: 2},
				},
			},
		},
		{
			name: "element: self-closing with conditional attribute with else block",
			input: `<hr style="padding: 10px"` + " " + `
			if true {
				class="itIsTrue"
			} else {
				class="itIsNotTrue"
			}
/>`,
			expected: &Element{
				Name: "hr",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 3, Line: 0, Col: 3},
				},
				Attributes: []Attribute{
					&ConstantAttribute{
						Value: "padding: 10px",
						Key: ConstantAttributeKey{
							Name: "style",
							NameRange: Range{
								From: Position{Index: 4, Line: 0, Col: 4},
								To:   Position{Index: 9, Line: 0, Col: 9},
							},
						},
					},
					&ConditionalAttribute{
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
							&ConstantAttribute{
								Value: "itIsTrue",
								Key: ConstantAttributeKey{
									Name: "class",
									NameRange: Range{
										From: Position{Index: 44, Line: 2, Col: 4},
										To:   Position{Index: 49, Line: 2, Col: 9},
									},
								},
							},
						},
						Else: []Attribute{
							&ConstantAttribute{
								Value: "itIsNotTrue",
								Key: ConstantAttributeKey{
									Name: "class",
									NameRange: Range{
										From: Position{Index: 77, Line: 4, Col: 4},
										To:   Position{Index: 82, Line: 4, Col: 9},
									},
								},
							},
						},
					},
				},
				IndentAttrs: true,
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 104, Line: 6, Col: 2},
				},
			},
		},
		{
			name: "element: open and close with conditional attribute",
			input: `<p style="padding: 10px"` + " " + `
			if true {
				class="itIsTrue"
			}
>Test</p>`,
			expected: &Element{
				Name: "p",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 2, Line: 0, Col: 2},
				},
				Attributes: []Attribute{
					&ConstantAttribute{
						Value: "padding: 10px",
						Key: ConstantAttributeKey{
							Name: "style",
							NameRange: Range{
								From: Position{Index: 3, Line: 0, Col: 3},
								To:   Position{Index: 8, Line: 0, Col: 8},
							},
						},
					},
					&ConditionalAttribute{
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
							&ConstantAttribute{
								Value: "itIsTrue",
								Key: ConstantAttributeKey{
									Name: "class",
									NameRange: Range{
										From: Position{Index: 43, Line: 2, Col: 4},
										To:   Position{Index: 48, Line: 2, Col: 9},
									},
								},
							},
						},
					},
				},
				IndentAttrs: true,
				Children: []Node{
					&Text{
						Value: "Test",
						Range: Range{
							From: Position{Index: 66, Line: 4, Col: 1},
							To:   Position{Index: 70, Line: 4, Col: 5},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 74, Line: 4, Col: 9},
				},
			},
		},
		{
			name:  "element: open and close",
			input: `<a></a>`,
			expected: &Element{
				Name: "a",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 2, Line: 0, Col: 2},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 7, Line: 0, Col: 7},
				},
			},
		},
		{
			name:  "element: open and close with text",
			input: `<a>The text</a>`,
			expected: &Element{
				Name: "a",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 2, Line: 0, Col: 2},
				},
				Children: []Node{
					&Text{
						Value: "The text",
						Range: Range{
							From: Position{Index: 3, Line: 0, Col: 3},
							To:   Position{Index: 11, Line: 0, Col: 11},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 15, Line: 0, Col: 15},
				},
			},
		},
		{
			name:  "element: with self-closing child element",
			input: `<a><b/></a>`,
			expected: &Element{
				Name: "a",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 2, Line: 0, Col: 2},
				},
				Children: []Node{
					&Element{
						Name: "b",
						NameRange: Range{
							From: Position{Index: 4, Line: 0, Col: 4},
							To:   Position{Index: 5, Line: 0, Col: 5},
						},
						Range: Range{
							From: Position{Index: 3, Line: 0, Col: 3},
							To:   Position{Index: 7, Line: 0, Col: 7},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 11, Line: 0, Col: 11},
				},
			},
		},
		{
			name:  "element: with non-self-closing child element",
			input: `<a><b></b></a>`,
			expected: &Element{
				Name: "a",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 2, Line: 0, Col: 2},
				},
				Children: []Node{
					&Element{
						Name: "b",
						NameRange: Range{
							From: Position{Index: 4, Line: 0, Col: 4},
							To:   Position{Index: 5, Line: 0, Col: 5},
						},
						Range: Range{
							From: Position{Index: 3, Line: 0, Col: 3},
							To:   Position{Index: 10, Line: 0, Col: 10},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 14, Line: 0, Col: 14},
				},
			},
		},
		{
			name:  "element: containing space",
			input: `<a> <b> </b> </a>`,
			expected: &Element{
				Name: "a",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 2, Line: 0, Col: 2},
				},
				Children: []Node{
					&Whitespace{Value: " "},
					&Element{
						Name: "b",
						NameRange: Range{
							From: Position{Index: 5, Line: 0, Col: 5},
							To:   Position{Index: 6, Line: 0, Col: 6},
						},

						Children: []Node{
							&Whitespace{Value: " "},
						},
						TrailingSpace: SpaceHorizontal,
						Range: Range{
							From: Position{Index: 4, Line: 0, Col: 4},
							To:   Position{Index: 13, Line: 0, Col: 13},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 17, Line: 0, Col: 17},
				},
			},
		},
		{
			name:  "element: with multiple child elements",
			input: `<a><b></b><c><d/></c></a>`,
			expected: &Element{
				Name: "a",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 2, Line: 0, Col: 2},
				},
				Children: []Node{
					&Element{
						Name: "b",
						NameRange: Range{
							From: Position{Index: 4, Line: 0, Col: 4},
							To:   Position{Index: 5, Line: 0, Col: 5},
						},
						Range: Range{
							From: Position{Index: 3, Line: 0, Col: 3},
							To:   Position{Index: 10, Line: 0, Col: 10},
						},
					},
					&Element{
						Name: "c",
						NameRange: Range{
							From: Position{Index: 11, Line: 0, Col: 11},
							To:   Position{Index: 12, Line: 0, Col: 12},
						},
						Children: []Node{
							&Element{
								Name: "d",
								NameRange: Range{
									From: Position{Index: 14, Line: 0, Col: 14},
									To:   Position{Index: 15, Line: 0, Col: 15},
								},
								Range: Range{
									From: Position{Index: 13, Line: 0, Col: 13},
									To:   Position{Index: 17, Line: 0, Col: 17},
								},
							},
						},
						Range: Range{
							From: Position{Index: 10, Line: 0, Col: 10},
							To:   Position{Index: 21, Line: 0, Col: 21},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 25, Line: 0, Col: 25},
				},
			},
		},
		{
			name:  "element: empty",
			input: `<div></div>`,
			expected: &Element{
				Name: "div",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 4, Line: 0, Col: 4},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 11, Line: 0, Col: 11},
				},
			},
		},
		{
			name:  "element: containing string expression",
			input: `<div>{ "test" }</div>`,
			expected: &Element{
				Name: "div",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 4, Line: 0, Col: 4},
				},
				Children: []Node{
					&StringExpression{
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
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 21, Line: 0, Col: 21},
				},
			},
		},
		{
			name:  "element: inputs can contain class attributes",
			input: `<input  type="email" id="email" name="email" class={ "a", "b", "c",  templ.KV("c", false)}	placeholder="your@email.com" autocomplete="off"/>`,
			expected: &Element{
				Name: "input",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 6, Line: 0, Col: 6},
				},
				Attributes: []Attribute{
					&ConstantAttribute{
						Value: "email",
						Key: ConstantAttributeKey{
							Name: "type",
							NameRange: Range{
								From: Position{Index: 8, Line: 0, Col: 8},
								To:   Position{Index: 12, Line: 0, Col: 12},
							},
						},
					},
					&ConstantAttribute{
						Value: "email",
						Key: ConstantAttributeKey{
							Name: "id",
							NameRange: Range{
								From: Position{Index: 21, Line: 0, Col: 21},
								To:   Position{Index: 23, Line: 0, Col: 23},
							},
						},
					},
					&ConstantAttribute{
						Value: "email",
						Key: ConstantAttributeKey{
							Name: "name",
							NameRange: Range{
								From: Position{Index: 32, Line: 0, Col: 32},
								To:   Position{Index: 36, Line: 0, Col: 36},
							},
						},
					},
					&ExpressionAttribute{
						Key: ConstantAttributeKey{
							Name: "class",
							NameRange: Range{
								From: Position{Index: 45, Line: 0, Col: 45},
								To:   Position{Index: 50, Line: 0, Col: 50},
							},
						},
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
					&ConstantAttribute{
						Value: "your@email.com",
						Key: ConstantAttributeKey{
							Name: "placeholder",
							NameRange: Range{
								From: Position{Index: 91, Line: 0, Col: 91},
								To:   Position{Index: 102, Line: 0, Col: 102},
							},
						},
					},
					&ConstantAttribute{
						Value: "off",
						Key: ConstantAttributeKey{
							Name: "autocomplete",
							NameRange: Range{
								From: Position{Index: 120, Line: 0, Col: 120},
								To:   Position{Index: 132, Line: 0, Col: 132},
							},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 140, Line: 0, Col: 140},
				},
			},
		},
		{
			name: "element: with multi-line attributes",
			input: `<input
	type="email"` + " " + `
	id="email"` + " " + `
	name="email"
></input>`,
			expected: &Element{
				Name:        "input",
				IndentAttrs: true,
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 6, Line: 0, Col: 6},
				},
				Attributes: []Attribute{
					&ConstantAttribute{
						Value: "email",
						Key: ConstantAttributeKey{
							Name: "type",
							NameRange: Range{
								From: Position{Index: 8, Line: 1, Col: 1},
								To:   Position{Index: 12, Line: 1, Col: 5},
							},
						},
					},
					&ConstantAttribute{
						Value: "email",
						Key: ConstantAttributeKey{
							Name: "id",
							NameRange: Range{
								From: Position{Index: 23, Line: 2, Col: 1},
								To:   Position{Index: 25, Line: 2, Col: 3},
							},
						},
					},
					&ConstantAttribute{
						Value: "email",
						Key: ConstantAttributeKey{
							Name: "name",
							NameRange: Range{
								From: Position{Index: 36, Line: 3, Col: 1},
								To:   Position{Index: 40, Line: 3, Col: 5},
							},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 58, Line: 4, Col: 9},
				},
			},
		},
		{
			name:  "element: self-closing with unquoted attribute",
			input: `<hr noshade=noshade/>`,
			expected: &Element{
				Name: "hr",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 3, Line: 0, Col: 3},
				},
				Attributes: []Attribute{
					&ConstantAttribute{
						Value: "noshade",
						Key: ConstantAttributeKey{
							Name: "noshade",
							NameRange: Range{
								From: Position{Index: 4, Line: 0, Col: 4},
								To:   Position{Index: 11, Line: 0, Col: 11},
							},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 21, Line: 0, Col: 21},
				},
			},
		},
		{
			name:  "element: self-closing with unquoted and other attributes",
			input: `<hr noshade=noshade disabled other-attribute={ false } />`,
			expected: &Element{
				Name: "hr",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 3, Line: 0, Col: 3},
				},
				Attributes: []Attribute{
					&ConstantAttribute{
						Value: "noshade",
						Key: ConstantAttributeKey{
							Name: "noshade",
							NameRange: Range{
								From: Position{Index: 4, Line: 0, Col: 4},
								To:   Position{Index: 11, Line: 0, Col: 11},
							},
						},
					},
					&BoolConstantAttribute{
						Key: ConstantAttributeKey{
							Name: "disabled",
							NameRange: Range{
								From: Position{Index: 20, Line: 0, Col: 20},
								To:   Position{Index: 28, Line: 0, Col: 28},
							},
						},
					},
					&ExpressionAttribute{
						Key: ConstantAttributeKey{
							Name: "other-attribute",
							NameRange: Range{
								From: Position{Index: 29, Line: 0, Col: 29},
								To:   Position{Index: 44, Line: 0, Col: 44},
							},
						},
						Expression: Expression{
							Value: "false",
							Range: Range{
								From: Position{
									Index: 47,
									Line:  0,
									Col:   47,
								},
								To: Position{
									Index: 52,
									Line:  0,
									Col:   52,
								},
							},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 57, Line: 0, Col: 57},
				},
			},
		},
		{
			name:  "element: space before name= { is still parsed as an expression",
			input: `<div hx-vals= { "test" }></div>`,
			expected: &Element{
				Name: "div",
				NameRange: Range{
					From: Position{Index: 1, Line: 0, Col: 1},
					To:   Position{Index: 4, Line: 0, Col: 4},
				},
				Attributes: []Attribute{
					&ExpressionAttribute{
						Key: ConstantAttributeKey{
							Name: "hx-vals",
							NameRange: Range{
								From: Position{Index: 5, Line: 0, Col: 5},
								To:   Position{Index: 12, Line: 0, Col: 12},
							},
						},
						Expression: Expression{
							Value: `"test"`,
							Range: Range{
								From: Position{
									Index: 16,
									Line:  0,
									Col:   16,
								},
								To: Position{
									Index: 22,
									Line:  0,
									Col:   22,
								},
							},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 31, Line: 0, Col: 31},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			result, matched, err := element.Parse(input)
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}
			if !matched {
				t.Fatalf("failed to parse at %d", input.Index())
			}
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Error(diff)
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
			expected: parse.Error("<a>: close tag not found",
				parse.Position{
					Index: 3,
					Line:  0,
					Col:   3,
				}),
		},
		{
			name:  "element: style must only contain text",
			input: `<style><button /></style>`,
			expected: parse.Error("<style>: invalid node contents: style elements must only contain text",
				parse.Position{
					Index: 0,
					Line:  0,
					Col:   0,
				}),
		},
		{
			name:  "element: style tags cannot contain non-text nodes",
			input: `<style>{ "value" }</style>`,
			expected: parse.Error("<style>: invalid node contents: style elements must only contain text",
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
	sb.WriteString(strings.Repeat("a", 16384))
	sb.WriteString("</div>")
	_, matched, err := element.Parse(parse.NewInput(sb.String()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !matched {
		t.Errorf("unexpected failure to parse")
	}
}

func FuzzElement(f *testing.F) {
	seeds := []string{
		`<br>`,
		`<a href="test" unquoted=unquoted/>`,
		`<input value={ "test" }/>`,
		`<div>{ "test" }</div>`,
		`<a unquoted=unquoted href="test" unquoted=unquoted>Test</a>`,
	}

	for _, tc := range seeds {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, input string) {
		_, _, _ = element.Parse(parse.NewInput(input))
	})
}

func TestElementFormatting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "element can be formatted",
			input:    `<div>Text</div>`,
			expected: `<div>Text</div>`,
		},
		{
			name:     "attribute escaping is not lost",
			input:    `<div data-line-data="[{&quot;x&quot;: 0, &quot;y&quot;: 1 }]"></div>`,
			expected: `<div data-line-data="[{&quot;x&quot;: 0, &quot;y&quot;: 1 }]"></div>`,
		},
		{
			name:     "constant attributes are not re-escaped - single quote",
			input:    `<div data-line-data='"'></div>`,
			expected: `<div data-line-data='"'></div>`,
		},
		{
			name:     "constant attributes are not re-escaped - double quote",
			input:    `<div data-line-data='"'></div>`,
			expected: `<div data-line-data='"'></div>`,
		},
		{
			name:  "constant attributes are not re-escaped - backslash",
			input: `<div data-line-data='\'></div>`,
			// templ formatting prefers double quotes.
			expected: `<div data-line-data="\"></div>`,
		},
		{
			name:  "constant attributes are not re-escaped - ampersand",
			input: `<div data-line-data='&'></div>`,
			// templ formatting prefers double quotes.
			expected: `<div data-line-data="&"></div>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			result, matched, err := element.Parse(input)
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}
			if !matched {
				t.Fatalf("failed to parse at %d", input.Index())
			}
			formatted := new(strings.Builder)
			if err = result.Write(formatted, 0); err != nil {
				t.Fatalf("failed to format: %v", err)
			}
			if diff := cmp.Diff(tt.expected, formatted.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}
