package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestGoCommentParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected *GoComment
	}{
		{
			name: "single line can have a newline at the end",
			input: `// single line comment
`,
			expected: &GoComment{
				Contents:  " single line comment",
				Multiline: false,
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 22, Line: 0, Col: 22},
				},
			},
		},
		{
			name:  "single line comments can terminate the file",
			input: `// single line comment`,
			expected: &GoComment{
				Contents:  " single line comment",
				Multiline: false,
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 22, Line: 0, Col: 22},
				},
			},
		},
		{
			name:  "multiline comments can be on one line",
			input: `/* multiline comment, on one line */`,
			expected: &GoComment{
				Contents:  " multiline comment, on one line ",
				Multiline: true,
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 36, Line: 0, Col: 36},
				},
			},
		},
		{
			name: "multiline comments can span lines",
			input: `/* multiline comment,
on multiple lines */`,
			expected: &GoComment{
				Contents:  " multiline comment,\non multiple lines ",
				Multiline: true,
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 42, Line: 1, Col: 20},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			result, ok, err := goComment.Parse(input)
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}
			if !ok {
				t.Fatalf("failed to parse at %d", input.Index())
			}
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestCommentParserErrors(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:  "unclosed multi-line Go comments result in an error",
			input: `/* unclosed Go comment`,
			expected: parse.Error("expected end comment literal '*/' not found",
				parse.Position{
					Index: 0,
					Line:  0,
					Col:   0,
				}),
		},
		{
			name:     "single-line Go comment with no newline is allowed",
			input:    `// Comment with no newline`,
			expected: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			_, _, err := goComment.Parse(input)
			if diff := cmp.Diff(tt.expected, err); diff != "" {
				t.Error(diff)
			}
		})
	}
}
