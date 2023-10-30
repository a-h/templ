package templ_internal

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var cssInputs = []any{
	[]string{"a", "b"},    // []string
	"c",                   // string
	ConstantCSSClass("d"), // ConstantCSSClass
	ComponentCSSClass{ID: "e", Class: ".e{color:red}"}, // ComponentCSSClass
	map[string]bool{"f": true, "ff": false},            // map[string]bool
	KeyValue[string, bool]{"g", true},                  // KeyValue[string, bool]
	KeyValue[string, bool]{"gg", false},                // KeyValue[string, bool]
	[]KeyValue[string, bool]{
		{"h", true},
		{"hh", false},
	}, // []KeyValue[string, bool]
	KeyValue[CSSClass, bool]{ConstantCSSClass("i"), true},   // KeyValue[CSSClass, bool]
	KeyValue[CSSClass, bool]{ConstantCSSClass("ii"), false}, // KeyValue[CSSClass, bool]
	KeyValue[ComponentCSSClass, bool]{ComponentCSSClass{
		ID:    "j",
		Class: ".j{color:red}",
	}, true}, // KeyValue[ComponentCSSClass, bool]
	KeyValue[ComponentCSSClass, bool]{ComponentCSSClass{
		ID:    "jj",
		Class: ".jj{color:red}",
	}, false}, // KeyValue[ComponentCSSClass, bool]
	CSSClasses{ConstantCSSClass("k")},                             // CSSClasses
	func() CSSClass { return ConstantCSSClass("l") },              // func() CSSClass
	CSSClass(ConstantCSSClass("m")),                               // CSSClass
	customClass{name: "n"},                                        // CSSClass
	KeyValue[ConstantCSSClass, bool]{ConstantCSSClass("o"), true}, // KeyValue[ConstantCSSClass, bool]
	[]KeyValue[ConstantCSSClass, bool]{
		{ConstantCSSClass("p"), true},
		{ConstantCSSClass("pp"), false},
	}, // []KeyValue[ConstantCSSClass, bool]
}

type customClass struct {
	name string
}

func (cc customClass) ClassName() string {
	return cc.name
}

func TestRenderCSS(t *testing.T) {
	c1 := ComponentCSSClass{
		ID:    "c1",
		Class: ".c1{color:red}",
	}
	c2 := ComponentCSSClass{
		ID:    "c2",
		Class: ".c2{color:blue}",
	}

	tests := []struct {
		name     string
		toIgnore []any
		toRender []any
		expected string
	}{
		{
			name:     "if none are ignored, everything is rendered",
			toIgnore: nil,
			toRender: []any{c1, c2},
			expected: `<style type="text/css">.c1{color:red}.c2{color:blue}</style>`,
		},
		{
			name: "if something outside the expected is ignored, if has no effect",
			toIgnore: []any{
				ComponentCSSClass{
					ID:    "c3",
					Class: SafeCSS(".c3{color:yellow}"),
				},
			},
			toRender: []any{c1, c2},
			expected: `<style type="text/css">.c1{color:red}.c2{color:blue}</style>`,
		},
		{
			name:     "if one is ignored, it's not rendered",
			toIgnore: []any{c1},
			toRender: []any{c1, c2},
			expected: `<style type="text/css">.c2{color:blue}</style>`,
		},
		{
			name: "if all are ignored, not even style tags are rendered",
			toIgnore: []any{
				c1,
				c2,
				ComponentCSSClass{
					ID:    "c3",
					Class: SafeCSS(".c3{color:yellow}"),
				},
			},
			toRender: []any{c1, c2},
			expected: ``,
		},
		{
			name:     "CSS classes are rendered",
			toIgnore: nil,
			toRender: cssInputs,
			expected: `<style type="text/css">.e{color:red}.j{color:red}</style>`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			b := new(bytes.Buffer)

			// Render twice, reusing the same context so that there's a memory of which classes have been rendered.
			ctx = InitializeContext(ctx)
			err := RenderCSSItems(ctx, b, tt.toIgnore...)
			if err != nil {
				t.Fatalf("failed to render initial CSS: %v", err)
			}

			// Now render again to check that only the expected classes were rendered.
			b.Reset()
			err = RenderCSSItems(ctx, b, tt.toRender...)
			if err != nil {
				t.Fatalf("failed to render CSS: %v", err)
			}

			if diff := cmp.Diff(tt.expected, b.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestClassSanitization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    `safe`,
			expected: `safe`,
		},
		{
			input:    `safe-name`,
			expected: "safe-name",
		},
		{
			input:    `safe_name`,
			expected: "safe_name",
		},
		{
			input:    `!unsafe`,
			expected: "--templ-css-class-safe-name",
		},
		{
			input:    `</style>`,
			expected: "--templ-css-class-safe-name",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			actual := Class(tt.input)
			if actual.ClassName() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual.ClassName())
			}
		})
	}
}

func TestClassesFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    []any
		expected string
	}{
		{
			name:     "safe constants are allowed",
			input:    []any{"a", "b", "c"},
			expected: "a b c",
		},
		{
			name:     "unsafe constants are filtered",
			input:    []any{"</style>", "b", "</style>"},
			expected: "--templ-css-class-safe-name b",
		},
		{
			name:     "legacy CSS types are supported",
			input:    []any{"a", SafeClass("b"), Class("c")},
			expected: "a b c",
		},
		{
			name: "CSS components are included in the output",
			input: []any{
				ComponentCSSClass{ID: "classA", Class: SafeCSS(".classA{background-color:white;}")},
				ComponentCSSClass{ID: "classB", Class: SafeCSS(".classB{background-color:green;}")},
				"c",
			},
			expected: "classA classB c",
		},
		{
			name: "optional classes can be applied with expressions",
			input: []any{
				"a",
				ComponentCSSClass{ID: "classA", Class: SafeCSS(".classA{background-color:white;}")},
				ComponentCSSClass{ID: "classB", Class: SafeCSS(".classB{background-color:green;}")},
				"c",
				map[string]bool{
					"a":      false,
					"classA": false,
					"classB": false,
					"c":      true,
					"d":      false,
				},
			},
			expected: "c",
		},
		{
			name: "unknown types for classes get rendered as --templ-css-class-unknown-type",
			input: []any{
				123,
				map[string]string{"test": "no"},
				false,
				"c",
			},
			expected: "--templ-css-class-unknown-type c",
		},
		{
			name: "string arrays are supported",
			input: []any{
				[]string{"a", "b", "c"},
				"d",
			},
			expected: "a b c d",
		},
		{
			name: "string arrays are checked for unsafe class names",
			input: []any{
				[]string{"a", "b", "c </style>"},
				"d",
			},
			expected: "a b c --templ-css-class-safe-name d",
		},
		{
			name: "strings are broken up",
			input: []any{
				"a </style>",
			},
			expected: "a --templ-css-class-safe-name",
		},
		{
			name: "if a CSSClasses is passed in, the nested CSSClasses are extracted",
			input: []any{
				Classes(
					"a",
					SafeClass("b"),
					Class("c"),
					ComponentCSSClass{
						ID:    "d",
						Class: "{}",
					},
				),
			},
			expected: "a b c d",
		},
		{
			name: "kv types can be used to show or hide classes",
			input: []any{
				"a",
				KeyValue[string, bool]{"b", true},
				"c",
				KeyValue[string, bool]{"c", false},
				KeyValue[CSSClass, bool]{SafeClass("d"), true},
				KeyValue[CSSClass, bool]{SafeClass("e"), false},
			},
			expected: "a b d",
		},
		{
			name: "an array of KV types can be used to show or hide classes",
			input: []any{
				"a",
				"c",
				[]KeyValue[string, bool]{
					{"b", true},
					{"c", false},
					{"d", true},
				},
			},
			expected: "a b d",
		},
		{
			name: "the brackets on component CSS function calls can be elided",
			input: []any{
				func() CSSClass {
					return ComponentCSSClass{
						ID:    "a",
						Class: "",
					}
				},
			},
			expected: "a",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := Classes(test.input...).String()
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestRenderScriptItems(t *testing.T) {
	s1 := ComponentScript{
		Name:     "s1",
		Function: "function s1() { return 'hello1'; }",
	}
	s2 := ComponentScript{
		Name:     "s2",
		Function: "function s2() { return 'hello2'; }",
	}
	tests := []struct {
		name     string
		toIgnore []ComponentScript
		toRender []ComponentScript
		expected string
	}{
		{
			name:     "if none are ignored, everything is rendered",
			toIgnore: nil,
			toRender: []ComponentScript{s1, s2},
			expected: `<script type="text/javascript">` + s1.Function + s2.Function + `</script>`,
		},
		{
			name: "if something outside the expected is ignored, if has no effect",
			toIgnore: []ComponentScript{
				{
					Name:     "s3",
					Function: "function s3() { return 'hello3'; }",
				},
			},
			toRender: []ComponentScript{s1, s2},
			expected: `<script type="text/javascript">` + s1.Function + s2.Function + `</script>`,
		},
		{
			name:     "if one is ignored, it's not rendered",
			toIgnore: []ComponentScript{s1},
			toRender: []ComponentScript{s1, s2},
			expected: `<script type="text/javascript">` + s2.Function + `</script>`,
		},
		{
			name: "if all are ignored, not even style tags are rendered",
			toIgnore: []ComponentScript{
				s1,
				s2,
				{
					Name:     "s3",
					Function: "function s3() { return 'hello3'; }",
				},
			},
			toRender: []ComponentScript{s1, s2},
			expected: ``,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			b := new(bytes.Buffer)

			// Render twice, reusing the same context so that there's a memory of which classes have been rendered.
			ctx = InitializeContext(ctx)
			err := RenderScriptItems(ctx, b, tt.toIgnore...)
			if err != nil {
				t.Fatalf("failed to render initial scripts: %v", err)
			}

			// Now render again to check that only the expected classes were rendered.
			b.Reset()
			err = RenderScriptItems(ctx, b, tt.toRender...)
			if err != nil {
				t.Fatalf("failed to render scripts: %v", err)
			}

			if diff := cmp.Diff(tt.expected, b.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}
