package format

import (
	"fmt"
	"strings"
	"testing"

	"github.com/a-h/templ/internal/prettier"
)

// createBenchmarkTemplate returns a templ file with componentCount components, each with
// constant attributes, plus a page that contains a style element and a script element.
func createBenchmarkTemplate(componentCount int) string {
	var sb strings.Builder
	sb.WriteString("package benchmark\n\n")
	for i := range componentCount {
		fmt.Fprintf(&sb, "templ C%d() {\n", i)
		fmt.Fprintf(&sb, "\t<div class=\"flex items-center p-4\" id=\"c%d\">\n", i)
		fmt.Fprintf(&sb, "\t\t<p style=\"color: red\">Component %d</p>\n", i)
		sb.WriteString("\t</div>\n}\n\n")
	}
	sb.WriteString("templ Page() {\n")
	sb.WriteString("\t<html>\n\t\t<head>\n")
	sb.WriteString("\t\t\t<style>\nbody { margin: 0; }\n</style>\n")
	sb.WriteString("\t\t\t<script>\nfunction greet(name) { console.log(\"hello \" + name) }\n</script>\n")
	sb.WriteString("\t\t</head>\n\t\t<body>\n\t\t\t@C0()\n\t\t</body>\n\t</html>\n}\n")
	return sb.String()
}

func BenchmarkTempl(b *testing.B) {
	if !prettier.IsAvailable(prettier.DefaultCommand()) {
		b.Skip("Prettier is not available, skipping formatting benchmarks")
	}
	for _, componentCount := range []int{1, 10, 50} {
		src := []byte(createBenchmarkTemplate(componentCount))
		b.Run(fmt.Sprintf("components=%d", componentCount), func(b *testing.B) {
			for b.Loop() {
				_, _, err := Templ(src, "", Config{PrettierRequired: true})
				if err != nil {
					b.Fatalf("failed to format: %v", err)
				}
			}
		})
	}
}
