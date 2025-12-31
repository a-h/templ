package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestFallthroughParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected *Fallthrough
	}{
		{
			name:  "basic fallthrough",
			input: "fallthrough\n",
			expected: &Fallthrough{
				Range: NewRange(parse.Position{Index: 0, Line: 0, Col: 0}, parse.Position{Index: 12, Line: 1, Col: 0}),
			},
		},
		{
			name:  "fallthrough with spaces before newline",
			input: "fallthrough    \n",
			expected: &Fallthrough{
				Range: NewRange(parse.Position{Index: 0, Line: 0, Col: 0}, parse.Position{Index: 16, Line: 1, Col: 0}),
			},
		},
		{
			name:  "fallthrough with tabs before newline",
			input: "fallthrough\t\t\t\n",
			expected: &Fallthrough{
				Range: NewRange(parse.Position{Index: 0, Line: 0, Col: 0}, parse.Position{Index: 15, Line: 1, Col: 0}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			result, ok, err := fallthroughExpression.Parse(input)
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

func TestFallthroughParserErrors(t *testing.T) {
	var tests = []struct {
		name          string
		input         string
		expectedError error
		expectedOK    bool
	}{
		{
			name:          "invalid fallthrough keyword",
			input:         `fallthroug`,
			expectedError: nil,
			expectedOK:    false,
		},
		{
			name:  "missing newline after fallthrough",
			input: `fallthrough some extra`,
			expectedError: parse.Error("expected newline after fallthrough", parse.Position{
				Index: 12,
				Line:  0,
				Col:   12,
			}),
			expectedOK: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			_, ok, err := fallthroughExpression.Parse(input)
			if ok != tt.expectedOK {
				t.Fatalf("expected ok to be %v, got %v", tt.expectedOK, ok)
			}
			if diff := cmp.Diff(tt.expectedError, err); diff != "" {
				t.Error(diff)
			}
		})
	}
}
