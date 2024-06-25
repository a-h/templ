package templ

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
)

type flushableErrorWriter struct {
	lastFlushPos    int
	pos             int
	sb              strings.Builder
	flushedSections []string
}

func (f *flushableErrorWriter) Write(p []byte) (n int, err error) {
	n, err = f.sb.Write(p)
	if err != nil {
		return
	}
	if n < len(p) {
		err = io.ErrShortWrite
	}
	f.pos += n
	return
}

func (f *flushableErrorWriter) Flush() error {
	f.flushedSections = append(f.flushedSections, f.sb.String()[f.lastFlushPos:f.pos])
	f.lastFlushPos = f.pos
	return nil
}

type flushableWriter struct {
	lastFlushPos    int
	pos             int
	sb              strings.Builder
	flushedSections []string
}

func (f *flushableWriter) Write(p []byte) (n int, err error) {
	n, err = f.sb.Write(p)
	if err != nil {
		return
	}
	if n < len(p) {
		err = io.ErrShortWrite
	}
	f.pos += n
	return
}

func (f *flushableWriter) Flush() {
	f.flushedSections = append(f.flushedSections, f.sb.String()[f.lastFlushPos:f.pos])
	f.lastFlushPos = f.pos
}

func TestFlush(t *testing.T) {
	t.Run("errors in child components are propagated", func(t *testing.T) {
		expectedErr := fmt.Errorf("test error")
		child := ComponentFunc(func(ctx context.Context, w io.Writer) error {
			return expectedErr
		})

		sb := new(strings.Builder)
		ctx := WithChildren(context.Background(), child)

		err := Flush().Render(ctx, sb)
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}
		if err != expectedErr {
			t.Fatalf("expected error to be %v, got %v", expectedErr, err)
		}
	})
	t.Run("can render to a flushable error writer", func(t *testing.T) {
		child := ComponentFunc(func(ctx context.Context, w io.Writer) error {
			_, err := w.Write([]byte("hello"))
			return err
		})

		b := &flushableErrorWriter{}
		ctx := WithChildren(context.Background(), child)

		// Render the FlushComponent to the buffer
		if err := Flush().Render(ctx, b); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(b.flushedSections) != 1 {
			t.Fatalf("expected 1 flushed section, got %d", len(b.flushedSections))
		}
		if b.flushedSections[0] != "hello" {
			t.Fatalf("expected flushed section to be 'hello', got %q", b.flushedSections[0])
		}
	})
	t.Run("can render to a flushable writer", func(t *testing.T) {
		child := ComponentFunc(func(ctx context.Context, w io.Writer) error {
			_, err := w.Write([]byte("hello"))
			return err
		})

		b := &flushableWriter{}
		ctx := WithChildren(context.Background(), child)

		// Render the FlushComponent to the buffer
		if err := Flush().Render(ctx, b); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(b.flushedSections) != 1 {
			t.Fatalf("expected 1 flushed section, got %d", len(b.flushedSections))
		}
		if b.flushedSections[0] != "hello" {
			t.Fatalf("expected flushed section to be 'hello', got %q", b.flushedSections[0])
		}
	})
	t.Run("non-flushable streams are a no-op", func(t *testing.T) {
		sb := new(strings.Builder)
		if err := Flush().Render(context.Background(), sb); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}
