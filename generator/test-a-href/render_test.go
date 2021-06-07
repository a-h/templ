package testahref

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const expected = `<a href="javascript:alert(&#39;unaffected&#39;);">Ignored</a>` +
	`<a href="about:invalid#TemplFailedSanitizationURL">Sanitized</a>` +
	`<a href="javascript:alert(&#39;should not be sanitized&#39;)">Unsanitized</a>`

func TestAHref(t *testing.T) {
	w := new(strings.Builder)
	err := render().Render(context.Background(), w)
	if err != nil {
		t.Errorf("failed to render: %v", err)
	}
	if diff := cmp.Diff(expected, w.String()); diff != "" {
		t.Error(diff)
	}
}
