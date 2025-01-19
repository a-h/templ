package parse

import (
	"fmt"
)

func Error(msg string, pos Position) ParseError {
	return ParseError{
		Msg: msg,
		Pos: pos,
	}
}

type ParseError struct {
	Msg string
	Pos Position
}

func (e ParseError) Error() string {
	return fmt.Sprintf("%s: %v", e.Msg, e.Pos)
}
