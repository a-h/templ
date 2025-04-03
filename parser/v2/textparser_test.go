package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestTextParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected *Text
	}{
		{
			name:  "Text ends at an element start",
			input: `abcdef<a href="https://example.com">More</a>`,
			expected: &Text{
				Value: "abcdef",
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 6, Line: 0, Col: 6},
				},
			},
		},
		{
			name:  "Text ends at a templ expression start",
			input: `abcdef{ "test" }`,
			expected: &Text{
				Value: "abcdef",
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 6, Line: 0, Col: 6},
				},
			},
		},
		{
			name:  "Text may contain spaces",
			input: `abcdef ghijk{ "test" }`,
			expected: &Text{
				Value: "abcdef ghijk",
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 12, Line: 0, Col: 12},
				},
			},
		},
		{
			name:  "Text may contain named references",
			input: `abcdef&nbsp;ghijk{ "test" }`,
			expected: &Text{
				Value: "abcdef&nbsp;ghijk",
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 17, Line: 0, Col: 17},
				},
			},
		},
		{
			name:  "Text may contain base 10 numeric references",
			input: `abcdef&#32;ghijk{ "test" }`,
			expected: &Text{
				Value: "abcdef&#32;ghijk",
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 16, Line: 0, Col: 16},
				},
			},
		},
		{
			name:  "Text may contain hexadecimal numeric references",
			input: `abcdef&#x20;ghijk{ "test" }`,
			expected: &Text{
				Value: "abcdef&#x20;ghijk",
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 17, Line: 0, Col: 17},
				},
			},
		},
		{
			name:  "Multiline text is collected line by line",
			input: "Line 1\nLine 2",
			expected: &Text{
				Value: "Line 1",
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 6, Line: 0, Col: 6},
				},
				TrailingSpace: "\n",
			},
		},
		{
			name:  "Multiline text is collected line by line (Windows)",
			input: "Line 1\r\nLine 2",
			expected: &Text{
				Value: "Line 1",
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 6, Line: 0, Col: 6},
				},
				TrailingSpace: "\n",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			actual, ok, err := textParser.Parse(input)
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
