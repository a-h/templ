package goexpression

import (
	"fmt"
	"go/token"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var ifTests = []testInput{
	{
		name:  "basic if",
		input: `true`,
	},
	{
		name:  "if function call",
		input: `pkg.Func()`,
	},
	{
		name:  "compound",
		input: "x := val(); x > 3",
	},
	{
		name:  "if multiple",
		input: `x && y && (!z)`,
	},
	{
		name:  "if with function literal",
		input: `f := func() bool { return true }; f()`,
	},
	{
		name:  "if with slice literal",
		input: `items := []string{"a", "b"}; len(items) > 0`,
	},
	{
		name:  "if with composite literal in parentheses",
		input: `v == (T{A: 1})`,
	},
}

func TestIf(t *testing.T) {
	prefix := "if "
	suffixes := []string{
		"{\n<div>\nif true content\n\t</div>}",
		" {\n<div>\nif true content\n\t</div>}",
	}
	for _, test := range ifTests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, prefix, suffix, If))
		}
	}
}

func FuzzIf(f *testing.F) {
	suffixes := []string{
		"{\n<div>\nif true content\n\t</div>}",
		" {\n<div>\nif true content\n\t</div>}",
	}
	for _, test := range ifTests {
		for _, suffix := range suffixes {
			f.Add("if " + test.input + suffix)
		}
	}
	f.Fuzz(func(t *testing.T, src string) {
		start, end, err := If([]byte(src))
		if err != nil {
			t.Skip()
			return
		}
		panicIfInvalid(src, start, end)
	})
}

func panicIfInvalid(src string, start, end int) {
	_ = src[start:end]
}

var forTests = []testInput{
	{
		name:  "three component",
		input: `i := 0; i < 100; i++`,
	},
	{
		name:  "three component, empty",
		input: `; ; i++`,
	},
	{
		name:  "while",
		input: `n < 5`,
	},
	{
		name:  "infinite",
		input: ``,
	},
	{
		name:  "range with index",
		input: `k, v := range m`,
	},
	{
		name:  "range with key only",
		input: `k := range m`,
	},
	{
		name:  "channel receive",
		input: `x := range channel`,
	},
	{
		name:  "range over slice literal",
		input: `_, v := range []string{"a", "b"}`,
	},
	{
		name:  "range over function call with function literal",
		input: `v := range filter(items, func(s string) bool { return s != "" })`,
	},
}

func TestFor(t *testing.T) {
	prefix := "for "
	suffixes := []string{
		" {\n<div>\nloop content\n\t</div>}",
	}
	for _, test := range forTests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, prefix, suffix, For))
		}
	}
}

func FuzzFor(f *testing.F) {
	suffixes := []string{
		"",
		" {",
		" {}",
		" {\n<div>\nloop content\n\t</div>}",
	}
	for _, test := range forTests {
		for _, suffix := range suffixes {
			f.Add("for " + test.input + suffix)
		}
	}
	f.Fuzz(func(t *testing.T, src string) {
		start, end, err := For([]byte(src))
		if err != nil {
			t.Skip()
			return
		}
		panicIfInvalid(src, start, end)
	})
}

var switchTests = []testInput{
	{
		name:  "switch",
		input: ``,
	},
	{
		name:  "switch with expression",
		input: `x`,
	},
	{
		name:  "switch with function call",
		input: `pkg.Func()`,
	},
	{
		name:  "type switch",
		input: `x := x.(type)`,
	},
	{
		name:  "switch with function literal",
		input: `f := func() int { return 1 }; f()`,
	},
}

func TestSwitch(t *testing.T) {
	prefix := "switch "
	suffixes := []string{
		" {\ncase 1:\n\t<div>\n\tcase 2:\n\t\t<div>\n\tdefault:\n\t\t<div>\n\t</div>}",
		" {\ndefault:\n\t<div>\n\t</div>}",
		" {\n}",
	}
	for _, test := range switchTests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, prefix, suffix, Switch))
		}
	}
}

func FuzzSwitch(f *testing.F) {
	suffixes := []string{
		"",
		" {",
		" {}",
		" {\n<div>\nloop content\n\t</div>}",
	}
	for _, test := range switchTests {
		for _, suffix := range suffixes {
			f.Add(test.input + suffix)
		}
	}
	f.Fuzz(func(t *testing.T, s string) {
		src := "switch " + s
		start, end, err := For([]byte(src))
		if err != nil {
			t.Skip()
			return
		}
		panicIfInvalid(src, start, end)
	})
}

var caseTests = []testInput{
	{
		name:  "case",
		input: `case 1:`,
	},
	{
		name:  "case with expression",
		input: `case x > 3:`,
	},
	{
		name:  "case with function call",
		input: `case pkg.Func():`,
	},
	{
		name:  "case with multiple expressions",
		input: `case x > 3, x < 4:`,
	},
	{
		name:  "case with multiple expressions and default",
		input: `case x > 3, x < 4, x == 5:`,
	},
	{
		name:  "case with type switch",
		input: `case bool:`,
	},
}

func TestCase(t *testing.T) {
	suffixes := []string{
		"\n<div>\ncase 1 content\n\t</div>\n\tcase 3:",
		"\ndefault:\n\t<div>\n\t</div>}",
		"\n}",
	}
	for _, test := range caseTests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, "", suffix, Case))
		}
	}
}

func FuzzCaseStandard(f *testing.F) {
	suffixes := []string{
		"",
		"\n<div>\ncase 1 content\n\t</div>\n\tcase 3:",
		"\ndefault:\n\t<div>\n\t</div>}",
		"\n}",
	}
	for _, test := range caseTests {
		for _, suffix := range suffixes {
			f.Add(test.input + suffix)
		}
	}
	f.Fuzz(func(t *testing.T, src string) {
		start, end, err := Case([]byte(src))
		if err != nil {
			t.Skip()
			return
		}
		panicIfInvalid(src, start, end)
	})
}

func TestCaseDefault(t *testing.T) {
	suffixes := []string{
		"\n<div>\ncase 1 content\n\t</div>\n\tcase 3:",
		"\ncase:\n\t<div>\n\t</div>}",
		"\n}",
	}
	tests := []testInput{
		{
			name:  "default",
			input: `default:`,
		},
		{
			name:  "default with padding",
			input: `default :`,
		},
		{
			name:  "default with padding",
			input: `default   :`,
		},
	}
	for _, test := range tests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, "", suffix, Case))
		}
	}
}

func FuzzCaseDefault(f *testing.F) {
	suffixes := []string{
		"",
		" ",
		"\n<div>\ncase 1 content\n\t</div>\n\tcase 3:",
		"\ncase:\n\t<div>\n\t</div>}",
		"\n}",
	}
	for _, suffix := range suffixes {
		f.Add("default:" + suffix)
	}
	f.Fuzz(func(t *testing.T, src string) {
		start, end, err := Case([]byte(src))
		if err != nil {
			t.Skip()
			return
		}
		panicIfInvalid(src, start, end)
	})
}

var expressionTests = []testInput{
	{
		name:  "string literal",
		input: `"hello"`,
	},
	{
		name:  "string literal with escape",
		input: `"hello\n"`,
	},
	{
		name:  "backtick string literal",
		input: "`hello`",
	},
	{
		name:  "backtick string literal containing double quote",
		input: "`hello" + `"` + `world` + "`",
	},
	{
		name:  "function call in package",
		input: `components.Other()`,
	},
	{
		name:  "slice index call",
		input: `components[0].Other()`,
	},
	{
		name:  "map index function call",
		input: `components["name"].Other()`,
	},
	{
		name:  "function literal",
		input: `components["name"].Other(func() bool { return true })`,
	},
	{
		name: "multiline function call",
		input: `component(map[string]string{
				"namea": "name_a",
			  "nameb": "name_b",
			})`,
	},
	{
		name:  "call with braces and brackets",
		input: `templates.New(test{}, other())`,
	},
	{
		name:  "bare variable",
		input: `component`,
	},
	{
		name:  "boolean expression",
		input: `direction == "newest"`,
	},
	{
		name:  "boolean expression with parens",
		input: `len(data.previousPageUrl) == 0`,
	},
	{
		name:  "string concat",
		input: `direction + "newest"`,
	},
	{
		name: "function call",
		input: `SplitRule(types.GroupMember{
    UserID:   uuid.NewString(),
    Username: "user me",
}, []types.GroupMember{
    {
    UserID:   uuid.NewString(),
    Username: "user 1",
    },
})`,
	},
}

func TestExpression(t *testing.T) {
	prefix := ""
	suffixes := []string{
		"",
		"}",
		"\t}",
		"  }",
	}
	for _, test := range expressionTests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, prefix, suffix, Expression))
		}
	}
}

var templExpressionTests = []testInput{
	{
		name:  "function call in package",
		input: `components.Other()`,
	},
	{
		name:  "slice index call",
		input: `components[0].Other()`,
	},
	{
		name:  "map index function call",
		input: `components["name"].Other()`,
	},
	{
		name: "multiline chain call",
		input: `components.
	Other()`,
	},
	{
		name:  "map index function call backtick literal",
		input: "components[`name" + `"` + "`].Other()",
	},
	{
		name:  "function literal",
		input: `components["name"].Other(func() bool { return true })`,
	},
	{
		name: "multiline function call",
		input: `component(map[string]string{
				"namea": "name_a",
			  "nameb": "name_b",
			})`,
	},
	{
		name: "function call with slice of complex types",
		input: `tabs([]TabData{
  {Name: "A"},
  {Name: "B"},
})`,
	},
	{
		name: "function call with slice of explicitly named complex types",
		input: `tabs([]TabData{
  TabData{Name: "A"},
  TabData{Name: "B"},
})`,
	},
	{
		name:  "function call with empty slice of strings",
		input: `Inner([]string{})`,
	},
	{
		name:  "function call with empty slice of maps",
		input: `Inner([]map[string]any{})`,
	},
	{
		name:  "function call with empty slice of anon structs",
		input: `Inner([]map[string]struct{}{})`,
	},
	{
		name: "function call with slice of pointers to complex types",
		input: `tabs([]*TabData{
  &{Name: "A"},
  &{Name: "B"},
})`,
	},
	{
		name: "function call with slice of pointers to explicitly named complex types",
		input: `tabs([]*TabData{
  &TabData{Name: "A"},
  &TabData{Name: "B"},
})`,
	},
	{
		name: "function call with array of explicit length",
		input: `tabs([2]TabData{
  {Name: "A"},
  {Name: "B"},
})`,
	},
	{
		name: "function call with array of inferred length",
		input: `tabs([...]TabData{
  {Name: "A"},
  {Name: "B"},
})`,
	},
	{
		name: "function call with function arg",
		input: `componentA(func(y []int) string {
		return "hi"
	})`,
	},
	{
		name: "function call with function called arg",
		input: `componentA(func(y []int) string {
		return "hi"
	}())`,
	},
	{
		name:  "call with braces and brackets",
		input: `templates.New(test{}, other())`,
	},
	{
		name:  "generic call",
		input: `templates.New(toString[[]int](data))`,
	},
	{
		name:  "struct method call",
		input: `typeName{}.Method()`,
	},
	{
		name:  "struct method call in other package",
		input: "layout.DefaultLayout{}.Compile()",
	},
	{
		name:  "bare variable",
		input: `component`,
	},
}

func TestTemplExpression(t *testing.T) {
	prefix := ""
	suffixes := []string{
		"",
		"}",
		"\t}",
		"  }",
		"</div>",
		"<p>/</p>",
		" just some text",
		" { <div>Child content</div> }",
	}
	for _, test := range templExpressionTests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, prefix, suffix, TemplExpression))
		}
	}
}

func FuzzTemplExpression(f *testing.F) {
	suffixes := []string{
		"",
		" }",
		" }}</a>\n}",
		"...",
	}
	for _, test := range expressionTests {
		for _, suffix := range suffixes {
			f.Add(test.input + suffix)
		}
	}
	f.Fuzz(func(t *testing.T, s string) {
		src := "switch " + s
		start, end, err := TemplExpression([]byte(src))
		if err != nil {
			t.Skip()
			return
		}
		panicIfInvalid(src, start, end)
	})
}

func FuzzExpression(f *testing.F) {
	suffixes := []string{
		"",
		" }",
		" }}</a>\n}",
		"...",
	}
	for _, test := range expressionTests {
		for _, suffix := range suffixes {
			f.Add(test.input + suffix)
		}
	}
	f.Fuzz(func(t *testing.T, s string) {
		src := "switch " + s
		start, end, err := Expression([]byte(src))
		if err != nil {
			t.Skip()
			return
		}
		panicIfInvalid(src, start, end)
	})
}

var sliceArgsTests = []testInput{
	{
		name:  "no input",
		input: ``,
	},
	{
		name:  "single input",
		input: `nil`,
	},
	{
		name:  "inputs to function call",
		input: `a, b, "c"`,
	},
	{
		name:  "function call in package",
		input: `components.Other()`,
	},
	{
		name:  "slice index call",
		input: `components[0].Other()`,
	},
	{
		name:  "map index function call",
		input: `components["name"].Other()`,
	},
	{
		name:  "function literal",
		input: `components["name"].Other(func() bool { return true })`,
	},
	{
		name: "multiline function call",
		input: `component(map[string]string{
				"namea": "name_a",
			  "nameb": "name_b",
			})`,
	},
	{
		name:  "package name, but no variable or function",
		input: `fmt.`,
	},
}

func TestSliceArgs(t *testing.T) {
	suffixes := []string{
		"",
		"}",
		"}</a>\n}\nvar x = []struct {}{}",
	}
	for _, test := range sliceArgsTests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), func(t *testing.T) {
				expr, err := SliceArgs([]byte(test.input + suffix))
				if err != nil {
					t.Errorf("failed to parse slice args: %v", err)
				}
				if diff := cmp.Diff(test.input, expr); diff != "" {
					t.Error(diff)
				}
			})
		}
	}
}

func FuzzSliceArgs(f *testing.F) {
	suffixes := []string{
		"",
		"}",
		" }",
		"}</a>\n}\nvar x = []struct {}{}",
	}
	for _, test := range sliceArgsTests {
		for _, suffix := range suffixes {
			f.Add(test.input + suffix)
		}
	}
	f.Fuzz(func(t *testing.T, s string) {
		_, err := SliceArgs([]byte(s))
		if err != nil {
			t.Skip()
			return
		}
	})
}

func TestChildren(t *testing.T) {
	prefix := ""
	suffixes := []string{
		" }",
		" } <div>Other content</div>",
		"", // End of file.
	}
	tests := []testInput{
		{
			name:  "children",
			input: `children...`,
		},
		{
			name:  "function",
			input: `components.Spread()...`,
		},
		{
			name:  "alternative variable",
			input: `components...`,
		},
		{
			name:  "index",
			input: `groups[0]...`,
		},
		{
			name:  "map",
			input: `components["name"]...`,
		},
		{
			name:  "map func key",
			input: `components[getKey(ctx)]...`,
		},
	}
	for _, test := range tests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, prefix, suffix, Expression))
		}
	}
}

var funcTests = []testInput{
	{
		name:  "void func",
		input: `myfunc()`,
	},
	{
		name:  "receiver func",
		input: `(r recv) myfunc()`,
	},
	{
		name:  "variadic parameter",
		input: `myfunc(items ...string)`,
	},
	{
		name:  "type parameters",
		input: `myfunc[T, U any](t T, u U)`,
	},
	{
		name:  "struct parameter",
		input: `myfunc(v struct{ X int })`,
	},
	{
		name:  "func parameter",
		input: `myfunc(fn func(a int) error)`,
	},
	{
		name:  "map of struct parameter",
		input: `myfunc(m map[string]struct{ Y int })`,
	},
	{
		name:  "generic struct parameter",
		input: `myfunc[T any](v struct{ Items []T })`,
	},
}

func FuzzFuncs(f *testing.F) {
	suffixes := []string{
		"",
		"}",
		" }",
		"}</a>\n}\nvar x = []struct {}{}",
	}
	for _, test := range funcTests {
		for _, suffix := range suffixes {
			f.Add(test.input + suffix)
		}
	}
	f.Fuzz(func(t *testing.T, s string) {
		_, _, err := Func([]byte(s))
		if err != nil {
			t.Skip()
			return
		}
	})
}

func TestFunc(t *testing.T) {
	suffixes := []string{
		"",
		"}",
		"}\nvar x = []struct {}{}",
		"}\nfunc secondFunc() {}",
		" {\n\t<p>{ \"}\" }</p>\n}\n\ntempl next() {\n\t<p>Next</p>\n}",
	}
	for _, test := range funcTests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), func(t *testing.T) {
				name, expr, err := Func([]byte(test.input + suffix))
				if err != nil {
					t.Errorf("failed to parse slice args: %v", err)
				}
				if diff := cmp.Diff(test.input, expr); diff != "" {
					t.Error(diff)
				}
				if diff := cmp.Diff("myfunc", name); diff != "" {
					t.Error(diff)
				}
			})
		}
	}
}

type testInput struct {
	name        string
	input       string
	expectedErr error
}

type extractor func(content []byte) (start, end int, err error)

func run(test testInput, prefix, suffix string, e extractor) func(t *testing.T) {
	return func(t *testing.T) {
		src := prefix + test.input + suffix
		start, end, err := e([]byte(src))
		if test.expectedErr == nil && err != nil {
			t.Fatalf("expected nil error got error type %T: %v", err, err)
		}
		if test.expectedErr != nil && err == nil {
			t.Fatalf("expected err %q, got %v", test.expectedErr.Error(), err)
		}
		if test.expectedErr != nil && err != nil && test.expectedErr.Error() != err.Error() {
			t.Fatalf("expected err %q, got %q", test.expectedErr.Error(), err.Error())
		}
		actual := src[start:end]
		if diff := cmp.Diff(test.input, actual); diff != "" {
			t.Error(diff)
		}
	}
}

func TestFindTerminatingToken(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{name: "an if statement block opens at the first brace", header: `if x `},
		{name: "a brace in a string does not open the block", header: `if a == "{" `},
		{name: "a brace in a comment does not open the block", header: `if a /* { */ == b `},
		{name: "a composite literal in parentheses does not open the block", header: `if x := (struct{}{}); x == y `},
		{name: "a function literal does not open the block", header: `if f := func() bool { return true }; f() `},
		{name: "a map literal does not open the block", header: `if x := map[string]int{"a": 1}; x["a"] == 1 `},
		{name: "a slice literal in a range clause does not open the block", header: `for _, v := range []string{"a", "b"} `},
		{name: "a struct slice literal in a range clause does not open the block", header: `for _, v := range []struct{ A int }{{1}} `},
		{name: "a function literal argument does not open the block", header: `for v := range filter(items, func(s string) bool { return s != "" }) `},
		{name: "a switch without a tag opens at the first brace", header: `switch `},
		{name: "a type switch opens after the type assertion", header: `switch x := x.(type) `},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, offset, err := findTerminatingToken(containerPrefix, []byte(tt.header+"{\n\t<p>{ v }</p>\n}\n<div>"), token.LBRACE, "{}\n}")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(len(tt.header), offset); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestFindTerminatingTokenReturnsErrorWhenHeaderIsInvalid(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "a syntax error before the brace is returned", content: "if x == == y {\n\t<p>A</p>\n}"},
		{name: "an illegal token before the brace is returned", content: "if x # y {\n\t<p>A</p>\n}"},
		{name: "content without a brace is not found", content: "if x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := findTerminatingToken(containerPrefix, []byte(tt.content), token.LBRACE, "{}\n}")
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}
}

func TestFindTerminatingTokenFindsClosingBrace(t *testing.T) {
	tests := []struct {
		name string
		args string
	}{
		{name: "a variable closes at the first brace", args: ` name `},
		{name: "a brace in a string does not close the expression", args: ` "}" `},
		{name: "a function literal does not close the expression", args: ` f(func() string { return "a" }) `},
		{name: "a composite literal does not close the expression", args: ` T{A: 1}.String() `},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, offset, err := findTerminatingToken("package main\nvar templ_args = []any{", []byte(tt.args+"}</p>\n}\n<div>"), token.RBRACE, "}")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(len(tt.args), offset); diff != "" {
				t.Error(diff)
			}
		})
	}
}
