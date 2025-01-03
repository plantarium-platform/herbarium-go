package manager

import (
	"fmt"
	"github.com/plantarium-platform/herbarium-go/internal/haproxy"
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"log"
	"strings"
	"sync"
	"sync/atomic"
)

// StemManagerInterface defines methods for managing stems.
type StemManagerInterface interface {
	RegisterStem(config models.StemConfig) error             // Adds a new stem to the system with explicit configuration.
	UnregisterStem(key storage.StemKey) error                // Removes a stem from the system.
	FetchStemInfo(key storage.StemKey) (*models.Stem, error) // Retrieves information about a specific stem.
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

// RegisterStem registers a new stem in the system.
func (s *StemManager) RegisterStem(config models.StemConfig) error {
	log.Printf("Starting registration for stem: Name=%s, Version=%s, URL=%s", config.Name, config.Version, config.URL)

	// Define the stem key
	stemKey := storage.StemKey{Name: config.Name, Version: config.Version}

	// Check if the stem already exists
	if _, err := s.StemRepo.FetchStem(stemKey); err == nil {
		log.Printf("Stem %s already exists in version %s. Aborting registration.", config.Name, config.Version)
		return fmt.Errorf(
			"Stem %s already exists in version %s. Please provide a new version or stop the previous one.",
			config.Name, config.Version,
		)
	}

	cleanURL := strings.TrimPrefix(config.URL, "/") // Remove leading slash
	err := s.HAProxyClient.BindStem(cleanURL)
	if err != nil {
		log.Printf("Failed to bind stem backend for URL %s: %v", config.URL, err)
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
		log.Printf("Failed to save stem %s to repository: %v", config.Name, err)
		return fmt.Errorf("failed to save stem to repository: %v", err)
	}

	// Start the minimum number of instances if specified
	if config.MinInstances != nil && *config.MinInstances > 0 {
		log.Printf("Starting %d leaf instances for stem %s (version %s)", *config.MinInstances, config.Name, config.Version)
		for i := 0; i < *config.MinInstances; i++ {
			_, err := s.LeafManager.StartLeaf(config.Name, config.Version)
			if err != nil {
				log.Printf("Failed to start leaf for stem %s version %s: %v", config.Name, config.Version, err)
				return fmt.Errorf("failed to start leaf for stem %s version %s: %v", config.Name, config.Version, err)
			}
		}
	}

	log.Printf("Successfully registered stem: Name=%s, Version=%s, URL=%s", config.Name, config.Version, config.URL)
	return nil
}

// UnregisterStem removes a stem from the system.
func (s *StemManager) UnregisterStem(key storage.StemKey) error {
	// Step 1: Fetch the stem
	stem, err := s.StemRepo.FetchStem(key)
	if err != nil {
		return fmt.Errorf("failed to fetch stem %s version %s: %v", key.Name, key.Version, err)
	}

	// Step 2: Retrieve all running leafs for the stem
	leafs, err := s.LeafManager.GetRunningLeafs(key)
	if err != nil {
		return fmt.Errorf("failed to retrieve running leafs for stem %s version %s: %v", key.Name, key.Version, err)
	}

	// Step 3: Stop all leafs in parallel
	var wg sync.WaitGroup
	var stopError atomic.Value // To capture the first error, if any
	for _, leaf := range leafs {
		wg.Add(1)
		go func(leafID string) {
			defer wg.Done()
			err := s.LeafManager.StopLeaf(key.Name, key.Version, leafID)
			if err != nil {
				stopError.Store(err) // Capture the error
			}
		}(leaf.ID)
	}
	wg.Wait()

	// Check if any errors occurred while stopping leafs
	if storedError := stopError.Load(); storedError != nil {
		return fmt.Errorf("failed to stop leafs for stem %s version %s: %v", key.Name, key.Version, storedError)
	}

	// Step 4: Remove stem from HAProxy
	err = s.HAProxyClient.UnbindStem(stem.HAProxyBackend)
	if err != nil {
		return fmt.Errorf("failed to unbind stem backend for %s: %v", stem.HAProxyBackend, err)
	}

	// Step 5: Remove stem from the repository
	err = s.StemRepo.DeleteStem(key)
	if err != nil {
		return fmt.Errorf("failed to remove stem %s version %s from repository: %v", key.Name, key.Version, err)
	}

	return nil
}

// FetchStemInfo retrieves information about a specific stem.
func (s *StemManager) FetchStemInfo(key storage.StemKey) (*models.Stem, error) {
	return s.StemRepo.FetchStem(key)
}
