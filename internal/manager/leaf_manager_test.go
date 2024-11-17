package manager

import (
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
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
