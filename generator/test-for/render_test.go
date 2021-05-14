package testfor

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var input = []string{"a", "b", "c"}

const expected = `<div>a</div><div>b</div><div>c</div>`

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
