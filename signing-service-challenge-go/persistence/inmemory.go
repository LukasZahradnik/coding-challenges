package persistence

import (
	"errors"
	"maps"
	"slices"
	"sync"
)

var (
	ErrItemExists   = errors.New("item already exists")
	ErrItemNotFound = errors.New("item not found")
)

type Store[T any] interface {
	Store(id string, item T) error
	Set(id string, item T) error
	Get(id string) (*T, error)
	List() ([]T, error)
	Lock(id string) error
	Unlock(id string) error
}

type InMemoryStore[T any] struct {
	storage map[string]T
	locks   map[string]*sync.RWMutex
	mu      sync.RWMutex
}

func NewInMemoryStore[T any]() *InMemoryStore[T] {
	return &InMemoryStore[T]{
		storage: make(map[string]T),
		locks:   make(map[string]*sync.RWMutex),
	}
}

func (s *InMemoryStore[T]) Store(id string, item T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.storage[id]; ok {
		return ErrItemExists
	}

	s.storage[id] = item
	s.locks[id] = &sync.RWMutex{}

	return nil
}

func (s *InMemoryStore[T]) Set(id string, item T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.storage[id] = item

	return nil
}

func (s *InMemoryStore[T]) Get(id string) (*T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.storage[id]
	if !ok {
		return nil, ErrItemNotFound
	}

	return &item, nil
}

func (s *InMemoryStore[T]) List() ([]T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.storage) == 0 {
		return []T{}, nil
	}

	return slices.Collect(maps.Values(s.storage)), nil
}

func (s *InMemoryStore[T]) Lock(id string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lock, ok := s.locks[id]
	if !ok {
		return ErrItemNotFound
	}

	lock.Lock()

	return nil
}

func (s *InMemoryStore[T]) Unlock(id string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lock, ok := s.locks[id]
	if !ok {
		return ErrItemNotFound
	}

	lock.Unlock()

	return nil
}
