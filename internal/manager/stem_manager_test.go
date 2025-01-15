package manager

import (
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

func TestStemManager_AddStemWithMinInstances(t *testing.T) {
	// Set up environment variable for root folder
	tempRootDir := "../../testdata"
	err := os.Setenv("PLANTARIUM_ROOT_FOLDER", tempRootDir)
	assert.NoError(t, err, "failed to set PLANTARIUM_ROOT_FOLDER environment variable")

	// Set up temporary log directory
	tempLogDir := "../../.test-logs"
	err = os.Setenv("PLANTARIUM_LOG_FOLDER", tempLogDir)
	assert.NoError(t, err, "failed to set PLANTARIUM_LOG_FOLDER environment variable")

	err = os.MkdirAll(tempLogDir, os.ModePerm)
	assert.NoError(t, err, "failed to create test log directory")
	herbariumDB := storage.GetHerbariumDB()
	herbariumDB.Clear()
	leafRepo := repos.NewLeafRepository(herbariumDB)
	stemRepo := repos.NewStemRepository(herbariumDB)

	mockHAProxyClient := new(MockHAProxyClient)
	leafManager := NewLeafManager(leafRepo, mockHAProxyClient, stemRepo)

	mockHAProxyClient.On("BindStem", "test").Return(nil)
	mockHAProxyClient.On("BindLeaf", mock.Anything, mock.Anything, "localhost", mock.AnythingOfType("int")).Return(nil)

	stemManager := NewStemManager(stemRepo, leafManager, mockHAProxyClient)

	minInstances := 2
	startMessage := "from 127.0.0.1"
	stemConfig := models.StemConfig{
		Name:         "ping-service-stem",
		URL:          "/test",
		Command:      determinePingCommand(), // Use ping command
		Env:          map[string]string{"ENV_VAR": "test"},
		Version:      "v1.0",
		MinInstances: &minInstances,
		StartMessage: &startMessage,
	}

	err = stemManager.RegisterStem(stemConfig)
	assert.NoError(t, err)

	stemKey := storage.StemKey{Name: "ping-service-stem", Version: "v1.0"}
	stem, err := stemRepo.FetchStem(stemKey)
	assert.NoError(t, err)
	assert.NotNil(t, stem)
	assert.Equal(t, "ping-service-stem", stem.Name)
	assert.Equal(t, "v1.0", stem.Version)

	assert.Equal(t, *stemConfig.MinInstances, len(stem.LeafInstances))

	for leafID, leaf := range stem.LeafInstances {
		assert.NotNil(t, leaf)
		assert.Equal(t, models.StatusRunning, leaf.Status)
		assert.Equal(t, "ping-service-stem", stem.Name)
		_, err := leafRepo.FindLeafByID(stemKey, leafID)
		assert.NoError(t, err)
	}

}
func TestStemManager_AddStemWithGraftNode(t *testing.T) {
	// Set up environment variable for root folder
	tempRootDir := "../../testdata"
	err := os.Setenv("PLANTARIUM_ROOT_FOLDER", tempRootDir)
	assert.NoError(t, err, "failed to set PLANTARIUM_ROOT_FOLDER environment variable")

	// Set up temporary log directory
	tempLogDir := "../../.test-logs"
	err = os.Setenv("PLANTARIUM_LOG_FOLDER", tempLogDir)
	assert.NoError(t, err, "failed to set PLANTARIUM_LOG_FOLDER environment variable")

	err = os.MkdirAll(tempLogDir, os.ModePerm)
	assert.NoError(t, err, "failed to create test log directory")
	herbariumDB := storage.GetHerbariumDB()
	herbariumDB.Clear()
	leafRepo := repos.NewLeafRepository(herbariumDB)
	stemRepo := repos.NewStemRepository(herbariumDB)

	mockHAProxyClient := new(MockHAProxyClient)
	leafManager := NewLeafManager(leafRepo, mockHAProxyClient, stemRepo)

	mockHAProxyClient.On("BindStem", "test").Return(nil)
	mockHAProxyClient.On("BindLeaf", mock.Anything, mock.Anything, "localhost", mock.AnythingOfType("int")).Return(nil)

	stemManager := NewStemManager(stemRepo, leafManager, mockHAProxyClient)

	stemConfig := models.StemConfig{
		Name:    "test-stem",
		URL:     "/test",
		Command: determinePingCommand(), // Use ping command
		Env:     map[string]string{"ENV_VAR": "test"},
		Version: "1.0.0",
	}

	err = stemManager.RegisterStem(stemConfig)
	assert.NoError(t, err)

	stemKey := storage.StemKey{Name: "test-stem", Version: "1.0.0"}
	stem, err := stemRepo.FetchStem(stemKey)
	assert.NoError(t, err)
	assert.NotNil(t, stem)
	assert.Equal(t, "test-stem", stem.Name)
	assert.Equal(t, "1.0.0", stem.Version)

	// Verify that no leaf instances exist
	assert.Equal(t, 0, len(stem.LeafInstances))

	// Verify that the graft node is set
	graftNode, err := leafRepo.GetGraftNode(stemKey)
	assert.NoError(t, err)
	assert.NotNil(t, graftNode)
	assert.Equal(t, "test-stem-1.0.0-graftnode", graftNode.ID)
	assert.Equal(t, models.StatusRunning, graftNode.Status)

	mockHAProxyClient.AssertExpectations(t)
}

func TestStemManager_AddStem_DuplicateError(t *testing.T) {
	herbariumDB := storage.GetHerbariumDB()
	leafRepo := repos.NewLeafRepository(herbariumDB)
	stemRepo := repos.NewStemRepository(herbariumDB)

	mockHAProxyClient := new(MockHAProxyClient)
	leafManager := NewLeafManager(leafRepo, mockHAProxyClient, stemRepo)

	stemManager := NewStemManager(stemRepo, leafManager, mockHAProxyClient)

	stemKey := storage.StemKey{Name: "test-stem", Version: "1.0.0"}
	herbariumDB.Stems[stemKey] = &models.Stem{
		Name:           "test-stem",
		Type:           models.StemTypeDeployment,
		HAProxyBackend: "test-backend",
		Version:        "1.0.0",
		LeafInstances: map[string]*models.Leaf{
			"leaf-1": {
				ID:            "leaf-1",
				Status:        models.StatusRunning,
				Port:          8000,
				PID:           12345,
				HAProxyServer: "haproxy-server",
			},
		},
		Config: &models.StemConfig{
			Name:    "test-stem",
			URL:     "/test",
			Command: "./run-test",
			Version: "1.0.0",
		},
	}

	stemConfig := models.StemConfig{
		Name:         "test-stem",
		URL:          "/test",
		Command:      "./run-test",
		Env:          map[string]string{"ENV_VAR": "test"},
		Version:      "1.0.0",
		MinInstances: nil,
	}

	err := stemManager.RegisterStem(stemConfig)
	assert.Error(t, err)
	assert.Equal(t, "Stem test-stem already exists in version 1.0.0. Please provide a new version or stop the previous one.", err.Error())
}

func TestStemManager_UnregisterStem(t *testing.T) {
	// Set up in-memory storage and repositories
	herbariumDB := storage.GetHerbariumDB()

	stemRepo := repos.NewStemRepository(herbariumDB)

	// Mock HAProxyClient and LeafManager
	mockHAProxyClient := new(MockHAProxyClient)
	mockLeafManager := new(MockLeafManager)

	// Create StemManager
	stemManager := NewStemManager(stemRepo, mockLeafManager, mockHAProxyClient)

	// Define stem details
	stemKey := storage.StemKey{Name: "test-stem", Version: "1.0.0"}
	stem := &models.Stem{
		Name:           stemKey.Name,
		Type:           models.StemTypeDeployment,
		HAProxyBackend: "/test",
		Version:        stemKey.Version,
		LeafInstances: map[string]*models.Leaf{
			"leaf1": {
				ID:            "leaf1",
				Status:        models.StatusRunning,
				Port:          8000,
				PID:           1234,
				HAProxyServer: "haproxy-leaf1",
			},
			"leaf2": {
				ID:            "leaf2",
				Status:        models.StatusRunning,
				Port:          8001,
				PID:           5678,
				HAProxyServer: "haproxy-leaf2",
			},
		},
	}
	herbariumDB.Stems[stemKey] = stem

	// Mock stopping leafs
	mockLeafManager.On("StopLeaf", stemKey.Name, stemKey.Version, "leaf1").Return(nil)
	mockLeafManager.On("StopLeaf", stemKey.Name, stemKey.Version, "leaf2").Return(nil)

	// Mock setup for GetRunningLeafs
	mockLeafManager.On("GetRunningLeafs", storage.StemKey{Name: "test-stem", Version: "1.0.0"}).
		Return([]models.Leaf{
			{
				ID:            "leaf1",
				Status:        models.StatusRunning,
				Port:          8000,
				PID:           12345,
				HAProxyServer: "haproxy-server-1",
			},
			{
				ID:            "leaf2",
				Status:        models.StatusRunning,
				Port:          8001,
				PID:           12346,
				HAProxyServer: "haproxy-server-2",
			},
		}, nil)

	// Mock HAProxy unbind
	mockHAProxyClient.On("UnbindStem", "/test").Return(nil)

	// Call UnregisterStem
	err := stemManager.UnregisterStem(stemKey)
	assert.NoError(t, err)

	// Verify all leafs are stopped
	mockLeafManager.AssertCalled(t, "StopLeaf", stemKey.Name, stemKey.Version, "leaf1")
	mockLeafManager.AssertCalled(t, "StopLeaf", stemKey.Name, stemKey.Version, "leaf2")

	// Verify HAProxy backend is unbound
	mockHAProxyClient.AssertCalled(t, "UnbindStem", "/test")

	// Verify stem is removed from in-memory database
	_, err = stemRepo.FetchStem(stemKey)
	assert.Error(t, err)
	assert.Equal(t, "stem test-stem with version 1.0.0 not found", err.Error())
}

func TestStemManager_FetchStemInfo(t *testing.T) {
	// Set up the in-memory storage
	herbariumDB := storage.GetHerbariumDB()
	stemRepo := repos.NewStemRepository(herbariumDB)

	// Initialize the StemManager with a real repository
	mockLeafManager := new(MockLeafManager) // Mock leaf manager (not used in this test)
	mockHAProxyClient := new(MockHAProxyClient)
	stemManager := NewStemManager(stemRepo, mockLeafManager, mockHAProxyClient)

	// Define a stem key
	stemKey := storage.StemKey{Name: "test-stem", Version: "1.0.0"}

	// Manually add a stem to the in-memory database
	stem := &models.Stem{
		Name:           stemKey.Name,
		Type:           models.StemTypeDeployment,
		WorkingURL:     "/test",
		HAProxyBackend: "/test",
		Version:        stemKey.Version,
		Environment:    map[string]string{"ENV_VAR": "test"},
		LeafInstances:  make(map[string]*models.Leaf),
		Config: &models.StemConfig{
			Name:    stemKey.Name,
			URL:     "/test",
			Command: "echo 'test'",
			Env:     map[string]string{"ENV_VAR": "test"},
			Version: stemKey.Version,
		},
	}
	err := stemRepo.SaveStem(stemKey, stem)
	assert.NoError(t, err, "failed to save stem to repository")

	// Call FetchStemInfo to retrieve the stem
	retrievedStem, err := stemManager.FetchStemInfo(stemKey)
	assert.NoError(t, err, "failed to fetch stem info")
	assert.NotNil(t, retrievedStem, "retrieved stem should not be nil")

	// Validate the retrieved stem data
	assert.Equal(t, stemKey.Name, retrievedStem.Name, "stem name should match")
	assert.Equal(t, stemKey.Version, retrievedStem.Version, "stem version should match")
	assert.Equal(t, "/test", retrievedStem.WorkingURL, "stem URL should match")
	assert.Equal(t, "/test", retrievedStem.HAProxyBackend, "stem HAProxy backend should match")
	assert.Equal(t, map[string]string{"ENV_VAR": "test"}, retrievedStem.Environment, "stem environment should match")
	assert.Equal(t, "echo 'test'", retrievedStem.Config.Command, "stem command should match")
}
