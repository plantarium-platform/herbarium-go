package repositories

import (
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"testing"
	"time"
)

func TestLeafRepository_AddLeaf(t *testing.T) {
	testStorage := storage.GetTestStorage()
	repo := NewLeafRepository(testStorage)

	// Add a new leaf to an existing stem
	err := repo.AddLeaf("system-service", "leaf-2", "haproxy-system", 2345, 8082, time.Now())
	if err != nil {
		t.Fatalf("failed to add leaf: %v", err)
	}

	// Verify that the leaf was added
	leaf, err := repo.FindLeafByID("system-service", "leaf-2")
	if err != nil {
		t.Fatalf("failed to find added leaf: %v", err)
	}

	if leaf.ID != "leaf-2" {
		t.Errorf("expected leaf ID to be leaf-2, got %s", leaf.ID)
	}
	if leaf.PID != 2345 {
		t.Errorf("expected leaf PID to be 2345, got %d", leaf.PID)
	}
}

func TestLeafRepository_RemoveLeaf(t *testing.T) {
	testStorage := storage.GetTestStorage()
	repo := NewLeafRepository(testStorage)

	// Remove an existing leaf
	err := repo.RemoveLeaf("user-deployment", "leaf-1")
	if err != nil {
		t.Fatalf("failed to remove leaf: %v", err)
	}

	// Verify that the leaf no longer exists
	_, err = repo.FindLeafByID("user-deployment", "leaf-1")
	if err == nil {
		t.Errorf("expected an error when finding removed leaf")
	}
}

func TestLeafRepository_FindLeafByID(t *testing.T) {
	testStorage := storage.GetTestStorage()
	repo := NewLeafRepository(testStorage)

	// Find an existing leaf
	leaf, err := repo.FindLeafByID("system-service", "leaf-1")
	if err != nil {
		t.Fatalf("failed to find leaf: %v", err)
	}

	if leaf.ID != "leaf-1" {
		t.Errorf("expected leaf ID to be leaf-1, got %s", leaf.ID)
	}

	// Try to find a non-existent leaf
	_, err = repo.FindLeafByID("system-service", "non-existent-leaf")
	if err == nil {
		t.Errorf("expected an error when finding non-existent leaf")
	}
}

func TestLeafRepository_ListLeafs(t *testing.T) {
	testStorage := storage.GetTestStorage()
	repo := NewLeafRepository(testStorage)

	// List all leafs for an existing stem
	leafs, err := repo.ListLeafs("system-service")
	if err != nil {
		t.Fatalf("failed to list leafs: %v", err)
	}

	// Verify the count of leafs
	if len(leafs) != 1 {
		t.Errorf("expected 1 leaf for system-service, got %d", len(leafs))
	}

	// Check that the leaf ID matches
	if leafs[0].ID != "leaf-1" {
		t.Errorf("expected leaf ID to be leaf-1, got %s", leafs[0].ID)
	}
}

func TestLeafRepository_UpdateLeafStatus(t *testing.T) {
	testStorage := storage.GetTestStorage()
	repo := NewLeafRepository(testStorage)

	// Update the status of an existing leaf
	err := repo.UpdateLeafStatus("system-service", "leaf-1", storage.StatusRunning)
	if err != nil {
		t.Fatalf("failed to update leaf status: %v", err)
	}

	// Verify that the status was updated
	leaf, err := repo.FindLeafByID("system-service", "leaf-1")
	if err != nil {
		t.Fatalf("failed to find leaf after status update: %v", err)
	}

	if leaf.Status != storage.StatusRunning {
		t.Errorf("expected leaf status to be RUNNING, got %s", leaf.Status)
	}
}

func TestLeafRepository_SetGraftNode(t *testing.T) {
	testStorage := storage.GetTestStorage()
	repo := NewLeafRepository(testStorage)

	graftNode := &storage.Leaf{
		ID:            "graft-leaf-new",
		PID:           3456,
		HAProxyServer: "haproxy-user",
		Port:          9093,
		Status:        storage.StatusStarting,
		Initialized:   time.Now(),
	}

	// Set the graft node for an existing stem
	err := repo.SetGraftNode("user-deployment", graftNode)
	if err != nil {
		t.Fatalf("failed to set graft node: %v", err)
	}

	// Verify that the graft node was set correctly
	retrievedNode, err := repo.GetGraftNode("user-deployment")
	if err != nil {
		t.Fatalf("failed to get graft node: %v", err)
	}

	if retrievedNode.ID != "graft-leaf-new" {
		t.Errorf("expected graft node ID to be graft-leaf-new, got %s", retrievedNode.ID)
	}
	if retrievedNode.PID != 3456 {
		t.Errorf("expected graft node PID to be 3456, got %d", retrievedNode.PID)
	}
}

func TestLeafRepository_GetGraftNode(t *testing.T) {
	testStorage := storage.GetTestStorage()
	repo := NewLeafRepository(testStorage)

	// Get an existing graft node
	graftNode, err := repo.GetGraftNode("user-deployment")
	if err != nil {
		t.Fatalf("failed to get graft node: %v", err)
	}

	if graftNode.ID != "graft-leaf" {
		t.Errorf("expected graft node ID to be graft-leaf, got %s", graftNode.ID)
	}

	// Try to get a graft node for a non-existent stem
	_, err = repo.GetGraftNode("non-existent-stem")
	if err == nil {
		t.Errorf("expected an error when getting graft node for non-existent stem")
	}
}

func TestLeafRepository_ClearGraftNode(t *testing.T) {
	testStorage := storage.GetTestStorage()
	repo := NewLeafRepository(testStorage)

	// Clear the graft node for an existing stem
	err := repo.ClearGraftNode("user-deployment")
	if err != nil {
		t.Fatalf("failed to clear graft node: %v", err)
	}

	// Verify that the graft node is cleared
	graftNode, err := repo.GetGraftNode("user-deployment")
	if err != nil {
		t.Fatalf("failed to get graft node after clearing: %v", err)
	}

	if graftNode != nil {
		t.Errorf("expected graft node to be nil after clearing, got %+v", graftNode)
	}
}
