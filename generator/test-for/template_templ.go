// Code generated by templ DO NOT EDIT.

package testfor

import "github.com/a-h/templ"
import "context"
import "io"

func render(items []string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		ctx, _ = templ.RenderedCSSClassesFromContext(ctx)
		for _, item := range items {
			_, err = io.WriteString(w, "<div>")
			if err != nil {
				return err
			}
			_, err = io.WriteString(w, templ.EscapeString(item))
			if err != nil {
				return err
			}
			_, err = io.WriteString(w, "</div>")
			if err != nil {
				return err
			}
		}
		return err
	})
}

