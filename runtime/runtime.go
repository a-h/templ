package runtime

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

type GeneratedComponentInput struct {
	Context context.Context
	Writer  io.Writer
}

func GeneratedTemplate(f func(GeneratedComponentInput) error) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return f(GeneratedComponentInput{ctx, w})
	})
}
