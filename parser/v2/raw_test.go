package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

var ignoredContent = `{
	fjkjkl: 123,
	{{
}`

func TestRawElementParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected RawElement
	}{
		{
			name:  "style tag",
			input: `<style type="text/css">contents</style>`,
			expected: RawElement{
				Name: "style",
				Attributes: []Attribute{
					ConstantAttribute{
						Name:  "type",
						Value: "text/css",
						NameRange: Range{
							From: Position{Index: 7, Line: 0, Col: 7},
							To:   Position{Index: 11, Line: 0, Col: 11},
						},
					},
				},
				Contents: "contents",
			},
		},
		{
			name:  "style tag containing mismatched braces",
			input: `<style type="text/css">` + ignoredContent + "</style>",
			expected: RawElement{
				Name: "style",
				Attributes: []Attribute{
					ConstantAttribute{
						Name:  "type",
						Value: "text/css",
						NameRange: Range{
							From: Position{Index: 7, Line: 0, Col: 7},
							To:   Position{Index: 11, Line: 0, Col: 11},
						},
					},
				},
				Contents: ignoredContent,
			},
		},
		{
			name:  "script tag",
			input: `<script type="vbscript">dim x = 1</script>`,
			expected: RawElement{
				Name: "script",
				Attributes: []Attribute{
					ConstantAttribute{
						Name:  "type",
						Value: "vbscript",
						NameRange: Range{
							From: Position{Index: 8, Line: 0, Col: 8},
							To:   Position{Index: 12, Line: 0, Col: 12},
						},
					},
				},
				Contents: "dim x = 1",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			actual, ok, err := rawElements.Parse(input)
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

func TestRawElementParserIsNotGreedy(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected RawElement
	}{
		{
			name:  "styles tag",
			input: `<styles></styles>`,
		},
		{
			name:  "scripts tag",
			input: `<scripts></scripts>`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			actual, ok, err := rawElements.Parse(input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ok {
				t.Fatalf("unexpected success for input %q", tt.input)
			}
			if actual != nil {
				t.Fatalf("expected nil Node got %v", actual)
			}
		})
	}
}
