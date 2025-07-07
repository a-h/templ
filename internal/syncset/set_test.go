package syncset

import "testing"

func TestSet(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		t.Run("Returns false for non-existent items", func(t *testing.T) {
			s := New[string]()
			if ok := s.Get("foo"); ok {
				t.Error("expected Get to return false for non-existent item")
			}
		})
	})
	t.Run("Set", func(t *testing.T) {
		s := New[string]()
		t.Run("Can set new item", func(t *testing.T) {
			s.Set("foo")
			if !s.Get("foo") {
				t.Error("expected Get to return true for item that was just set")
			}
		})
	})
	t.Run("Delete", func(t *testing.T) {
		s := New[string]()
		t.Run("Returns false for non-existent items", func(t *testing.T) {
			if deleted := s.Delete("foo"); deleted {
				t.Error("expected Delete to return false for non-existent item")
			}
		})
		t.Run("Returns true for existing items", func(t *testing.T) {
			s.Set("foo")
			if deleted := s.Delete("foo"); !deleted {
				t.Error("expected Delete to return true for existing item")
			}
			if ok := s.Get("foo"); ok {
				t.Error("expected Get to return false for item that was just deleted")
			}
		})
	})
	t.Run("Count", func(t *testing.T) {
		t.Run("Returns 0 for empty set", func(t *testing.T) {
			s := New[string]()
			if count := s.Count(); count != 0 {
				t.Errorf("expected Count to return 0 for empty set, got %d", count)
			}
		})
		t.Run("Returns correct count for non-empty set", func(t *testing.T) {
			s := New[string]()
			s.Set("foo")
			s.Set("bar")
			if count := s.Count(); count != 2 {
				t.Errorf("expected Count to return 2 for set with two items, got %d", count)
			}
		})
		t.Run("Returns correct count after deletions", func(t *testing.T) {
			s := New[string]()
			s.Set("foo")
			s.Set("bar")
			s.Delete("foo")
			if count := s.Count(); count != 1 {
				t.Errorf("expected Count to return 1 after deleting one item, got %d", count)
			}
		})
		t.Run("Returns correct count after multiple deletions", func(t *testing.T) {
			s := New[string]()
			s.Set("foo")
			s.Set("bar")
			s.Set("baz")
			s.Delete("foo")
			s.Delete("bar")
			if count := s.Count(); count != 1 {
				t.Errorf("expected Count to return 1 after deleting two items, got %d", count)
			}
		})
	})
}
