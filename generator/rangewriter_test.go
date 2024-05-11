package generator

import (
	"bytes"
	"testing"

	"github.com/a-h/templ/parser/v2"
	"github.com/google/go-cmp/cmp"
)

func TestRangeWriter(t *testing.T) {
	w := new(bytes.Buffer)
	rw := NewRangeWriter(w)
	t.Run("indices are zero bound", func(t *testing.T) {
		if diff := cmp.Diff(parser.NewPosition(0, 0, 0), rw.Current); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("writing characters increases the col position", func(t *testing.T) {
		if _, err := rw.Write("abc"); err != nil {
			t.Fatalf("failed to write: %v", err)
		}
		if diff := cmp.Diff(parser.NewPosition(3, 0, 3), rw.Current); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("newline characters implement carriage return", func(t *testing.T) {
		if _, err := rw.Write("\n1"); err != nil {
			t.Fatalf("failed to write: %v", err)
		}
		if diff := cmp.Diff(parser.NewPosition(5, 1, 1), rw.Current); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("multi-byte characters count as 3, because that's their UTF8 representation", func(t *testing.T) {
		if _, err := rw.Write("\n你"); err != nil {
			t.Fatalf("failed to write: %v", err)
		}
		if diff := cmp.Diff(parser.NewPosition(9, 2, 3), rw.Current); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("a range is returned from each write", func(t *testing.T) {
		if _, err := rw.Write("\n"); err != nil {
			t.Fatalf("failed to write: %v", err)
		}
		r, err := rw.Write("test")
		if err != nil {
			t.Fatalf("expected successful write, got error: %v", err)
		}
		if diff := cmp.Diff(parser.NewPosition(10, 3, 0), r.From); diff != "" {
			t.Errorf("unexpected from:\n%s", diff)
		}
		if diff := cmp.Diff(parser.NewPosition(14, 3, 4), r.To); diff != "" {
			t.Errorf("unexpected to:\n%s", diff)
		}
	})
}
