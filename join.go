package templ

import (
	"context"
	"io"
)

// Join returns a single `templ.Component` that will render provided components in order.
// If any of the components return an error the Join component will immediately return with the error.
func Join(components ...Component) Component {
	return ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		for _, c := range components {
			if err = c.Render(ctx, w); err != nil {
				return err
			}
		}
		return nil
	})
}
