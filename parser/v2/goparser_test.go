package parser

import (
	"fmt"
	"strings"
	"testing"
)

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

func generateTemplFile(components, rowsPerComponent int, row string) string {
	var sb strings.Builder
	sb.WriteString("package main\n\n")
	for i := range components {
		fmt.Fprintf(&sb, "templ Component%d(name string, count int) {\n", i)
		for j := range rowsPerComponent {
			fmt.Fprintf(&sb, row, i, j)
		}
		sb.WriteString("}\n\n")
	}
	return sb.String()
}

func BenchmarkParseStringManyComponents(b *testing.B) {
	rows := []struct {
		name   string
		format string
	}{
		{name: "element", format: "\t<p>Some HTML content for component %d row %d.</p>\n"},
		{name: "if", format: "\tif count > %d {\n\t\t<p>Row %d.</p>\n\t}\n"},
		{name: "for range over slice literal", format: "\tfor range []string{\"%d\", \"%d\"} {\n\t\t<p>Row.</p>\n\t}\n"},
	}
	for _, row := range rows {
		for _, components := range []int{10, 40, 80} {
			src := generateTemplFile(components, 40, row.format)
			b.Run(fmt.Sprintf("row=%s/components=%d", row.name, components), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if _, err := ParseString(src); err != nil {
						b.Fatalf("unexpected error: %v", err)
					}
				}
			})
		}
	}
}
