package redis

import (
	"sync"
	"time"
)

type Store interface {
	Set(key string, value string, expiryMs *int)
	Get(key string) (string, bool)
	Delete(key string)
}

type storeItem struct {
	value     string
	expiresAt *time.Time
}

type Nower func() time.Time

type InMemoryStore struct {
	data  map[string]storeItem
	mu    sync.RWMutex
	nower Nower
}

func NewInMemoryStore(opts ...func(*InMemoryStore)) Store {
	store := &InMemoryStore{
		data:  map[string]storeItem{},
		nower: time.Now,
	}

	for _, opt := range opts {
		opt(store)
	}

	return store
}

func WithNower(nower Nower) func(*InMemoryStore) {
	return func(s *InMemoryStore) {
		s.nower = nower
	}
}

func (s *InMemoryStore) Set(key string, value string, expiryMs *int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := storeItem{}

	if expiryMs != nil {
		now := s.nower()
		expiresAt := now.Add(time.Duration(*expiryMs) * time.Millisecond)
		item.expiresAt = &expiresAt
	}

	item.value = value
	s.data[key] = item
}

func (s *InMemoryStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, found := s.data[key]
	if !found {
		return "", false
	}

	if item.expiresAt == nil {
		return item.value, found
	}

	now := s.nower()
	if now.After(*item.expiresAt) {
		return "", false
	}

	return item.value, found
}

func (s *InMemoryStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}
