package parser

import (
	"errors"
	"testing"
)

func TestNewTrailingSpace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected TrailingSpace
	}{
		{name: "empty input is no space", input: "", expected: SpaceNone},
		{name: "spaces and tabs are horizontal space", input: " \t ", expected: SpaceHorizontal},
		{name: "a single newline is vertical space", input: "\n", expected: SpaceVertical},
		{name: "a single CRLF is vertical space", input: "\r\n", expected: SpaceVertical},
		{name: "a newline surrounded by indentation is vertical space", input: "  \n\t\t", expected: SpaceVertical},
		{name: "two newlines are double vertical space", input: "\n\n", expected: SpaceVerticalDouble},
		{name: "two CRLFs are double vertical space", input: "\r\n\r\n", expected: SpaceVerticalDouble},
		{name: "a blank line containing indentation is double vertical space", input: "\n\t\t\n\t", expected: SpaceVerticalDouble},
		{name: "many newlines are double vertical space", input: "\n\n\n\n", expected: SpaceVerticalDouble},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := NewTrailingSpace(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual != tt.expected {
				t.Errorf("got %q, expected %q", actual, tt.expected)
			}
		})
	}
}

func TestNewTrailingSpaceReturnsAnErrorForNonSpaceCharacters(t *testing.T) {
	_, err := NewTrailingSpace("\n\na")
	if !errors.Is(err, ErrNonSpaceCharacter) {
		t.Errorf("got %v, expected %v", err, ErrNonSpaceCharacter)
	}
}
