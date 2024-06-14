package runtime

import (
	"bytes"
	"context"
	"io"

	"github.com/a-h/templ"
)

func WriterToBuffer(w io.Writer) (*bytes.Buffer, bool, func()) {
	buffer, ok := w.(*bytes.Buffer)
	if !ok {
		buffer = templ.GetBuffer()
		return buffer, false, func() {
			templ.ReleaseBuffer(buffer)
		}
	}
	return buffer, true, func() {}
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
