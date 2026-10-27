package testblanklines

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBlankLinesDoNotChangeTheRenderedOutput(t *testing.T) {
	items := []string{"a", "b"}

	without := new(strings.Builder)
	if err := WithoutBlankLines(items).Render(context.Background(), without); err != nil {
		t.Fatalf("failed to render template without blank lines: %v", err)
	}
	with := new(strings.Builder)
	if err := WithBlankLines(items).Render(context.Background(), with); err != nil {
		t.Fatalf("failed to render template with blank lines: %v", err)
	}

	if diff := cmp.Diff(without.String(), with.String()); diff != "" {
		t.Error(diff)
	}
}
