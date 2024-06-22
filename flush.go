package templ

import (
	"context"
	"fmt"
	"io"
)

func Flush() FlushComponent {
	return FlushComponent{}
}

type FlushComponent struct {
}

type flusher interface {
	Flush() error
}

func (f FlushComponent) Render(ctx context.Context, w io.Writer) (err error) {
	if err = GetChildren(ctx).Render(ctx, w); err != nil {
		return err
	}
	b, isTemplBuffer := w.(flusher)
	if !isTemplBuffer {
		return fmt.Errorf("unable to flush, writer is not flushable")
	}
	return b.Flush()
}
