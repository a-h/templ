package parser

import (
	"testing"

	"github.com/a-h/lexical/input"
	"github.com/google/go-cmp/cmp"
)

func TestTemplateParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected HTMLTemplate
	}{
		{
			name: "template: no parameters",
			input: `templ Name() {
}`,
			expected: HTMLTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  1,
							Col:   6,
						},
						To: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
					},
				},
				Parameters: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 11,
							Line:  1,
							Col:   11,
						},
						To: Position{
							Index: 11,
							Line:  1,
							Col:   11,
						},
					},
				},
				Children: []Node{},
			},
		},
		{
			name: "template: no spaces",
			input: `templ Name(){
}`,
			expected: HTMLTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  1,
							Col:   6,
						},
						To: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
					},
				},
				Parameters: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 11,
							Line:  1,
							Col:   11,
						},
						To: Position{
							Index: 11,
							Line:  1,
							Col:   11,
						},
					},
				},
				Children: []Node{},
			},
		},
		{
			name: "template: single parameter",
			input: `templ Name(p Parameter) {
}`,
			expected: HTMLTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  1,
							Col:   6,
						},
						To: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
					},
				},
				Parameters: Expression{
					Value: "p Parameter",
					Range: Range{
						From: Position{
							Index: 11,
							Line:  1,
							Col:   11,
						},
						To: Position{
							Index: 22,
							Line:  1,
							Col:   22,
						},
					},
				},
				Children: []Node{},
			},
		},
		{
			name: "template: containing element",
			input: `templ Name(p Parameter) {
<span>{ "span content" }</span>
}`,
			expected: HTMLTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  1,
							Col:   6,
						},
						To: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
					},
				},
				Parameters: Expression{
					Value: "p Parameter",
					Range: Range{
						From: Position{
							Index: 11,
							Line:  1,
							Col:   11,
						},
						To: Position{
							Index: 22,
							Line:  1,
							Col:   22,
						},
					},
				},
				Children: []Node{
					Element{
						Name:       "span",
						Attributes: []Attribute{},
						Children: []Node{
							StringExpression{
								Expression: Expression{
									Value: `"span content"`,
									Range: Range{
										From: Position{
											Index: 34,
											Line:  2,
											Col:   8,
										},
										To: Position{
											Index: 48,
											Line:  2,
											Col:   22,
										},
									},
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
			input: `templ Name(p Parameter) {
<div>
  { "div content" }
  <span>
	{ "span content" }
  </span>
</div>
}`,
			expected: HTMLTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  1,
							Col:   6,
						},
						To: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
					},
				},
				Parameters: Expression{
					Value: "p Parameter",
					Range: Range{
						From: Position{
							Index: 11,
							Line:  1,
							Col:   11,
						},
						To: Position{
							Index: 22,
							Line:  1,
							Col:   22,
						},
					},
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
									Range: Range{
										From: Position{
											Index: 36,
											Line:  3,
											Col:   4,
										},
										To: Position{
											Index: 49,
											Line:  3,
											Col:   17,
										},
									},
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
											Range: Range{
												From: Position{
													Index: 64,
													Line:  5,
													Col:   3,
												},
												To: Position{
													Index: 78,
													Line:  5,
													Col:   17,
												},
											},
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
			input: `templ Name(p Parameter) {
	if p.Test {
		<span>
			{ "span content" }
		</span>
	}
}`,
			expected: HTMLTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  1,
							Col:   6,
						},
						To: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
					},
				},
				Parameters: Expression{
					Value: "p Parameter",
					Range: Range{
						From: Position{
							Index: 11,
							Line:  1,
							Col:   11,
						},
						To: Position{
							Index: 22,
							Line:  1,
							Col:   22,
						},
					},
				},
				Children: []Node{
					Whitespace{Value: "\t"},
					IfExpression{
						Expression: Expression{
							Value: `if p.Test {`,
							Range: Range{
								From: Position{
									Index: 30,
									Line:  2,
									Col:   4,
								},
								To: Position{
									Index: 37,
									Line:  2,
									Col:   11,
								},
							},
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
											Range: Range{
												From: Position{
													Index: 53,
													Line:  4,
													Col:   5,
												},
												To: Position{
													Index: 67,
													Line:  4,
													Col:   19,
												},
											},
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
		{
			name: "template: inputs",
			input: `templ Name(p Parameter) {
	<input type="text" value="a" />
	<input type="text" value="b" />
}`,
			expected: HTMLTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  1,
							Col:   6,
						},
						To: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
					},
				},
				Parameters: Expression{
					Value: "p Parameter",
					Range: Range{
						From: Position{
							Index: 11,
							Line:  1,
							Col:   11,
						},
						To: Position{
							Index: 22,
							Line:  1,
							Col:   22,
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
					},
					Whitespace{Value: "\n\t"},
					Element{
						Name: "input",
						Attributes: []Attribute{
							ConstantAttribute{Name: "type", Value: "text"},
							ConstantAttribute{Name: "value", Value: "b"},
						},
					},
					Whitespace{Value: "\n"},
				},
			},
		},
		{
			name: "template: doctype",
			input: `templ Name() {
<!DOCTYPE html>
}`,
			expected: HTMLTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 6,
							Line:  1,
							Col:   6,
						},
						To: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
					},
				},
				Parameters: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 11,
							Line:  1,
							Col:   11,
						},
						To: Position{
							Index: 11,
							Line:  1,
							Col:   11,
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
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := input.NewFromString(tt.input)
			result := template.Parse(input)
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
