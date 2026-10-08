package runtime

import (
	"context"
	"errors"

	"github.com/a-h/templ"
	"github.com/a-h/templ/internal/htmlattr"
)

// attributeValue is the set of types that an attribute expression can return.
type attributeValue interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~complex64 | ~complex128 |
		~string | ~bool |
		templ.ComponentScript
}

// AttributeExpression is the result of an attribute expression. Generated code
// creates one with NewAttributeExpression, so that an expression that returns
// a value and an error, e.g. { getURL() }, can be resolved by a method that
// also takes a context.
type AttributeExpression[T attributeValue] struct {
	Value T
	Err   error
}

// NewAttributeExpression returns the result of an attribute expression.
func NewAttributeExpression[T attributeValue](v T, errs ...error) AttributeExpression[T] {
	return AttributeExpression[T]{
		Value: v,
		Err:   errors.Join(errs...),
	}
}

// Resolve converts the value to an HTML-escaped string. A
// templ.ComponentScript is converted to its function call.
func (e AttributeExpression[T]) Resolve(ctx context.Context) (string, error) {
	return e.resolve(ctx, htmlattr.ContextDefault)
}

// ResolveURL converts the value of an attribute that contains a URL to an
// HTML-escaped string. A templ.SafeURL is not sanitized. All other values are
// sanitized with templ.URL.
func (e AttributeExpression[T]) ResolveURL(ctx context.Context) (string, error) {
	return e.resolve(ctx, htmlattr.ContextURL)
}

// ResolveAnimationValue converts the value of an SVG animation attribute, such
// as the to attribute of a set element, to an HTML-escaped string. SVG
// animations can set the href of their parent element, so each item in the
// semicolon separated list is sanitized with templ.URL. A templ.SafeURL is not
// sanitized.
func (e AttributeExpression[T]) ResolveAnimationValue(ctx context.Context) (string, error) {
	return e.resolve(ctx, htmlattr.ContextAnimationValue)
}

// ResolveDynamic converts the value of an attribute with a name that is only
// known at render time to an HTML-escaped string. The value is sanitized based
// on the name, in the same way as the value of a spread attribute.
func (e AttributeExpression[T]) ResolveDynamic(ctx context.Context, elementName, name string) (string, error) {
	return e.resolve(ctx, htmlattr.Classify(elementName, SanitizeAttributeName(name)))
}

func (e AttributeExpression[T]) resolve(ctx context.Context, attrContext htmlattr.Context) (string, error) {
	if e.Err != nil {
		return "", e.Err
	}
	return resolveValue(ctx, attrContext, e.Value)
}
