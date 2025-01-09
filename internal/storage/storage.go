package storage

import (
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"sync"
)

// StemKey represents a composite key for identifying stems by name and version.
type StemKey struct {
	Name    string
	Version string
}

// HerbariumDB is a singleton in-memory storage for managing Stems and their associated leaf instances.
type HerbariumDB struct {
	Stems map[StemKey]*models.Stem // Map of Stems, keyed by composite key
	mu    sync.RWMutex             // Mutex to handle concurrent access safely
}

// instance is the singleton instance of HerbariumDB.
var instance *HerbariumDB
var once sync.Once

// GetHerbariumDB returns the singleton instance of HerbariumDB.
func GetHerbariumDB() *HerbariumDB {
	once.Do(func() {
		instance = &HerbariumDB{
			Stems: make(map[StemKey]*models.Stem),
		}
	})
	return instance
}

// WithLock executes fn while holding the write lock.
func (s *HerbariumDB) WithLock(fn func() error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fn()
}

// WithRLock executes fn while holding the read lock.
func (s *HerbariumDB) WithRLock(fn func() error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fn()
}

func (s *HerbariumDB) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Stems = make(map[StemKey]*models.Stem)
}
