package templ_test

import (
	"context"
	"io"
	"sync"
	"testing"

	"github.com/a-h/templ"
)

func TestWithChildrenConcurrentSafety(t *testing.T) {
	ctx := templ.InitializeContext(context.Background())
	child := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error { return nil })

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			newCtx := templ.WithChildren(ctx, child)
			_ = templ.GetChildren(newCtx)
		})
	}
	wg.Wait()
}
