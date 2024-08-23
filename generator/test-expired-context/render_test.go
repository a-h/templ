package testexpiredcontext

import (
	"context"
	"io"
	"testing"
	"time"
)

func Test(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now())
	defer cancel()
	component := DummyComponent()
	err := component.Render(ctx, io.Discard)
	if err == nil {
		t.Error("Expected a deadline exceeded error, but got nil")
	}
}
