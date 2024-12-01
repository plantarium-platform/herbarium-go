package manager

import (
	"github.com/plantarium-platform/herbarium-go/internal/haproxy"
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
)

// StemManagerInterface defines methods for managing stems.
type StemManagerInterface interface {
	AddStem(config models.StemConfig) error                 // Adds a new stem to the system with explicit configuration.
	RemoveStem(name, version string) error                  // Removes a stem from the system.
	GetStemInfo(name, version string) (*models.Stem, error) // Retrieves information about a specific stem.
}

// StemManager is an implementation of StemManagerInterface.
type StemManager struct {
	StemRepo      *repos.StemRepository
	LeafManager   LeafManagerInterface
	HAProxyClient haproxy.HAProxyClientInterface
}

// NewStemManager creates a new instance of StemManager.
func NewStemManager(stemRepo *repos.StemRepository, leafManager LeafManagerInterface, haProxyClient haproxy.HAProxyClientInterface) *StemManager {
	return &StemManager{
		StemRepo:      stemRepo,
		LeafManager:   leafManager,
		HAProxyClient: haProxyClient,
	}
}

// AddStem adds a new stem with the given configuration.
func (s *StemManager) AddStem(config models.StemConfig) error {
	// TODO: Add logic for adding a stem using the provided configuration.
	return nil
}

// RemoveStem removes a stem by its name and version.
func (s *StemManager) RemoveStem(name, version string) error {
	// TODO: Add logic for removing a stem.
	return nil
}

// GetStemInfo retrieves information about a specific stem by name and version.
func (s *StemManager) GetStemInfo(name, version string) (*models.Stem, error) {
	// TODO: Add logic for retrieving stem info.
	return nil, nil
}
