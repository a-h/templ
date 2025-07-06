package syncmap

import "sync"

func New[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		m:  make(map[K]V),
		mu: sync.RWMutex{},
	}
}

type Map[K comparable, V any] struct {
	m  map[K]V
	mu sync.RWMutex
}

func (m *Map[K, V]) Get(key K) (v V, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok = m.m[key]
	return v, ok
}

func (m *Map[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[key] = value
}

func (m *Map[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.m, key)
}

func (m *Map[K, V]) CompareAndSwap(key K, shouldUpdate func(previous, updated V) bool, value V) (swapped bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.m[key]
	if ok && !shouldUpdate(v, value) {
		return false
	}
	m.m[key] = value
	return true
}

func UpdateIfChanged[V comparable](previous, updated V) bool {
	return previous != updated
}
