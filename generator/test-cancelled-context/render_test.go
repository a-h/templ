package testcancelledcontext

import (
	"context"
	"io"
	"testing"
)

func Test(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := EmptyComponent().Render(ctx, io.Discard)
	if err != context.Canceled {
		t.Errorf("expected deadline exceeded, got %v (%T)", err, err)
	}
}
