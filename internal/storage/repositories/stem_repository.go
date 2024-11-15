package repositories

import (
	"fmt"
	"github.com/plantarium-platform/herbarium-go/internal/storage"
)

// StemRepositoryInterface defines methods for managing stems.
type StemRepositoryInterface interface {
	AddStem(name, stemType, workingURL, haproxyBackend, version string, envVars map[string]string, config *storage.ServiceConfig) error
	RemoveStem(name string) error
	FindStemByName(name string) (*storage.Stem, error)
	ListStems() ([]*storage.Stem, error)
	ReplaceStem(name, newVersion string, newConfig *storage.ServiceConfig) error
}

// StemRepository is an implementation of StemRepositoryInterface.
type StemRepository struct {
	storage *storage.LeafStorage
}

// NewStemRepository initializes a new StemRepository with the provided storage.
func NewStemRepository(storage *storage.LeafStorage) *StemRepository {
	return &StemRepository{
		storage: storage,
	}
}

// AddStem adds a new stem to the storage.
func (r *StemRepository) AddStem(name, stemType, workingURL, haproxyBackend, version string,
	envVars map[string]string, config *storage.ServiceConfig) error {

	return r.storage.WithLock(func() error {
		if _, exists := r.storage.Stems[name]; exists {
			return fmt.Errorf("stem %s already exists", name)
		}

		r.storage.Stems[name] = &storage.Stem{
			Name:           name,
			Type:           storage.StemType(stemType),
			WorkingURL:     workingURL,
			HAProxyBackend: haproxyBackend,
			Version:        version,
			Environment:    envVars,
			LeafInstances:  make(map[string]*storage.Leaf),
			Config:         config,
		}

		return nil
	})
}

// RemoveStem removes a stem from the storage.
func (r *StemRepository) RemoveStem(name string) error {
	return r.storage.WithLock(func() error {
		if _, exists := r.storage.Stems[name]; !exists {
			return fmt.Errorf("stem %s not found", name)
		}

		delete(r.storage.Stems, name)
		return nil
	})
}

// FindStemByName retrieves a stem by its name.
func (r *StemRepository) FindStemByName(name string) (*storage.Stem, error) {
	var stem *storage.Stem
	err := r.storage.WithRLock(func() error {
		var exists bool
		stem, exists = r.storage.Stems[name]
		if !exists {
			return fmt.Errorf("stem %s not found", name)
		}
		return nil
	})
	return stem, err
}

// ListStems lists all stems in the storage.
func (r *StemRepository) ListStems() ([]*storage.Stem, error) {
	var stems []*storage.Stem
	err := r.storage.WithRLock(func() error {
		stems = make([]*storage.Stem, 0, len(r.storage.Stems))
		for _, stem := range r.storage.Stems {
			stems = append(stems, stem)
		}
		return nil
	})
	return stems, err
}

// ReplaceStem replaces an existing stem with a new version.
func (r *StemRepository) ReplaceStem(name, newVersion string, newConfig *storage.ServiceConfig) error {
	return r.storage.WithLock(func() error {
		stem, exists := r.storage.Stems[name]
		if !exists {
			return fmt.Errorf("stem %s not found", name)
		}

		// Preserve existing leaf instances and environment while updating version and config
		stem.Version = newVersion
		stem.Config = newConfig

		return nil
	})
}
