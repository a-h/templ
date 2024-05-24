package templ

import (
	"context"
	"io"
	"sync/atomic"
)

// onceHandleIndex is used to identify unique once handles in a program run.
var onceHandleIndex int64

type OnceOpt func(*OnceHandle)

// WithOnceComponent sets the component to be rendered once per context.
// This can be used instead of setting the children of the `Once` method,
// for example, if creating a code component outside of a templ HTML template.
func WithComponent(c Component) OnceOpt {
	return func(o *OnceHandle) {
		o.c = c
	}
}

// NewOnceHandle creates a OnceHandle used to ensure that the children of its
// `Once` method are only rendered once per context.
func NewOnceHandle(opts ...OnceOpt) *OnceHandle {
	oh := &OnceHandle{
		id: atomic.AddInt64(&onceHandleIndex, 1),
	}
	for _, opt := range opts {
		opt(oh)
	}
	return oh
}

// OnceHandle is used to ensure that the children of its `Once` method are are only
// rendered once per context.
type OnceHandle struct {
	// id is used to identify which instance of the OnceHandle is being used.
	// The OnceHandle can't be an empty struct, because:
	//
	//  | Two distinct zero-size variables may
	//  | have the same address in memory
	//
	// https://go.dev/ref/spec#Size_and_alignment_guarantees
	id int64
	// c is the component to be rendered once per context.
	// if c is nil, the children of the `Once` method are rendered.
	c Component
}

// Once returns a component that renders its children once per context.
func (o *OnceHandle) Once() Component {
	return ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, v := getContext(ctx)
		if v.getHasBeenRendered(o) {
			return nil
		}
		v.setHasBeenRendered(o)
		if o.c != nil {
			return o.c.Render(ctx, w)
		}
		return GetChildren(ctx).Render(ctx, w)
	})
}
