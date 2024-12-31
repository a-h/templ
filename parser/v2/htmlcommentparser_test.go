package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestHTMLCommentParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected HTMLComment
	}{
		{
			name:  "comment - single line",
			input: `<!-- single line comment -->`,
			expected: HTMLComment{
				Contents: " single line comment ",
			},
		},
		{
			name:  "comment - no whitespace",
			input: `<!--no whitespace between sequence open and close-->`,
			expected: HTMLComment{
				Contents: "no whitespace between sequence open and close",
			},
		},
		{
			name: "comment - multiline",
			input: `<!-- multiline
								comment
					-->`,
			expected: HTMLComment{
				Contents: ` multiline
								comment
					`,
			},
		},
		{
			name:  "comment - with tag",
			input: `<!-- <p class="test">tag</p> -->`,
			expected: HTMLComment{
				Contents: ` <p class="test">tag</p> `,
			},
		},
		{
			name:  "comments can contain tags",
			input: `<!-- <div> hello world </div> -->`,
			expected: HTMLComment{
				Contents: ` <div> hello world </div> `,
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
				t.Error(diff)
			}
		})
	}
}

func TestHTMLCommentParserErrors(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:  "unclosed HTML comment",
			input: `<!-- unclosed HTML comment`,
			expected: parse.Error("expected end comment literal '-->' not found",
				parse.Position{
					Index: 0,
					Line:  0,
					Col:   0,
				}),
		},
		{
			name:  "comment in comment",
			input: `<!-- <-- other --> -->`,
			expected: parse.Error("comment contains invalid sequence '--'", parse.Position{
				Index: 8,
				Line:  0,
				Col:   8,
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
