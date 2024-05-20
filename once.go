package templ

import (
	"context"
	"io"
)

var _ Component = OnceComponent{}

// Once is a component that renders its children once per context.
func Once(id string) OnceComponent {
	return OnceComponent{
		ID: id,
	}
}

// OnceComponent is a component that renders its children once per context.
type OnceComponent struct {
	ID string
}

func (o OnceComponent) Render(ctx context.Context, w io.Writer) (err error) {
	_, v := getContext(ctx)
	if v.getHasOnceBeenRendered(o.ID) {
		return nil
	}
	v.setHasOnceBeenRendered(o.ID)
	return GetChildren(ctx).Render(ctx, w)
}
