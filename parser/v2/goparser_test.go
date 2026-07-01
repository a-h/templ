package parser

import (
	"fmt"
	"strings"
	"testing"
)

func TestExtractFuncDeclSignature(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected string
	}{
		{
			name:     "a simple declaration is extracted up to the body brace",
			src:      "Component() {\n\t<p>Hello</p>\n}",
			expected: "Component() ",
		},
		{
			name:     "parameters are included in the extracted declaration",
			src:      "Component(name string, count int) {\n}",
			expected: "Component(name string, count int) ",
		},
		{
			name:     "a variadic parameter is included",
			src:      "Component(items ...string) {\n}",
			expected: "Component(items ...string) ",
		},
		{
			name:     "a single type parameter is included",
			src:      "Component[T any]() {\n}",
			expected: "Component[T any]() ",
		},
		{
			name:     "multiple type parameters are included",
			src:      "Component[T, U any](t T, u U) {\n}",
			expected: "Component[T, U any](t T, u U) ",
		},
		{
			name:     "a struct-typed parameter does not terminate scanning at its opening brace",
			src:      "Component(v struct{ X int }) {\n}",
			expected: "Component(v struct{ X int }) ",
		},
		{
			name:     "an empty struct-typed parameter does not terminate scanning",
			src:      "Component(v struct{}) {\n}",
			expected: "Component(v struct{}) ",
		},
		{
			name:     "a func-typed parameter is included",
			src:      "Component(fn func(a int) error) {\n}",
			expected: "Component(fn func(a int) error) ",
		},
		{
			name:     "a map-typed parameter with braces does not terminate scanning",
			src:      "Component(m map[string]struct{ Y int }) {\n}",
			expected: "Component(m map[string]struct{ Y int }) ",
		},
		{
			name:     "a generic struct-typed parameter combining brackets and braces is handled",
			src:      "Component[T any](v struct{ Items []T }) {\n}",
			expected: "Component[T any](v struct{ Items []T }) ",
		},
		{
			name:     "a comment before the body brace does not terminate scanning",
			src:      "Component() /* body */ {\n}",
			expected: "Component() /* body */ ",
		},
		{
			name:     "a brace inside a string literal parameter default does not terminate scanning",
			src:      "Component(s string) {\n\t<p>{ \"}\" }</p>\n}",
			expected: "Component(s string) ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractFuncDeclSignature(tt.src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("got %q, expected %q", got, tt.expected)
			}
		})
	}
}

func TestExtractFuncDeclSignatureReturnsErrorWhenBodyBraceIsMissing(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{name: "an empty source has no body brace", src: ""},
		{name: "a signature without a body has no body brace", src: "Component()"},
		{name: "an unclosed parameter list has no body brace", src: "Component(name string"},
		{name: "an illegal token fails before the body brace is reached", src: "Component(name #string) {\n}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extractFuncDeclSignature(tt.src)
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}
}

func TestParseStringExtractsComplexTemplSignatures(t *testing.T) {
	tests := []struct {
		name         string
		src          string
		expectedExpr string
	}{
		{
			name:         "a generic component signature is extracted",
			src:          "package main\n\ntempl Component[T any](items []T) {\n\t<p>Hello</p>\n}\n",
			expectedExpr: "Component[T any](items []T)",
		},
		{
			name:         "a struct-typed parameter is extracted",
			src:          "package main\n\ntempl Component(v struct{ X int }) {\n\t<p>Hello</p>\n}\n",
			expectedExpr: "Component(v struct{ X int })",
		},
		{
			name:         "a func-typed parameter is extracted",
			src:          "package main\n\ntempl Component(fn func(a int) error) {\n\t<p>Hello</p>\n}\n",
			expectedExpr: "Component(fn func(a int) error)",
		},
		{
			name:         "multiple type parameters are extracted",
			src:          "package main\n\ntempl Component[T, U any](t T, u U) {\n\t<p>Hello</p>\n}\n",
			expectedExpr: "Component[T, U any](t T, u U)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf, err := ParseString(tt.src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var found bool
			for _, n := range tf.Nodes {
				tmpl, ok := n.(*HTMLTemplate)
				if !ok {
					continue
				}
				found = true
				if tmpl.Expression.Value != tt.expectedExpr {
					t.Errorf("got expression %q, expected %q", tmpl.Expression.Value, tt.expectedExpr)
				}
			}
			if !found {
				t.Fatal("expected an HTMLTemplate node, found none")
			}
		})
	}
}

func FuzzExtractFuncDeclSignature(f *testing.F) {
	seeds := []string{
		"Component() {\n}",
		"Component[T any](v struct{ X int }) {\n}",
		"Component(fn func(a int) error) {\n}",
		"Component(",
		"",
		"{}",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, src string) {
		decl, err := extractFuncDeclSignature(src)
		if err != nil {
			return
		}
		if !strings.HasPrefix(src, decl) {
			t.Fatalf("returned declaration %q is not a prefix of source %q", decl, src)
		}
		if len(decl) >= len(src) || src[len(decl)] != '{' {
			t.Fatalf("returned declaration %q must be immediately followed by an opening brace in source %q", decl, src)
		}
	})
}

func BenchmarkExtractFuncDeclSignature(b *testing.B) {
	src := "Component[T any](name string, items []T, v struct{ X int }) {\n\t<p>Body</p>\n}"
	b.ReportAllocs()
	for b.Loop() {
		if _, err := extractFuncDeclSignature(src); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func generateTemplFile(components, rowsPerComponent int) string {
	var sb strings.Builder
	sb.WriteString("package main\n\n")
	for i := range components {
		fmt.Fprintf(&sb, "templ Component%d(name string, count int) {\n", i)
		for j := range rowsPerComponent {
			fmt.Fprintf(&sb, "\t<p>Some HTML content for component %d row %d.</p>\n", i, j)
		}
		sb.WriteString("}\n\n")
	}
	return sb.String()
}

func BenchmarkParseStringManyComponents(b *testing.B) {
	for _, components := range []int{10, 40, 80} {
		src := generateTemplFile(components, 40)
		b.Run(fmt.Sprintf("components=%d", components), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, err := ParseString(src); err != nil {
					b.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}
