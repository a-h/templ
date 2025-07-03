package syncmap

import "testing"

func TestMap(t *testing.T) {
	t.Run("Can Set and Get values", func(t *testing.T) {
		m := New[string, int]()
		m.Set("key1", 42)
		if v, ok := m.Get("key1"); !ok || v != 42 {
			t.Errorf("Expected value 42 for key 'key1', got %d", v)
		}
	})
	t.Run("Can Delete values", func(t *testing.T) {
		m := New[string, int]()
		m.Set("key1", 42)
		m.Delete("key1")
		if _, ok := m.Get("key1"); ok {
			t.Error("Expected key 'key1' to be deleted")
		}
	})
	t.Run("CompareAndSwap", func(t *testing.T) {
		t.Run("Swaps if condition is met", func(t *testing.T) {
			m := New[string, int]()
			m.Set("key1", 42)
			swapped := m.CompareAndSwap("key1", func(previous, updated int) bool {
				return updated > previous
			}, 50)
			if !swapped {
				t.Error("Expected CompareAndSwap to succeed")
			}
			if v, ok := m.Get("key1"); !ok || v != 50 {
				t.Errorf("Expected value 50 for key 'key1', got %d", v)
			}
		})
		t.Run("Does not swap value if condition is not met", func(t *testing.T) {
			m := New[string, int]()
			m.Set("key1", 42)
			swapped := m.CompareAndSwap("key1", func(previous, updated int) bool {
				return updated > previous
			}, 30)
			if swapped {
				t.Error("Expected CompareAndSwap to fail")
			}
			if v, ok := m.Get("key1"); !ok || v != 42 {
				t.Errorf("Expected value 42 for key 'key1', got %d", v)
			}
		})
		t.Run("Swaps value if it does not exist", func(t *testing.T) {
			m := New[string, int]()
			swapped := m.CompareAndSwap("key1", func(previous, updated int) bool {
				return previous < updated
			}, 50)
			if !swapped {
				t.Error("Expected CompareAndSwap to succeed for non-existing key")
			}
			if v, ok := m.Get("key1"); !ok || v != 50 {
				t.Errorf("Expected value 50 for key 'key1', got %d", v)
			}
		})
		t.Run("UpdateIfChanged", func(t *testing.T) {
			t.Run("Swaps if the value is different", func(t *testing.T) {
				m := New[string, int]()
				m.Set("key1", 42)
				swapped := m.CompareAndSwap("key1", UpdateIfChanged, 50)
				if !swapped {
					t.Error("Expected CompareAndSwap to succeed with UpdateIfChanged")
				}
				if v, ok := m.Get("key1"); !ok || v != 50 {
					t.Errorf("Expected value 50 for key 'key1', got %d", v)
				}
			})
			t.Run("Does not swap if the value is the same", func(t *testing.T) {
				m := New[string, int]()
				m.Set("key1", 42)
				swapped := m.CompareAndSwap("key1", UpdateIfChanged, 42)
				if swapped {
					t.Error("Expected CompareAndSwap to fail with UpdateIfChanged for same value")
				}
				if v, ok := m.Get("key1"); !ok || v != 42 {
					t.Errorf("Expected value 42 for key 'key1', got %d", v)
				}
			})
		})
	})
}
