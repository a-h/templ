package conditionalattributeselseif

import (
	"fmt"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

func render(t *testing.T, c templ.Component) string {
	t.Helper()
	var w strings.Builder
	if err := c.Render(t.Context(), &w); err != nil {
		t.Fatal(err)
	}
	return w.String()
}

func TestBranchSelection(t *testing.T) {
	for selected, branch := range []string{"if", "first", "second", "else"} {
		t.Run(branch, func(t *testing.T) {
			var checked []int
			actual := render(t, branches(func(i int) bool {
				checked = append(checked, i)
				return i == selected
			}))
			if want := fmt.Sprintf(`<div data-branch="%s"></div>`, branch); actual != want {
				t.Errorf("got %q, want %q", actual, want)
			}
			wantChecked := []int{0, 1, 2}
			if selected < 3 {
				wantChecked = wantChecked[:selected+1]
			}
			if diff := cmp.Diff(wantChecked, checked); diff != "" {
				t.Errorf("conditions were not short-circuited (-want +got):\n%s", diff)
			}
		})
	}
}

func TestInitializerAndNestedConditionWithoutElse(t *testing.T) {
	for _, tt := range []struct{ value, want string }{
		{"", `<div></div>`},
		{"x", `<div data-value="short"></div>`},
		{"a&b", `<div data-value="a&amp;b"></div>`},
	} {
		t.Run(tt.value, func(t *testing.T) {
			if got := render(t, withoutElse(tt.value)); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestElseIfCSS(t *testing.T) {
	got := render(t, withCSS(true))
	if !strings.HasPrefix(got, "<style") || !strings.Contains(got, "color:red;") || !strings.Contains(got, `</style><div class="`) {
		t.Errorf("expected CSS before the opening tag, got %q", got)
	}
	if got := render(t, withCSS(false)); got != `<div></div>` {
		t.Errorf("unselected branch emitted CSS: %q", got)
	}
}

func TestElseIfScript(t *testing.T) {
	got := render(t, withScript(true))
	if !strings.HasPrefix(got, "<script") || !strings.Contains(got, `alert("clicked")`) || !strings.Contains(got, `</script><button onclick="`) {
		t.Errorf("expected script before the opening tag, got %q", got)
	}
	if got := render(t, withScript(false)); got != `<button></button>` {
		t.Errorf("unselected branch emitted a script: %q", got)
	}
}

func TestElseIfSpreadAndDynamicKeySanitization(t *testing.T) {
	for _, tt := range []struct {
		name      string
		component templ.Component
		want      string
	}{
		{"spread", withSpread(true, templ.Attributes{"title": `a"&b`}), `<div title="a&#34;&amp;b"></div>`},
		{"unselected spread", withSpread(false, templ.Attributes{"title": "unused"}), `<div></div>`},
		{"dynamic URL", withDynamicKey(true, "href", "javascript:alert(1)"), `<a href="about:invalid#TemplFailedSanitizationURL"></a>`},
		{"unselected dynamic key", withDynamicKey(false, "href", "unused"), `<a></a>`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := render(t, tt.component); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
