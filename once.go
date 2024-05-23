package templ

import (
	"context"
	"io"
	"sync/atomic"
)

// onceHandleIndex is used to identify unique once handles in a program run.
var onceHandleIndex int64

// NewOnceHandle creates a OnceHandle used to ensure that the children of its
// `Once` method are only rendered once per context.
func NewOnceHandle() *OnceHandle {
	return &OnceHandle{
		id: atomic.AddInt64(&onceHandleIndex, 1),
	}
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
}

// Once returns a component that renders its children once per context.
func (o *OnceHandle) Once() Component {
	return ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, v := getContext(ctx)
		if v.getHasBeenRendered(o) {
			return nil
		}
		v.setHasBeenRendered(o)
		return GetChildren(ctx).Render(ctx, w)
	})
}
