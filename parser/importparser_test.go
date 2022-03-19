package parser

import (
	"testing"

	"github.com/a-h/lexical/input"
	"github.com/google/go-cmp/cmp"
)

func TestImportParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name:  "import: named",
			input: `{% import name "github.com/a-h/something" %}`,
			expected: Import{
				Expression: Expression{
					Value: `name "github.com/a-h/something"`,
					Range: Range{
						From: Position{
							Index: 10,
							Line:  1,
							Col:   10,
						},
						To: Position{
							Index: 41,
							Line:  1,
							Col:   41,
						},
					},
				},
			},
		},
		{
			name:  "import: default",
			input: `{% import "github.com/a-h/something" %}`,
			expected: Import{
				Expression: Expression{
					Value: `"github.com/a-h/something"`,
					Range: Range{
						From: Position{
							Index: 10,
							Line:  1,
							Col:   10,
						},
						To: Position{
							Index: 36,
							Line:  1,
							Col:   36,
						},
					},
				},
			},
		},
		{
			name:  "import: no spaces",
			input: `{%import "github.com/a-h/something"%}`,
			expected: Import{
				Expression: Expression{
					Value: `"github.com/a-h/something"`,
					Range: Range{
						From: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
						To: Position{
							Index: 35,
							Line:  1,
							Col:   35,
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
			result := newImportParser().Parse(input)
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
