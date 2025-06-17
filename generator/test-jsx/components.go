package testjsx

import (
	"context"
	"fmt"
	"io"
	
	"github.com/a-h/templ"
)

// TestComponent is a component with named parameters for testing symbol resolution
func TestComponent(title string, count int, enabled bool) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if enabled {
			_, err := fmt.Fprintf(w, "%s - Count: %d", title, count)
			return err
		}
		_, err := fmt.Fprint(w, title)
		return err
	})
}

// AnotherComponent demonstrates parameter order matters
func AnotherComponent(description string, name string, age int) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := fmt.Fprintf(w, "%s (%d): %s", name, age, description)
		return err
	})
}