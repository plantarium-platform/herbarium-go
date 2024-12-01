package repos

import (
	"fmt"
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
)

// StemRepositoryInterface defines methods for managing stems.
type StemRepositoryInterface interface {
	AddStem(key storage.StemKey, stemType, workingURL, haproxyBackend string, envVars map[string]string, config *models.StemConfig) error
	RemoveStem(key storage.StemKey) error
	FindStem(key storage.StemKey) (*models.Stem, error)
	ListStems() ([]*models.Stem, error)
	ReplaceStem(key storage.StemKey, newVersion string, newConfig *models.StemConfig) error
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
func (r *StemRepository) AddStem(key storage.StemKey, stemType, workingURL, haproxyBackend string,
	envVars map[string]string, config *models.StemConfig) error {

	return r.storage.WithLock(func() error {
		if _, exists := r.storage.Stems[key]; exists {
			return fmt.Errorf("stem %s with version %s already exists", key.Name, key.Version)
		}

		r.storage.Stems[key] = &models.Stem{
			Name:           key.Name,
			Type:           models.StemType(stemType),
			WorkingURL:     workingURL,
			HAProxyBackend: haproxyBackend,
			Version:        key.Version,
			Environment:    envVars,
			LeafInstances:  make(map[string]*models.Leaf),
			Config:         config,
		}

		return nil
	})
}

// RemoveStem removes a stem from the storage.
func (r *StemRepository) RemoveStem(key storage.StemKey) error {
	return r.storage.WithLock(func() error {
		if _, exists := r.storage.Stems[key]; !exists {
			return fmt.Errorf("stem %s with version %s not found", key.Name, key.Version)
		}

		delete(r.storage.Stems, key)
		return nil
	})
}

// FindStem retrieves a stem by its composite key.
func (r *StemRepository) FindStem(key storage.StemKey) (*models.Stem, error) {
	var stem *models.Stem
	err := r.storage.WithRLock(func() error {
		var exists bool
		stem, exists = r.storage.Stems[key]
		if !exists {
			return fmt.Errorf("stem %s with version %s not found", key.Name, key.Version)
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
func (r *StemRepository) ReplaceStem(key storage.StemKey, newVersion string, newConfig *models.StemConfig) error {
	return r.storage.WithLock(func() error {
		stem, exists := r.storage.Stems[key]
		if !exists {
			return fmt.Errorf("stem %s with version %s not found", key.Name, key.Version)
		}

		// Preserve existing leaf instances and environment while updating version and config
		stem.Version = newVersion
		stem.Config = newConfig

		return nil
	})
}
