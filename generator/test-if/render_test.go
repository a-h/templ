package testif

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const expected = `True`

func TestRender(t *testing.T) {
	w := new(strings.Builder)
	err := render(data{}).Render(context.Background(), w)
	if err != nil {
		t.Errorf("failed to render: %v", err)
	}
	if diff := cmp.Diff(expected, w.String()); diff != "" {
		t.Error(diff)
	}
}
