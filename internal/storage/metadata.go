package storage

import (
	"sync"
)

type FileMetadata struct {
	FileName    string
	TotalChunks int
}

type MetadataStore struct {
	data map[string]FileMetadata
	mu   sync.RWMutex
}

func NewMetadataStore() *MetadataStore {
	return &MetadataStore{
		data: make(map[string]FileMetadata),
	}
}

func (m *MetadataStore) Set(fileID string, meta FileMetadata) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[fileID] = meta
}

func (m *MetadataStore) Get(fileID string) (FileMetadata, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	meta, exists := m.data[fileID]
	return meta, exists
}
