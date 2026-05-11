package testtextinlineexpression

import (
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

func TestTextInlineExpression(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    templ.Component
		expected string
	}{
		{
			name:     "inline component after text renders on the same line",
			input:    InlineComponentAfterText(),
			expected: InlineComponentAfterTextExpected,
		},
		{
			name:     "email address renders as plain text",
			input:    EmailAddress(),
			expected: EmailAddressExpected,
		},
		{
			name:     "inline component between text preserves surrounding words",
			input:    InlineComponentBetweenText(),
			expected: InlineComponentBetweenTextExpected,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := new(strings.Builder)
			err := test.input.Render(context.Background(), w)
			if err != nil {
				t.Fatalf("failed to render: %v", err)
			}
			if diff := cmp.Diff(test.expected, w.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}
