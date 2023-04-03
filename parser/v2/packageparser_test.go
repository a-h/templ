package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestPackageParserErrors(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected parse.ParseError
	}{
		{
			name:  "unterminated package",
			input: "package ",
			expected: parse.Error(
				"package literal not terminated",
				parse.Position{
					Index: 8,
					Line:  0,
					Col:   8,
				},
			),
		},
		{
			name:  "unterminated package, new line",
			input: "package \n",
			expected: parse.Error(
				"package literal not terminated",
				parse.Position{
					Index: 0,
					Line:  0,
					Col:   0,
				},
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			pi := parse.NewInput(tt.input)
			_, ok, err := pkg.Parse(pi)
			if ok {
				t.Errorf("expected parsing to fail, but it succeeded")
			}
			if diff := cmp.Diff(tt.expected, err); diff != "" {
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
			input: "package parser\n",
			expected: Package{
				Expression: Expression{
					Value: "package parser",
					Range: Range{
						From: Position{
							Index: 0,
							Line:  0,
							Col:   0,
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
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			actual, ok, err := pkg.Parse(input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !ok {
				t.Fatalf("unexpected failure for input %q", tt.input)
			}
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Error(diff)
			}
		})
	}
}
