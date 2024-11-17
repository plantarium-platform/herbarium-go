package repositories

import (
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"testing"
)

func TestStemRepository_AddStem(t *testing.T) {
	testStorage := storage.GetTestStorage()
	repo := NewStemRepository(testStorage)

	// Add a new stem
	err := repo.AddStem("test-stem", string(storage.StemTypeDeployment), "http://localhost:7070",
		"haproxy-test", "1.0.1", map[string]string{"TEST_ENV": "true"}, &models.ServiceConfig{})
	if err != nil {
		t.Fatalf("failed to add stem: %v", err)
	}

	// Verify that the stem was added
	stem, err := repo.FindStemByName("test-stem")
	if err != nil {
		t.Fatalf("failed to find added stem: %v", err)
	}

	if stem.Name != "test-stem" {
		t.Errorf("expected stem name to be test-stem, got %s", stem.Name)
	}
	if stem.Version != "1.0.1" {
		t.Errorf("expected stem version to be 1.0.1, got %s", stem.Version)
	}
}

func TestStemRepository_RemoveStem(t *testing.T) {
	testStorage := storage.GetTestStorage()
	repo := NewStemRepository(testStorage)

	// Remove an existing stem
	err := repo.RemoveStem("system-service")
	if err != nil {
		t.Fatalf("failed to remove stem: %v", err)
	}

	// Verify that the stem no longer exists
	_, err = repo.FindStemByName("system-service")
	if err == nil {
		t.Errorf("expected an error when finding removed stem")
	}
}

func TestStemRepository_FindStemByName(t *testing.T) {
	testStorage := storage.GetTestStorage()
	repo := NewStemRepository(testStorage)

	// Find an existing stem
	stem, err := repo.FindStemByName("user-deployment")
	if err != nil {
		t.Fatalf("failed to find stem: %v", err)
	}

	if stem.Name != "user-deployment" {
		t.Errorf("expected stem name to be user-deployment, got %s", stem.Name)
	}

	// Try to find a non-existent stem
	_, err = repo.FindStemByName("non-existent-stem")
	if err == nil {
		t.Errorf("expected an error when finding non-existent stem")
	}
}

func TestStemRepository_ListStems(t *testing.T) {
	testStorage := storage.GetTestStorage()
	repo := NewStemRepository(testStorage)

	// List all stems
	stems, err := repo.ListStems()
	if err != nil {
		t.Fatalf("failed to list stems: %v", err)
	}

	// Verify the count of stems
	if len(stems) != 2 {
		t.Errorf("expected 2 stems, got %d", len(stems))
	}

	// Verify specific stems are in the list
	found := false
	for _, stem := range stems {
		if stem.Name == "system-service" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find system-service in the stems list")
	}
}

func TestStemRepository_ReplaceStem(t *testing.T) {
	testStorage := storage.GetTestStorage()
	repo := NewStemRepository(testStorage)

	// Replace an existing stem with a new version
	err := repo.ReplaceStem("user-deployment", "1.1.0", &models.ServiceConfig{})
	if err != nil {
		t.Fatalf("failed to replace stem: %v", err)
	}

	// Verify that the stem was updated
	stem, err := repo.FindStemByName("user-deployment")
	if err != nil {
		t.Fatalf("failed to find updated stem: %v", err)
	}

	if stem.Version != "1.1.0" {
		t.Errorf("expected stem version to be 1.1.0, got %s", stem.Version)
	}
}
