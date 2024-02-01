package parser

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTemplateFileParser(t *testing.T) {
	t.Run("requests migration of legacy formats", func(t *testing.T) {
		input := `{% package templates %}
`
		_, err := ParseString(input)
		if err == nil {
			t.Error("expected ErrLegacyFileFormat, got nil")
		}
		if err != ErrLegacyFileFormat {
			t.Errorf("expected ErrLegacyFileFormat, got %v", err)
		}
	})
	t.Run("but can accept a package expression, if one is provided", func(t *testing.T) {
		input := `package goof

templ Hello() {
	Hello
}`
		tf, err := ParseString(input)
		if err != nil {
			t.Fatalf("failed to parse template, with t.Fatalf(parser %v", err)
		}
		if len(tf.Nodes) != 1 {
			t.Errorf("expected 2 nodes, got %+v", tf.Nodes)
		}
		if tf.Package.Expression.Value != "package goof" {
			t.Errorf("expected \"goof\", got %q", tf.Package.Expression.Value)
		}
	})
	t.Run("can start with comments", func(t *testing.T) {
		input := `// Example comment.
package goof

templ Hello() {
	Hello
}`
		tf, err := ParseString(input)
		if err != nil {
			t.Fatalf("failed to parse template, with t.Fatalf(parser %v", err)
		}
		if len(tf.Nodes) != 1 {
			t.Errorf("expected 2 node, got %d nodes with content %+v", len(tf.Nodes), tf.Nodes)
		}
	})
	t.Run("template files can end with Go expressions", func(t *testing.T) {
		input := `package goof

const x = "123"

templ Hello() {
	Hello
}

const y = "456"
`
		tf, err := ParseString(input)
		if err != nil {
			t.Fatalf("failed to parse template, with t.Fatalf(parser %v", err)
		}
		if len(tf.Nodes) != 3 {
			var nodeTypes []string
			for _, n := range tf.Nodes {
				nodeTypes = append(nodeTypes, reflect.TypeOf(n).Name())
			}
			t.Fatalf("expected 3 nodes, got %d nodes, %v", len(tf.Nodes), nodeTypes)
		}
		expr, isGoExpression := tf.Nodes[0].(TemplateFileGoExpression)
		if !isGoExpression {
			t.Errorf("0: expected expression, got %t", tf.Nodes[2])
		}
		if expr.Expression.Value != `const x = "123"` {
			t.Errorf("0: unexpected expression: %q", expr.Expression.Value)
		}
		expr, isGoExpression = tf.Nodes[2].(TemplateFileGoExpression)
		if !isGoExpression {
			t.Errorf("2: expected expression, got %t", tf.Nodes[2])
		}
		if expr.Expression.Value != `const y = "456"` {
			t.Errorf("2: unexpected expression: %q", expr.Expression.Value)
		}
	})
	t.Run("template files can end with string literals", func(t *testing.T) {
		input := `package goof

const x = "123"

templ Hello() {
	Hello
}

const y = ` + "`456`"
		tf, err := ParseString(input)
		if err != nil {
			t.Fatalf("failed to parse template, with t.Fatalf(parser %v", err)
		}
		if len(tf.Nodes) != 3 {
			var nodeTypes []string
			for _, n := range tf.Nodes {
				nodeTypes = append(nodeTypes, reflect.TypeOf(n).Name())
			}
			t.Fatalf("expected 3 nodes, got %d nodes, %v", len(tf.Nodes), nodeTypes)
		}
		expr, isGoExpression := tf.Nodes[0].(TemplateFileGoExpression)
		if !isGoExpression {
			t.Errorf("0: expected expression, got %t", tf.Nodes[2])
		}
		if expr.Expression.Value != `const x = "123"` {
			t.Errorf("0: unexpected expression: %q", expr.Expression.Value)
		}
		expr, isGoExpression = tf.Nodes[2].(TemplateFileGoExpression)
		if !isGoExpression {
			t.Errorf("2: expected expression, got %t", tf.Nodes[2])
		}
		if expr.Expression.Value != "const y = `456`" {
			t.Errorf("2: unexpected expression: %q", expr.Expression.Value)
		}
	})
}

func TestTemplateFileRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "template files can be round tripped",
			input: `package goof

templ Hello() {
	Hello
}
`,
			expected: `package goof

templ Hello() {
	Hello
}
`,
		},
		{
			name: "template files can start with comments",
			input: `// Go comment
package goof

templ Hello() {
	Hello
}
`,
			expected: `// Go comment
package goof

templ Hello() {
	Hello
}
`,
		},
		{
			name: "template files can start with comments, mixed with whitespace",
			input: `
// Go comment

package goof

templ Hello() {
	Hello
}
`,
			expected: `
// Go comment

package goof

templ Hello() {
	Hello
}
`,
		},
		{
			name: "template files can start with multiline comments",
			input: `
/********************
* multiline message *
********************/

package goof

templ Hello() {
	Hello
}
`,
			expected: `
/********************
* multiline message *
********************/

package goof

templ Hello() {
	Hello
}
`,
		},
		{
			name: "template files can start with comments, mixed with whitespace",
			input: `
// Go comment

/* Multiline comment on a single line */

/*

Multi-line comment on multiple lines

*/

package goof

templ Hello() {
	Hello
}
`,
			expected: `
// Go comment

/* Multiline comment on a single line */

/*

Multi-line comment on multiple lines

*/

package goof

templ Hello() {
	Hello
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("failed to parse template file: %v", err)
			}

			sb := new(strings.Builder)
			err = tf.Write(sb)
			if err != nil {
				t.Fatalf("failed to write template file: %v", err)
			}
			output := sb.String()

			if diff := cmp.Diff(tt.expected, output); diff != "" {
				t.Errorf("unexpected output (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDefaultPackageName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard filename",
			input:    "/files/on/disk/header.templ",
			expected: "disk",
		},
		{
			name:     "path that starts with numbers",
			input:    "/files/on/123disk/header.templ",
			expected: "main",
		},
		{
			name:     "path that includes hyphens",
			input:    "/files/on/disk-drive/header.templ",
			expected: "main",
		},
		{
			name:     "relative path",
			input:    "header.templ",
			expected: "main",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := getDefaultPackageName(tt.input)
			if actual != tt.expected {
				t.Errorf("expected %q got %q", tt.expected, actual)
			}
		})
	}
}
