package parser

import (
	"bytes"
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestForExpressionParser(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expected     interface{}
		expectedHTML string
	}{
		{
			name: "for: simple",
			input: `for _, item := range p.Items {
					<div>{ item }</div>
				}`,
			expectedHTML: `                              
	<div>        </div>
 `,
			expected: ForExpression{
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
					Whitespace{Value: "\t\t\t\t\t"},
					Element{
						Name: "div",
						NameRange: Range{
							From: Position{Index: 37, Line: 1, Col: 6},
							To:   Position{Index: 40, Line: 1, Col: 9},
						},
						Children: []Node{
							StringExpression{
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
					},
				},
			},
		},
		{
			name: "for: simple, without spaces",
			input: `for _, item := range p.Items{
					<div>{ item }</div>
				}`,
			expectedHTML: `                              
	<div>        </div>
 `,
			expected: ForExpression{
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
					Whitespace{Value: "\t\t\t\t\t"},
					Element{
						Name: "div",
						NameRange: Range{
							From: Position{Index: 36, Line: 1, Col: 6},
							To:   Position{Index: 39, Line: 1, Col: 9},
						},
						Children: []Node{
							StringExpression{
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
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			actual, ok, err := forExpression.Parse(input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !ok {
				t.Fatalf("unexpected failure for input %q", tt.input)
			}
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Error(diff)
			}

			w := new(bytes.Buffer)
			cw := NewContextWriter(w, WriteContextHTML)
			if err := actual.Write(cw, 0); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			actualHTML := w.String()
			if diff := cmp.Diff(tt.expectedHTML, actualHTML); diff != "" {
				t.Error(diff)

				t.Errorf("input:\n%s", displayWhitespaceChars(tt.input))
				t.Errorf("expected:\n%s", displayWhitespaceChars(tt.expectedHTML))
				t.Errorf("got:\n%s", displayWhitespaceChars(actualHTML))
			}
			if diff := cmp.Diff(getLineLengths(tt.input), getLineLengths(tt.expectedHTML)); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestIncompleteFor(t *testing.T) {
	t.Run("no opening brace", func(t *testing.T) {
		input := parse.NewInput(`for with no brace`)
		_, _, err := forExpression.Parse(input)
		if err.Error() != "for: unterminated (missing closing '{\\n') - https://templ.guide/syntax-and-usage/statements#incomplete-statements: line 0, col 0" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("capitalised For", func(t *testing.T) {
		input := parse.NewInput(`For with no brace`)
		_, ok, err := forExpression.Parse(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatal("expected a non match")
		}
	})
}
