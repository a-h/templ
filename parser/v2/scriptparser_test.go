package parser

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	_ "embed"

	"github.com/a-h/parse"
	"github.com/a-h/templ/runtime"
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
			if diff := cmp.Diff(string(expected), actual.String()); diff != "" {
				t.Error(diff)
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
			name:  "no content",
			input: `<script></script>`,
			expected: &ScriptElement{
				OpenTagRange:  Range{To: Position{Index: 8, Col: 8}},
				CloseTagRange: Range{From: Position{Index: 8, Col: 8}, To: Position{Index: 17, Col: 17}},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 17, Line: 0, Col: 17},
				},
			},
		},
		{
			name:  "vbscript",
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
						ValueRange: Range{
							From: Position{Index: 14, Line: 0, Col: 14},
							To:   Position{Index: 22, Line: 0, Col: 22},
						},
						Range: Range{
							From: Position{Index: 8, Line: 0, Col: 8},
							To:   Position{Index: 23, Line: 0, Col: 23},
						},
					},
				},
				Contents: []ScriptContents{
					NewScriptContentsScriptCode("dim x = 1"),
				},
				OpenTagRange:  Range{To: Position{Index: 24, Col: 24}},
				CloseTagRange: Range{From: Position{Index: 33, Col: 33}, To: Position{Index: 42, Col: 42}},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 42, Line: 0, Col: 42},
				},
			},
		},
		{
			name:  "go expression",
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
						Range: Range{
							From: Position{Index: 8, Line: 0, Col: 8},
							To:   Position{Index: 18, Line: 0, Col: 18},
						},
					}, ScriptContentsContextExpression),
				},
				OpenTagRange:  Range{To: Position{Index: 8, Col: 8}},
				CloseTagRange: Range{From: Position{Index: 18, Col: 18}, To: Position{Index: 27, Col: 27}},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 27, Line: 0, Col: 27},
				},
			},
		},
		{
			name:  "go expression with explicit type",
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
					ValueRange: Range{
						From: Position{Index: 14, Line: 0, Col: 14},
						To:   Position{Index: 29, Line: 0, Col: 29},
					},
					Range: Range{
						From: Position{Index: 8, Line: 0, Col: 8},
						To:   Position{Index: 30, Line: 0, Col: 30},
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
						Range: Range{
							From: Position{Index: 31, Line: 0, Col: 31},
							To:   Position{Index: 41, Line: 0, Col: 41},
						},
					}, ScriptContentsContextExpression),
				},
				OpenTagRange:  Range{To: Position{Index: 31, Col: 31}},
				CloseTagRange: Range{From: Position{Index: 41, Col: 41}, To: Position{Index: 50, Col: 50}},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 50, Line: 0, Col: 50},
				},
			},
		},
		{
			name:  "go expression with module type",
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
					ValueRange: Range{
						From: Position{Index: 14, Line: 0, Col: 14},
						To:   Position{Index: 20, Line: 0, Col: 20},
					},
					Range: Range{
						From: Position{Index: 8, Line: 0, Col: 8},
						To:   Position{Index: 21, Line: 0, Col: 21},
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
						Range: Range{
							From: Position{Index: 22, Line: 0, Col: 22},
							To:   Position{Index: 32, Line: 0, Col: 32},
						},
					}, ScriptContentsContextExpression),
				},
				OpenTagRange:  Range{To: Position{Index: 22, Col: 22}},
				CloseTagRange: Range{From: Position{Index: 32, Col: 32}, To: Position{Index: 41, Col: 41}},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 41, Line: 0, Col: 41},
				},
			},
		},
		{
			name:  "go expression with javascript type",
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
					ValueRange: Range{
						From: Position{Index: 14, Line: 0, Col: 14},
						To:   Position{Index: 24, Line: 0, Col: 24},
					},
					Range: Range{
						From: Position{Index: 8, Line: 0, Col: 8},
						To:   Position{Index: 25, Line: 0, Col: 25},
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
						Range: Range{
							From: Position{Index: 26, Line: 0, Col: 26},
							To:   Position{Index: 36, Line: 0, Col: 36},
						},
					}, ScriptContentsContextExpression),
				},
				OpenTagRange:  Range{To: Position{Index: 26, Col: 26}},
				CloseTagRange: Range{From: Position{Index: 36, Col: 36}, To: Position{Index: 45, Col: 45}},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 45, Line: 0, Col: 45},
				},
			},
		},
		{
			name: "go expression - multiline 1",
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
						Range: Range{
							From: Position{Index: 9, Line: 1, Col: 0},
							To:   Position{Index: 20, Line: 2, Col: 0},
						},
					}, ScriptContentsContextExpression),
				},
				OpenTagRange:  Range{To: Position{Index: 8, Col: 8}},
				CloseTagRange: Range{From: Position{Index: 20, Line: 2}, To: Position{Index: 29, Line: 2, Col: 9}},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 29, Line: 2, Col: 9},
				},
			},
		},
		{
			name:  "go expression in single quoted string",
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
						Range: Range{
							From: Position{Index: 17, Line: 0, Col: 17},
							To:   Position{Index: 27, Line: 0, Col: 27},
						},
					}, ScriptContentsContextString),
					NewScriptContentsScriptCode("';"),
				},
				OpenTagRange:  Range{To: Position{Index: 8, Col: 8}},
				CloseTagRange: Range{From: Position{Index: 29, Col: 29}, To: Position{Index: 38, Col: 38}},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 38, Line: 0, Col: 38},
				},
			},
		},
		{
			name:  "go expression in double quoted string",
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
						Range: Range{
							From: Position{Index: 17, Line: 0, Col: 17},
							To:   Position{Index: 27, Line: 0, Col: 27},
						},
					}, ScriptContentsContextString),
					NewScriptContentsScriptCode("\";"),
				},
				OpenTagRange:  Range{To: Position{Index: 8, Col: 8}},
				CloseTagRange: Range{From: Position{Index: 29, Col: 29}, To: Position{Index: 38, Col: 38}},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 38, Line: 0, Col: 38},
				},
			},
		},
		{
			name: "go expression in double quoted multiline string",
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
						Range: Range{
							From: Position{Index: 34, Line: 1, Col: 0},
							To:   Position{Index: 45, Line: 1, Col: 11},
						},
					}, ScriptContentsContextString),
					NewScriptContentsScriptCode("\\\nto see if it works\";"),
				},
				OpenTagRange:  Range{To: Position{Index: 8, Col: 8}},
				CloseTagRange: Range{From: Position{Index: 67, Line: 2, Col: 20}, To: Position{Index: 76, Line: 2, Col: 29}},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 76, Line: 2, Col: 29},
				},
			},
		},
		{
			name:  "go expression in backtick quoted string",
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
						Range: Range{
							From: Position{Index: 17, Line: 0, Col: 17},
							To:   Position{Index: 27, Line: 0, Col: 27},
						},
					}, ScriptContentsContextTemplateLiteral),
					NewScriptContentsScriptCode("`;"),
				},
				OpenTagRange:  Range{To: Position{Index: 8, Col: 8}},
				CloseTagRange: Range{From: Position{Index: 29, Col: 29}, To: Position{Index: 38, Col: 38}},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 38, Line: 0, Col: 38},
				},
			},
		},
		{
			name: "single line commented out go expressions are ignored",
			input: `<script>
// {{ name }}
</script>`,
			expected: &ScriptElement{
				Contents: []ScriptContents{
					NewScriptContentsScriptCode("\n"),
					NewScriptContentsScriptCode("// {{ name }}\n"),
				},
				OpenTagRange:  Range{To: Position{Index: 8, Col: 8}},
				CloseTagRange: Range{From: Position{Index: 23, Line: 2}, To: Position{Index: 32, Line: 2, Col: 9}},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 32, Line: 2, Col: 9},
				},
			},
		},
		{
			name: "single line comments after expressions are allowed",
			input: `<script>
const category = path.split('/')[2]; // example comment
</script>`,
			expected: &ScriptElement{
				Contents: []ScriptContents{
					NewScriptContentsScriptCode("\nconst category = path.split('/')[2]; "),
					NewScriptContentsScriptCode("// example comment\n"),
				},
				OpenTagRange:  Range{To: Position{Index: 8, Col: 8}},
				CloseTagRange: Range{From: Position{Index: 65, Line: 2}, To: Position{Index: 74, Line: 2, Col: 9}},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 74, Line: 2, Col: 9},
				},
			},
		},
		{
			name: "multiline commented out go expressions are ignored",
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
				OpenTagRange:  Range{To: Position{Index: 8, Col: 8}},
				CloseTagRange: Range{From: Position{Index: 70, Line: 4}, To: Position{Index: 79, Line: 4, Col: 9}},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 79, Line: 4, Col: 9},
				},
			},
		},
		{
			name: "non js content is parsed raw",
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
					ValueRange: Range{
						From: Position{Index: 14, Line: 0, Col: 14},
						To:   Position{Index: 30, Line: 0, Col: 30},
					},
					Range: Range{
						From: Position{Index: 8, Line: 0, Col: 8},
						To:   Position{Index: 31, Line: 0, Col: 31},
					},
				}},
				Contents: []ScriptContents{
					NewScriptContentsScriptCode("\nset tier_1 to #tier-1's value\n"),
				},
				OpenTagRange:  Range{To: Position{Index: 32, Col: 32}},
				CloseTagRange: Range{From: Position{Index: 63, Line: 2}, To: Position{Index: 72, Line: 2, Col: 9}},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 72, Line: 2, Col: 9},
				},
			},
		},
		{
			name: "regexp expressions",
			input: `<script>
const result = call(1000 / 10, {{ data }}, 1000 / 10);
</script>`,
			expected: &ScriptElement{
				Contents: []ScriptContents{
					NewScriptContentsScriptCode("\nconst result = call(1000 / 10, "),
					NewScriptContentsGo(&GoCode{
						Expression: Expression{
							Value: "data",
							Range: Range{
								From: Position{Index: 43, Line: 1, Col: 34},
								To:   Position{Index: 47, Line: 1, Col: 38},
							},
						},
						Range: Range{
							From: Position{Index: 40, Line: 1, Col: 31},
							To:   Position{Index: 50, Line: 1, Col: 41},
						},
					}, ScriptContentsContextExpression),
					NewScriptContentsScriptCode(", 1000 / 10);\n"),
				},
				OpenTagRange:  Range{To: Position{Index: 8, Col: 8}},
				CloseTagRange: Range{From: Position{Index: 64, Line: 2}, To: Position{Index: 73, Line: 2, Col: 9}},
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 73, Line: 2, Col: 9},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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

// scriptContentsContextEscape mirrors the Context-to-function mapping in
// generator.go, so that the test below can show what a given Context
// actually produces in rendered output, rather than asserting on the
// Context enum value itself.
func scriptContentsContextEscape(c ScriptContentsContext, value string) (string, error) {
	switch c {
	case ScriptContentsContextString:
		return runtime.ScriptContentInsideStringLiteral(value)
	case ScriptContentsContextTemplateLiteral:
		return runtime.ScriptContentInsideTemplateLiteral(value)
	default:
		return runtime.ScriptContentOutsideStringLiteral(value)
	}
}

// TestScriptElementRendersGoExpressionsWithPositionAppropriateEscaping covers
// GHSA-jq3r-rjqp-mwgj. Each case supplies the literal value "${x}" for every
// {{ }} expression in the input, and asserts the exact text that ends up in
// the rendered <script> body. "${x}" is used because it shows the difference
// a position's escaping makes: '$' and '{' have no special meaning inside an
// ordinary quoted string or a JSON-encoded value, but inside a backtick
// quoted template literal they must be hex-escaped, or the browser would
// evaluate "${x}" as a JavaScript substitution instead of treating it as text.
func TestScriptElementRendersGoExpressionsWithPositionAppropriateEscaping(t *testing.T) {
	const value = "${x}"
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "a value directly inside a backtick template literal is hex-escaped so ${x} cannot be evaluated",
			input:    "<script>var a = `{{ name }}`;</script>",
			expected: `var a = ` + "`" + `\u0024\u007bx\u007d` + "`" + `;`,
		},
		{
			name:     "a value inside a dollar-brace substitution is JSON encoded, so it is quoted rather than evaluated",
			input:    "<script>var a = `${ {{ name }} }`;</script>",
			expected: `var a = ` + "`" + `${ "${x}"}` + "`" + `;`,
		},
		{
			name:     "a value after a dollar-brace substitution has closed is template-literal text again",
			input:    "<script>var a = `${ {{ x }} }{{ y }}`;</script>",
			expected: `var a = ` + "`" + `${ "${x}"}\u0024\u007bx\u007d` + "`" + `;`,
		},
		{
			name:     "a value inside a double quoted string is left alone, since $ and { aren't special there",
			input:    `<script>var a = "{{ name }}";</script>`,
			expected: `var a = "${x}";`,
		},
		{
			name:     "a value outside any string literal is JSON encoded",
			input:    "<script>var a = {{ name }};</script>",
			expected: `var a = "${x}";`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := parse.NewInput(tt.input)
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
			for _, c := range se.Contents {
				if c.Value != nil {
					actual.WriteString(*c.Value)
					continue
				}
				escaped, err := scriptContentsContextEscape(c.Context, value)
				if err != nil {
					t.Fatalf("unexpected error escaping value: %v", err)
				}
				actual.WriteString(escaped)
			}
			if diff := cmp.Diff(tt.expected, actual.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestScriptElementRegexpParser(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expected   string
		expectedOK bool
	}{
		{
			name:       "no content is considered to be a comment",
			input:      `//`,
			expectedOK: false,
		},
		{
			name:       "must not be multiline",
			input:      "/div>\n</div>",
			expectedOK: false,
		},
		{
			name:       "match a single char",
			input:      `/a/`,
			expected:   `/a/`,
			expectedOK: true,
		},
		{
			name:       "match a simple regex",
			input:      `/a|b/`,
			expected:   `/a|b/`,
			expectedOK: true,
		},
		{
			name:       "match a complex regex",
			input:      `/a(b|c)*d{2,4}/`,
			expected:   `/a(b|c)*d{2,4}/`,
			expectedOK: true,
		},
		{
			name:       "match a regex with flags",
			input:      `/a/i`,
			expected:   `/a/i`,
			expectedOK: true,
		},
		{
			name:       "match a regex with multiple flags",
			input:      `/a/gmi`,
			expected:   `/a/gmi`,
			expectedOK: true,
		},
		{
			name:       "escaped slashes",
			input:      `/a\/b\/c/`,
			expected:   `/a\/b\/c/`,
			expectedOK: true,
		},
		{
			name:       "no match: missing closing slash",
			input:      `/a|b`,
			expected:   "",
			expectedOK: false,
		},
		{
			name:       "no match: missing opening slash",
			input:      `a|b/`,
			expected:   "",
			expectedOK: false,
		},
		{
			name:       "must not contain interpolated go expressions",
			input:      `/a{{ b }}/`,
			expected:   "",
			expectedOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := parse.NewInput(tt.input)
			result, ok, err := regexpLiteral.Parse(input)
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}
			if ok != tt.expectedOK {
				t.Fatalf("expected ok to be %v, got %v", tt.expectedOK, ok)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
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
