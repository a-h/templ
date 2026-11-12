package predeclared

import (
	"fmt"
	"testing"
)

type namedString string

func TestFormatString(t *testing.T) {
	tests := []struct {
		name       string
		input      any
		expected   string
		expectedOK bool
	}{
		{name: "strings are returned unchanged", input: "a <b>", expected: "a <b>", expectedOK: true},
		{name: "ints are formatted like fmt.Sprint", input: -123, expected: fmt.Sprint(-123), expectedOK: true},
		{name: "int64s are formatted like fmt.Sprint", input: int64(1) << 40, expected: fmt.Sprint(int64(1) << 40), expectedOK: true},
		{name: "bools are formatted like fmt.Sprint", input: true, expected: fmt.Sprint(true), expectedOK: true},
		{name: "floats are not formatted", input: 1.5, expected: "", expectedOK: false},
		{name: "named strings are not formatted", input: namedString("a"), expected: "", expectedOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, ok := FormatString(tt.input)
			if ok != tt.expectedOK {
				t.Fatalf("expected ok %v, got %v", tt.expectedOK, ok)
			}
			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}
