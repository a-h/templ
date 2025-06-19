package mod

import (
	"context"
	"fmt"
	"io"

	"github.com/a-h/templ"
)

type StructComponent struct {
	Name  string
	Child templ.Component
	Attrs templ.Attributer
}

func (c *StructComponent) Render(ctx context.Context, w io.Writer) error {
	if _, err := fmt.Fprint(w, "<div class=\"struct-component\""); err != nil {
		return err
	}
	if c.Attrs != nil {
		if err := templ.RenderAttributes(ctx, w, c.Attrs); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprint(w, ">"); err != nil {
		return err
	}
	if c.Child != nil {
		if err := c.Child.Render(ctx, w); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprint(w, c.Name); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, "</div>"); err != nil {
		return err
	}
	return nil
}
