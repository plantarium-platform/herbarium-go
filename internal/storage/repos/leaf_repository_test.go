package repos

import (
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"testing"
)

func TestStemRepository_AddStem(t *testing.T) {
	testStorage := storage.GetTestStorage()
	repo := NewStemRepository(testStorage)

	// Create a composite key for the stem
	stemKey := storage.StemKey{Name: "test-stem", Version: "1.0.1"}

	// Create a new stem object
	stem := &models.Stem{
		Name:           stemKey.Name,
		Type:           models.StemTypeDeployment,
		WorkingURL:     "http://localhost:7070",
		HAProxyBackend: "haproxy-test",
		Version:        stemKey.Version,
		Environment:    map[string]string{"TEST_ENV": "true"},
		Config:         &models.StemConfig{},
	}

	// Save the stem to the repository
	err := repo.SaveStem(stemKey, stem)
	if err != nil {
		t.Fatalf("failed to add stem: %v", err)
	}

	// Verify that the stem was added
	fetchedStem, err := repo.FetchStem(stemKey)
	if err != nil {
		t.Fatalf("failed to find added stem: %v", err)
	}

	if fetchedStem.Name != "test-stem" {
		t.Errorf("expected stem name to be test-stem, got %s", fetchedStem.Name)
	}
	if fetchedStem.Version != "1.0.1" {
		t.Errorf("expected stem version to be 1.0.1, got %s", fetchedStem.Version)
	}
}

func TestStemRepository_RemoveStem(t *testing.T) {
	testStorage := storage.GetTestStorage()
	repo := NewStemRepository(testStorage)

	// Create a composite key for the stem
	stemKey := storage.StemKey{Name: "system-service", Version: "1.0.0"}

	// Remove an existing stem
	err := repo.DeleteStem(stemKey)
	if err != nil {
		t.Fatalf("failed to remove stem: %v", err)
	}

	// Verify that the stem no longer exists
	_, err = repo.FetchStem(stemKey)
	if err == nil {
		t.Errorf("expected an error when finding removed stem")
	}
}

func TestStemRepository_FindStem(t *testing.T) {
	testStorage := storage.GetTestStorage()
	repo := NewStemRepository(testStorage)

	// Create a composite key for the stem
	stemKey := storage.StemKey{Name: "user-deployment", Version: "1.0.0"}

	// Find an existing stem
	stem, err := repo.FetchStem(stemKey)
	if err != nil {
		t.Fatalf("failed to find stem: %v", err)
	}

	if stem.Name != "user-deployment" {
		t.Errorf("expected stem name to be user-deployment, got %s", stem.Name)
	}

	// Try to find a non-existent stem
	nonExistentKey := storage.StemKey{Name: "non-existent-stem", Version: "1.0.0"}
	_, err = repo.FetchStem(nonExistentKey)
	if err == nil {
		t.Errorf("expected an error when finding non-existent stem")
	}
}

func TestStemRepository_ListStems(t *testing.T) {
	testStorage := storage.GetTestStorage()
	repo := NewStemRepository(testStorage)

	// List all stems
	stems, err := repo.GetAllStems()
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

	// Create a composite key for the stem
	stemKey := storage.StemKey{Name: "user-deployment", Version: "1.0.0"}

	// Replace an existing stem with a new version
	err := repo.UpdateStem(stemKey, "1.1.0", &models.StemConfig{})
	if err != nil {
		t.Fatalf("failed to replace stem: %v", err)
	}

	// Verify that the stem was updated
	stem, err := repo.FetchStem(stemKey)
	if err != nil {
		t.Fatalf("failed to find updated stem: %v", err)
	}

	if stem.Version != "1.1.0" {
		t.Errorf("expected stem version to be 1.1.0, got %s", stem.Version)
	}
}
