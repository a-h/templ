package templ

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

func MustNewRenderLock(opts ...RenderLockOpt) *RenderLock {
	provider, err := NewRenderLock(opts...)
	if err != nil {
		panic(err)
	}
	return provider
}

type RenderLockOpt func(o *RenderLock) error

func WithLockID(id string) RenderLockOpt {
	return func(o *RenderLock) error {
		o.ID = id
		return nil
	}
}

func WithLockIDFunction(f func() (string, error)) RenderLockOpt {
	return func(o *RenderLock) error {
		o.IDFunction = f
		return nil
	}
}

// NewRenderLock proivdes a component that renders its children once per context.
func NewRenderLock(opts ...RenderLockOpt) (once *RenderLock, err error) {
	once = &RenderLock{
		IDFunction: generateLockID,
	}
	for _, opt := range opts {
		err = opt(once)
		if err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}
	if once.ID == "" {
		once.ID, err = once.IDFunction()
		if err != nil {
			return nil, fmt.Errorf("failed to generate ID: %w", err)
		}
	}
	return once, nil
}

func generateLockID() (id string, err error) {
	h := sha256.New()
	_, err = io.CopyN(h, rand.Reader, 128)
	if err != nil {
		return "", fmt.Errorf("failed to generate lock ID: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

type RenderLock struct {
	ID         string
	IDFunction func() (string, error)
}

func (o *RenderLock) Once() Component {
	return ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, v := getContext(ctx)
		if v.getHasOnceBeenRendered(o.ID) {
			return nil
		}
		v.setHasOnceBeenRendered(o.ID)
		return GetChildren(ctx).Render(ctx, w)
	})
}
