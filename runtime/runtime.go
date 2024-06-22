package runtime

import (
	"bytes"
	"context"
	"io"

	"github.com/a-h/templ"
)

func WriterToBuffer(w io.Writer) (buf *bytes.Buffer, ok bool, release func()) {
	buf, ok = w.(*bytes.Buffer)
	if !ok {
		buf = templ.GetBuffer()
		return buf, false, func() {
			templ.ReleaseBuffer(buf)
		}
	}
	return buf, true, func() {}
}

type GeneratedComponentInput struct {
	Context context.Context
	Writer  io.Writer
}

func GeneratedTemplate(f func(GeneratedComponentInput) error) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return f(GeneratedComponentInput{ctx, w})
	})
}
