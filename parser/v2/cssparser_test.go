package parser

import (
	"bytes"
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestExpressionCSSPropertyParser(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    ExpressionCSSProperty
		expectedCSS string
	}{
		{
			name:        "css: single constant property",
			input:       `background-color: { constants.BackgroundColor };`,
			expectedCSS: "background-color: '                           ';\n",
			expected: ExpressionCSSProperty{
				Name: "background-color",
				Value: StringExpression{
					Expression: Expression{
						Value: "constants.BackgroundColor",
						Range: Range{
							From: Position{
								Index: 20,
								Line:  0,
								Col:   20,
							},
							To: Position{
								Index: 45,
								Line:  0,
								Col:   45,
							},
						},
					},
				},
			},
		},
		{
			name:        "css: single constant property with windows newlines",
			input:       "background-color:\r\n{ constants.BackgroundColor };\r\n",
			expectedCSS: "background-color: '                           ';\n",
			expected: ExpressionCSSProperty{
				Name: "background-color",
				Value: StringExpression{
					Expression: Expression{
						Value: "constants.BackgroundColor",
						Range: Range{
							From: Position{
								Index: 21,
								Line:  1,
								Col:   2,
							},
							To: Position{
								Index: 46,
								Line:  1,
								Col:   27,
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
			input := parse.NewInput(tt.input + "\n")
			result, ok, err := expressionCSSPropertyParser.Parse(input)
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}
			if !ok {
				t.Fatalf("failed to parse at %d", input.Index())
			}
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf(diff)
			}

			w := new(bytes.Buffer)
			cw := NewContextWriter(w, WriteContextCSS)
			if err := result.Write(cw, 0); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			actualCSS := w.String()
			if diff := cmp.Diff(tt.expectedCSS, actualCSS); diff != "" {
				t.Error(diff)

				t.Errorf("input:\n%s", displayWhitespaceChars(tt.input))
				t.Errorf("expected:\n%s", displayWhitespaceChars(tt.expectedCSS))
				t.Errorf("got:\n%s", displayWhitespaceChars(actualCSS))
			}
			if diff := cmp.Diff(getLineLengths(tt.input), getLineLengths(actualCSS)); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestConstantCSSPropertyParser(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    ConstantCSSProperty
		expectedCSS string
	}{
		{
			name:        "css: single constant property",
			input:       `background-color: #ffffff;`,
			expectedCSS: "background-color: #ffffff;\n",
			expected: ConstantCSSProperty{
				Name:  "background-color",
				Value: "#ffffff",
			},
		},
		{
			name:        "css: single constant webkit property",
			input:       `-webkit-text-stroke-color: #ffffff;`,
			expectedCSS: "-webkit-text-stroke-color: #ffffff;\n",
			expected: ConstantCSSProperty{
				Name:  "-webkit-text-stroke-color",
				Value: "#ffffff",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input + "\n")
			result, ok, err := constantCSSPropertyParser.Parse(input)
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}
			if !ok {
				t.Fatalf("failed to parse at %d", input.Index())
			}
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf(diff)
			}

			w := new(bytes.Buffer)
			cw := NewContextWriter(w, WriteContextCSS)
			if err := result.Write(cw, 0); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			actualCSS := w.String()
			if diff := cmp.Diff(tt.expectedCSS, actualCSS); diff != "" {
				t.Error(diff)

				t.Errorf("input:\n%s", displayWhitespaceChars(tt.input))
				t.Errorf("expected:\n%s", displayWhitespaceChars(tt.expectedCSS))
				t.Errorf("got:\n%s", displayWhitespaceChars(actualCSS))
			}
			if diff := cmp.Diff(getLineLengths(tt.input), getLineLengths(actualCSS)); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestCSSParser(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    CSSTemplate
		expectedCSS string
	}{
		{
			name: "css: no parameters, no content",
			input: `css Name() {
}`,
			expectedCSS: `            
 `,
			expected: CSSTemplate{
				Name: "Name",
				Expression: Expression{
					Value: "Name()",
					Range: Range{
						From: Position{
							Index: 4,
							Line:  0,
							Col:   4,
						},
						To: Position{
							Index: 10,
							Line:  0,
							Col:   10,
						},
					},
				},
				Properties: []CSSProperty{},
			},
		},
		{
			name: "css: without spaces",
			input: `css Name() {
}`,
			expectedCSS: `            
 `,
			expected: CSSTemplate{
				Name: "Name",
				Expression: Expression{
					Value: "Name()",
					Range: Range{
						From: Position{
							Index: 4,
							Line:  0,
							Col:   4,
						},
						To: Position{
							Index: 10,
							Line:  0,
							Col:   10,
						},
					},
				},
				Properties: []CSSProperty{},
			},
		},
		{
			name: "css: single constant property",
			input: `css Name() {
background-color: #ffffff;
}`,
			expectedCSS: `            
	background-color: #ffffff;
 `,
			expected: CSSTemplate{
				Name: "Name",
				Expression: Expression{
					Value: "Name()",
					Range: Range{
						From: Position{
							Index: 4,
							Line:  0,
							Col:   4,
						},
						To: Position{
							Index: 10,
							Line:  0,
							Col:   10,
						},
					},
				},
				Properties: []CSSProperty{
					ConstantCSSProperty{
						Name:  "background-color",
						Value: "#ffffff",
					},
				},
			},
		},
		{
			name: "css: single expression property",
			input: `css Name() {
background-color: { constants.BackgroundColor };
}`,
			expectedCSS: `            
	background-color: '                           ';
 `,
			expected: CSSTemplate{
				Name: "Name",
				Expression: Expression{
					Value: "Name()",
					Range: Range{
						From: Position{
							Index: 4,
							Line:  0,
							Col:   4,
						},
						To: Position{
							Index: 10,
							Line:  0,
							Col:   10,
						},
					},
				},
				Properties: []CSSProperty{
					ExpressionCSSProperty{
						Name: "background-color",
						Value: StringExpression{
							Expression: Expression{
								Value: "constants.BackgroundColor",
								Range: Range{
									From: Position{
										Index: 33,
										Line:  1,
										Col:   20,
									},
									To: Position{
										Index: 58,
										Line:  1,
										Col:   45,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "css: single expression with parameter",
			input: `css Name(prop string) {
background-color: { prop };
}`,
			expectedCSS: `                       
	background-color: '      ';
 `,
			expected: CSSTemplate{
				Name: "Name",
				Expression: Expression{
					Value: "Name(prop string)",
					Range: Range{
						From: Position{
							Index: 4,
							Line:  0,
							Col:   4,
						},
						To: Position{
							Index: 21,
							Line:  0,
							Col:   21,
						},
					},
				},
				Properties: []CSSProperty{
					ExpressionCSSProperty{
						Name: "background-color",
						Value: StringExpression{
							Expression: Expression{
								Value: "prop",
								Range: Range{
									From: Position{
										Index: 44,
										Line:  1,
										Col:   20,
									},
									To: Position{
										Index: 48,
										Line:  1,
										Col:   24,
									},
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
			result, ok, err := cssParser.Parse(input)
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}
			if !ok {
				t.Fatalf("failed to parse at %d", input.Index())
			}
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf(diff)
			}

			w := new(bytes.Buffer)
			cw := NewContextWriter(w, WriteContextCSS)
			if err := result.Write(cw, 0); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			actualCSS := w.String()
			if diff := cmp.Diff(tt.expectedCSS, actualCSS); diff != "" {
				t.Error(diff)

				t.Errorf("input:\n%s", displayWhitespaceChars(tt.input))
				t.Errorf("expected:\n%s", displayWhitespaceChars(tt.expectedCSS))
				t.Errorf("got:\n%s", displayWhitespaceChars(actualCSS))
			}
			if diff := cmp.Diff(getLineLengths(tt.input), getLineLengths(actualCSS)); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
