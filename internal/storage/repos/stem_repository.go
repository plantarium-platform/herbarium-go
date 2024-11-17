package repos

import (
	"fmt"
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
)

// StemRepositoryInterface defines methods for managing stems.
type StemRepositoryInterface interface {
	AddStem(name, stemType, workingURL, haproxyBackend, version string, envVars map[string]string, config *models.ServiceConfig) error
	RemoveStem(name string) error
	FindStemByName(name string) (*models.Stem, error)
	ListStems() ([]*models.Stem, error)
	ReplaceStem(name, newVersion string, newConfig *models.ServiceConfig) error
}

// StemRepository is an implementation of StemRepositoryInterface.
type StemRepository struct {
	storage *storage.HerbariumDB
}

// NewStemRepository initializes a new StemRepository with the provided storage.
func NewStemRepository(storage *storage.HerbariumDB) *StemRepository {
	return &StemRepository{
		storage: storage,
	}
}

// AddStem adds a new stem to the storage.
func (r *StemRepository) AddStem(name, stemType, workingURL, haproxyBackend, version string,
	envVars map[string]string, config *models.ServiceConfig) error {

	return r.storage.WithLock(func() error {
		if _, exists := r.storage.Stems[name]; exists {
			return fmt.Errorf("stem %s already exists", name)
		}

		r.storage.Stems[name] = &models.Stem{
			Name:           name,
			Type:           models.StemType(stemType),
			WorkingURL:     workingURL,
			HAProxyBackend: haproxyBackend,
			Version:        version,
			Environment:    envVars,
			LeafInstances:  make(map[string]*models.Leaf),
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
func (r *StemRepository) FindStemByName(name string) (*models.Stem, error) {
	var stem *models.Stem
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
func (r *StemRepository) ListStems() ([]*models.Stem, error) {
	var stems []*models.Stem
	err := r.storage.WithRLock(func() error {
		stems = make([]*models.Stem, 0, len(r.storage.Stems))
		for _, stem := range r.storage.Stems {
			stems = append(stems, stem)
		}
		return nil
	})
	return stems, err
}

// ReplaceStem replaces an existing stem with a new version.
func (r *StemRepository) ReplaceStem(name, newVersion string, newConfig *models.ServiceConfig) error {
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
