package testwhitespacearoundgokeywords

import (
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

func TestTextWhitespace(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    templ.Component
		expected string
	}{
		{
			name:     "whitespace is consistent in a true evaluated if statement",
			input:    WhitespaceIsConsistentInIf(true, false),
			expected: WhitespaceIsConsistentInTrueIfExpected,
		},
		{
			name:     "whitespace is consistent in a true evaluated else if statement",
			input:    WhitespaceIsConsistentInIf(false, true),
			expected: WhitespaceIsConsistentInTrueElseIfExpected,
		},
		{
			name:     "whitespace is consistent in a true evaluated else statement",
			input:    WhitespaceIsConsistentInIf(false, false),
			expected: WhitespaceIsConsistentInTrueElseExpected,
		},
		{
			name:     "whitespace is consistent in a false evaluated if statement",
			input:    WhitespaceIsConsistentInFalseIf(),
			expected: WhitespaceIsConsistentInFalseIfExpected,
		},
		{
			name:     "whitespace is consistent in a switch statement with a true case",
			input:    WhitespaceIsConsistentInSwitch(1),
			expected: WhitespaceIsConsistentInOneSwitchExpected,
		},
		{
			name:     "whitespace is consistent in a switch statement with a default case",
			input:    WhitespaceIsConsistentInSwitch(2),
			expected: WhitespaceIsConsistentInDefaultSwitchExpected,
		},
		{
			name:     "whitespace is consistent in a switch statement with no default case and no true cases",
			input:    WhitespaceIsConsistentInSwitchNoDefault(),
			expected: WhitespaceIsConsistentInSwitchNoDefaultExpected,
		},
		{
			name:     "whitespace is consistent in a for statement that runs 0 times",
			input:    WhitespaceIsConsistentInFor(0),
			expected: WhitespaceIsConsistentInForZeroExpected,
		},
		{
			name:     "whitespace is consistent in a for statement that runs 1 times",
			input:    WhitespaceIsConsistentInFor(1),
			expected: WhitespaceIsConsistentInForOneExpected,
		},
		{
			name:     "whitespace is consistent in a for statement that runs 3 times",
			input:    WhitespaceIsConsistentInFor(3),
			expected: WhitespaceIsConsistentInForThreeExpected,
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
