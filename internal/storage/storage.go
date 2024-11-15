package storage

import "sync"

// LeafStorage is a singleton in-memory storage for managing Stems and their associated leaf instances.
type LeafStorage struct {
	Stems map[string]*Stem // Map of Stems, each containing its own leaf instances
	mu    sync.RWMutex     // Mutex to handle concurrent access safely
}

// instance is the singleton instance of LeafStorage.
var instance *LeafStorage
var once sync.Once

// GetLeafStorage returns the singleton instance of LeafStorage.
func GetLeafStorage() *LeafStorage {
	once.Do(func() {
		instance = &LeafStorage{
			Stems: make(map[string]*Stem),
		}
	})
	return instance
}

// WithLock executes fn while holding the write lock
func (s *LeafStorage) WithLock(fn func() error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fn()
}

// WithRLock executes fn while holding the read lock
func (s *LeafStorage) WithRLock(fn func() error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fn()
}
