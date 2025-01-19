package parse

import (
	"sort"
)

// NewInput creates an input from the given string.
func NewInput(s string) *Input {
	ip := &Input{
		s:         s,
		charIndex: 0,
	}
	for i, r := range s {
		if r == '\n' {
			ip.newLines = append(ip.newLines, i)
		}
	}
	return ip
}

// InputString is an input used by parsers. It stores the current location
// and character positions.
type Input struct {
	s         string
	charIndex int
	// character positions of new line characters.
	newLines []int
}

func (in *Input) Peek(n int) (s string, ok bool) {
	if in.charIndex+n > len(in.s) {
		return
	}
	if n < 0 {
		return in.s[in.charIndex:], true
	}
	return in.s[in.charIndex : in.charIndex+n], true
}

func (in *Input) Take(n int) (s string, ok bool) {
	if in.charIndex+n > len(in.s) {
		return
	}
	from := in.charIndex
	in.charIndex += n
	return in.s[from:in.charIndex], true
}

// Position returns the zero-bound index, line and column number of the current position within the stream.
func (in *Input) Position() Position {
	return in.PositionAt(in.charIndex)
}

// Position returns the zero-bound index, line and column number of the current position within the stream.
func (in *Input) PositionAt(index int) Position {
	lineIndex := sort.Search(len(in.newLines), func(lineIndex int) bool {
		return index <= in.newLines[lineIndex]
	})
	var previousLineEnd int
	if lineIndex > 0 {
		previousLineEnd = in.newLines[lineIndex-1] + 1
	}
	colIndex := index - previousLineEnd
	return Position{Index: index, Line: lineIndex, Col: colIndex}
}

// Index returns the current character index of the parser input.
func (in *Input) Index() int {
	return in.charIndex
}

// Seek to a position in the string.
func (in *Input) Seek(index int) (ok bool) {
	if index < 0 || index > len(in.s) {
		return
	}
	in.charIndex = index
	return true
}
