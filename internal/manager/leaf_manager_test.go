package manager

import (
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"testing"
	"time"

	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/stretchr/testify/assert"
)

// TestLeafManager_StartLeaf tests the StartLeaf method in LeafManager.
func TestLeafManager_StartLeaf(t *testing.T) {
	// Set up real in-memory repository
	leafStorage := &storage.HerbariumDB{}
	leafRepo := repos.NewLeafRepository(leafStorage)
	manager := NewLeafManager(leafRepo)

	// Call StartLeaf
	leafID, err := manager.StartLeaf("stemName", "1.0.0")

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, leafID) // Verify a leaf ID was generated

	// Verify repository state
	leafs, err := leafRepo.ListLeafs("stemName")
	assert.NoError(t, err)
	assert.Len(t, leafs, 1)              // Verify a leaf was added
	assert.Equal(t, leafID, leafs[0].ID) // Verify leaf ID matches
}

// TestLeafManager_StopLeaf tests the StopLeaf method in LeafManager.
func TestLeafManager_StopLeaf(t *testing.T) {
	// Set up real in-memory repository
	leafStorage := &storage.HerbariumDB{}
	leafRepo := repos.NewLeafRepository(leafStorage)
	manager := NewLeafManager(leafRepo)

	// Add a leaf to the repository
	leafID := "leaf123"
	err := leafRepo.AddLeaf("stemName", leafID, "haproxy-server", 12345, 8080, time.Now())
	assert.NoError(t, err)

	// Call StopLeaf
	err = manager.StopLeaf(leafID)
	assert.NoError(t, err)

	// Verify repository state
	leafs, err := leafRepo.ListLeafs("stemName")
	assert.NoError(t, err)
	assert.Empty(t, leafs) // Verify the leaf was removed
}

// TestLeafManager_GetRunningLeafs tests the GetRunningLeafs method in LeafManager.
func TestLeafManager_GetRunningLeafs(t *testing.T) {
	// Set up real in-memory repository
	leafStorage := &storage.HerbariumDB{}
	leafRepo := repos.NewLeafRepository(leafStorage)
	manager := NewLeafManager(leafRepo)

	// Add multiple leafs to the repository
	err := leafRepo.AddLeaf("stemName", "leaf1", "haproxy-server", 12345, 8080, time.Now())
	assert.NoError(t, err)
	err = leafRepo.AddLeaf("stemName", "leaf2", "haproxy-server", 12346, 8081, time.Now())
	assert.NoError(t, err)

	// Call GetRunningLeafs
	leafs, err := manager.GetRunningLeafs("stemName", "1.0.0")
	assert.NoError(t, err)

	// Verify repository state
	assert.Len(t, leafs, 2)               // Verify two leafs are returned
	assert.Equal(t, "leaf1", leafs[0].ID) // Verify first leaf ID
	assert.Equal(t, "leaf2", leafs[1].ID) // Verify second leaf ID
}

// TestStartLeafInternal_Success tests the positive scenario where leaf is started successfully.
func TestStartLeafInternal_Success(t *testing.T) {
	// Initialize in-memory DB (real database instance)
	storage := &storage.HerbariumDB{
		Stems: make(map[string]*models.Stem),
	}

	// Create a stem with no leaves initially
	stemName := "test-java-service"
	stemVersion := "v1.1"
	stem := &models.Stem{
		Name:           stemName,
		Type:           models.StemTypeDeployment,
		WorkingURL:     "/hello",
		HAProxyBackend: "java-backend",
		Version:        stemVersion,
		Environment: map[string]string{
			"GLOBAL_VAR": "production",
		},
		LeafInstances: make(map[string]*models.Leaf),
		GraftNodeLeaf: nil,
		Config: &models.StemConfig{
			Name:    "test-java-service",
			URL:     "/hello",
			Command: "java -jar hello-service.jar",
			Env: map[string]string{
				"GLOBAL_VAR": "production",
			},
			Version: "v1.1", // Or whichever version is relevant
			Dependencies: []struct {
				Name   string `yaml:"name"`
				Schema string `yaml:"schema"`
			}{
				{
					Name:   "postgres",
					Schema: "test",
				},
			},
		},
	}

	// Add the stem to the DB
	storage.Stems[stemName] = stem

	// Create the leaf manager with the real DB
	leafRepo := repos.NewLeafRepository(storage)
	leafManager := NewLeafManager(leafRepo)

	// Generate the leafID and port dynamically
	leafID := "test-leaf-123"
	leafPort := 8080

	// Call the internal method to start the leaf
	err := leafManager.startLeafInternal(stemName, stemVersion, leafID, leafPort, stem.Config)

	// Assert no error occurred
	assert.NoError(t, err)

	// Fetch the leaf from the repository and check its values
	leaf, err := leafRepo.FindLeafByID(stemName, leafID)

	// Assert leaf was added and status is updated to RUNNING
	assert.NoError(t, err)
	assert.NotNil(t, leaf)
	assert.Equal(t, leaf.Status, models.StatusRunning)
	assert.Equal(t, leaf.HAProxyServer, "java-backend")
	assert.Equal(t, leaf.Port, leafPort)
}
