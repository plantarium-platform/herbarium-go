package manager

import (
	"fmt"
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"log"
	"os"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestStartLeafWithPingService(t *testing.T) {
	// Mock the time to a fixed value using monkey patch
	fakeTime := time.Date(2023, 01, 01, 12, 0, 0, 0, time.UTC)
	patch := monkey.Patch(time.Now, func() time.Time { return fakeTime })
	t.Cleanup(patch.Unpatch) // Ensure monkey patch is undone after test

	// Set the environment variable for the log folder
	tempLogDir := "../../.test-logs"
	err := os.Setenv("PLANTARIUM_LOG_FOLDER", tempLogDir)
	assert.NoError(t, err, "failed to set PLANTARIUM_LOG_FOLDER environment variable")

	// Ensure the log folder exists
	err = os.MkdirAll(tempLogDir, os.ModePerm)
	assert.NoError(t, err, "failed to create test log directory")

	// Set up real in-memory repository
	leafStorage := storage.GetHerbariumDB() // Access singleton HerbariumDB
	leafRepo := repos.NewLeafRepository(leafStorage)

	// Create stem repository
	stemRepo := repos.NewStemRepository(leafStorage)

	// Define the stem and leaf information
	stemName := "ping-service-stem"
	stemVersion := "v1.0"
	leafPort := 8000

	// Hardcode the leafID based on fakeTime (time.Unix(1672574400))
	leafID := "ping-service-stem-v1.0-1672574400" // Hardcoded value

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
			Command: "ping 127.0.0.1 -t", // Using localhost to avoid external dependencies
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
	mockHAProxyClient.On("BindLeaf", "ping-backend", leafID, "127.0.0.1", leafPort).Return(nil)

	// Create the LeafManager with the mock HAProxyClient and stemRepo
	leafManager := NewLeafManager(leafRepo, mockHAProxyClient, stemRepo)

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
	assert.Equal(t, leaf.HAProxyServer, leafID)
	assert.Equal(t, leaf.Port, leafPort)

	// Check that the PID is set (this assumes the process has started successfully)
	assert.Greater(t, leaf.PID, 0)

	// Sleep for 100ms to allow the ping command to write to the log file
	time.Sleep(100 * time.Millisecond)

	// Check the log file for ping results
	logFilePath := fmt.Sprintf("%s/%s.log", tempLogDir, leafID)
	_, err = os.Stat(logFilePath) // Check if the log file exists
	assert.NoError(t, err, "log file should exist")

	logFile, err := os.Open(logFilePath)
	assert.NoError(t, err)

	// Read and check the log file contents
	logContents := make([]byte, 1024)
	n, err := logFile.Read(logContents)
	assert.NoError(t, err)

	// Print the contents of the log file for debugging purposes
	log.Printf("Log file contents: %s", string(logContents[:n]))

	// Check that the ping output is present (without assertions for now)
	assert.Contains(t, string(logContents[:n]), "from 127.0.0.1")
	logFile.Close()

	// Setup cleanup for all created resources
	t.Cleanup(func() {
		// Stop the process if running
		if leaf != nil {
			err := stopProcessByPID(leaf.PID)
			if err != nil {
				log.Printf("Failed to stop process with PID %d: %v", leaf.PID, err)
			}
		}

		// Remove the temporary log directory
		err = os.RemoveAll(tempLogDir)
		if err != nil {
			log.Printf("Failed to remove temporary log directory %s: %v", tempLogDir, err)
		}

		// Unset the environment variable
		os.Unsetenv("PLANTARIUM_LOG_FOLDER")
	})
}

func TestLeafManager_GetRunningLeafs(t *testing.T) {
	// Set up real in-memory repository
	leafStorage := storage.GetHerbariumDB() // Access singleton HerbariumDB
	leafRepo := repos.NewLeafRepository(leafStorage)

	// Create stem repository
	stemRepo := repos.NewStemRepository(leafStorage) // Create stem repo (based on the same storage)

	// Create and add a sample stem to the DB
	stemName := "ping-service-stem"
	stemVersion := "v1.0"
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
	leafStorage.Stems[stemName] = stem

	// Mock HAProxyClient
	mockHAProxyClient := new(MockHAProxyClient)

	// Create the LeafManager with the mock HAProxyClient and stemRepo
	leafManager := NewLeafManager(leafRepo, mockHAProxyClient, stemRepo)

	// Add multiple leafs to the repository (manually in this test scenario)
	err := leafRepo.AddLeaf(stemName, "leaf1", "haproxy-server", 12345, 8080, time.Now())
	assert.NoError(t, err)
	err = leafRepo.AddLeaf(stemName, "leaf2", "haproxy-server", 12346, 8081, time.Now())
	assert.NoError(t, err)

	// Call GetRunningLeafs
	leafs, err := leafManager.GetRunningLeafs(stemName, "1.0.0")
	assert.NoError(t, err)

	// Verify repository state
	assert.Len(t, leafs, 2)               // Verify two leafs are returned
	assert.Equal(t, "leaf1", leafs[0].ID) // Verify first leaf ID
	assert.Equal(t, "leaf2", leafs[1].ID) // Verify second leaf ID
}

func TestStartLeafInternal_Success(t *testing.T) {
	// Initialize in-memory DB (real database instance)
	storage := &storage.HerbariumDB{
		Stems: make(map[string]*models.Stem),
	}

	// Create stem repository
	stemRepo := repos.NewStemRepository(storage) // Create stem repo (based on the same storage)

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

	// Create the leaf manager with the real DB and mock HAProxy client
	leafRepo := repos.NewLeafRepository(storage)
	mockHAProxyClient := new(MockHAProxyClient)
	leafManager := NewLeafManager(leafRepo, mockHAProxyClient, stemRepo)

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

func stopProcessByPID(pid int) error {
	// Look up the process using its PID
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process with PID %d: %v", pid, err)
	}

	// Terminate the process
	err = process.Kill()
	if err != nil {
		return fmt.Errorf("failed to kill process with PID %d: %v", pid, err)
	}

	// Wait for the process to exit to make sure it's fully terminated
	_, err = process.Wait()
	if err != nil {
		return fmt.Errorf("failed to wait for process with PID %d to exit: %v", pid, err)
	}

	return nil
}
