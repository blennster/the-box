package internal

import "sync"

type Storage[K any, V any] interface {
	// Get a value from the store. The returned function is for unlocking the underlying mutex
	Load(key K) (V, bool)
	Store(key K, value V)
	Delete(key string)
}

type InMemoryStore[K comparable, V any] struct {
	m sync.Map
}

func NewInMemoryStore[K comparable, V any]() InMemoryStore[K, V] {
	return InMemoryStore[K, V]{
		m: sync.Map{},
	}
}

// Add implements Storage.
func (s *InMemoryStore[K, V]) Store(key K, value V) {
	s.m.Store(key, value)
}

// Get implements Storage.
func (s *InMemoryStore[K, V]) Load(key K) (value V, ok bool) {
	val, ok := s.m.Load(key)

	if ok {
		return val.(V), ok
	}

	var v V
	return v, ok
}

// Remove implements Storage.
func (s *InMemoryStore[K, V]) Delete(key K) {
	s.m.Delete(key)
}
