package parser

import (
	"testing"

	"github.com/a-h/lexical/input"
	"github.com/google/go-cmp/cmp"
)

func TestTextParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected Text
	}{
		{
			name:  "Text ends at an element start",
			input: `abcdef<a href="https://example.com">More</a>`,
			expected: Text{
				Value: "abcdef",
			},
		},
		{
			name:  "Text ends at a templ expression start",
			input: `abcdef{%= "test" %}`,
			expected: Text{
				Value: "abcdef",
			},
		},
		{
			name:  "Text may contain spaces",
			input: `abcdef ghijk{%= "test" %}`,
			expected: Text{
				Value: "abcdef ghijk",
			},
		},
		{
			name:  "Text may contain named references",
			input: `abcdef&nbsp;ghijk{%= "test" %}`,
			expected: Text{
				Value: "abcdef&nbsp;ghijk",
			},
		},
		{
			name:  "Text may contain base 10 numeric references",
			input: `abcdef&#32;ghijk{%= "test" %}`,
			expected: Text{
				Value: "abcdef&#32;ghijk",
			},
		},
		{
			name:  "Text may contain hexadecimal numeric references",
			input: `abcdef&#x20;ghijk{%= "test" %}`,
			expected: Text{
				Value: "abcdef&#x20;ghijk",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := input.NewFromString(tt.input)
			result := newTextParser().Parse(input)
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
