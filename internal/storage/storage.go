package storage

import (
	"sync"
)

type InMemoryStorage struct {
	data map[string]URLData
	mu   sync.RWMutex
	UUID int
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]URLData),
	}
}

type URLData struct {
	OriginalURL string
	OwnerID     string
}

func (s *InMemoryStorage) GetUUID() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.UUID
}

func (s *InMemoryStorage) Save(shortURL, originalURL, owner string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[shortURL] = URLData{OriginalURL: originalURL, OwnerID: owner}
	s.UUID += 1
	return nil
}

func (s *InMemoryStorage) Find(shortURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	urlData, exists := s.data[shortURL]
	if !exists {
		return "", false
	}
	return urlData.OriginalURL, true
}

func (s *InMemoryStorage) CreateTable() error {
	return nil
}
