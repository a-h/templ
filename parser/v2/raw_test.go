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
