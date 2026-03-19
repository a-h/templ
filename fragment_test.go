package templ_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

func TestFragment_ClearsChildrenBeforeRendering(t *testing.T) {
	// Simulates the generated code pattern for a component that always renders { children... }.
	// e.g. templ Wrapper() { <div class="wrapper">{ children... }</div> }
	wrapper := func() templ.Component {
		return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			children := templ.GetChildren(ctx)
			ctx = templ.ClearChildren(ctx)
			if _, err := io.WriteString(w, `<div class="wrapper">`); err != nil {
				return err
			}
			if err := children.Render(ctx, w); err != nil {
				return err
			}
			_, err := io.WriteString(w, `</div>`)
			return err
		})
	}

	// Simulates: templ Layout() { <main>@templ.Fragment("content") { { children... } }</main> }
	layout := func() templ.Component {
		return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			children := templ.GetChildren(ctx)
			ctx = templ.ClearChildren(ctx)
			if _, err := io.WriteString(w, `<main>`); err != nil {
				return err
			}
			fragmentChildren := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
				return children.Render(ctx, w)
			})
			if err := templ.Fragment("content").Render(templ.WithChildren(ctx, fragmentChildren), w); err != nil {
				return err
			}
			_, err := io.WriteString(w, `</main>`)
			return err
		})
	}

	// Simulates: templ Page() { @Layout() { @Wrapper()  <p>actual content</p> } }
	// Wrapper is called WITHOUT a children block, so no WithChildren call is generated.
	page := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		ctx = templ.ClearChildren(ctx)
		pageContent := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			if err := wrapper().Render(ctx, w); err != nil {
				return err
			}
			_, err := io.WriteString(w, `<p>actual content</p>`)
			return err
		})
		return layout().Render(templ.WithChildren(ctx, pageContent), w)
	})

	t.Run("components without children block render empty children", func(t *testing.T) {
		w := new(bytes.Buffer)
		if err := page.Render(context.Background(), w); err != nil {
			t.Fatalf("failed to render: %v", err)
		}
		expected := `<main><div class="wrapper"></div><p>actual content</p></main>`
		if diff := cmp.Diff(expected, w.String()); diff != "" {
			t.Error(diff)
		}
	})
}

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
		w := new(bytes.Buffer)
		if err := templ.RenderFragments(context.Background(), w, fragmentPage, "fragment"); err != nil {
			t.Fatalf("failed to render: %v", err)
		}

		// Note that the fragment should have been written to the output.
		if w.String() != "fragment_contents" {
			t.Errorf("expected output 'fragment_contents', got '%s'", w.String())
		}
	})
}
