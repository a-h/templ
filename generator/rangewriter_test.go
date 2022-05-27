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
		if diff := cmp.Diff(parser.NewPosition(), rw.Current); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("writing characters increases the col position", func(t *testing.T) {
		rw.Write("abc")
		if diff := cmp.Diff(parser.NewPositionFromValues(3, 0, 3), rw.Current); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("newline characters implement carriage return", func(t *testing.T) {
		rw.Write("\n1")
		if diff := cmp.Diff(parser.NewPositionFromValues(5, 1, 1), rw.Current); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("multi-byte characters count as a single column position", func(t *testing.T) {
		rw.Write("\n你")
		if diff := cmp.Diff(parser.NewPositionFromValues(9, 2, 1), rw.Current); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("a range is returned from each write", func(t *testing.T) {
		rw.Write("\n")
		r, err := rw.Write("test")
		if err != nil {
			t.Fatalf("expected successful write, got error: %v", err)
		}
		if diff := cmp.Diff(parser.NewPositionFromValues(10, 3, 0), r.From); diff != "" {
			t.Errorf("unexpected from:\n%s", diff)
		}
		if diff := cmp.Diff(parser.NewPositionFromValues(14, 3, 4), r.To); diff != "" {
			t.Errorf("unexpected to:\n%s", diff)
		}
	})
}
