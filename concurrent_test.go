package templ_test

import (
	"context"
	"io"
	"sync"
	"testing"

	"github.com/a-h/templ"
)

// TestWithChildrenConcurrentSafety validates that WithChildren doesn't cause
// data races when called concurrently. This would fail before the fix that
// moved children storage to a separate context key.
func TestWithChildrenConcurrentSafety(t *testing.T) {
	ctx := templ.InitializeContext(context.Background())
	child := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error { return nil })

	// Spawn goroutines that all call WithChildren concurrently
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// This should not cause data races
			newCtx := templ.WithChildren(ctx, child)
			_ = templ.GetChildren(newCtx)
		}()
	}
	wg.Wait()
}
