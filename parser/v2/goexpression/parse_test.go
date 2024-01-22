package goexpression

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIf(t *testing.T) {
	prefix := "if "
	suffixes := []string{
		"{\n<div>\nif true content\n\t</div>}",
	}
	tests := []testInput{
		{
			name:  "basic if",
			input: `true `,
		},
		{
			name:  "if function call",
			input: `pkg.Func() `,
		},
		{
			name:  "compound",
			input: "x := val(); x > 3 ",
		},
		{
			name:  "if multiple",
			input: `x && y && (!z) `,
		},
	}
	for _, test := range tests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, prefix, suffix, If))
		}
	}
}

func TestElse(t *testing.T) {
	prefix := "else "
	suffixes := []string{
		"{\n<div>\nelse content\n\t</div>}",
	}
	tests := []testInput{
		{
			name:  "else",
			input: ``,
		},
		{
			name:  "else with spacing",
			input: `    `,
		},
	}
	for _, test := range tests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, prefix, suffix, Else))
		}
	}
}

func TestElseIf(t *testing.T) {
	prefix := "else if "
	suffixes := []string{
		"{\n<div>\nelse content\n\t</div>}",
	}
	tests := []testInput{
		{
			name:  "boolean",
			input: `true `,
		},
		{
			name:  "boolean with spacing",
			input: `   true `,
		},
		{
			name:  "func",
			input: `pkg.Func() `,
		},
		{
			name:  "expression",
			input: "x > 3 ",
		},
		{
			name:  "multiple",
			input: `x && y && (!z) `,
		},
	}
	for _, test := range tests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, prefix, suffix, ElseIf))
		}
	}
}

func TestFor(t *testing.T) {
	prefix := "for "
	suffixes := []string{
		"{\n<div>\nloop content\n\t</div>}",
	}
	tests := []testInput{
		{
			name:  "three component",
			input: `i := 0; i < 100; i++ `,
		},
		{
			name:  "three component, empty",
			input: `; ; i++ `,
		},
		{
			name:  "while",
			input: `n < 5 `,
		},
		{
			name:  "infinite",
			input: ``,
		},
		{
			name:  "range with index",
			input: `k, v := range m `,
		},
		{
			name:  "range with key only",
			input: `k := range m `,
		},
		{
			name:  "channel receive",
			input: `x := range channel `,
		},
	}
	for _, test := range tests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, prefix, suffix, For))
		}
	}
}

func TestSwitch(t *testing.T) {
	prefix := "switch "
	suffixes := []string{
		"{\ncase 1:\n\t<div>\n\tcase 2:\n\t\t<div>\n\tdefault:\n\t\t<div>\n\t</div>}",
		"{\ndefault:\n\t<div>\n\t</div>}",
		"{\n}",
	}
	tests := []testInput{
		{
			name:  "switch",
			input: ``,
		},
		{
			name:  "switch with expression",
			input: `x `,
		},
		{
			name:  "switch with function call",
			input: `pkg.Func() `,
		},
		{
			name:  "type switch",
			input: `x := x.(type) `,
		},
	}
	for _, test := range tests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, prefix, suffix, Switch))
		}
	}
}

func TestCase(t *testing.T) {
	prefix := "case "
	suffixes := []string{
		":\n<div>\ncase 1 content\n\t</div>\n\tcase 3:",
		":\ndefault:\n\t<div>\n\t</div>}",
		":\n}",
	}
	tests := []testInput{
		{
			name:  "case",
			input: `1`,
		},
		{
			name:  "case with expression",
			input: `x > 3`,
		},
		{
			name:  "case with function call",
			input: `pkg.Func()`,
		},
		{
			name:  "case with multiple expressions",
			input: `x > 3, x < 4`,
		},
		{
			name:  "case with multiple expressions and default",
			input: `x > 3, x < 4, x == 5`,
		},
		{
			name:  "case with type switch",
			input: `bool`,
		},
	}
	for _, test := range tests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, prefix, suffix, Case))
		}
	}
}

func TestCaseDefault(t *testing.T) {
	prefix := "default"
	suffixes := []string{
		":\n<div>\ncase 1 content\n\t</div>\n\tcase 3:",
		":\ndefault:\n\t<div>\n\t</div>}",
		":\n}",
	}
	tests := []testInput{
		{
			name:  "default",
			input: ``,
		},
	}
	for _, test := range tests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, prefix, suffix, Case))
		}
	}
}

func TestExpression(t *testing.T) {
	prefix := ""
	suffixes := []string{
		"}",
	}
	tests := []testInput{
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
	}
	for _, test := range tests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, prefix, suffix, Expression))
		}
	}
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

type testInput struct {
	name        string
	input       string
	expectedErr error
}

type extractor func(content string) (start, end, length int, err error)

func run(test testInput, prefix, suffix string, e extractor) func(t *testing.T) {
	return func(t *testing.T) {
		src := prefix + test.input + suffix
		start, end, _, err := e(src)
		if test.expectedErr == nil && err != nil {
			t.Fatalf("expected nil error, got %v, %T", err, err)
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
