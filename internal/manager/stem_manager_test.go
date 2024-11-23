package manager

/*
TODO TBD Later

import (
	"testing"

	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"github.com/stretchr/testify/assert"
)

// TestStemManager_AddStem tests the AddStem method with real StemRepository and MockLeafManager.
func TestStemManager_AddStem(t *testing.T) {
	// Initialize a real in-memory DB for testing
	stemStorage := &storage.HerbariumDB{}
	stemRepo := repos.NewStemRepository(stemStorage)

	// Use the mock from test_util
	mockLeafManager := new(MockLeafManager)
	manager := NewStemManager(stemRepo, mockLeafManager)

	// Set up expectations for MockLeafManager if needed
	mockLeafManager.On("GetRunningLeafs", "stemName", "1.0.0").Return([]models.Leaf{}, nil)

	// Call the method under test
	err := manager.AddStem("stemName", "1.0.0")
	assert.NoError(t, err)

	// Optionally, assert calls made to the mock
	mockLeafManager.AssertCalled(t, "GetRunningLeafs", "stemName", "1.0.0")
}

// TestStemManager_RemoveStem tests the RemoveStem method with real StemRepository and MockLeafManager.
func TestStemManager_RemoveStem(t *testing.T) {
	// Initialize a real in-memory DB for testing
	stemStorage := &storage.HerbariumDB{}
	stemRepo := repos.NewStemRepository(stemStorage)

	// Use the mock from test_util
	mockLeafManager := new(MockLeafManager)
	manager := NewStemManager(stemRepo, mockLeafManager)

	// Set up expectations for MockLeafManager if needed
	mockLeafManager.On("GetRunningLeafs", "stemName", "1.0.0").Return([]models.Leaf{}, nil)

	// Call the method under test
	err := manager.RemoveStem("stemName", "1.0.0")
	assert.NoError(t, err)

	// Optionally, assert calls made to the mock
	mockLeafManager.AssertCalled(t, "GetRunningLeafs", "stemName", "1.0.0")
}

// TestStemManager_GetStemInfo tests the GetStemInfo method with real StemRepository and MockLeafManager.
func TestStemManager_GetStemInfo(t *testing.T) {
	// Initialize a real in-memory DB for testing
	stemStorage := &storage.HerbariumDB{}
	stemRepo := repos.NewStemRepository(stemStorage)

	// Use the mock from test_util
	mockLeafManager := new(MockLeafManager)
	manager := NewStemManager(stemRepo, mockLeafManager)

	// Set up expectations for MockLeafManager if needed
	mockLeafManager.On("GetRunningLeafs", "stemName", "1.0.0").Return([]models.Leaf{}, nil)

	// Call the method under test
	stem, err := manager.GetStemInfo("stemName", "1.0.0")
	assert.NoError(t, err)
	assert.NotNil(t, stem)

	// Optionally, assert calls made to the mock
	mockLeafManager.AssertCalled(t, "GetRunningLeafs", "stemName", "1.0.0")
}
*/
