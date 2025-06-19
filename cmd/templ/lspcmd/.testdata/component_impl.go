package testelementcomponent

import (
	"context"
	"fmt"
	"io"
)

type ComponentImpl struct{}

func (c ComponentImpl) Render(ctx context.Context, w io.Writer) (err error) {
	_, err = fmt.Fprint(w, "<div>Component Rendered</div>")
	return
}