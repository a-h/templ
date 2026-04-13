package templ_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/a-h/templ"
)

func TestConcurrentContextThreadSafety(t *testing.T) {
	script := templ.ComponentScript{
		Name:     "concurrentScript",
		Function: `function concurrent() {}`,
	}
	cssClass := templ.ComponentCSSClass{
		ID:    "concurrentClass",
		Class: templ.SafeCSS(".concurrentClass { color: blue; }"),
	}

	tests := []struct {
		name   string
		render func(ctx context.Context) error
	}{
		{
			name: "concurrent renders are thread-safe for scripts",
			render: func(ctx context.Context) error {
				var buf bytes.Buffer
				return templ.RenderScriptItems(ctx, &buf, script)
			},
		},
		{
			name: "concurrent renders are thread-safe for CSS",
			render: func(ctx context.Context) error {
				var buf bytes.Buffer
				return templ.RenderCSSItems(ctx, &buf, cssClass)
			},
		},
		{
			name: "concurrent renders are thread-safe for Once handles",
			render: func(ctx context.Context) error {
				var buf bytes.Buffer
				handle := templ.NewOnceHandle()
				return handle.Once().Render(ctx, &buf)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := templ.InitializeContext(context.Background())
			ctx = templ.WithChildren(ctx, templ.Raw("content"))

			var wg sync.WaitGroup
			errs := make([]error, 10)

			for i := range 10 {
				wg.Go(func() {
					errs[i] = tt.render(ctx)
				})
			}

			wg.Wait()

			if err := errors.Join(errs...); err != nil {
				t.Errorf("concurrent render failed: %v", err)
			}
		})
	}
}

func TestConcurrentDeduplication(t *testing.T) {
	t.Run("scripts are deduplicated across concurrent renders", func(t *testing.T) {
		ctx := templ.InitializeContext(context.Background())

		script := templ.ComponentScript{
			Name:     "sharedScript",
			Function: `function shared() { console.log("shared"); }`,
		}

		var wg sync.WaitGroup
		outputs := make([]string, 10)
		errs := make([]error, 10)

		for i := range 10 {
			wg.Go(func() {
				var buf bytes.Buffer
				if err := templ.RenderScriptItems(ctx, &buf, script); err != nil {
					errs[i] = err
					return
				}
				outputs[i] = buf.String()
			})
		}

		wg.Wait()

		if err := errors.Join(errs...); err != nil {
			t.Fatalf("concurrent render failed: %v", err)
		}

		// Exactly one goroutine should have rendered the script.
		var renderedCount int
		for _, output := range outputs {
			if strings.Contains(output, script.Function) {
				renderedCount++
			}
		}

		if renderedCount != 1 {
			t.Errorf("expected exactly 1 goroutine to render script, got %d", renderedCount)
		}
	})

	t.Run("CSS classes are deduplicated across concurrent renders", func(t *testing.T) {
		ctx := templ.InitializeContext(context.Background())

		cssClass := templ.ComponentCSSClass{
			ID:    "sharedClass",
			Class: templ.SafeCSS(".sharedClass { color: red; }"),
		}

		var wg sync.WaitGroup
		outputs := make([]string, 10)
		errs := make([]error, 10)

		for i := range 10 {
			wg.Go(func() {
				var buf bytes.Buffer
				if err := templ.RenderCSSItems(ctx, &buf, cssClass); err != nil {
					errs[i] = err
					return
				}
				outputs[i] = buf.String()
			})
		}

		wg.Wait()

		if err := errors.Join(errs...); err != nil {
			t.Fatalf("concurrent render failed: %v", err)
		}

		// Exactly one goroutine should have rendered the CSS.
		var renderedCount int
		for _, output := range outputs {
			if strings.Contains(output, string(cssClass.Class)) {
				renderedCount++
			}
		}

		if renderedCount != 1 {
			t.Errorf("expected exactly 1 goroutine to render CSS, got %d", renderedCount)
		}
	})

	t.Run("Once handles are deduplicated across concurrent renders", func(t *testing.T) {
		ctx := templ.InitializeContext(context.Background())
		ctx = templ.WithChildren(ctx, templ.Raw("once content"))

		handle := templ.NewOnceHandle()
		onceComponent := handle.Once()

		var wg sync.WaitGroup
		outputs := make([]string, 10)
		errs := make([]error, 10)

		for i := range 10 {
			wg.Go(func() {
				var buf bytes.Buffer
				if err := onceComponent.Render(ctx, &buf); err != nil {
					errs[i] = err
					return
				}
				outputs[i] = buf.String()
			})
		}

		wg.Wait()

		if err := errors.Join(errs...); err != nil {
			t.Fatalf("concurrent render failed: %v", err)
		}

		// Exactly one goroutine should have rendered the content.
		var renderedCount int
		for _, output := range outputs {
			if output == "once content" {
				renderedCount++
			}
		}

		if renderedCount != 1 {
			t.Errorf("expected exactly 1 goroutine to render Once content, got %d", renderedCount)
		}
	})
}

func TestSequentialDeduplication(t *testing.T) {
	t.Run("component scripts are not rendered more than once in sequence", func(t *testing.T) {
		ctx := templ.InitializeContext(context.Background())

		script := templ.ComponentScript{
			Name:     "testScript",
			Function: `function test() {}`,
		}

		var buf bytes.Buffer
		if err := templ.RenderScriptItems(ctx, &buf, script); err != nil {
			t.Fatalf("first render failed: %v", err)
		}

		if !strings.Contains(buf.String(), script.Function) {
			t.Errorf("first render should output script")
		}

		buf.Reset()
		if err := templ.RenderScriptItems(ctx, &buf, script); err != nil {
			t.Fatalf("second render failed: %v", err)
		}

		if buf.String() != "" {
			t.Errorf("second render should not output script, got: %q", buf.String())
		}
	})

	t.Run("nonce is preserved in context", func(t *testing.T) {
		expectedNonce := "test-nonce-123"
		ctx := templ.WithNonce(context.Background(), expectedNonce)

		actualNonce := templ.GetNonce(ctx)
		if actualNonce != expectedNonce {
			t.Errorf("expected nonce %q, got %q", expectedNonce, actualNonce)
		}
	})
}
