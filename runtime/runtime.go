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

// SetTemplArg stores value in m under key only if the key is not already set.
// This ensures that when components are nested, the outermost (first) component's
// arguments take precedence over any inner components with same-named parameters.
func SetTemplArg(m map[string]any, key string, value any) {
	if _, exists := m[key]; !exists {
		m[key] = value
	}
}
