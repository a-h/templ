package parser

import (
	"testing"

	"github.com/a-h/lexical/input"
	"github.com/google/go-cmp/cmp"
)

func TestExpressionCSSPropertyParser(t *testing.T) {
	var tests = []struct {
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
								Line:  1,
								Col:   20,
							},
							To: Position{
								Index: 45,
								Line:  1,
								Col:   45,
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
			input := input.NewFromString(tt.input + "\n")
			result := newExpressionCSSPropertyParser().Parse(input)
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

func TestConstantCSSPropertyParser(t *testing.T) {
	var tests = []struct {
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
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := input.NewFromString(tt.input + "\n")
			result := newConstantCSSPropertyParser().Parse(input)
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

func TestCSSParser(t *testing.T) {
	var tests = []struct {
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
							Line:  1,
							Col:   4,
						},
						To: Position{
							Index: 8,
							Line:  1,
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
							Line:  1,
							Col:   4,
						},
						To: Position{
							Index: 8,
							Line:  1,
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
							Line:  1,
							Col:   4,
						},
						To: Position{
							Index: 8,
							Line:  1,
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
							Line:  1,
							Col:   4,
						},
						To: Position{
							Index: 8,
							Line:  1,
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
										Line:  2,
										Col:   20,
									},
									To: Position{
										Index: 58,
										Line:  2,
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
			input := input.NewFromString(tt.input)
			result := newCSSParser().Parse(input)
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
