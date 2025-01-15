package manager

import (
	"fmt"
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/stretchr/testify/mock"
	"log"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestStartLeafWithPingService(t *testing.T) {
	fakeTime := time.Date(2023, 01, 01, 12, 0, 0, 0, time.UTC)
	patch := monkey.Patch(time.Now, func() time.Time { return fakeTime })
	t.Cleanup(patch.Unpatch)

	tempLogDir := "../../.test-logs"
	err := os.Setenv("PLANTARIUM_LOG_FOLDER", tempLogDir)
	assert.NoError(t, err, "failed to set PLANTARIUM_LOG_FOLDER environment variable")
	tempRootDir := "../../testdata"
	err = os.Setenv("PLANTARIUM_ROOT_FOLDER", tempRootDir)
	assert.NoError(t, err, "failed to set PLANTARIUM_ROOT_FOLDER environment variable")
	err = os.MkdirAll(tempLogDir, os.ModePerm)
	assert.NoError(t, err, "failed to create test log directory")

	leafStorage := storage.GetHerbariumDB()
	leafRepo := repos.NewLeafRepository(leafStorage)
	stemRepo := repos.NewStemRepository(leafStorage)

	stemKey := storage.StemKey{Name: "ping-service-stem", Version: "v1.0"}
	leafPort := 8000
	leafID := "ping-service-stem-v1.0-1672574400000000000"
	startMessage := "from 127.0.0.1"
	stem := &models.Stem{
		Name:           stemKey.Name,
		Type:           models.StemTypeDeployment,
		WorkingURL:     "/ping",
		HAProxyBackend: "ping-backend",
		Version:        stemKey.Version,
		Environment: map[string]string{
			"GLOBAL_VAR": "production",
		},
		LeafInstances: make(map[string]*models.Leaf),
		Config: &models.StemConfig{
			Name:         "ping-service",
			URL:          "/ping",
			Command:      determinePingCommand(),
			StartMessage: &startMessage,
			Env: map[string]string{
				"GLOBAL_VAR": "production",
			},
			Version: stemKey.Version,
		},
	}

	leafStorage.Stems[stemKey] = stem

	mockHAProxyClient := new(MockHAProxyClient)
	mockHAProxyClient.On("BindLeaf", "ping-backend", leafID, "localhost", leafPort).Return(nil)

	leafManager := NewLeafManager(leafRepo, mockHAProxyClient, stemRepo)

	leafIDReturned, err := leafManager.StartLeaf(stemKey.Name, stemKey.Version, nil)
	assert.NoError(t, err)
	assert.Equal(t, leafID, leafIDReturned)

	mockHAProxyClient.AssertExpectations(t)

	leaf, err := leafRepo.FindLeafByID(stemKey, leafID)
	assert.NoError(t, err)
	assert.NotNil(t, leaf)
	assert.Equal(t, leaf.Status, models.StatusRunning)
	assert.Equal(t, leaf.HAProxyServer, leafID)
	assert.Equal(t, leaf.Port, leafPort)

	assert.Greater(t, leaf.PID, 0)

	time.Sleep(200 * time.Millisecond)
	logFilePath := fmt.Sprintf("%s/%s.log", tempLogDir, leafID)
	_, err = os.Stat(logFilePath)
	assert.NoError(t, err, "log file should exist")

	// Read the contents of the log file
	logFileContents, err := os.ReadFile(logFilePath)
	assert.NoError(t, err, "failed to read log file contents")

	// Validate the log file contents
	log.Printf("Log file contents:\n%s", string(logFileContents))
	assert.Contains(t, string(logFileContents), "from 127.0.0.1")

	t.Cleanup(func() {
		if leaf != nil {
			err := stopProcessByPID(leaf.PID)
			if err != nil {
				log.Printf("Failed to stop process with PID %d: %v", leaf.PID, err)
			}
		}

		err = os.RemoveAll(tempLogDir)
		if err != nil {
			log.Printf("Failed to remove temporary log directory %s: %v", tempLogDir, err)
		}

		os.Unsetenv("PLANTARIUM_LOG_FOLDER")
	})
}

func determinePingCommand() string {
	switch runtime.GOOS {
	case "windows":
		return "ping 127.0.0.1 -t" // Run indefinitely on Windows
	default:
		return "ping 127.0.0.1" // Runs with default behavior on Unix-like systems
	}
}

func TestLeafManager_GetRunningLeafs(t *testing.T) {
	leafStorage := storage.GetHerbariumDB()
	leafRepo := repos.NewLeafRepository(leafStorage)
	stemRepo := repos.NewStemRepository(leafStorage)

	stemKey := storage.StemKey{Name: "ping-service-stem", Version: "v1.0"}
	stem := &models.Stem{
		Name:           stemKey.Name,
		Type:           models.StemTypeDeployment,
		WorkingURL:     "/ping",
		HAProxyBackend: "ping-backend",
		Version:        stemKey.Version,
		Environment: map[string]string{
			"GLOBAL_VAR": "production",
		},
		LeafInstances: make(map[string]*models.Leaf),
		Config: &models.StemConfig{
			Name:    "ping-service",
			URL:     "/ping",
			Command: determinePingCommand(),
			Env: map[string]string{
				"GLOBAL_VAR": "production",
			},
			Version: stemKey.Version,
		},
	}

	leafStorage.Stems[stemKey] = stem

	mockHAProxyClient := new(MockHAProxyClient)
	leafManager := NewLeafManager(leafRepo, mockHAProxyClient, stemRepo)

	err := leafRepo.AddLeaf(stemKey, "leaf1", "haproxy-server", 12345, 8080, time.Now())
	assert.NoError(t, err)
	err = leafRepo.AddLeaf(stemKey, "leaf2", "haproxy-server", 12346, 8081, time.Now())
	assert.NoError(t, err)

	leafs, err := leafManager.GetRunningLeafs(stemKey)
	assert.NoError(t, err)

	assert.Len(t, leafs, 2)
	assert.Equal(t, "leaf1", leafs[0].ID)
	assert.Equal(t, "leaf2", leafs[1].ID)
}

func stopProcessByPID(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process with PID %d: %v", pid, err)
	}

	err = process.Kill()
	if err != nil {
		return fmt.Errorf("failed to kill process with PID %d: %v", pid, err)
	}

	_, err = process.Wait()
	if err != nil {
		return fmt.Errorf("failed to wait for process with PID %d to exit: %v", pid, err)
	}

	return nil
}

func TestStopLeaf(t *testing.T) {
	// Set up an in-memory storage and repositories
	leafStorage := storage.GetHerbariumDB()
	leafRepo := repos.NewLeafRepository(leafStorage)
	stemRepo := repos.NewStemRepository(leafStorage)

	// Define the stem key and leaf information
	stemKey := storage.StemKey{Name: "test-stem", Version: "v1.0"}
	leafID := "test-leaf-123"
	leafPort := 8000

	// Start a ping process and get its PID
	cmd := exec.Command("ping", "localhost", "-t")
	err := cmd.Start()
	assert.NoError(t, err, "failed to start ping process")

	pid := cmd.Process.Pid

	// Ensure the ping process is killed after the test
	defer func() {
		err := cmd.Process.Kill()
		if err != nil {
			log.Printf("Failed to kill ping process with PID %d: %v", pid, err)
		}
	}()

	// Manually add the stem and leaf to the in-memory database
	stem := &models.Stem{
		Name:           stemKey.Name,
		Type:           models.StemTypeDeployment,
		HAProxyBackend: "test-backend",
		Version:        stemKey.Version,
		LeafInstances: map[string]*models.Leaf{
			leafID: {
				ID:            leafID,
				Status:        models.StatusRunning,
				Port:          leafPort,
				PID:           pid,
				HAProxyServer: "haproxy-server",
			},
		},
	}
	leafStorage.Stems[stemKey] = stem

	// Mock HAProxyClient
	mockHAProxyClient := new(MockHAProxyClient)
	mockHAProxyClient.On("UnbindLeaf", "test-backend", "haproxy-server").Return(nil)

	// Create the LeafManager
	leafManager := NewLeafManager(leafRepo, mockHAProxyClient, stemRepo)

	// Stop the leaf
	err = leafManager.StopLeaf(stemKey.Name, stemKey.Version, leafID)
	assert.NoError(t, err, "failed to stop leaf")

	// Verify HAProxyClient UnbindLeaf was called with correct arguments
	mockHAProxyClient.AssertCalled(t, "UnbindLeaf", "test-backend", "haproxy-server")

	// Verify that the leaf is removed directly in the in-memory database
	stemInDB, exists := leafStorage.Stems[stemKey]
	assert.True(t, exists, "stem should still exist in the database")
	assert.Empty(t, stemInDB.LeafInstances, "stem should have no leaf instances remaining")
}

func TestStartGraftNodeLeaf(t *testing.T) {
	// Mock time for consistent ID generation
	fakeTime := time.Date(2023, 01, 01, 12, 0, 0, 0, time.UTC)
	patch := monkey.Patch(time.Now, func() time.Time { return fakeTime })
	t.Cleanup(patch.Unpatch)

	// Setup temporary log directory
	tempLogDir := "../../.test-logs"
	err := os.Setenv("PLANTARIUM_LOG_FOLDER", tempLogDir)
	assert.NoError(t, err, "failed to set PLANTARIUM_LOG_FOLDER environment variable")

	err = os.MkdirAll(tempLogDir, os.ModePerm)
	assert.NoError(t, err, "failed to create test log directory")

	// Setup in-memory storage and repositories
	leafStorage := storage.GetHerbariumDB()
	leafStorage.Clear()
	leafRepo := repos.NewLeafRepository(leafStorage)
	stemRepo := repos.NewStemRepository(leafStorage)

	stemKey := storage.StemKey{Name: "test-stem", Version: "1.0.0"}
	stem := &models.Stem{
		Name:           stemKey.Name,
		Type:           models.StemTypeDeployment,
		WorkingURL:     "/test",
		HAProxyBackend: "test-backend",
		Version:        stemKey.Version,
		Environment: map[string]string{
			"ENV_VAR": "test",
		},
		Config: &models.StemConfig{
			Name:    "test-service",
			URL:     "/test",
			Command: determinePingCommand(),
			Env: map[string]string{
				"ENV_VAR": "test",
			},
			Version: stemKey.Version,
		},
	}
	leafStorage.Stems[stemKey] = stem

	// Mock HAProxyClient
	mockHAProxyClient := new(MockHAProxyClient)
	mockHAProxyClient.On("BindStem", "test-backend").Return(nil)
	mockHAProxyClient.On("ReplaceLeaf", "test-backend", "test-stem-1.0.0-graftnode", mock.Anything, "localhost", mock.AnythingOfType("int")).Run(func(args mock.Arguments) {
		log.Printf("ReplaceLeaf called with args: %v", args)
	}).Return(nil)

	mockHAProxyClient.On("BindLeaf", "test-backend", "test-stem-1.0.0-graftnode", "localhost", mock.AnythingOfType("int")).Run(func(args mock.Arguments) {
		log.Printf("BindLeaf called with args: %v", args)
	}).Return(nil)
	// Create the LeafManager
	leafManager := NewLeafManager(leafRepo, mockHAProxyClient, stemRepo)

	// Test StartGraftNodeLeaf
	graftNodeID, err := leafManager.StartGraftNodeLeaf(stemKey.Name, stemKey.Version)
	assert.NoError(t, err, "failed to start graft node leaf")
	assert.Equal(t, "test-stem-1.0.0-graftnode", graftNodeID)

	// Verify graft node in the repository
	graftNode, err := leafRepo.GetGraftNode(stemKey)
	assert.NoError(t, err)
	assert.NotNil(t, graftNode)
	assert.Equal(t, graftNode.ID, "test-stem-1.0.0-graftnode")
	assert.Equal(t, graftNode.Status, models.StatusRunning)

	t.Cleanup(func() {
		err = os.RemoveAll(tempLogDir)
		if err != nil {
			log.Printf("Failed to remove temporary log directory %s: %v", tempLogDir, err)
		}

		os.Unsetenv("PLANTARIUM_LOG_FOLDER")
	})
}
