package parser

import (
	"testing"

	"github.com/a-h/lexical/input"
	"github.com/google/go-cmp/cmp"
)

func TestScriptTemplateParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected ScriptTemplate
	}{
		{
			name: "script: no parameters, no content",
			input: `script Name() {
}`,
			expected: ScriptTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 7,
							Line:  1,
							Col:   7,
						},
						To: Position{
							Index: 11,
							Line:  1,
							Col:   11,
						},
					},
				},
				Parameters: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 12,
							Line:  1,
							Col:   12,
						},
						To: Position{
							Index: 12,
							Line:  1,
							Col:   12,
						},
					},
				},
			},
		},
		{
			name: "script: no spaces",
			input: `script Name(){
}`,
			expected: ScriptTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 7,
							Line:  1,
							Col:   7,
						},
						To: Position{
							Index: 11,
							Line:  1,
							Col:   11,
						},
					},
				},
				Parameters: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 12,
							Line:  1,
							Col:   12,
						},
						To: Position{
							Index: 12,
							Line:  1,
							Col:   12,
						},
					},
				},
			},
		},
		{
			name: "script: containing a JS variable",
			input: `script Name() {
var x = "x";
}`,
			expected: ScriptTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 7,
							Line:  1,
							Col:   7,
						},
						To: Position{
							Index: 11,
							Line:  1,
							Col:   11,
						},
					},
				},
				Parameters: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 12,
							Line:  1,
							Col:   12,
						},
						To: Position{
							Index: 12,
							Line:  1,
							Col:   12,
						},
					},
				},
				Value: `var x = "x";` + "\n",
			},
		},
		{
			name: "script: single argument",
			input: `script Name(value string) {
console.log(value);
}`,
			expected: ScriptTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 7,
							Line:  1,
							Col:   7,
						},
						To: Position{
							Index: 11,
							Line:  1,
							Col:   11,
						},
					},
				},
				Parameters: Expression{
					Value: "value string",
					Range: Range{
						From: Position{
							Index: 12,
							Line:  1,
							Col:   12,
						},
						To: Position{
							Index: 24,
							Line:  1,
							Col:   24,
						},
					},
				},
				Value: `console.log(value);` + "\n",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := input.NewFromString(tt.input)
			result := newScriptTemplateParser().Parse(input)
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
