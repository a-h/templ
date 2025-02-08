package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestScriptElementParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ScriptElement
	}{
		{
			name:  "script: no content",
			input: `<script></script>`,
			expected: ScriptElement{
				Name: "script",
			},
		},
		{
			name:  "script: vbscript",
			input: `<script type="vbscript">dim x = 1</script>`,
			expected: ScriptElement{
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
				Contents: []ScriptContents{
					NewScriptContentsJS("dim x = 1"),
				},
			},
		},
		{
			name:  "script: go expression",
			input: `<script>{{ name }}</script>`,
			expected: ScriptElement{
				Name: "script",
				Contents: []ScriptContents{
					NewScriptContentsGo(GoCode{
						Expression: Expression{
							Value: "name",
							Range: Range{
								From: Position{Index: 11, Line: 0, Col: 11},
								To:   Position{Index: 15, Line: 0, Col: 15},
							},
						},
					}, false),
				},
			},
		},
		{
			name: "script: go expression - multiline 1",
			input: `<script>
{{ name }}
</script>`,
			expected: ScriptElement{
				Name: "script",
				Contents: []ScriptContents{
					NewScriptContentsJS("\n"),
					NewScriptContentsGo(GoCode{
						Expression: Expression{
							Value: "name",
							Range: Range{
								From: Position{Index: 12, Line: 1, Col: 3},
								To:   Position{Index: 16, Line: 1, Col: 7},
							},
						},
						TrailingSpace: SpaceVertical,
					}, false),
				},
			},
		},
		{
			name:  "script: go expression in single quoted string",
			input: `<script>var x = '{{ name }}';</script>`,
			expected: ScriptElement{
				Name: "script",
				Contents: []ScriptContents{
					NewScriptContentsJS("var x = '"),
					NewScriptContentsGo(GoCode{
						Expression: Expression{
							Value: "name",
							Range: Range{
								From: Position{Index: 20, Line: 0, Col: 20},
								To:   Position{Index: 24, Line: 0, Col: 24},
							},
						},
					}, true),
					NewScriptContentsJS("';"),
				},
			},
		},
		{
			name:  "script: go expression in double quoted string",
			input: `<script>var x = "{{ name }}";</script>`,
			expected: ScriptElement{
				Name: "script",
				Contents: []ScriptContents{
					NewScriptContentsJS("var x = \""),
					NewScriptContentsGo(GoCode{
						Expression: Expression{
							Value: "name",
							Range: Range{
								From: Position{Index: 20, Line: 0, Col: 20},
								To:   Position{Index: 24, Line: 0, Col: 24},
							},
						},
					}, true),
					NewScriptContentsJS("\";"),
				},
			},
		},
		{
			name: "script: go expression in double quoted multiline string",
			input: `<script>var x = "This is a test \
{{ name }} \
to see if it works";</script>`,
			expected: ScriptElement{
				Name: "script",
				Contents: []ScriptContents{
					NewScriptContentsJS("var x = \"This is a test \\\n"),
					NewScriptContentsGo(GoCode{
						Expression: Expression{
							Value: "name",
							Range: Range{
								From: Position{Index: 37, Line: 1, Col: 3},
								To:   Position{Index: 41, Line: 1, Col: 7},
							},
						},
						TrailingSpace: SpaceHorizontal,
					}, true),
					NewScriptContentsJS("\\\nto see if it works\";"),
				},
			},
		},
		{
			name:  "script: go expression in backtick quoted string",
			input: `<script>var x = ` + "`" + "{{ name }}" + "`" + `;</script>`,
			expected: ScriptElement{
				Name: "script",
				Contents: []ScriptContents{
					NewScriptContentsJS("var x = `"),
					NewScriptContentsGo(GoCode{
						Expression: Expression{
							Value: "name",
							Range: Range{
								From: Position{Index: 20, Line: 0, Col: 20},
								To:   Position{Index: 24, Line: 0, Col: 24},
							},
						},
					}, true),
					NewScriptContentsJS("`;"),
				},
			},
		},
		{
			name: "script: single line commented out go expressions are ignored",
			input: `<script>
// {{ name }}
</script>`,
			expected: ScriptElement{
				Name: "script",
				Contents: []ScriptContents{
					NewScriptContentsJS("\n"),
					NewScriptContentsJS("// {{ name }}\n"),
				},
			},
		},
		{
			name: "script: multiline commented out go expressions are ignored",
			input: `<script>
/* There's some content
{{ name }}
but it's commented out */
</script>`,
			expected: ScriptElement{
				Name: "script",
				Contents: []ScriptContents{
					NewScriptContentsJS("\n"),
					NewScriptContentsJS("/* There's some content\n{{ name }}\nbut it's commented out */\n"),
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			result, ok, err := scriptElement.Parse(input)
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
