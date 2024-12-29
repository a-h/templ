package goexpression

import (
	"fmt"
	"go/token"
)

type Stack[T any] []T

func (s *Stack[T]) Push(v T) {
	*s = append(*s, v)
}

func (s *Stack[T]) Pop() (v T) {
	if len(*s) == 0 {
		return v
	}
	v = (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return v
}

func (s *Stack[T]) Peek() (v T) {
	if len(*s) == 0 {
		return v
	}
	return (*s)[len(*s)-1]
}

var goTokenOpenToClose = map[token.Token]token.Token{
	token.LPAREN: token.RPAREN,
	token.LBRACE: token.RBRACE,
	token.LBRACK: token.RBRACK,
}

var goTokenCloseToOpen = map[token.Token]token.Token{
	token.RPAREN: token.LPAREN,
	token.RBRACE: token.LBRACE,
	token.RBRACK: token.LBRACK,
}

type ErrUnbalanced struct {
	Token token.Token
}

func (e ErrUnbalanced) Error() string {
	return fmt.Sprintf("unbalanced '%s'", e.Token)
}

func NewExpressionParser() *ExpressionParser {
	return &ExpressionParser{
		Stack:    make(Stack[token.Token], 0),
		Previous: token.PERIOD,
		Fns:      make(Stack[int], 0),
	}
}

type ExpressionParser struct {
	Stack    Stack[token.Token]
	End      int
	Previous token.Token
	Fns      Stack[int] // Stack of function depths.
}

func (ep *ExpressionParser) setEnd(pos token.Pos, tok token.Token, lit string) {
	ep.End = int(pos) + len(tokenString(tok, lit)) - 1
}

func (ep *ExpressionParser) hasSpaceBeforeCurrentToken(pos token.Pos) bool {
	return (int(pos) - 1) > ep.End
}

func (ep *ExpressionParser) isTopLevel() bool {
	return len(ep.Fns) == 0 && len(ep.Stack) == 0
}

func (ep *ExpressionParser) Insert(
	pos token.Pos,
	tok token.Token,
	lit string,
) (stop bool, err error) {
	defer func() {
		ep.Previous = tok
	}()

	// If we've reach the end of the file, terminate reading.
	if tok == token.EOF {
		// If the EOF was reached, but we're not at the top level, we must have an unbalanced expression.
		if !ep.isTopLevel() {
			return true, ErrUnbalanced{ep.Stack.Pop()}
		}
		return true, nil
	}

	// Handle function literals e.g. func() { fmt.Println("Hello") }
	// By pushing the current depth onto the stack, we prevent stopping
	// until we've closed the function.
	if tok == token.FUNC {
		ep.Fns.Push(len(ep.Stack))
		ep.setEnd(pos, tok, lit)
		return false, nil
	}
	// If we're opening a pair, we don't stop until we've closed it.
	if _, isOpener := goTokenOpenToClose[tok]; isOpener {
		// If we're at an open brace, at the top level, where a space has been used, stop.
		if tok == token.LBRACE && ep.isTopLevel() {
			// Previous was paren, e.g. () {
			if ep.Previous == token.RPAREN {
				return true, nil
			}
			// Previous was ident that isn't a type.
			// In `name {`, `name` is considered to be a variable.
			// In `name{`, `name` is considered to be a type name.
			if ep.Previous == token.IDENT && ep.hasSpaceBeforeCurrentToken(pos) {
				return true, nil
			}
		}
		ep.Stack.Push(tok)
		ep.setEnd(pos, tok, lit)
		return false, nil
	}
	if opener, isCloser := goTokenCloseToOpen[tok]; isCloser {
		if len(ep.Stack) == 0 {
			// We've got a close token, but there's nothing to close, so we must be done.
			return true, nil
		}
		actual := ep.Stack.Pop()
		if !isCloser {
			return false, ErrUnbalanced{tok}
		}
		if actual != opener {
			return false, ErrUnbalanced{tok}
		}
		if tok == token.RBRACE {
			// If we're closing a function, pop the function depth.
			if len(ep.Stack) == ep.Fns.Peek() {
				ep.Fns.Pop()
			}
		}
		ep.setEnd(pos, tok, lit)
		return false, nil
	}
	// If we're in a function literal slice, or pair, we allow anything until we close it.
	if len(ep.Fns) > 0 || len(ep.Stack) > 0 {
		ep.setEnd(pos, tok, lit)
		return false, nil
	}
	// We allow an ident to follow a period or a closer.
	// e.g. "package.name", "typeName{field: value}.name()".
	// or "call().name", "call().name()".
	// But not "package .name" or "typeName{field: value} .name()".
	if tok == token.IDENT && (ep.Previous == token.PERIOD || isCloser(ep.Previous)) {
		if isCloser(ep.Previous) && ep.hasSpaceBeforeCurrentToken(pos) {
			// This token starts later than the last ending, which means
			// there's a space.
			return true, nil
		}
		ep.setEnd(pos, tok, lit)
		return false, nil
	}
	if tok == token.PERIOD && (ep.Previous == token.IDENT || isCloser(ep.Previous)) {
		ep.setEnd(pos, tok, lit)
		return false, nil
	}

	// No match, so stop.
	return true, nil
}

func tokenString(tok token.Token, lit string) string {
	if tok.IsKeyword() || tok.IsOperator() {
		return tok.String()
	}
	return lit
}

func isCloser(tok token.Token) bool {
	_, ok := goTokenCloseToOpen[tok]
	return ok
}
