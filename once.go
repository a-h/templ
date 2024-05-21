package templ

import (
	"context"
	"io"
)

var _ Component = OnceComponent[string]{}
var _ Component = OnceComponent[int]{}

// Once is a component that renders its children once per context.
func Once[T comparable](id T) OnceComponent[T] {
	return OnceComponent[T]{
		ID: id,
	}
}

// OnceComponent is a component that renders its children once per context.
type OnceComponent[T comparable] struct {
	ID T
}

func (o OnceComponent[T]) Render(ctx context.Context, w io.Writer) (err error) {
	_, v := getContext(ctx)
	if v.getHasOnceBeenRendered(o.ID) {
		return nil
	}
	v.setHasOnceBeenRendered(o.ID)
	return GetChildren(ctx).Render(ctx, w)
}
