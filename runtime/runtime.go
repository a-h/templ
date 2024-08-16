package runtime

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

// GeneratedComponentInput is used to avoid generated code needing to import the `context` and `io` packages.
type GeneratedComponentInput struct {
	Context context.Context
	Writer  io.Writer
}

// GeneratedTemplate is used to avoid generated code needing to import the `context` and `io` packages.
func GeneratedTemplate(f func(GeneratedComponentInput) error) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return f(GeneratedComponentInput{ctx, w})
	})
}
