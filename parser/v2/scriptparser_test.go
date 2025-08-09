package parser

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	_ "embed"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
)

func TestScriptElementParserPlain(t *testing.T) {
	files, _ := filepath.Glob("scriptparsertestdata/*.txt")
	if len(files) == 0 {
		t.Errorf("no test files found")
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			a, err := txtar.ParseFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if len(a.Files) != 2 {
				t.Fatalf("expected 2 files, got %d", len(a.Files))
			}

			input := parse.NewInput(clean(a.Files[0].Data))
			result, ok, err := scriptElement.Parse(input)
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}
			if !ok {
				t.Fatalf("failed to parse at %d", input.Index())
			}

			se, isScriptElement := result.(*ScriptElement)
			if !isScriptElement {
				t.Fatalf("expected ScriptElement, got %T", result)
			}

			var actual strings.Builder
			for _, content := range se.Contents {
				if content.GoCode != nil {
					t.Fatalf("expected plain text, got GoCode")
				}
				if content.Value == nil {
					t.Fatalf("expected plain text, got nil")
				}
				actual.WriteString(*content.Value)
			}

			expected := clean(a.Files[1].Data)
			if diff := cmp.Diff(actual.String(), string(expected)); diff != "" {
				t.Fatalf("%s:\n%s", file, diff)
			}
		})
	}
}

func TestScriptElementParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *ScriptElement
	}{
		{
			name:  "script: no content",
			input: `<script></script>`,
			expected: &ScriptElement{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 17, Line: 0, Col: 17},
				},
			},
		},
		{
			name:  "script: vbscript",
			input: `<script type="vbscript">dim x = 1</script>`,
			expected: &ScriptElement{
				Attributes: []Attribute{
					&ConstantAttribute{
						Value: "vbscript",
						Key: ConstantAttributeKey{
							Name: "type",
							NameRange: Range{
								From: Position{Index: 8, Line: 0, Col: 8},
								To:   Position{Index: 12, Line: 0, Col: 12},
							},
						},
					},
				},
				Contents: []ScriptContents{
					NewScriptContentsScriptCode("dim x = 1"),
				},
			},
		},
		{
			name:  "script: go expression",
			input: `<script>{{ name }}</script>`,
			expected: &ScriptElement{
				Contents: []ScriptContents{
					NewScriptContentsGo(&GoCode{
						Expression: Expression{
							Value: "name",
							Range: Range{
								From: Position{Index: 11, Line: 0, Col: 11},
								To:   Position{Index: 15, Line: 0, Col: 15},
							},
						},
					}, false),
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 27, Line: 0, Col: 27},
				},
			},
		},
		{
			name:  "script: go expression with explicit type",
			input: `<script type="text/javascript">{{ name }}</script>`,
			expected: &ScriptElement{
				Attributes: []Attribute{&ConstantAttribute{
					Value: "text/javascript",
					Key: ConstantAttributeKey{
						Name: "type", NameRange: Range{
							From: Position{Index: 8, Line: 0, Col: 8},
							To:   Position{Index: 12, Line: 0, Col: 12},
						},
					},
				}},
				Contents: []ScriptContents{
					NewScriptContentsGo(&GoCode{
						Expression: Expression{
							Value: "name",
							Range: Range{
								From: Position{Index: 34, Line: 0, Col: 34},
								To:   Position{Index: 38, Line: 0, Col: 38},
							},
						},
					}, false),
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 50, Line: 0, Col: 50},
				},
			},
		},
		{
			name:  "script: go expression with module type",
			input: `<script type="module">{{ name }}</script>`,
			expected: &ScriptElement{
				Attributes: []Attribute{&ConstantAttribute{
					Value: "module",
					Key: ConstantAttributeKey{
						Name: "type", NameRange: Range{
							From: Position{Index: 8, Line: 0, Col: 8},
							To:   Position{Index: 12, Line: 0, Col: 12},
						},
					},
				}},
				Contents: []ScriptContents{
					NewScriptContentsGo(&GoCode{
						Expression: Expression{
							Value: "name",
							Range: Range{
								From: Position{Index: 25, Line: 0, Col: 25},
								To:   Position{Index: 29, Line: 0, Col: 29},
							},
						},
					}, false),
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 41, Line: 0, Col: 41},
				},
			},
		},
		{
			name:  "script: go expression with javascript type",
			input: `<script type="javascript">{{ name }}</script>`,
			expected: &ScriptElement{
				Attributes: []Attribute{&ConstantAttribute{
					Value: "javascript",
					Key: ConstantAttributeKey{
						Name: "type", NameRange: Range{
							From: Position{Index: 8, Line: 0, Col: 8},
							To:   Position{Index: 12, Line: 0, Col: 12},
						},
					},
				}},
				Contents: []ScriptContents{
					NewScriptContentsGo(&GoCode{
						Expression: Expression{
							Value: "name",
							Range: Range{
								From: Position{Index: 29, Line: 0, Col: 29},
								To:   Position{Index: 33, Line: 0, Col: 33},
							},
						},
					}, false),
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 45, Line: 0, Col: 45},
				},
			},
		},
		{
			name: "script: go expression - multiline 1",
			input: `<script>
{{ name }}
</script>`,
			expected: &ScriptElement{
				Contents: []ScriptContents{
					NewScriptContentsScriptCode("\n"),
					NewScriptContentsGo(&GoCode{
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
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 29, Line: 2, Col: 9},
				},
			},
		},
		{
			name:  "script: go expression in single quoted string",
			input: `<script>var x = '{{ name }}';</script>`,
			expected: &ScriptElement{
				Contents: []ScriptContents{
					NewScriptContentsScriptCode("var x = '"),
					NewScriptContentsGo(&GoCode{
						Expression: Expression{
							Value: "name",
							Range: Range{
								From: Position{Index: 20, Line: 0, Col: 20},
								To:   Position{Index: 24, Line: 0, Col: 24},
							},
						},
					}, true),
					NewScriptContentsScriptCode("';"),
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 38, Line: 0, Col: 38},
				},
			},
		},
		{
			name:  "script: go expression in double quoted string",
			input: `<script>var x = "{{ name }}";</script>`,
			expected: &ScriptElement{
				Contents: []ScriptContents{
					NewScriptContentsScriptCode("var x = \""),
					NewScriptContentsGo(&GoCode{
						Expression: Expression{
							Value: "name",
							Range: Range{
								From: Position{Index: 20, Line: 0, Col: 20},
								To:   Position{Index: 24, Line: 0, Col: 24},
							},
						},
					}, true),
					NewScriptContentsScriptCode("\";"),
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 38, Line: 0, Col: 38},
				},
			},
		},
		{
			name: "script: go expression in double quoted multiline string",
			input: `<script>var x = "This is a test \
{{ name }} \
to see if it works";</script>`,
			expected: &ScriptElement{
				Contents: []ScriptContents{
					NewScriptContentsScriptCode("var x = \"This is a test \\\n"),
					NewScriptContentsGo(&GoCode{
						Expression: Expression{
							Value: "name",
							Range: Range{
								From: Position{Index: 37, Line: 1, Col: 3},
								To:   Position{Index: 41, Line: 1, Col: 7},
							},
						},
						TrailingSpace: SpaceHorizontal,
					}, true),
					NewScriptContentsScriptCode("\\\nto see if it works\";"),
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 76, Line: 2, Col: 29},
				},
			},
		},
		{
			name:  "script: go expression in backtick quoted string",
			input: `<script>var x = ` + "`" + "{{ name }}" + "`" + `;</script>`,
			expected: &ScriptElement{
				Contents: []ScriptContents{
					NewScriptContentsScriptCode("var x = `"),
					NewScriptContentsGo(&GoCode{
						Expression: Expression{
							Value: "name",
							Range: Range{
								From: Position{Index: 20, Line: 0, Col: 20},
								To:   Position{Index: 24, Line: 0, Col: 24},
							},
						},
					}, true),
					NewScriptContentsScriptCode("`;"),
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 38, Line: 0, Col: 38},
				},
			},
		},
		{
			name: "script: single line commented out go expressions are ignored",
			input: `<script>
// {{ name }}
</script>`,
			expected: &ScriptElement{
				Contents: []ScriptContents{
					NewScriptContentsScriptCode("\n"),
					NewScriptContentsScriptCode("// {{ name }}\n"),
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 32, Line: 2, Col: 9},
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
			expected: &ScriptElement{
				Contents: []ScriptContents{
					NewScriptContentsScriptCode("\n"),
					NewScriptContentsScriptCode("/* There's some content\n{{ name }}\nbut it's commented out */\n"),
				},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 79, Line: 4, Col: 9},
				},
			},
		},
		{
			name: "script: non js content is parsed raw",
			input: `<script type="text/hyperscript">
set tier_1 to #tier-1's value
</script>`,
			expected: &ScriptElement{
				Attributes: []Attribute{&ConstantAttribute{
					Value: "text/hyperscript",
					Key: ConstantAttributeKey{
						Name: "type", NameRange: Range{
							From: Position{Index: 8, Line: 0, Col: 8},
							To:   Position{Index: 12, Line: 0, Col: 12},
						},
					},
				}},
				Contents: []ScriptContents{
					NewScriptContentsScriptCode("\nset tier_1 to #tier-1's value\n"),
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

func FuzzScriptParser(f *testing.F) {
	files, _ := filepath.Glob("scriptparsertestdata/*.txt")
	if len(files) == 0 {
		f.Errorf("no test files found")
	}
	for _, file := range files {
		a, err := txtar.ParseFile(file)
		if err != nil {
			f.Fatal(err)
		}
		if len(a.Files) != 2 {
			f.Fatalf("expected 2 files, got %d", len(a.Files))
		}
		f.Add(clean(a.Files[0].Data))
	}

	f.Fuzz(func(t *testing.T, input string) {
		_, _, _ = scriptElement.Parse(parse.NewInput(input))
	})
}

func clean(b []byte) string {
	b = bytes.ReplaceAll(b, []byte("$\n"), []byte("\n"))
	b = bytes.TrimSuffix(b, []byte("\n"))
	return string(b)
}
