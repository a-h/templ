package runtime

import (
	"errors"
	"net/http/httptest"
	"testing"
)

var wasClosed bool

type closable struct {
	*httptest.ResponseRecorder
}

func (c *closable) Close() error {
	wasClosed = true
	return nil
}

func TestBuffer(t *testing.T) {
	underlying := httptest.NewRecorder()
	w, _ := GetBuffer(&closable{underlying})
	t.Run("can write to a buffer", func(t *testing.T) {
		if _, err := w.Write([]byte("A")); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("can write a string to a buffer", func(t *testing.T) {
		if _, err := w.WriteString("A"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("can flush a buffer", func(t *testing.T) {
		if err := w.Flush(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("can close a buffer", func(t *testing.T) {
		if err := w.Close(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !wasClosed {
			t.Error("expected the underlying writer to be closed")
		}
	})
	t.Run("can get the size of a buffer", func(t *testing.T) {
		if w.Size() != DefaultBufferSize {
			t.Errorf("expected %d, got %d", DefaultBufferSize, w.Size())
		}
	})
	t.Run("can reset a buffer", func(t *testing.T) {
		w.Reset(underlying)
	})
	if underlying.Body.String() != "AA" {
		t.Errorf("expected %q, got %q", "AA", underlying.Body.String())
	}
}

type failStream struct {
}

var errTest = errors.New("test error")

func (f *failStream) Write(p []byte) (n int, err error) {
	return 0, errTest
}

func (f *failStream) Close() error {
	return errTest
}

func TestBufferErrors(t *testing.T) {
	w, _ := GetBuffer(&failStream{})
	t.Run("close errors are returned", func(t *testing.T) {
		if err := w.Close(); err != errTest {
			t.Errorf("expected %v, got %v", errTest, err)
		}
	})
}
