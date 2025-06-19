package mod

import (
	"context"
	"fmt"
	"io"
)

type StructComponent struct {
	Name string
}

func (c *StructComponent) Render(ctx context.Context, w io.Writer) error {
	_, err := fmt.Fprintf(w, "<div class=\"struct-component\">%s</div>", c.Name)
	return err
}
