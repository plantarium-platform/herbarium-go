package repos

import (
	"fmt"
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"time"
)

// LeafRepositoryInterface defines methods for managing leaves.
type LeafRepositoryInterface interface {
	AddLeaf(stemKey storage.StemKey, leafID, haproxyServer string, pid, port int, initialized time.Time) error
	RemoveLeaf(stemKey storage.StemKey, leafID string) error
	FindLeafByID(stemKey storage.StemKey, leafID string) (*models.Leaf, error)
	ListLeafs(stemKey storage.StemKey) ([]*models.Leaf, error)
	UpdateLeafStatus(stemKey storage.StemKey, leafID string, status models.LeafStatus) error
	SetGraftNode(stemKey storage.StemKey, graftNode *models.Leaf) error
	GetGraftNode(stemKey storage.StemKey) (*models.Leaf, error)
	ClearGraftNode(stemKey storage.StemKey) error
}

// LeafRepository is an implementation of LeafRepositoryInterface.
type LeafRepository struct {
	storage *storage.HerbariumDB
}

// NewLeafRepository initializes a new LeafRepository with the provided storage.
func NewLeafRepository(storage *storage.HerbariumDB) *LeafRepository {
	return &LeafRepository{
		storage: storage,
	}
}

// getStem is a helper to get a stem with error checking using StemKey.
func (r *LeafRepository) getStem(stemKey storage.StemKey) (*models.Stem, error) {
	stem, exists := r.storage.Stems[stemKey]
	if !exists {
		return nil, fmt.Errorf("stem %s with version %s not found", stemKey.Name, stemKey.Version)
	}
	return stem, nil
}

// AddLeaf adds a new leaf to a specified stem.
func (r *LeafRepository) AddLeaf(stemKey storage.StemKey, leafID, haproxyServer string, pid, port int, initialized time.Time) error {
	return r.storage.WithLock(func() error {
		stem, err := r.getStem(stemKey)
		if err != nil {
			return err
		}

		if _, exists := stem.LeafInstances[leafID]; exists {
			return fmt.Errorf("leaf %s already exists in stem %s version %s", leafID, stemKey.Name, stemKey.Version)
		}

		stem.LeafInstances[leafID] = &models.Leaf{
			ID:            leafID,
			PID:           pid,
			HAProxyServer: haproxyServer,
			Port:          port,
			Status:        models.StatusRunning,
			Initialized:   initialized,
		}

		return nil
	})
}

// RemoveLeaf removes a leaf from a specified stem.
func (r *LeafRepository) RemoveLeaf(stemKey storage.StemKey, leafID string) error {
	return r.storage.WithLock(func() error {
		stem, err := r.getStem(stemKey)
		if err != nil {
			return err
		}

		if _, exists := stem.LeafInstances[leafID]; !exists {
			return fmt.Errorf("leaf %s not found in stem %s version %s", leafID, stemKey.Name, stemKey.Version)
		}

		delete(stem.LeafInstances, leafID)
		return nil
	})
}

// FindLeafByID finds a leaf by its ID within a specified stem.
func (r *LeafRepository) FindLeafByID(stemKey storage.StemKey, leafID string) (*models.Leaf, error) {
	var leaf *models.Leaf // Declare leaf outside the closure
	err := r.storage.WithRLock(func() error {
		stem, err := r.getStem(stemKey)
		if err != nil {
			return err
		}

		foundLeaf, exists := stem.LeafInstances[leafID]
		if !exists {
			return fmt.Errorf("leaf %s not found in stem %s version %s", leafID, stemKey.Name, stemKey.Version)
		}

		leaf = foundLeaf // Assign to the outer variable
		return nil
	})
	return leaf, err
}

// ListLeafs lists all leafs for a specified stem.
func (r *LeafRepository) ListLeafs(stemKey storage.StemKey) (leafs []*models.Leaf, err error) {
	err = r.storage.WithRLock(func() error {
		stem, err := r.getStem(stemKey)
		if err != nil {
			return err
		}

		leafs = make([]*models.Leaf, 0, len(stem.LeafInstances))
		for _, leaf := range stem.LeafInstances {
			leafs = append(leafs, leaf)
		}

		return nil
	})
	return leafs, err
}

// UpdateLeafStatus updates the status of a specified leaf.
func (r *LeafRepository) UpdateLeafStatus(stemKey storage.StemKey, leafID string, status models.LeafStatus) error {
	return r.storage.WithLock(func() error {
		stem, err := r.getStem(stemKey)
		if err != nil {
			return err
		}

		leaf, exists := stem.LeafInstances[leafID]
		if !exists {
			return fmt.Errorf("leaf %s not found in stem %s version %s", leafID, stemKey.Name, stemKey.Version)
		}

		leaf.Status = status
		return nil
	})
}

// SetGraftNode sets a graft node for a specified stem.
func (r *LeafRepository) SetGraftNode(stemKey storage.StemKey, graftNode *models.Leaf) error {
	return r.storage.WithLock(func() error {
		stem, err := r.getStem(stemKey)
		if err != nil {
			return err
		}

		stem.GraftNodeLeaf = graftNode
		return nil
	})
}

// GetGraftNode retrieves the graft node for a specified stem.
func (r *LeafRepository) GetGraftNode(stemKey storage.StemKey) (graftNode *models.Leaf, err error) {
	err = r.storage.WithRLock(func() error {
		stem, err := r.getStem(stemKey)
		if err != nil {
			return err
		}

		graftNode = stem.GraftNodeLeaf
		return nil
	})
	return graftNode, err
}

// ClearGraftNode clears the graft node for a specified stem.
func (r *LeafRepository) ClearGraftNode(stemKey storage.StemKey) error {
	return r.storage.WithLock(func() error {
		stem, err := r.getStem(stemKey)
		if err != nil {
			return err
		}

		stem.GraftNodeLeaf = nil
		return nil
	})
}
