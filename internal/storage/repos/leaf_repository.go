package repos

import (
	"fmt"
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"time"
)

// LeafRepositoryInterface defines methods for managing leaves.
type LeafRepositoryInterface interface {
	AddLeaf(stemName, leafID, haproxyServer string, pid, port int, initialized time.Time) error
	RemoveLeaf(stemName, leafID string) error
	FindLeafByID(stemName, leafID string) (*models.Leaf, error)
	ListLeafs(stemName string) ([]*models.Leaf, error)
	UpdateLeafStatus(stemName string, leafID string, status models.LeafStatus) error
	SetGraftNode(stemName string, graftNode *models.Leaf) error
	GetGraftNode(stemName string) (*models.Leaf, error)
	ClearGraftNode(stemName string) error
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

// getStem is a helper to get a stem with error checking
func (r *LeafRepository) getStem(stemName string) (*models.Stem, error) {
	stem, exists := r.storage.Stems[stemName]
	if !exists {
		return nil, fmt.Errorf("stem %s not found", stemName)
	}
	return stem, nil
}

// AddLeaf adds a new leaf to a specified stem.
func (r *LeafRepository) AddLeaf(stemName, leafID, haproxyServer string, pid, port int, initialized time.Time) error {
	return r.storage.WithLock(func() error {
		stem, err := r.getStem(stemName)
		if err != nil {
			return err
		}

		if _, exists := stem.LeafInstances[leafID]; exists {
			return fmt.Errorf("leaf %s already exists in stem %s", leafID, stemName)
		}

		stem.LeafInstances[leafID] = &models.Leaf{
			ID:            leafID,
			PID:           pid,
			HAProxyServer: haproxyServer,
			Port:          port,
			Status:        models.StatusStarting,
			Initialized:   initialized,
		}

		return nil
	})
}

// RemoveLeaf removes a leaf from a specified stem.
func (r *LeafRepository) RemoveLeaf(stemName, leafID string) error {
	return r.storage.WithLock(func() error {
		stem, err := r.getStem(stemName)
		if err != nil {
			return err
		}

		if _, exists := stem.LeafInstances[leafID]; !exists {
			return fmt.Errorf("leaf %s not found in stem %s", leafID, stemName)
		}

		delete(stem.LeafInstances, leafID)
		return nil
	})
}

// FindLeafByID finds a leaf by its ID within a specified stem.
func (r *LeafRepository) FindLeafByID(stemName, leafID string) (*models.Leaf, error) {
	var leaf *models.Leaf // Declare leaf outside the closure
	err := r.storage.WithRLock(func() error {
		stem, err := r.getStem(stemName)
		if err != nil {
			return err
		}

		foundLeaf, exists := stem.LeafInstances[leafID]
		if !exists {
			return fmt.Errorf("leaf %s not found in stem %s", leafID, stemName)
		}

		leaf = foundLeaf // Assign to the outer variable
		return nil
	})
	return leaf, err
}

// ListLeafs lists all leafs for a specified stem.
func (r *LeafRepository) ListLeafs(stemName string) (leafs []*models.Leaf, err error) {
	err = r.storage.WithRLock(func() error {
		stem, err := r.getStem(stemName)
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
func (r *LeafRepository) UpdateLeafStatus(stemName, leafID string, status models.LeafStatus) error {
	return r.storage.WithLock(func() error {
		stem, err := r.getStem(stemName)
		if err != nil {
			return err
		}

		leaf, exists := stem.LeafInstances[leafID]
		if !exists {
			return fmt.Errorf("leaf %s not found in stem %s", leafID, stemName)
		}

		leaf.Status = status
		return nil
	})
}

// SetGraftNode sets a graft node for a specified stem.
func (r *LeafRepository) SetGraftNode(stemName string, graftNode *models.Leaf) error {
	return r.storage.WithLock(func() error {
		stem, err := r.getStem(stemName)
		if err != nil {
			return err
		}

		stem.GraftNodeLeaf = graftNode
		return nil
	})
}

// GetGraftNode retrieves the graft node for a specified stem.
func (r *LeafRepository) GetGraftNode(stemName string) (graftNode *models.Leaf, err error) {
	err = r.storage.WithRLock(func() error {
		stem, err := r.getStem(stemName)
		if err != nil {
			return err
		}

		graftNode = stem.GraftNodeLeaf
		return nil
	})
	return graftNode, err
}

// ClearGraftNode clears the graft node for a specified stem.
func (r *LeafRepository) ClearGraftNode(stemName string) error {
	return r.storage.WithLock(func() error {
		stem, err := r.getStem(stemName)
		if err != nil {
			return err
		}

		stem.GraftNodeLeaf = nil
		return nil
	})
}
