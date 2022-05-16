package parser

import (
	"testing"

	"github.com/a-h/lexical/input"
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
			input := input.NewFromString(tt.input)
			result := rawElements(input)
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
