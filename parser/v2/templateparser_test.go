package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestTemplateParser(t *testing.T) {
	var tests = []struct {
		name        string
		input       string
		expected    HTMLTemplate
		expectError bool
	}{
		{
			name: "template: no parameters",
			input: `templ Name() {
}`,
			expected: HTMLTemplate{
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
			expected: HTMLTemplate{
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
			expected: HTMLTemplate{
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
			expected: HTMLTemplate{
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
			name: "template: containing element",
			input: `templ Name(p Parameter) {
<span>{ "span content" }</span>
}`,
			expected: HTMLTemplate{
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
					Element{
						Name: "span",
						Children: []Node{
							StringExpression{
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
			expected: HTMLTemplate{
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
					Element{
						Name: "div",
						Children: []Node{
							Whitespace{Value: "\n  "},
							StringExpression{
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
							Element{
								Name: "span",
								Children: []Node{
									Whitespace{Value: "\n\t"},
									StringExpression{
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
							},
						},
						IndentChildren: true,
						TrailingSpace:  SpaceVertical,
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
			expected: HTMLTemplate{
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
					Whitespace{Value: "\t"},
					IfExpression{
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
							Whitespace{Value: "\t\t"},
							Element{
								Name: "span",
								Children: []Node{
									Whitespace{"\n\t\t\t"},
									StringExpression{
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
			name: "template: inputs",
			input: `templ Name(p Parameter) {
	<input type="text" value="a" />
	<input type="text" value="b" />
}`,
			expected: HTMLTemplate{
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
					Whitespace{Value: "\t"},
					Element{
						Name: "input",
						Attributes: []Attribute{
							ConstantAttribute{Name: "type", Value: "text"},
							ConstantAttribute{Name: "value", Value: "a"},
						},
						TrailingSpace: SpaceVertical,
					},
					Element{
						Name: "input",
						Attributes: []Attribute{
							ConstantAttribute{Name: "type", Value: "text"},
							ConstantAttribute{Name: "value", Value: "b"},
						},
						TrailingSpace: SpaceVertical,
					},
				},
			},
		},
		{
			name: "template: doctype",
			input: `templ Name() {
<!DOCTYPE html>
}`,
			expected: HTMLTemplate{
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
					DocType{
						Value: "html",
					},
					Whitespace{Value: "\n"},
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
			expected:    HTMLTemplate{},
			expectError: true,
		},
		{
			name: "template: can contain inline templ elements",
			input: `templ x() {
 <a href="/"> @Icon("home", Inline) Home</a>
}`,
			expected: HTMLTemplate{
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
					Whitespace{
						Value: " ",
					},
					Element{
						Name: "a",
						Attributes: []Attribute{
							ConstantAttribute{
								Name:  "href",
								Value: "/",
							},
						},
						Children: []Node{
							Whitespace{Value: " "},
							TemplElementExpression{
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
							},
							Whitespace{Value: " "},
							Text{Value: "Home"},
						},
						TrailingSpace: SpaceVertical,
					},
				},
			},
		},
		{
			name: "template: can contain single line comments",
			input: `templ x() {
	// Comment
}`,
			expected: HTMLTemplate{
				Expression: Expression{
					Value: "x()",
					Range: Range{
						From: Position{Index: 6, Line: 0, Col: 6},
						To:   Position{Index: 9, Line: 0, Col: 9},
					},
				},
				Children: []Node{
					Whitespace{Value: "\t"},
					GoComment{Contents: " Comment", Multiline: false},
				},
			},
		},
		{
			name: "template: can contain block comments on the same line",
			input: `templ x() {
	/* Comment */
}`,
			expected: HTMLTemplate{
				Expression: Expression{
					Value: "x()",
					Range: Range{
						From: Position{Index: 6, Line: 0, Col: 6},
						To:   Position{Index: 9, Line: 0, Col: 9},
					},
				},
				Children: []Node{
					Whitespace{Value: "\t"},
					GoComment{Contents: " Comment ", Multiline: true},
					Whitespace{Value: "\n"},
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
			expected: HTMLTemplate{
				Expression: Expression{
					Value: "x()",
					Range: Range{
						From: Position{Index: 6, Line: 0, Col: 6},
						To:   Position{Index: 9, Line: 0, Col: 9},
					},
				},
				Children: []Node{
					Whitespace{Value: "\t"},
					GoComment{Contents: " Line 1\n\t\t Line 2\n\t", Multiline: true},
					Whitespace{Value: "\n"},
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
			expected: HTMLTemplate{
				Expression: Expression{
					Value: "x()",
					Range: Range{
						From: Position{Index: 6, Line: 0, Col: 6},
						To:   Position{Index: 9, Line: 0, Col: 9},
					},
				},
				Children: []Node{
					Whitespace{Value: "\t"},
					HTMLComment{Contents: " Single line "},
					Whitespace{Value: "\n\t"},
					HTMLComment{Contents: " \n\t\tMultiline\n\t"},
					Whitespace{Value: "\n"},
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
			expected: HTMLTemplate{
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
					Whitespace{Value: "\t\t"},
					Element{
						Name: "span",
						Attributes: []Attribute{SpreadAttributes{
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
							Whitespace{"\n\t\t\t"},
							ChildrenExpression{},
							Whitespace{Value: "\n\t\t"},
						},
						IndentChildren: true,
						TrailingSpace:  SpaceVertical,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			actual, ok, err := template.Parse(input)
			diff := cmp.Diff(tt.expected, actual)
			switch {
			case tt.expectError && err == nil:
				t.Errorf("expected an error got nil: %+v", actual)
			case !tt.expectError && err != nil:
				t.Errorf("unexpected error: %v", err)
			case tt.expectError && ok:
				t.Errorf("Success=%v want=%v", ok, !tt.expectError)
			case !tt.expectError && diff != "":
				t.Errorf(diff)
			}
		})
	}
}

func TestTemplateParserErrors(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "template: containing element",
			input: `templ Name(p Parameter) {
<span
}`,
			expected: "<span>: malformed open element: line 2, col 0",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			_, ok, err := template.Parse(input)
			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.expected)
			}
			if ok {
				t.Error("expected failure, but got success")
			}
			if diff := cmp.Diff(tt.expected, err.Error()); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
