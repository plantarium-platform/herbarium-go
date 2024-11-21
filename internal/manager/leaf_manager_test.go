package manager

import (
	"fmt"
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"os"
	"testing"
	"time"

	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestStartLeafWithPingService(t *testing.T) {
	// Set up real in-memory repository
	leafStorage := storage.GetHerbariumDB() // Access singleton HerbariumDB
	leafRepo := repos.NewLeafRepository(leafStorage)

	// Define the stem and leaf information
	stemName := "ping-service-stem"
	stemVersion := "v1.0"
	leafID := "test-leaf-123"
	leafPort := 8080

	// Define the configuration for the stem with a ping command
	stem := &models.Stem{
		Name:           stemName,
		Type:           models.StemTypeDeployment,
		WorkingURL:     "/ping",
		HAProxyBackend: "ping-backend",
		Version:        stemVersion,
		Environment: map[string]string{
			"GLOBAL_VAR": "production",
		},
		LeafInstances: make(map[string]*models.Leaf),
		GraftNodeLeaf: nil,
		Config: &models.StemConfig{
			Name:    "ping-service",
			URL:     "/ping",
			Command: "ping 127.0.0.1", // Using localhost to avoid external dependencies
			Env: map[string]string{
				"GLOBAL_VAR": "production",
			},
			Version: "v1.0",
		},
	}

	// Add the stem to the DB
	leafStorage.Stems[stemName] = stem

	// Mock HAProxyClient
	mockHAProxyClient := new(MockHAProxyClient)
	mockHAProxyClient.On("BindLeaf", "ping-backend", "ping-service", "127.0.0.1", leafPort).Return(nil)

	// Create the LeafManager with the mock HAProxyClient
	leafManager := NewLeafManager(leafRepo, mockHAProxyClient)

	// Start the leaf
	leafIDReturned, err := leafManager.StartLeaf(stemName, stemVersion)
	assert.NoError(t, err)
	assert.Equal(t, leafID, leafIDReturned)

	// Verify that BindLeaf was called with the expected arguments
	mockHAProxyClient.AssertExpectations(t)

	// Verify leaf creation in the repository
	leaf, err := leafRepo.FindLeafByID(stemName, leafID)
	assert.NoError(t, err)
	assert.NotNil(t, leaf)
	assert.Equal(t, leaf.Status, models.StatusRunning)
	assert.Equal(t, leaf.HAProxyServer, "ping-backend")
	assert.Equal(t, leaf.Port, leafPort)

	// Check that the PID is set (this assumes the process has started successfully)
	assert.Greater(t, leaf.PID, 0)

	// Check the log file for ping results
	logFilePath := fmt.Sprintf("%s.log", leafID)
	_, err = os.Stat(logFilePath) // Check if the log file exists
	assert.NoError(t, err, "log file should exist")

	logFile, err := os.Open(logFilePath)
	assert.NoError(t, err)
	defer logFile.Close()

	// Read and check the log file contents
	logContents := make([]byte, 1024)
	n, err := logFile.Read(logContents)
	assert.NoError(t, err)
	assert.Contains(t, string(logContents[:n]), "64 bytes from 127.0.0.1") // Check that ping output is present

	// Clean up the log file
	err = os.Remove(logFilePath)
	assert.NoError(t, err)
}

// TestLeafManager_GetRunningLeafs tests the GetRunningLeafs method in LeafManager.
func TestLeafManager_GetRunningLeafs(t *testing.T) {
	// Set up real in-memory repository
	leafStorage := &storage.HerbariumDB{}
	leafRepo := repos.NewLeafRepository(leafStorage)
	mockHAProxyClient := new(MockHAProxyClient)
	manager := NewLeafManager(leafRepo, mockHAProxyClient)

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
	mockHAProxyClient := new(MockHAProxyClient)
	leafManager := NewLeafManager(leafRepo, mockHAProxyClient)

	// Generate the leafID and port dynamically
	leafID := "test-leaf-123"
	leafPort := 8080

	// Call the internal method to start the leaf
	pid, err := leafManager.startLeafInternal(stemName, stemVersion, leafID, leafPort, stem.Config)

	// Assert no error occurred
	assert.NoError(t, err)

	// Fetch the leaf from the repository and check its values
	leaf, err := leafRepo.FindLeafByID(stemName, leafID)

	// Assert leaf was added and status is updated to RUNNING
	assert.NoError(t, err)
	assert.NotNil(t, leaf)
	assert.NotNil(t, pid)
	assert.Equal(t, leaf.Status, models.StatusRunning)
	assert.Equal(t, leaf.HAProxyServer, "java-backend")
	assert.Equal(t, leaf.Port, leafPort)
}
