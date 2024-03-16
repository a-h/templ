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

func (ep *ExpressionParser) Insert(pos token.Pos, tok token.Token, lit string) (stop bool, err error) {
	defer func() {
		ep.Previous = tok
	}()
	if tok == token.FUNC {
		// The next open brace will be the body of a function literal, so push the fn depth.
		ep.Fns.Push(len(ep.Stack))
		ep.End = int(pos) + len(tokenString(tok, lit)) - 1
		return false, nil
	}
	// Opening a pair can be done after an ident, but it can also be a func literal.
	// e.g. "name()", or "name(func() bool { return true })".
	if _, ok := goTokenOpenToClose[tok]; ok {
		if tok == token.LBRACE {
			if ep.Previous != token.IDENT {
				return true, nil
			}
			hasSpace := (int(pos) - 1) > ep.End
			if hasSpace && len(ep.Fns) == 0 {
				// There's a space, and we're not in a function so stop.
				return true, nil
			}
		}
		ep.Stack.Push(tok)
		ep.End = int(pos) + len(tokenString(tok, lit)) - 1
		return false, nil
	}
	// Closing a pair.
	if expected, ok := goTokenCloseToOpen[tok]; ok {
		if len(ep.Stack) == 0 {
			// We've got a close token, but there's nothing to close, so we must be done.
			return true, nil
		}
		actual := ep.Stack.Pop()
		if !ok {
			return false, ErrUnbalanced{tok}
		}
		if actual != expected {
			return false, ErrUnbalanced{tok}
		}
		// If we're closing a function, pop the function depth.
		if tok == token.RBRACE && len(ep.Stack) == ep.Fns.Peek() {
			ep.Fns.Pop()
		}
		ep.End = int(pos) + len(tokenString(tok, lit)) - 1
		return false, nil
	}
	// If we're within a pair, we allow anything.
	if len(ep.Stack) > 0 {
		ep.End = int(pos) + len(tokenString(tok, lit)) - 1
		return false, nil
	}
	// We allow a period to follow an ident or a closer.
	// e.g. "package.name" or "typeName{field: value}.name()".
	if tok == token.PERIOD && (ep.Previous == token.IDENT || isCloser(ep.Previous)) {
		ep.End = int(pos) + len(tokenString(tok, lit)) - 1
		return false, nil
	}
	// We allow an ident to follow a period or a closer.
	// e.g. "package.name", "typeName{field: value}.name()".
	// or "call().name", "call().name()".
	// But not "package .name" or "typeName{field: value} .name()".
	if tok == token.IDENT && (ep.Previous == token.PERIOD || isCloser(ep.Previous)) {
		if (int(pos) - 1) > ep.End {
			// There's a space, so stop.
			return true, nil
		}
		ep.End = int(pos) + len(tokenString(tok, lit)) - 1
		return false, nil
	}
	// Anything else returns stop=true.
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
