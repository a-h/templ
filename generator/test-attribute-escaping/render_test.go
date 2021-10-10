package testhtml

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const expected = `<div>` +
	`<a href="about:invalid#TemplFailedSanitizationURL"` +
	`</div>`

func TestHTML(t *testing.T) {
	w := new(strings.Builder)
	err := BasicTemplate(`javascript: alert("xss");`).Render(context.Background(), w)
	if err != nil {
		t.Errorf("failed to render: %v", err)
	}
	if diff := cmp.Diff(expected, w.String()); diff != "" {
		t.Error(diff)
	}
}
