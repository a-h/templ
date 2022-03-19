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
			input: `{% script Name() %}
{% endscript %}`,
			expected: ScriptTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 10,
							Line:  1,
							Col:   10,
						},
						To: Position{
							Index: 14,
							Line:  1,
							Col:   14,
						},
					},
				},
				Parameters: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 15,
							Line:  1,
							Col:   15,
						},
						To: Position{
							Index: 15,
							Line:  1,
							Col:   15,
						},
					},
				},
			},
		},
		{
			name: "script: no spaces",
			input: `{%script Name()%}
{% endscript %}`,
			expected: ScriptTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 9,
							Line:  1,
							Col:   9,
						},
						To: Position{
							Index: 13,
							Line:  1,
							Col:   13,
						},
					},
				},
				Parameters: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 14,
							Line:  1,
							Col:   14,
						},
						To: Position{
							Index: 14,
							Line:  1,
							Col:   14,
						},
					},
				},
			},
		},
		{
			name: "script: containing a JS variable",
			input: `{% script Name() %}
var x = "x";
{% endscript %}`,
			expected: ScriptTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 10,
							Line:  1,
							Col:   10,
						},
						To: Position{
							Index: 14,
							Line:  1,
							Col:   14,
						},
					},
				},
				Parameters: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 15,
							Line:  1,
							Col:   15,
						},
						To: Position{
							Index: 15,
							Line:  1,
							Col:   15,
						},
					},
				},
				Value: `var x = "x";` + "\n",
			},
		},
		{
			name: "script: single argument",
			input: `{% script Name(value string) %}
console.log(value);
{% endscript %}`,
			expected: ScriptTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 10,
							Line:  1,
							Col:   10,
						},
						To: Position{
							Index: 14,
							Line:  1,
							Col:   14,
						},
					},
				},
				Parameters: Expression{
					Value: "value string",
					Range: Range{
						From: Position{
							Index: 15,
							Line:  1,
							Col:   15,
						},
						To: Position{
							Index: 27,
							Line:  1,
							Col:   27,
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
