package templ_test

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/a-h/templ"
)

// TestConcurrentContextThreadSafety tests thread-safe access with mutex
func TestConcurrentContextThreadSafety(t *testing.T) {
	t.Run("concurrent renders are thread-safe for scripts", func(t *testing.T) {
		ctx := templ.InitializeContext(context.Background())

		script := templ.ComponentScript{
			Name:     "concurrentScript",
			Function: `function concurrent() {}`,
		}

		var wg sync.WaitGroup
		errors := make(chan error, 10)

		// Launch 10 goroutines trying to render the same script concurrently
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				var buf bytes.Buffer
				if err := templ.RenderScriptItems(ctx, &buf, script); err != nil {
					errors <- err
				}
			}()
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Errorf("concurrent render failed: %v", err)
		}
	})

	t.Run("concurrent renders are thread-safe for CSS", func(t *testing.T) {
		ctx := templ.InitializeContext(context.Background())

		cssClass := templ.ComponentCSSClass{
			ID:    "concurrentClass",
			Class: templ.SafeCSS(".concurrentClass { color: blue; }"),
		}

		var wg sync.WaitGroup
		errors := make(chan error, 10)

		// Launch 10 goroutines trying to render the same CSS concurrently
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				var buf bytes.Buffer
				if err := templ.RenderCSSItems(ctx, &buf, cssClass); err != nil {
					errors <- err
				}
			}()
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Errorf("concurrent render failed: %v", err)
		}
	})

	t.Run("concurrent renders are thread-safe for Once handles", func(t *testing.T) {
		ctx := templ.InitializeContext(context.Background())
		ctx = templ.WithChildren(ctx, templ.Raw("content"))

		handle := templ.NewOnceHandle()
		onceComponent := handle.Once()

		var wg sync.WaitGroup
		errors := make(chan error, 10)

		// Launch 10 goroutines trying to render the same Once handle concurrently
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				var buf bytes.Buffer
				if err := onceComponent.Render(ctx, &buf); err != nil {
					errors <- err
				}
			}()
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Errorf("concurrent render failed: %v", err)
		}
	})
}

// TestConcurrentDeduplication verifies that deduplication works correctly
// when multiple goroutines render to the same shared context
func TestConcurrentDeduplication(t *testing.T) {
	t.Run("scripts are deduplicated across concurrent renders", func(t *testing.T) {
		ctx := templ.InitializeContext(context.Background())

		script := templ.ComponentScript{
			Name:     "sharedScript",
			Function: `function shared() { console.log("shared"); }`,
		}

		var wg sync.WaitGroup
		outputs := make([]string, 10)
		errors := make(chan error, 10)

		// Launch 10 goroutines all trying to render the same script
		for i := 0; i < 10; i++ {
			wg.Add(1)
			idx := i
			go func() {
				defer wg.Done()
				var buf bytes.Buffer
				if err := templ.RenderScriptItems(ctx, &buf, script); err != nil {
					errors <- err
					return
				}
				outputs[idx] = buf.String()
			}()
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Fatalf("concurrent render failed: %v", err)
		}

		// Exactly one goroutine should have rendered the script
		renderedCount := 0
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
		errors := make(chan error, 10)

		// Launch 10 goroutines all trying to render the same CSS
		for i := 0; i < 10; i++ {
			wg.Add(1)
			idx := i
			go func() {
				defer wg.Done()
				var buf bytes.Buffer
				if err := templ.RenderCSSItems(ctx, &buf, cssClass); err != nil {
					errors <- err
					return
				}
				outputs[idx] = buf.String()
			}()
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Fatalf("concurrent render failed: %v", err)
		}

		// Exactly one goroutine should have rendered the CSS
		renderedCount := 0
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
		errors := make(chan error, 10)

		// Launch 10 goroutines all trying to render the same Once handle
		for i := 0; i < 10; i++ {
			wg.Add(1)
			idx := i
			go func() {
				defer wg.Done()
				var buf bytes.Buffer
				if err := onceComponent.Render(ctx, &buf); err != nil {
					errors <- err
					return
				}
				outputs[idx] = buf.String()
			}()
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Fatalf("concurrent render failed: %v", err)
		}

		// Exactly one goroutine should have rendered the content
		renderedCount := 0
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

// TestSequentialDeduplication verifies that deduplication still works
// in the sequential (non-concurrent) case
func TestSequentialDeduplication(t *testing.T) {
	t.Run("scripts deduplicate in sequential renders", func(t *testing.T) {
		ctx := templ.InitializeContext(context.Background())

		script := templ.ComponentScript{
			Name:     "testScript",
			Function: `function test() {}`,
		}

		// First render - should output script
		var buf1 bytes.Buffer
		if err := templ.RenderScriptItems(ctx, &buf1, script); err != nil {
			t.Fatalf("first render failed: %v", err)
		}

		if !strings.Contains(buf1.String(), script.Function) {
			t.Errorf("first render should output script")
		}

		// Second render - should not output script (already rendered)
		var buf2 bytes.Buffer
		if err := templ.RenderScriptItems(ctx, &buf2, script); err != nil {
			t.Fatalf("second render failed: %v", err)
		}

		if buf2.String() != "" {
			t.Errorf("second render should not output script, got: %q", buf2.String())
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
