package repos

import (
	"fmt"
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
)

// StemRepositoryInterface defines methods for managing stems.
type StemRepositoryInterface interface {
	SaveStem(key storage.StemKey, stem *models.Stem) error
	DeleteStem(key storage.StemKey) error
	FetchStem(key storage.StemKey) (*models.Stem, error)
	GetAllStems() ([]*models.Stem, error)
	UpdateStem(key storage.StemKey, newVersion string, newConfig *models.StemConfig) error
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

// RegisterStem saves a new stem to the storage.
func (r *StemRepository) SaveStem(key storage.StemKey, stem *models.Stem) error {
	return r.storage.WithLock(func() error {
		if _, exists := r.storage.Stems[key]; exists {
			return fmt.Errorf("stem %s with version %s already exists", key.Name, key.Version)
		}

		r.storage.Stems[key] = stem
		return nil
	})
}

// UnregisterStem removes a stem from the storage.
func (r *StemRepository) DeleteStem(key storage.StemKey) error {
	return r.storage.WithLock(func() error {
		if _, exists := r.storage.Stems[key]; !exists {
			return fmt.Errorf("stem %s with version %s not found", key.Name, key.Version)
		}

		delete(r.storage.Stems, key)
		return nil
	})
}

// FindStem retrieves a stem by its composite key.
func (r *StemRepository) FetchStem(key storage.StemKey) (*models.Stem, error) {
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
func (r *StemRepository) GetAllStems() ([]*models.Stem, error) {
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
func (r *StemRepository) UpdateStem(key storage.StemKey, newVersion string, newConfig *models.StemConfig) error {
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
