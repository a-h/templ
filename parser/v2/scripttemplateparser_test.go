package parser

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestScriptTemplateParser(t *testing.T) {
	var tests = []struct {
		name       string
		input      string
		expected   ScriptTemplate
		expectedJS string
	}{
		{
			name: "script: no parameters, no content",
			input: `script Name() {
}`,
			expectedJS: `               
 `,
			expected: ScriptTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 7,
							Line:  0,
							Col:   7,
						},
						To: Position{
							Index: 11,
							Line:  0,
							Col:   11,
						},
					},
				},
				Parameters: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 12,
							Line:  0,
							Col:   12,
						},
						To: Position{
							Index: 12,
							Line:  0,
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
			expectedJS: `               
 `,
			expected: ScriptTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 7,
							Line:  0,
							Col:   7,
						},
						To: Position{
							Index: 11,
							Line:  0,
							Col:   11,
						},
					},
				},
				Parameters: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 12,
							Line:  0,
							Col:   12,
						},
						To: Position{
							Index: 12,
							Line:  0,
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
			expectedJS: `               
var x = "x";
 `,
			expected: ScriptTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 7,
							Line:  0,
							Col:   7,
						},
						To: Position{
							Index: 11,
							Line:  0,
							Col:   11,
						},
					},
				},
				Parameters: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 12,
							Line:  0,
							Col:   12,
						},
						To: Position{
							Index: 12,
							Line:  0,
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
			expectedJS: `                           
console.log(value);
 `,
			expected: ScriptTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 7,
							Line:  0,
							Col:   7,
						},
						To: Position{
							Index: 11,
							Line:  0,
							Col:   11,
						},
					},
				},
				Parameters: Expression{
					Value: "value string",
					Range: Range{
						From: Position{
							Index: 12,
							Line:  0,
							Col:   12,
						},
						To: Position{
							Index: 24,
							Line:  0,
							Col:   24,
						},
					},
				},
				Value: `console.log(value);` + "\n",
			},
		},
		{
			name: "script: comment with single quote",
			input: `script Name() {
	//'
}`,
			expectedJS: `               
	//'
 `,
			expected: ScriptTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 7,
							Line:  0,
							Col:   7,
						},
						To: Position{
							Index: 11,
							Line:  0,
							Col:   11,
						},
					},
				},
				Parameters: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 12,
							Line:  0,
							Col:   12,
						},
						To: Position{
							Index: 12,
							Line:  0,
							Col:   12,
						},
					},
				},
				Value: `	//'` + "\n",
			},
		},
		{
			name: "script: empty assignment",
			input: `script Name() {
  let x = '';
}`,
			expectedJS: `               
  let x = '';
 `,
			expected: ScriptTemplate{
				Name: Expression{
					Value: "Name",
					Range: Range{
						From: Position{
							Index: 7,
							Line:  0,
							Col:   7,
						},
						To: Position{
							Index: 11,
							Line:  0,
							Col:   11,
						},
					},
				},
				Value: `  let x = '';` + "\n",
				Parameters: Expression{
					Value: "",
					Range: Range{
						From: Position{
							Index: 12,
							Line:  0,
							Col:   12,
						},
						To: Position{
							Index: 12,
							Line:  0,
							Col:   12,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		suffixes := []string{"", " Trailing '", ` Trailing "`, "\n// More content."}
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", tt.name, i), func(t *testing.T) {
				input := parse.NewInput(tt.input + suffix)
				actual, ok, err := scriptTemplateParser.Parse(input)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !ok {
					t.Fatalf("unexpected failure for input %q", tt.input)
				}
				if diff := cmp.Diff(tt.expected, actual); diff != "" {
					t.Error(diff)
				}
				actualSuffix, _ := input.Peek(-1)
				if diff := cmp.Diff(suffix, actualSuffix); diff != "" {
					t.Error("unexpected suffix")
					t.Error(diff)
				}

				w := new(bytes.Buffer)
				cw := NewContextWriter(w, WriteContextJS)
				if err := actual.Write(cw, 0); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				actualJS := w.String()
				if diff := cmp.Diff(tt.expectedJS, actualJS); diff != "" {
					t.Error(diff)

					t.Errorf("input:\n%s", displayWhitespaceChars(tt.input))
					t.Errorf("expected:\n%s", displayWhitespaceChars(tt.expectedJS))
					t.Errorf("got:\n%s", displayWhitespaceChars(actualJS))
				}
				if diff := cmp.Diff(getLineLengths(tt.input), getLineLengths(actualJS)); diff != "" {
					t.Errorf(diff)
				}
			})
		}
	}
}
