package parser

import (
	"io"
	"testing"

	"github.com/a-h/lexical/input"
	"github.com/google/go-cmp/cmp"
)

func TestPackageParserErrors(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected ParseError
	}{
		{
			name:  "unterminated package",
			input: "{% package ",
			expected: newParseError(
				"package literal not terminated",
				Position{
					Index: 11,
					Line:  1,
					Col:   11,
				},
				Position{
					Index: 12,
					Line:  1,
					Col:   11,
				},
			),
		},
		{
			name:  "unterminated package, new line",
			input: "{% package \n%}",
			expected: newParseError(
				"package literal not terminated",
				Position{
					Index: 11,
					Line:  1,
					Col:   11,
				},
				Position{
					Index: 11,
					Line:  1,
					Col:   11,
				},
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			pi := input.NewFromString(tt.input)
			actual := newPackageParser().Parse(pi)
			if actual.Success {
				t.Errorf("expected parsing to fail, but it succeeded")
			}
			if diff := cmp.Diff(tt.expected, actual.Error); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestPackageParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name:  "package: standard",
			input: `{% package parser %}`,
			expected: Package{
				Expression: Expression{
					Value: "parser",
					Range: Range{
						From: Position{
							Index: 11,
							Line:  1,
							Col:   11,
						},
						To: Position{
							Index: 17,
							Line:  1,
							Col:   17,
						},
					},
				},
			},
		},
		{
			name:  "package: no spaces",
			input: `{%package parser%}`,
			expected: Package{
				Expression: Expression{
					Value: "parser",
					Range: Range{
						From: Position{
							Index: 10,
							Line:  1,
							Col:   10,
						},
						To: Position{
							Index: 16,
							Line:  1,
							Col:   16,
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
			result := newPackageParser().Parse(input)
			if result.Error != nil && result.Error != io.EOF {
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
