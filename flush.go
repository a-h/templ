package templ

import (
	"context"
	"io"
)

// Flush flushes the output buffer after all its child components have been rendered.
func Flush() FlushComponent {
	return FlushComponent{}
}

type FlushComponent struct {
}

type flusherError interface {
	Flush() error
}

type flusher interface {
	Flush()
}

func (f FlushComponent) Render(ctx context.Context, w io.Writer) (err error) {
	if err = GetChildren(ctx).Render(ctx, w); err != nil {
		return err
	}
	switch w := w.(type) {
	case flusher:
		w.Flush()
		return nil
	case flusherError:
		return w.Flush()
	}
	return nil
}
