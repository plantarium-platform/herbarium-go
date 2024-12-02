package manager

import (
	"fmt"
	"github.com/plantarium-platform/herbarium-go/internal/haproxy"
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"strings"
	"sync"
	"sync/atomic"
)

// StemManagerInterface defines methods for managing stems.
type StemManagerInterface interface {
	RegisterStem(config models.StemConfig) error              // Adds a new stem to the system with explicit configuration.
	UnregisterStem(name, version string) error                // Removes a stem from the system.
	FetchStemInfo(name, version string) (*models.Stem, error) // Retrieves information about a specific stem.
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
func (s *StemManager) RegisterStem(config models.StemConfig) error {
	// Define the stem key
	stemKey := storage.StemKey{Name: config.Name, Version: config.Version}

	// Check if the stem already exists
	if _, err := s.StemRepo.FetchStem(stemKey); err == nil {
		return fmt.Errorf(
			"Stem %s already exists in version %s. Please provide a new version or stop the previous one.",
			config.Name, config.Version,
		)
	}
	cleanURL := strings.TrimPrefix(config.URL, "/") // Remove leading slash
	err := s.HAProxyClient.BindStem(cleanURL)
	if err != nil {
		return fmt.Errorf("failed to bind stem backend for URL %s: %v", config.URL, err)
	}

	// Create the new stem object
	stem := &models.Stem{
		Name:           config.Name,
		Type:           models.StemTypeDeployment,
		WorkingURL:     config.URL,
		HAProxyBackend: config.URL, // Use URL as the HAProxy backend identifier
		Version:        config.Version,
		Environment:    config.Env,
		LeafInstances:  make(map[string]*models.Leaf),
		Config:         &config,
	}

	// Save the stem to the repository
	err = s.StemRepo.SaveStem(stemKey, stem)
	if err != nil {
		return fmt.Errorf("failed to save stem to repository: %v", err)
	}

	// Start the minimum number of instances if specified
	if config.MinInstances != nil && *config.MinInstances > 0 {
		for i := 0; i < *config.MinInstances; i++ {
			_, err := s.LeafManager.StartLeaf(config.Name, config.Version)
			if err != nil {
				return fmt.Errorf("failed to start leaf for stem %s version %s: %v", config.Name, config.Version, err)
			}
		}
	}

	return nil
}

// UnregisterStem removes a stem by its name and version.
func (s *StemManager) UnregisterStem(name, version string) error {
	// Step 1: Create a StemKey and fetch the stem
	stemKey := storage.StemKey{Name: name, Version: version}
	stem, err := s.StemRepo.FetchStem(stemKey)
	if err != nil {
		return fmt.Errorf("failed to fetch stem %s version %s: %v", name, version, err)
	}

	// Step 2: Retrieve all running leafs for the stem
	leafs, err := s.LeafManager.GetRunningLeafs(stemKey)
	if err != nil {
		return fmt.Errorf("failed to retrieve running leafs for stem %s version %s: %v", name, version, err)
	}

	// Step 3: Stop all leafs in parallel
	var wg sync.WaitGroup
	var stopError atomic.Value // To capture the first error, if any
	for _, leaf := range leafs {
		wg.Add(1)
		go func(leafID string) {
			defer wg.Done()
			err := s.LeafManager.StopLeaf(name, version, leafID)
			if err != nil {
				stopError.Store(err) // Capture the error
			}
		}(leaf.ID)
	}
	wg.Wait()

	// Check if any errors occurred while stopping leafs
	if storedError := stopError.Load(); storedError != nil {
		return fmt.Errorf("failed to stop leafs for stem %s version %s: %v", name, version, storedError)
	}

	// Step 4: Remove stem from HAProxy
	err = s.HAProxyClient.UnbindStem(stem.HAProxyBackend)
	if err != nil {
		return fmt.Errorf("failed to unbind stem backend for %s: %v", stem.HAProxyBackend, err)
	}

	// Step 5: Remove stem from the in-memory database
	err = s.StemRepo.DeleteStem(stemKey)
	if err != nil {
		return fmt.Errorf("failed to remove stem %s version %s from repository: %v", name, version, err)
	}

	return nil
}

// GetStemInfo retrieves information about a specific stem by name and version.
func (s *StemManager) FetchStemInfo(name, version string) (*models.Stem, error) {
	// TODO: Add logic for retrieving stem info.
	return nil, nil
}
