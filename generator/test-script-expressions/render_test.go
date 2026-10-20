package testscriptexpressions

import (
	"context"
	_ "embed"
	"os"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/a-h/templ/generator/htmldiff"
)

//go:embed expected.html
var expected string

func Test(t *testing.T) {
	component := AllTests()

	actual, diff, err := htmldiff.Diff(component, expected)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "" {
		if err := os.WriteFile("actual.html", []byte(actual), 0644); err != nil {
			t.Errorf("failed to write actual.html: %v", err)
		}
		t.Error(diff)
	}
}

func TestScriptExpressionsAreEscapedForTheirPosition(t *testing.T) {
	tests := []struct {
		name      string
		component templ.Component
		expected  string
	}{
		{
			name:      "data after the condition of an if statement is JSON encoded",
			component: RenderAfterIfCondition("alert(1)"),
			expected:  `<script>if (ok) /"/.test(s); var a = "alert(1)";</script>`,
		},
		{
			name:      "data after division of a property named return is JSON encoded",
			component: RenderAfterPropertyNamedReturn("alert(1)"),
			expected:  `<script>var a = b.return / 2; var c = "/"; var d = "alert(1)";</script>`,
		},
		{
			name:      "data after division of a postfix increment is JSON encoded",
			component: RenderAfterPostfixIncrement("alert(1)"),
			expected:  `<script>var a = i++ / 2; var c = "/"; var d = "alert(1)";</script>`,
		},
		{
			name:      "data after an HTML open comment is JSON encoded",
			component: RenderAfterHTMLOpenComment("alert(1)"),
			expected:  "<script>\n\t<!-- \"\n\tvar a = \"alert(1)\";\n\t</script>",
		},
		{
			name:      "data in a regexp can not end the regexp",
			component: RenderInRegexp(`/;alert(1);//`),
			expected:  `<script>var r = /a\/;alert\(1\);\/\//;</script>`,
		},
		{
			name:      "data after division of a number with a trailing decimal point is JSON encoded",
			component: RenderAfterNumberWithTrailingDecimalPoint("alert(1)"),
			expected:  `<script>var a = 1. / 2; var c = "/"; var d = "alert(1)";</script>`,
		},
		{
			name:      "empty data at the start of a regexp renders an empty group",
			component: RenderAtRegexpStart(""),
			expected:  `<script>var r = /(?:)/;</script>`,
		},
		{
			name:      "data at the start of a regexp follows an empty group",
			component: RenderAtRegexpStart("+"),
			expected:  `<script>var r = /(?:)\+/;</script>`,
		},
		{
			name:      "empty data at the start of a regexp before a quantifier does not start a comment",
			component: RenderAtRegexpStartBeforeQuantifier(""),
			expected:  `<script>var r = /(?:)*/;</script>`,
		},
		{
			name:      "empty data in a character class adds no characters to the class",
			component: RenderInRegexpClass(""),
			expected:  `<script>var r = /[]/;</script>`,
		},
		{
			name:      "a script start tag in a regexp after an HTML open comment is escaped",
			component: RenderInRegexpAfterHTMLOpenComment("<script>"),
			expected:  "<script><!--\nvar r = /a\\u003cscript\\u003e/;</script>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual strings.Builder
			if err := tt.component.Render(context.Background(), &actual); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual.String() != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, actual.String())
			}
		})
	}
}
