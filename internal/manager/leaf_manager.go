package manager

import (
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
)

// LeafManagerInterface defines methods for managing leafs.
type LeafManagerInterface interface {
	StartLeaf(stemName, version string) (string, error)              // Starts a new leaf instance.
	StopLeaf(leafID string) error                                    // Stops a specific leaf instance.
	GetRunningLeafs(stemName, version string) ([]models.Leaf, error) // Retrieves all running leafs for a stem.
}

// LeafManager manages leaf instances and interacts with the Leaf repository.
type LeafManager struct {
	LeafRepo repos.LeafRepositoryInterface
}

// NewLeafManager creates a new LeafManager with the given repository.
func NewLeafManager(leafRepo repos.LeafRepositoryInterface) *LeafManager {
	return &LeafManager{
		LeafRepo: leafRepo,
	}
}

// StartLeaf starts a new leaf instance for the given stem and version.
func (l *LeafManager) StartLeaf(stemName, version string) (string, error) {
	// Method signature only - no implementation here.
	return "", nil
}

// StopLeaf stops a specific leaf instance by its ID.
func (l *LeafManager) StopLeaf(leafID string) error {
	// Method signature only - no implementation here.
	return nil
}

// GetRunningLeafs retrieves all running leafs for a given stem and version.
func (l *LeafManager) GetRunningLeafs(stemName, version string) ([]models.Leaf, error) {
	// Method signature only - no implementation here.
	return nil, nil
}
