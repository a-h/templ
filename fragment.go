package templ

import (
	"context"
	"io"
	"slices"
)

// RenderFragments renders the specified fragments to w.
func RenderFragments(ctx context.Context, w io.Writer, c Component, ids ...any) error {
	ctx = context.WithValue(ctx, fragmentContextKey, &FragmentContext{
		W:   w,
		IDs: ids,
	})
	return c.Render(ctx, io.Discard)
}

type fragmentContextKeyType int

const fragmentContextKey fragmentContextKeyType = iota

// FragmentContext is used to control rendering of fragments within a template.
type FragmentContext struct {
	W      io.Writer
	IDs    []any
	Active bool
}

// Fragment defines a fragment within a template that can be rendered conditionally based on the id.
// You can use it to render a specific part of a page, e.g. to reduce the amount of HTML returned from a HTMX-initiated request.
// Any non-matching contents of the template are rendered, but discarded by the FramentWriter.
func Fragment(id any) Component {
	return &fragment{
		ID: id,
	}
}

type fragment struct {
	ID any
}

func (f *fragment) Render(ctx context.Context, w io.Writer) (err error) {
	// If not in a fragment context, if we're a child fragment, or in a mismatching fragment context, render children normally.
	fragmentCtx := getFragmentContext(ctx)
	if fragmentCtx == nil || fragmentCtx.Active || !slices.Contains(fragmentCtx.IDs, f.ID) {
		return GetChildren(ctx).Render(ctx, w)
	}

	// Instruct child fragments to render their contents normally, because the writer
	// passed to them is already the FragmentContext's writer.
	fragmentCtx.Active = true
	defer func() {
		fragmentCtx.Active = false
	}()
	return GetChildren(ctx).Render(ctx, fragmentCtx.W)
}

// getFragmentContext retrieves the FragmentContext from the provided context. It returns nil if no
// FragmentContext is found or if the context value is of an unexpected type.
func getFragmentContext(ctx context.Context) *FragmentContext {
	ctxValue := ctx.Value(fragmentContextKey)
	if ctxValue == nil {
		return nil
	}
	v, ok := ctxValue.(*FragmentContext)
	if !ok {
		return nil
	}
	return v
}
