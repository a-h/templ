package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestExpressionCSSPropertyParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ExpressionCSSProperty
	}{
		{
			name:  "css: single constant property",
			input: `background-color: { constants.BackgroundColor };`,
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
			name:  "css: single constant property with windows newlines",
			input: "background-color:\r\n{ constants.BackgroundColor };\r\n",
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
		})
	}
}

func TestConstantCSSPropertyParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ConstantCSSProperty
	}{
		{
			name:  "css: single constant property",
			input: `background-color: #ffffff;`,
			expected: ConstantCSSProperty{
				Name:  "background-color",
				Value: "#ffffff",
			},
		},
		{
			name:  "css: single constant webkit property",
			input: `-webkit-text-stroke-color: #ffffff;`,
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
		})
	}
}

func TestCSSParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected CSSTemplate
	}{
		{
			name: "css: no parameters, no content",
			input: `css Name() {
}`,
			expected: CSSTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 4,
							Line:  0,
							Col:   4,
						},
						To: Position{
							Index: 8,
							Line:  0,
							Col:   8,
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
			expected: CSSTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 4,
							Line:  0,
							Col:   4,
						},
						To: Position{
							Index: 8,
							Line:  0,
							Col:   8,
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
			expected: CSSTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 4,
							Line:  0,
							Col:   4,
						},
						To: Position{
							Index: 8,
							Line:  0,
							Col:   8,
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
			expected: CSSTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 4,
							Line:  0,
							Col:   4,
						},
						To: Position{
							Index: 8,
							Line:  0,
							Col:   8,
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
		})
	}
}
