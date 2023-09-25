package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestCommentParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected Comment
	}{
		{
			name:  "comment - single line",
			input: `<!-- single line comment -->`,
			expected: Comment{
				Contents: " single line comment ",
			},
		},
		{
			name:  "comment - no whitespace",
			input: `<!--no whitespace between sequence open and close-->`,
			expected: Comment{
				Contents: "no whitespace between sequence open and close",
			},
		},
		{
			name: "comment - multiline",
			input: `<!-- multiline
								comment
					-->`,
			expected: Comment{
				Contents: ` multiline
								comment
					`,
			},
		},
		{
			name:  "comment - with tag",
			input: `<!-- <p class="test">tag</p> -->`,
			expected: Comment{
				Contents: ` <p class="test">tag</p> `,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			result, ok, err := htmlComment.Parse(input)
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

func TestCommentParserErrors(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:  "unclosed HTML comment",
			input: `<!-- unclosed HTML comment`,
			expected: parse.Error("expected end comment sequence not present",
				parse.Position{
					Index: 26,
					Line:  0,
					Col:   26,
				}),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			_, _, err := htmlComment.Parse(input)
			if diff := cmp.Diff(tt.expected, err); diff != "" {
				t.Error(diff)
			}
		})
	}
}
