package runtime

import (
	"context"
	"errors"

	"github.com/a-h/templ/internal/htmlattr"
)

// SrcdocExpression is the result of the expression of an iframe srcdoc
// attribute. Unlike AttributeExpression, the value can be a templ.Component.
type SrcdocExpression[T any] struct {
	Value T
	Err   error
}

// NewSrcdocExpression returns the result of the expression of an iframe srcdoc
// attribute.
func NewSrcdocExpression[T any](v T, errs ...error) SrcdocExpression[T] {
	return SrcdocExpression[T]{
		Value: v,
		Err:   errors.Join(errs...),
	}
}

// Resolve converts the value to an HTML-escaped string. The browser parses the
// value as an HTML document, so a templ.Component is rendered as HTML, while
// strings, booleans, and numbers are escaped as text. Other types return an
// error.
func (e SrcdocExpression[T]) Resolve(ctx context.Context) (string, error) {
	if e.Err != nil {
		return "", e.Err
	}
	return resolveValue(ctx, htmlattr.ContextSrcdoc, e.Value)
}
