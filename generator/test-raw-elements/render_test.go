package testrawelements

import (
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

func TestRawElements(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    templ.Component
		expected string
	}{
		{
			name:     "style",
			input:    StyleElement(),
			expected: StyleElementExpected,
		},
		{
			name:     "script",
			input:    ScriptElement(),
			expected: ScriptElementExpected,
		},
	} {
		w := new(strings.Builder)
		err := test.input.Render(context.Background(), w)
		if err != nil {
			t.Errorf("failed to render: %v", err)
		}
		if diff := cmp.Diff(test.expected, w.String()); diff != "" {
			t.Error(diff)
		}
	}
}
