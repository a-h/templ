package templ_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/a-h/templ"
)

func TestFragment(t *testing.T) {
	fragmentPage := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "page_contents"); err != nil {
			return err
		}
		fragmentContents := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			_, err := io.WriteString(w, "fragment_contents")
			return err
		})
		return templ.Fragment("fragment").Render(templ.WithChildren(ctx, fragmentContents), w)
	})
	t.Run("can render without a HTTP handler", func(t *testing.T) {
		// Set up the context.
		ctxW := new(bytes.Buffer)
		ctx := templ.WithFragmentContext(context.Background(), ctxW, "fragment")

		// Render the template.
		w := new(bytes.Buffer)
		if err := fragmentPage.Render(ctx, w); err != nil {
			t.Fatalf("failed to render: %v", err)
		}

		// Note that the fragment should not have been written to the output.
		if w.String() != "page_contents" {
			t.Errorf("expected output 'page_contents', got '%s'", w.String())
		}
		if ctxW.String() != "fragment_contents" {
			t.Errorf("expected output 'fragment_contents', got '%s'", ctxW.String())
		}
	})
}
