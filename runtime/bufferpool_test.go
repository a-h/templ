package runtime

import (
	"bytes"
	"testing"
)

func TestBufferPool(t *testing.T) {
	t.Run("can get a buffer from the pool", func(t *testing.T) {
		w, existing := GetBuffer(new(bytes.Buffer))
		if w == nil {
			t.Error("expected a buffer, got nil")
		}
		if existing {
			t.Error("expected a new buffer, got an existing buffer")
		}
		err := ReleaseBuffer(w)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("can get an existing buffer from the pool", func(t *testing.T) {
		w, existing := GetBuffer(new(bytes.Buffer))
		if w == nil {
			t.Error("expected a buffer, got nil")
		}
		if existing {
			t.Error("expected a new buffer, got an existing buffer")
		}

		w, existing = GetBuffer(w)
		if w == nil {
			t.Error("expected a buffer, got nil")
		}
		if !existing {
			t.Error("expected an existing buffer, got a new buffer")
		}

		err := ReleaseBuffer(w)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("can release any writer without error", func(t *testing.T) {
		err := ReleaseBuffer(new(bytes.Buffer))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("attempting to buffer a nil writer returns nil", func(t *testing.T) {
		w, existing := GetBuffer(nil)
		if w != nil {
			t.Error("expected nil, got a buffer")
		}
		if existing {
			t.Error("expected nil, got an existing buffer")
		}
	})
}
