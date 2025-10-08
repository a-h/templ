package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestSwitchExpressionParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *SwitchExpression
	}{
		{
			name: "switch: simple",
			input: `switch "stringy" {
}`,
			expected: &SwitchExpression{
				Expression: Expression{
					Value: `"stringy"`,
					Range: Range{
						From: Position{
							Index: 7,
							Line:  0,
							Col:   7,
						},
						To: Position{
							Index: 16,
							Line:  0,
							Col:   16,
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 20, Line: 1, Col: 1},
				},
			},
		},
		{
			name: "switch: default only",
			input: `switch "stringy" {
default:
	<span>
	  { "span content" }
	</span>
}`,
			expected: &SwitchExpression{
				Expression: Expression{
					Value: `"stringy"`,
					Range: Range{
						From: Position{
							Index: 7,
							Line:  0,
							Col:   7,
						},
						To: Position{
							Index: 16,
							Line:  0,
							Col:   16,
						},
					},
				},
				Cases: []CaseExpression{
					{
						Expression: Expression{
							Value: "default:",
							Range: Range{
								From: Position{
									Index: 19,
									Line:  1,
									Col:   0,
								},
								To: Position{
									Index: 27,
									Line:  1,
									Col:   8,
								},
							},
						},
						Children: []Node{
							&Whitespace{Value: "\t"},
							&Element{
								Name: "span",
								NameRange: Range{
									From: Position{Index: 30, Line: 2, Col: 2},
									To:   Position{Index: 34, Line: 2, Col: 6},
								},
								Children: []Node{
									&Whitespace{Value: "\n\t  "},
									&StringExpression{
										Expression: Expression{
											Value: `"span content"`,
											Range: Range{
												From: Position{
													Index: 41,
													Line:  3,
													Col:   5,
												},
												To: Position{
													Index: 55,
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
									From: Position{Index: 29, Line: 2, Col: 1},
									To:   Position{Index: 67, Line: 5, Col: 0},
								},
							},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 68, Line: 5, Col: 1},
				},
			},
		},
		{
			name: "switch: one case",
			input: `switch "stringy" {
	case "stringy":
<span>
  { "span content" }
</span>
}`,
			expected: &SwitchExpression{
				Expression: Expression{
					Value: `"stringy"`,
					Range: Range{
						From: Position{
							Index: 7,
							Line:  0,
							Col:   7,
						},
						To: Position{
							Index: 16,
							Line:  0,
							Col:   16,
						},
					},
				},
				Cases: []CaseExpression{
					{
						Expression: Expression{
							Value: "case \"stringy\":",
							Range: Range{
								From: Position{
									Index: 20,
									Line:  1,
									Col:   1,
								},
								To: Position{
									Index: 35,
									Line:  1,
									Col:   16,
								},
							},
						},
						Children: []Node{
							&Element{
								Name: "span",
								NameRange: Range{
									From: Position{Index: 37, Line: 2, Col: 1},
									To:   Position{Index: 41, Line: 2, Col: 5},
								},
								Children: []Node{
									&Whitespace{Value: "\n  "},
									&StringExpression{
										Expression: Expression{
											Value: `"span content"`,
											Range: Range{
												From: Position{
													Index: 47,
													Line:  3,
													Col:   4,
												},
												To: Position{
													Index: 61,
													Line:  3,
													Col:   18,
												},
											},
										},
										TrailingSpace: SpaceVertical,
									},
								},
								IndentChildren: true,
								TrailingSpace:  SpaceVertical,
								Range: Range{
									From: Position{Index: 36, Line: 2, Col: 0},
									To:   Position{Index: 72, Line: 5, Col: 0},
								},
							},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 73, Line: 5, Col: 1},
				},
			},
		},
		{
			name: "switch: two cases",
			input: `switch "stringy" {
	case "a":
		{ "A" }
	case "b":
		{ "B" }
}`,
			expected: &SwitchExpression{
				Expression: Expression{
					Value: `"stringy"`,
					Range: Range{
						From: Position{
							Index: 7,
							Line:  0,
							Col:   7,
						},
						To: Position{
							Index: 16,
							Line:  0,
							Col:   16,
						},
					},
				},
				Cases: []CaseExpression{
					{
						Expression: Expression{
							Value: "case \"a\":",
							Range: Range{
								From: Position{
									Index: 20,
									Line:  1,
									Col:   1,
								},
								To: Position{
									Index: 29,
									Line:  1,
									Col:   10,
								},
							},
						},
						Children: []Node{
							&Whitespace{
								Value: "\t\t",
							},
							&StringExpression{
								Expression: Expression{
									Value: `"A"`,
									Range: Range{
										From: Position{
											Index: 34,
											Line:  2,
											Col:   4,
										},
										To: Position{
											Index: 37,
											Line:  2,
											Col:   7,
										},
									},
								},
								TrailingSpace: SpaceVertical,
							},
						},
					},
					{
						Expression: Expression{
							Value: "case \"b\":",
							Range: Range{
								From: Position{
									Index: 41,
									Line:  3,
									Col:   1,
								},
								To: Position{
									Index: 50,
									Line:  3,
									Col:   10,
								},
							},
						},
						Children: []Node{
							&Whitespace{
								Value: "\t\t",
							},
							&StringExpression{
								Expression: Expression{
									Value: `"B"`,
									Range: Range{
										From: Position{
											Index: 55,
											Line:  4,
											Col:   4,
										},
										To: Position{
											Index: 58,
											Line:  4,
											Col:   7,
										},
									},
								},
								TrailingSpace: SpaceVertical,
							},
						},
					},
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 62, Line: 5, Col: 1},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			actual, matched, err := switchExpression.Parse(input)
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

func TestIncompleteSwitch(t *testing.T) {
	t.Run("no opening brace", func(t *testing.T) {
		input := parse.NewInput(`switch with no brace`)
		_, matched, err := switchExpression.Parse(input)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		if !matched {
			t.Fatal("expected a match, because we started with the text 'switch'")
		}
		pe, isParseError := err.(parse.ParseError)
		if !isParseError {
			t.Fatalf("expected a parse error, got %T", err)
		}
		if pe.Msg != "switch: "+unterminatedMissingCurly {
			t.Errorf("unexpected error: %v", err)
		}
		if pe.Pos.Line != 0 {
			t.Errorf("unexpected line: %d", pe.Pos.Line)
		}
	})
	t.Run("capitalised Switch", func(t *testing.T) {
		input := parse.NewInput(`Switch with no brace`)
		_, matched, err := switchExpression.Parse(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if matched {
			t.Fatal("expected a non match")
		}
	})
}
