package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/a-h/templ/cfg"
	"github.com/google/go-cmp/cmp"
)

func TestGoCodeParser(t *testing.T) {
	flagVal := cfg.Experiment.RawGo
	cfg.Experiment.RawGo = true
	defer func() {
		cfg.Experiment.RawGo = flagVal
	}()

	tests := []struct {
		name     string
		input    string
		expected GoCode
	}{
		{
			name:  "basic expression",
			input: `{{ p := "this" }}`,
			expected: GoCode{
				Expression: Expression{
					Value: `p := "this"`,
					Range: Range{
						From: Position{
							Index: 3,
							Line:  0,
							Col:   3,
						},
						To: Position{
							Index: 14,
							Line:  0,
							Col:   14,
						},
					},
				},
			},
		},
		{
			name:  "basic expression, no space",
			input: `{{p:="this"}}`,
			expected: GoCode{
				Expression: Expression{
					Value: `p:="this"`,
					Range: Range{
						From: Position{
							Index: 2,
							Line:  0,
							Col:   2,
						},
						To: Position{
							Index: 11,
							Line:  0,
							Col:   11,
						},
					},
				},
			},
		},
		{
			name: "multiline function decl",
			input: `{{
				p := func() {
					dosomething()
				}
			}}`,
			expected: GoCode{
				Expression: Expression{
					Value: `
				p := func() {
					dosomething()
				}`,
					Range: Range{
						From: Position{
							Index: 2,
							Line:  0,
							Col:   2,
						},
						To: Position{
							Index: 45,
							Line:  3,
							Col:   5,
						},
					},
				},
				Multiline: true,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			an, ok, err := goCode.Parse(input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !ok {
				t.Fatalf("unexpected failure for input %q", tt.input)
			}
			actual := an.(GoCode)
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Error(diff)
			}

			// Check the index.
			cut := tt.input[actual.Expression.Range.From.Index:actual.Expression.Range.To.Index]
			if tt.expected.Expression.Value != cut {
				t.Errorf("range, expected %q, got %q", tt.expected.Expression.Value, cut)
			}
		})
	}
}
