package parser

import (
	"testing"

	"github.com/a-h/lexical/input"
	"github.com/a-h/lexical/parse"
	"github.com/google/go-cmp/cmp"
)

func TestNewLine(t *testing.T) {
	input := input.NewFromString(`a
aaa`)
	var crlf = parse.String("\r\n")
	var lf = parse.Rune('\n')
	result := parse.StringUntil(parse.Or(crlf, lf))(input)
	s := result.Item.(string)
	if s != "a" {
		t.Errorf("expected a, got %q", s)
	}
}

func TestNewLineStream(t *testing.T) {
	input := input.NewFromString("A\nB\n\nD")
	testPosition(t, "start", Position{Index: 0, Line: 1, Col: 0}, NewPositionFromInput(input))
	r, err := input.Advance()
	testRune(t, "advance 1", 'A', r, nil, err)
	testPosition(t, "advance 1", Position{Index: 1, Line: 1, Col: 1}, NewPositionFromInput(input))
	r, err = input.Advance()
	testRune(t, "advance to newline", '\n', r, nil, err)
	testPosition(t, "advance to newline", Position{Index: 2, Line: 2, Col: 0}, NewPositionFromInput(input))
	r, err = input.Retreat()
	testRune(t, "retreat from newline", 'A', r, nil, err)
	testPosition(t, "retreat from newline", Position{Index: 1, Line: 1, Col: 1}, NewPositionFromInput(input))
}

func testPosition(t *testing.T, name string, expected, got Position) {
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Error(name)
		t.Errorf(diff)
	}
}

func testRune(t *testing.T, name string, expectedRune, gotRune rune, expectedError, gotError error) {
	if expectedRune != gotRune {
		t.Errorf("%s: expected rune %q, got %q", name, string(expectedRune), string(gotRune))
	}
	if expectedError != gotError {
		t.Errorf("%s: expected error %v, got %v", name, expectedError, gotError)
	}
}
