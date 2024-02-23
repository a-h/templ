package components

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

func insertUrlInJs(baseURL string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		s := "<script>var base_url = '" + baseURL + "';</script>"

		_, err := io.WriteString(w, s)

		return err
	})
}
