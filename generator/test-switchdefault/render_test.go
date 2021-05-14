package testswitchdefault

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var input = "b"

const expected = `it was something else`

func TestRender(t *testing.T) {
	w := new(strings.Builder)
	err := template(input).Render(context.Background(), w)
	if err != nil {
		t.Errorf("failed to render: %v", err)
	}
	if diff := cmp.Diff(expected, w.String()); diff != "" {
		t.Error(diff)
	}
}
