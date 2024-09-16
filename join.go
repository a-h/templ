package templ

import (
	"context"
	"io"
)

// Pass any number of templ.Components to get a single templ.Component with the components rendered
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
