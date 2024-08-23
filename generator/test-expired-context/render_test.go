package testexpiredcontext

import (
	"context"
	"io"
	"testing"
	"time"
)

func Test(t *testing.T) {
	ctx, _ := context.WithDeadline(context.Background(), time.Now())
	component := DummyComponent()
	err := component.Render(ctx, io.Discard)
	if err == nil {
		t.Error("Expected a deadline exceeded error, but got nil")
	}
}
