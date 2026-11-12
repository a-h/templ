package stringwriter

import (
	"errors"
	"strings"
	"testing"
)

type failingWriter struct {
	writes int
}

func (w *failingWriter) Write(p []byte) (n int, err error) {
	w.writes++
	return 0, errors.New("write failed")
}

func TestWriteStrings(t *testing.T) {
	t.Run("strings are written in order", func(t *testing.T) {
		var sb strings.Builder
		if err := WriteStrings(&sb, "a", "", "b", "c"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sb.String() != "abc" {
			t.Errorf("expected %q, got %q", "abc", sb.String())
		}
	})
	t.Run("writing stops at the first error", func(t *testing.T) {
		w := &failingWriter{}
		if err := WriteStrings(w, "a", "b"); err == nil {
			t.Fatal("expected an error")
		}
		if w.writes != 1 {
			t.Errorf("expected 1 write, got %d", w.writes)
		}
	})
}
