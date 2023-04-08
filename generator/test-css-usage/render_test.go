package testcssusage

import (
	"context"
	_ "embed"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

//go:embed expected.html
var expected string

func TestHTML(t *testing.T) {
	w := new(strings.Builder)
	err := ThreeButtons().Render(context.Background(), w)
	if err != nil {
		t.Errorf("failed to render: %v", err)
	}
	if diff := cmp.Diff(strings.TrimSpace(expected), w.String()); diff != "" {
		t.Error(diff)
	}
}
