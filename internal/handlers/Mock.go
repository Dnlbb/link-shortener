package handlers

import (
	"sync"
)

type URLData struct {
	OriginalURL string
	OwnerID     string
}

type MockRepository struct {
	data map[string]URLData
	mu   sync.RWMutex
	UUID int
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		data: make(map[string]URLData),
	}
}

func (m *MockRepository) GetUUID() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.UUID
}

func (m *MockRepository) Save(shortURL, originalURL, owner string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[shortURL] = URLData{OriginalURL: originalURL, OwnerID: owner}
	return nil
}

func (m *MockRepository) Find(shortURL string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	urlData, exists := m.data[shortURL]
	if !exists {
		return "", false
	}
	return urlData.OriginalURL, true
}

func (m *MockRepository) CreateTable() error {
	return nil
}
