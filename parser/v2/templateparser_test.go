package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestTemplateParser(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *HTMLTemplate
		expectError bool
	}{
		{
			name: "template: no parameters",
			input: `templ Name() {
}`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 16, Line: 1, Col: 1},
				},
				Expression: Expression{
					Value: "Name()",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  0,
							Col:   6,
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
		{
			name: "template: with receiver",
			input: `templ (data Data) Name() {
}`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 28, Line: 1, Col: 1},
				},
				Expression: Expression{
					Value: "(data Data) Name()",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  0,
							Col:   6,
						},
						To: Position{
							Index: 24,
							Line:  0,
							Col:   24,
						},
					},
				},
			},
		},
		{
			name: "template: no spaces",
			input: `templ Name(){
}`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 15, Line: 1, Col: 1},
				},
				Expression: Expression{
					Value: "Name()",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  0,
							Col:   6,
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
		{
			name: "template: single parameter",
			input: `templ Name(p Parameter) {
}`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 27, Line: 1, Col: 1},
				},
				Expression: Expression{
					Value: "Name(p Parameter)",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  0,
							Col:   6,
						},
						To: Position{
							Index: 23,
							Line:  0,
							Col:   23,
						},
					},
				},
			},
		},
		{
			name: "template: can have multiline params",
			input: `templ Multiline(
	params expense,
) {
}`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 39, Line: 3, Col: 1},
				},
				Expression: Expression{
					Value: "Multiline(\n\tparams expense,\n)",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  0,
							Col:   6,
						},
						To: Position{
							Index: 35,
							Line:  2,
							Col:   1,
						},
					},
				},
			},
		},
		{
			name: "template: containing element",
			input: `templ Name(p Parameter) {
<span>{ "span content" }</span>
}`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 59, Line: 2, Col: 1},
				},
				Expression: Expression{
					Value: "Name(p Parameter)",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  0,
							Col:   6,
						},
						To: Position{
							Index: 23,
							Line:  0,
							Col:   23,
						},
					},
				},
				Children: []Node{
					&Element{
						Name: "span",
						NameRange: Range{
							From: Position{Index: 27, Line: 1, Col: 1},
							To:   Position{Index: 31, Line: 1, Col: 5},
						},
						Children: []Node{
							&StringExpression{
								Expression: Expression{
									Value: `"span content"`,
									Range: Range{
										From: Position{
											Index: 34,
											Line:  1,
											Col:   8,
										},
										To: Position{
											Index: 48,
											Line:  1,
											Col:   22,
										},
									},
								},
							},
						},
						TrailingSpace: SpaceVertical,
						Range: Range{
							From: Position{Index: 26, Line: 1, Col: 0},
							To:   Position{Index: 58, Line: 2, Col: 0},
						},
					},
				},
			},
		},
		{
			name:  "template: containing element - no spacing",
			input: `templ Name(p Parameter) { <span>{ "span content" }</span> }`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 59, Line: 0, Col: 59},
				},
				Expression: Expression{
					Value: "Name(p Parameter)",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  0,
							Col:   6,
						},
						To: Position{
							Index: 23,
							Line:  0,
							Col:   23,
						},
					},
				},
				Children: []Node{
					&Element{
						Name: "span",
						NameRange: Range{
							From: Position{Index: 27, Line: 0, Col: 27},
							To:   Position{Index: 31, Line: 0, Col: 31},
						},
						Children: []Node{
							&StringExpression{
								Expression: Expression{
									Value: `"span content"`,
									Range: Range{
										From: Position{
											Index: 34,
											Line:  0,
											Col:   34,
										},
										To: Position{
											Index: 48,
											Line:  0,
											Col:   48,
										},
									},
								},
							},
						},
						TrailingSpace: SpaceHorizontal,
						Range: Range{
							From: Position{Index: 26, Line: 0, Col: 26},
							To:   Position{Index: 58, Line: 0, Col: 58},
						},
					},
				},
			},
		},
		{
			name: "template: containing nested elements",
			input: `templ Name(p Parameter) {
<div>
  { "div content" }
  <span>
	{ "span content" }
  </span>
</div>
}`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 99, Line: 7, Col: 1},
				},
				Expression: Expression{
					Value: "Name(p Parameter)",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  0,
							Col:   6,
						},
						To: Position{
							Index: 23,
							Line:  0,
							Col:   23,
						},
					},
				},
				Children: []Node{
					&Element{
						Name: "div",
						NameRange: Range{
							From: Position{Index: 27, Line: 1, Col: 1},
							To:   Position{Index: 30, Line: 1, Col: 4},
						},
						Children: []Node{
							&Whitespace{Value: "\n  "},
							&StringExpression{
								Expression: Expression{
									Value: `"div content"`,
									Range: Range{
										From: Position{
											Index: 36,
											Line:  2,
											Col:   4,
										},
										To: Position{
											Index: 49,
											Line:  2,
											Col:   17,
										},
									},
								},
								TrailingSpace: SpaceVertical,
							},
							&Element{
								Name: "span",
								NameRange: Range{
									From: Position{Index: 55, Line: 3, Col: 3},
									To:   Position{Index: 59, Line: 3, Col: 7},
								},
								Children: []Node{
									&Whitespace{Value: "\n\t"},
									&StringExpression{
										Expression: Expression{
											Value: `"span content"`,
											Range: Range{
												From: Position{
													Index: 64,
													Line:  4,
													Col:   3,
												},
												To: Position{
													Index: 78,
													Line:  4,
													Col:   17,
												},
											},
										},
										TrailingSpace: SpaceVertical,
									},
								},
								IndentChildren: true,
								TrailingSpace:  SpaceVertical,
								Range: Range{
									From: Position{Index: 54, Line: 3, Col: 2},
									To:   Position{Index: 91, Line: 6, Col: 0},
								},
							},
						},
						IndentChildren: true,
						TrailingSpace:  SpaceVertical,
						Range: Range{
							From: Position{Index: 26, Line: 1, Col: 0},
							To:   Position{Index: 98, Line: 7, Col: 0},
						},
					},
				},
			},
		},
		{
			name: "template: containing if element",
			input: `templ Name(p Parameter) {
	if p.Test {
		<span>
			{ "span content" }
		</span>
	}
}`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 84, Line: 6, Col: 1},
				},
				Expression: Expression{
					Value: "Name(p Parameter)",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  0,
							Col:   6,
						},
						To: Position{
							Index: 23,
							Line:  0,
							Col:   23,
						},
					},
				},
				Children: []Node{
					&Whitespace{Value: "\t"},
					&IfExpression{
						Expression: Expression{
							Value: `p.Test`,
							Range: Range{
								From: Position{
									Index: 30,
									Line:  1,
									Col:   4,
								},
								To: Position{
									Index: 36,
									Line:  1,
									Col:   10,
								},
							},
						},
						Then: []Node{
							&Whitespace{Value: "\t\t"},
							&Element{
								Name: "span",
								NameRange: Range{
									From: Position{Index: 42, Line: 2, Col: 3},
									To:   Position{Index: 46, Line: 2, Col: 7},
								},
								Children: []Node{
									&Whitespace{"\n\t\t\t"},
									&StringExpression{
										Expression: Expression{
											Value: `"span content"`,
											Range: Range{
												From: Position{
													Index: 53,
													Line:  3,
													Col:   5,
												},
												To: Position{
													Index: 67,
													Line:  3,
													Col:   19,
												},
											},
										},
										TrailingSpace: SpaceVertical,
									},
								},
								IndentChildren: true,
								TrailingSpace:  SpaceVertical,
								Range: Range{
									From: Position{Index: 41, Line: 2, Col: 2},
									To:   Position{Index: 81, Line: 5, Col: 1},
								},
							},
						},
						Range: Range{
							From: Position{Index: 27, Line: 1, Col: 1},
							To:   Position{Index: 82, Line: 5, Col: 2},
						},
					},
					&Whitespace{
						Value: "\n",
					},
				},
			},
		},
		{
			name: "template: inputs",
			input: `templ Name(p Parameter) {
	<input type="text" value="a" />
	<input type="text" value="b" />
}`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 93, Line: 3, Col: 1},
				},
				Expression: Expression{
					Value: "Name(p Parameter)",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  0,
							Col:   6,
						},
						To: Position{
							Index: 23,
							Line:  0,
							Col:   23,
						},
					},
				},
				Children: []Node{
					&Whitespace{Value: "\t"},
					&Element{
						Name: "input",
						NameRange: Range{
							From: Position{Index: 28, Line: 1, Col: 2},
							To:   Position{Index: 33, Line: 1, Col: 7},
						},
						Attributes: []Attribute{
							&ConstantAttribute{
								Value: "text",
								Key: ConstantAttributeKey{
									Name: "type",
									NameRange: Range{
										From: Position{Index: 34, Line: 1, Col: 8},
										To:   Position{Index: 38, Line: 1, Col: 12},
									},
								},
							},
							&ConstantAttribute{
								Value: "a",
								Key: ConstantAttributeKey{
									Name: "value",
									NameRange: Range{
										From: Position{Index: 46, Line: 1, Col: 20},
										To:   Position{Index: 51, Line: 1, Col: 25},
									},
								},
							},
						},
						TrailingSpace: SpaceVertical,
						Range: Range{
							From: Position{Index: 27, Line: 1, Col: 1},
							To:   Position{Index: 60, Line: 2, Col: 1},
						},
					},
					&Element{
						Name: "input",
						NameRange: Range{
							From: Position{Index: 61, Line: 2, Col: 2},
							To:   Position{Index: 66, Line: 2, Col: 7},
						},
						Attributes: []Attribute{
							&ConstantAttribute{
								Value: "text",
								Key: ConstantAttributeKey{
									Name: "type",
									NameRange: Range{
										From: Position{Index: 67, Line: 2, Col: 8},
										To:   Position{Index: 71, Line: 2, Col: 12},
									},
								},
							},
							&ConstantAttribute{
								Value: "b",
								Key: ConstantAttributeKey{
									Name: "value",
									NameRange: Range{
										From: Position{Index: 79, Line: 2, Col: 20},
										To:   Position{Index: 84, Line: 2, Col: 25},
									},
								},
							},
						},
						TrailingSpace: SpaceVertical,
						Range: Range{
							From: Position{Index: 60, Line: 2, Col: 1},
							To:   Position{Index: 92, Line: 3, Col: 0},
						},
					},
				},
			},
		},
		{
			name: "template: doctype",
			input: `templ Name() {
<!DOCTYPE html>
}`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 32, Line: 2, Col: 1},
				},
				Expression: Expression{
					Value: "Name()",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  0,
							Col:   6,
						},
						To: Position{
							Index: 12,
							Line:  0,
							Col:   12,
						},
					},
				},
				Children: []Node{
					&DocType{
						Value: "html",
					},
					&Whitespace{Value: "\n"},
				},
			},
		},
		{
			name: "template: incomplete open tag",
			input: `templ Name() {
				        <div
						{"some string"}
					</div>
}`,
			expected:    &HTMLTemplate{},
			expectError: true,
		},
		{
			name: "template: can contain inline templ elements",
			input: `templ x() {
 <a href="/"> @Icon("home", Inline) Home</a>
}`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 58, Line: 2, Col: 1},
				},
				Expression: Expression{
					Value: "x()",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  0,
							Col:   6,
						},
						To: Position{
							Index: 9,
							Line:  0,
							Col:   9,
						},
					},
				},
				Children: []Node{
					&Whitespace{
						Value: " ",
					},
					&Element{
						Name: "a",
						NameRange: Range{
							From: Position{Index: 14, Line: 1, Col: 2},
							To:   Position{Index: 15, Line: 1, Col: 3},
						},
						Attributes: []Attribute{
							&ConstantAttribute{
								Value: "/",
								Key: ConstantAttributeKey{
									Name: "href",
									NameRange: Range{
										From: Position{Index: 16, Line: 1, Col: 4},
										To:   Position{Index: 20, Line: 1, Col: 8},
									},
								},
							},
						},
						Children: []Node{
							&Whitespace{Value: " "},
							&TemplElementExpression{
								Expression: Expression{
									Value: `Icon("home", Inline)`,
									Range: Range{
										From: Position{
											Index: 27,
											Line:  1,
											Col:   15,
										},
										To: Position{
											Index: 47,
											Line:  1,
											Col:   35,
										},
									},
								},
								Range: Range{
									From: Position{Index: 26, Line: 1, Col: 14},
									To:   Position{Index: 47, Line: 1, Col: 35},
								},
							},
							&Whitespace{Value: " "},
							&Text{
								Value: "Home",
								Range: Range{
									From: Position{Index: 48, Line: 1, Col: 36},
									To:   Position{Index: 52, Line: 1, Col: 40},
								},
							},
						},
						TrailingSpace: SpaceVertical,
						Range: Range{
							From: Position{Index: 13, Line: 1, Col: 1},
							To:   Position{Index: 57, Line: 2, Col: 0},
						},
					},
				},
			},
		},
		{
			name: "template: can contain single line comments",
			input: `templ x() {
	// Comment
}`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 25, Line: 2, Col: 1},
				},
				Expression: Expression{
					Value: "x()",
					Range: Range{
						From: Position{Index: 6, Line: 0, Col: 6},
						To:   Position{Index: 9, Line: 0, Col: 9},
					},
				},
				Children: []Node{
					&Whitespace{Value: "\t"},
					&GoComment{
						Contents:  " Comment",
						Multiline: false,
						Range: Range{
							From: Position{Index: 13, Line: 1, Col: 1},
							To:   Position{Index: 23, Line: 1, Col: 11},
						},
					},
					&Whitespace{Value: "\n"},
				},
			},
		},
		{
			name: "template: can contain block comments on the same line",
			input: `templ x() {
	/* Comment */
}`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 28, Line: 2, Col: 1},
				},
				Expression: Expression{
					Value: "x()",
					Range: Range{
						From: Position{Index: 6, Line: 0, Col: 6},
						To:   Position{Index: 9, Line: 0, Col: 9},
					},
				},
				Children: []Node{
					&Whitespace{Value: "\t"},
					&GoComment{
						Contents:  " Comment ",
						Multiline: true,
						Range: Range{
							From: Position{Index: 13, Line: 1, Col: 1},
							To:   Position{Index: 26, Line: 1, Col: 14},
						},
					},
					&Whitespace{Value: "\n"},
				},
			},
		},
		{
			name: "template: can contain block comments on multiple lines",
			input: `templ x() {
	/* Line 1
		 Line 2
	*/
}`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 38, Line: 4, Col: 1},
				},
				Expression: Expression{
					Value: "x()",
					Range: Range{
						From: Position{Index: 6, Line: 0, Col: 6},
						To:   Position{Index: 9, Line: 0, Col: 9},
					},
				},
				Children: []Node{
					&Whitespace{Value: "\t"},
					&GoComment{
						Contents:  " Line 1\n\t\t Line 2\n\t",
						Multiline: true,
						Range: Range{
							From: Position{Index: 13, Line: 1, Col: 1},
							To:   Position{Index: 36, Line: 3, Col: 3},
						},
					},
					&Whitespace{Value: "\n"},
				},
			},
		},
		{
			name: "template: can contain HTML comments",
			input: `templ x() {
	<!-- Single line -->
	<!--
		Multiline
	-->
}`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 58, Line: 5, Col: 1},
				},
				Expression: Expression{
					Value: "x()",
					Range: Range{
						From: Position{Index: 6, Line: 0, Col: 6},
						To:   Position{Index: 9, Line: 0, Col: 9},
					},
				},
				Children: []Node{
					&Whitespace{Value: "\t"},
					&HTMLComment{
						Contents: " Single line ",
						Range: Range{
							From: Position{Index: 13, Line: 1, Col: 1},
							To:   Position{Index: 33, Line: 1, Col: 21},
						},
					},
					&Whitespace{Value: "\n\t"},
					&HTMLComment{
						Contents: "\n\t\tMultiline\n\t",
						Range: Range{
							From: Position{Index: 35, Line: 2, Col: 1},
							To:   Position{Index: 56, Line: 4, Col: 4},
						},
					},
					&Whitespace{Value: "\n"},
				},
			},
		},
		{
			name: "template: containing spread attributes and children expression",
			input: `templ Name(children templ.Attributes) {
		<span { children... }>
			{ children... }
		</span>
}`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 95, Line: 4, Col: 1},
				},
				Expression: Expression{
					Value: "Name(children templ.Attributes)",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  0,
							Col:   6,
						},
						To: Position{
							Index: 37,
							Line:  0,
							Col:   37,
						},
					},
				},
				Children: []Node{
					&Whitespace{Value: "\t\t"},
					&Element{
						Name: "span",
						NameRange: Range{
							From: Position{Index: 43, Line: 1, Col: 3},
							To:   Position{Index: 47, Line: 1, Col: 7},
						},
						Attributes: []Attribute{&SpreadAttributes{
							Expression{
								Value: "children",
								Range: Range{
									From: Position{
										Index: 50,
										Line:  1,
										Col:   10,
									},
									To: Position{
										Index: 58,
										Line:  1,
										Col:   18,
									},
								},
							},
						}},
						Children: []Node{
							&Whitespace{"\n\t\t\t"},
							&ChildrenExpression{},
							&Whitespace{Value: "\n\t\t"},
						},
						IndentChildren: true,
						TrailingSpace:  SpaceVertical,
						Range: Range{
							From: Position{Index: 42, Line: 1, Col: 2},
							To:   Position{Index: 94, Line: 4, Col: 0},
						},
					},
				},
			},
		},
		{
			name: "template: void element closers are ignored",
			input: `templ Name() {
	<br></br><br>
}`,
			expected: &HTMLTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 31, Line: 2, Col: 1},
				},
				Expression: Expression{
					Value: "Name()",
					Range: Range{
						From: Position{Index: 6, Line: 0, Col: 6},
						To:   Position{Index: 12, Line: 0, Col: 12},
					},
				},
				Children: []Node{
					&Whitespace{Value: "\t"},
					&Element{
						Name: "br",
						NameRange: Range{
							From: Position{Index: 17, Line: 1, Col: 2},
							To:   Position{Index: 19, Line: 1, Col: 4},
						},
						TrailingSpace: SpaceNone,
						Range: Range{
							From: Position{Index: 16, Line: 1, Col: 1},
							To:   Position{Index: 25, Line: 1, Col: 10},
						},
					},
					&Element{
						Name: "br",
						NameRange: Range{
							From: Position{Index: 26, Line: 1, Col: 11},
							To:   Position{Index: 28, Line: 1, Col: 13},
						},
						TrailingSpace: SpaceVertical,
						Range: Range{
							From: Position{Index: 25, Line: 1, Col: 10},
							To:   Position{Index: 30, Line: 2, Col: 0},
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
			actual, matched, err := template.Parse(input)
			diff := cmp.Diff(tt.expected, actual)
			switch {
			case tt.expectError && err == nil:
				t.Errorf("expected an error got nil: %+v", actual)
			case !tt.expectError && err != nil:
				t.Errorf("unexpected error: %v", err)
			case !tt.expectError && diff != "":
				t.Error(diff)
			}
			if !matched {
				t.Error("expected match, but got no match")
			}
		})
	}
}

func TestTemplateParserErrors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "template: containing element",
			input: `templ Name(p Parameter) {
<span
}`,
			expected: "<span>: malformed open element: line 3, col 0",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			_, matched, err := template.Parse(input)
			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.expected)
			}
			if !matched {
				t.Error("expected match, because there is a partial template")
			}
			if diff := cmp.Diff(tt.expected, err.Error()); diff != "" {
				t.Error(diff)
			}
		})
	}
}
