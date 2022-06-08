package testtemplelement

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var input = []string{"a", "b", "c"}

const expected = `<p>header</p>some text<p>footer</p>`

func TestFor(t *testing.T) {
	w := new(strings.Builder)
	err := template().Render(context.Background(), w)
	if err != nil {
		t.Errorf("failed to render: %v", err)
	}
	if diff := cmp.Diff(expected, w.String()); diff != "" {
		t.Error(diff)
	}
}
