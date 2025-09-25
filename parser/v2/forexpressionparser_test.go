package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestForExpressionParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{
			name: "for: infinite",
			input: `for {
					Ever
			}`,
			expected: &ForExpression{
				Expression: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 4,
							Line:  0,
							Col:   4,
						},
						To: Position{
							Index: 4,
							Line:  0,
							Col:   4,
						},
					},
				},
				Children: []Node{
					&Whitespace{Value: "\t\t\t\t\t"},
					&Text{
						Range: Range{
							From: Position{Index: 11, Line: 1, Col: 5},
							To:   Position{Index: 15, Line: 1, Col: 9},
						},
						Value:         "Ever",
						TrailingSpace: "\n",
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 20, Line: 2, Col: 4},
				},
			},
		},
		{
			name: "for: three clause",
			input: `for i := 0; i < 10; i ++ {
					Ever
			}`,
			expected: &ForExpression{
				Expression: Expression{
					Value: "i := 0; i < 10; i ++",
					Range: Range{
						From: Position{
							Index: 4,
							Line:  0,
							Col:   4,
						},
						To: Position{
							Index: 24,
							Line:  0,
							Col:   24,
						},
					},
				},
				Children: []Node{
					&Whitespace{Value: "\t\t\t\t\t"},
					&Text{
						Range: Range{
							From: Position{Index: 32, Line: 1, Col: 5},
							To:   Position{Index: 36, Line: 1, Col: 9},
						},
						Value:         "Ever",
						TrailingSpace: "\n",
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 41, Line: 2, Col: 4},
				},
			},
		},
		{
			name: "for: use existing variables",
			input: `for x, y = range []int{1, 2} {
					Ever
			}`,
			expected: &ForExpression{
				Expression: Expression{
					Value: "x, y = range []int{1, 2}",
					Range: Range{
						From: Position{
							Index: 4,
							Line:  0,
							Col:   4,
						},
						To: Position{
							Index: 28,
							Line:  0,
							Col:   28,
						},
					},
				},
				Children: []Node{
					&Whitespace{Value: "\t\t\t\t\t"},
					&Text{
						Range: Range{
							From: Position{Index: 36, Line: 1, Col: 5},
							To:   Position{Index: 40, Line: 1, Col: 9},
						},
						Value:         "Ever",
						TrailingSpace: "\n",
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 45, Line: 2, Col: 4},
				},
			},
		},
		{
			name: "for: empty first variable",
			input: `for _, item := range p.Items {
					<div>{ item }</div>
				}`,
			expected: &ForExpression{
				Expression: Expression{
					Value: `_, item := range p.Items`,
					Range: Range{
						From: Position{
							Index: 4,
							Line:  0,
							Col:   4,
						},
						To: Position{
							Index: 28,
							Line:  0,
							Col:   28,
						},
					},
				},
				Children: []Node{
					&Whitespace{Value: "\t\t\t\t\t"},
					&Element{
						Name: "div",
						NameRange: Range{
							From: Position{Index: 37, Line: 1, Col: 6},
							To:   Position{Index: 40, Line: 1, Col: 9},
						},
						Children: []Node{
							&StringExpression{
								Expression: Expression{
									Value: `item`,
									Range: Range{
										From: Position{
											Index: 43,
											Line:  1,
											Col:   12,
										},
										To: Position{
											Index: 47,
											Line:  1,
											Col:   16,
										},
									},
								},
							},
						},
						TrailingSpace: SpaceVertical,
						Range: Range{
							From: Position{Index: 36, Line: 1, Col: 5},
							To:   Position{Index: 60, Line: 2, Col: 4},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 61, Line: 2, Col: 5},
				},
			},
		},
		{
			name: "for: empty first variable, without spaces",
			input: `for _, item := range p.Items{
					<div>{ item }</div>
				}`,
			expected: &ForExpression{
				Expression: Expression{
					Value: `_, item := range p.Items`,
					Range: Range{
						From: Position{
							Index: 4,
							Line:  0,
							Col:   4,
						},
						To: Position{
							Index: 28,
							Line:  0,
							Col:   28,
						},
					},
				},
				Children: []Node{
					&Whitespace{Value: "\t\t\t\t\t"},
					&Element{
						Name: "div",
						NameRange: Range{
							From: Position{Index: 36, Line: 1, Col: 6},
							To:   Position{Index: 39, Line: 1, Col: 9},
						},
						Children: []Node{
							&StringExpression{
								Expression: Expression{
									Value: `item`,
									Range: Range{
										From: Position{
											Index: 42,
											Line:  1,
											Col:   12,
										},
										To: Position{
											Index: 46,
											Line:  1,
											Col:   16,
										},
									},
								},
							},
						},
						TrailingSpace: SpaceVertical,
						Range: Range{
							From: Position{Index: 35, Line: 1, Col: 5},
							To:   Position{Index: 59, Line: 2, Col: 4},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 60, Line: 2, Col: 5},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			actual, matched, err := forExpression.Parse(input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !matched {
				t.Fatalf("unexpected failure for input %q", tt.input)
			}
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestForExpressionParserNegatives(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "for: wrong case",
			input: `For `, // not a for expression
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			_, matched, err := forExpression.Parse(input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if matched {
				t.Fatal("unexpected success")
			}
		})
	}
}

func TestForExpressionParserIncomplete(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "for: just writing normal text",
			input: `for this is`, // not a for expression, use <span>for</span> or `{ "for" }` to escape.
		},
		{
			name:  "for: numbers are not variables",
			input: `for 1`,
		},
		{
			name:  "for: $ is not a variable",
			input: `for $`,
		},
		{
			name:  "for: bare",
			input: `for `,
		},
		{
			name:  "for: infinite",
			input: `for {`,
		},
		{
			name:  "for: variable i started",
			input: `for i`,
		},
		{
			name:  "for: variable _ started",
			input: `for _`,
		},
		{
			name:  "for: variable _ ended with comma",
			input: `for _,`,
		},
		{
			name:  "for: variable asd ended with comma",
			input: `for asd,`,
		},
		{
			name:  "for: three clause starting",
			input: `for i `,
		},
		{
			name:  "for: three clause starting with assignment",
			input: `for i :=`,
		},
		{
			name:  "for: three clause starting with initial value",
			input: `for i := 0;`,
		},
		{
			name:  "for: boolean expression",
			input: `for i <`, // 10
		},
		{
			name:  "for: range",
			input: `for i := range`,
		},
		{
			name:  "for: k, v",
			input: `for k, v`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			_, matched, err := forExpression.Parse(input)
			if err == nil {
				t.Fatal("partial for should not be parsed successfully, but got nil")
			}
			if !matched {
				t.Fatal("expected to be detected as a for loop, but wasn't")
			}
		})
	}
}
