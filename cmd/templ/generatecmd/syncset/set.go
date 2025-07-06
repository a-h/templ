package syncset

import "sync"

func New[T comparable]() *Set[T] {
	return &Set[T]{
		m:  make(map[T]struct{}),
		mu: sync.RWMutex{},
	}
}

type Set[T comparable] struct {
	m  map[T]struct{}
	mu sync.RWMutex
}

func (s *Set[T]) Get(key T) (ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok = s.m[key]
	return ok
}

func (s *Set[T]) Set(key T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = struct{}{}
}

func (s *Set[T]) Delete(key T) (deleted bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, deleted = s.m[key]
	delete(s.m, key)
	return deleted
}

func (s *Set[T]) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.m)
}
