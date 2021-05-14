package testswitch

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var input = "a"

const expected = `it was &#39;a&#39;`

func TestRender(t *testing.T) {
	w := new(strings.Builder)
	err := render(input).Render(context.Background(), w)
	if err != nil {
		t.Errorf("failed to render: %v", err)
	}
	if diff := cmp.Diff(expected, w.String()); diff != "" {
		t.Error(diff)
	}
}
