package manager

import (
	"fmt"
	"github.com/plantarium-platform/herbarium-go/internal/haproxy"
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"log"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// LeafManagerInterface defines methods for managing leafs.
type LeafManagerInterface interface {
	StartLeaf(stemName, version string) (string, error)         // Starts a new leaf instance.
	StopLeaf(stemName, version, leafID string) error            // Stops a specific leaf instance.
	GetRunningLeafs(key storage.StemKey) ([]models.Leaf, error) // Retrieves all running leafs for a stem.
}

// LeafManager manages leaf instances and interacts with the Leaf repository and HAProxy client.
type LeafManager struct {
	LeafRepo      repos.LeafRepositoryInterface
	StemRepo      repos.StemRepositoryInterface
	HAProxyClient haproxy.HAProxyClientInterface
}

// NewLeafManager creates a new LeafManager with the given repository and HAProxy client.
func NewLeafManager(leafRepo repos.LeafRepositoryInterface, haproxyClient haproxy.HAProxyClientInterface, stemRepo repos.StemRepositoryInterface) *LeafManager {
	return &LeafManager{
		LeafRepo:      leafRepo,
		StemRepo:      stemRepo,
		HAProxyClient: haproxyClient,
	}
}

// FindAvailablePort starts from a given base port and finds the first available port.
func findAvailablePort(startPort int) (int, error) {
	for port := startPort; port < 65535; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close() // Port is available
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports found")
}

// StartLeaf starts a new leaf instance for the given stem and version.
//
// Steps:
//
//  1. **Generate a Unique Leaf ID**: A unique identifier for the leaf instance is crea
//
//  1. **Generate a Unique Leaf ID**: A unique identifier for the leaf instance is created
//     based on the stem name, version, and the current timestamp. This ensures each instance
//     has a distinct ID for identification purposes.
//     Example format: `<stemName>-<version>-<timestamp>`.
//
//  2. **Find an Available Port**: The method identifies the first available network port
//     starting from a predefined base (8000 in this case). This port will be assigned to the
//     new leaf instance to avoid conflicts with other running processes.
//
//  3. **Retrieve the Stem Configuration**: The method queries the stem repository (`StemRepo`)
//     to fetch the configuration for the specified stem. If the stem is not found, or there
//     are issues with the repository, an error is returned.
//
//  4. **Start the Leaf Process**: The `startLeafInternal` method is called to execute the
//     process associated with the leaf. This method:
//     - Spawns a new OS-level process using the command specified in the stem configuration.
//     - Redirects the process output (stdout and stderr) to a log file for later analysis.
//     - Returns the Process ID (PID) of the running process if successful.
//     If the process fails to start, an error is returned.
//
//  5. **Bind the Leaf to HAProxy**: The HAProxy client (`HAProxyClient`) binds the leaf
//     instance to the appropriate HAProxy backend specified in the stem configuration.
//     The backend is responsible for routing traffic to the leaf. If binding fails, the
//     method ensures proper error reporting.
//
//  6. **Persist Leaf Details**: The leaf's metadata, including its ID, PID, assigned port,
//     and creation timestamp, is stored in the leaf repository (`LeafRepo`) under the
//     associated stem. If this operation fails, the method returns an error but considers
//     the leaf started (since the process and HAProxy binding were successful).
//
//  7. **Return the Leaf ID**: Upon successful execution of all the above steps, the method
//     returns the generated leaf ID to the caller.
//
// Errors:
// - Returns errors for issues such as:
//   - Finding an available port.
//   - Fetching stem configuration from the repository.
//   - Starting the leaf process.
//   - Binding the leaf to HAProxy.
//   - Persisting the leaf details in the repository.
//
// Example Workflow:
// 1. A request to start a new leaf for `ping-service-stem` version `v1.0` is made.
// 2. A leaf ID is generated: `ping-service-stem-v1.0-1672574400`.
// 3. Port 8000 is found to be available and assigned.
// 4. The process is started, and a PID (e.g., 12345) is obtained.
// 5. HAProxy binds the leaf to the `ping-backend` backend on `127.0.0.1:8000`.
// 6. The repository saves the leaf details under `ping-service-stem`.
// 7. The method returns the leaf ID `ping-service-stem-v1.0-1672574400`.
func (l *LeafManager) StartLeaf(stemName, version string) (string, error) {
	// Generate a unique leaf ID based on the stem name, version, and current timestamp
	leafID := fmt.Sprintf("%s-%s-%d", stemName, version, time.Now().UnixNano())

	// Find the first available port starting from 8000
	leafPort, err := findAvailablePort(8000)
	if err != nil {
		return "", fmt.Errorf("failed to find an available port: %v", err)
	}

	// Use StemKey to retrieve the stem configuration from the database
	stemKey := storage.StemKey{Name: stemName, Version: version}
	stem, err := l.StemRepo.FetchStem(stemKey)
	if err != nil {
		return "", fmt.Errorf("failed to find stem configuration: %v", err)
	}

	// Start the leaf process and get the PID
	pid, err := l.startLeafInternal(stemName, version, leafID, leafPort, stem.Config)
	if err != nil {
		return "", fmt.Errorf("failed to start leaf process: %v", err)
	}

	// Bind the leaf service to HAProxy
	err = l.HAProxyClient.BindLeaf(stem.HAProxyBackend, leafID, "127.0.0.1", leafPort)
	if err != nil {
		return "", fmt.Errorf("failed to bind leaf to HAProxy: %v", err)
	}

	// Save leaf details to the repository
	err = l.LeafRepo.AddLeaf(stemKey, leafID, leafID, pid, leafPort, time.Now())
	if err != nil {
		return "", fmt.Errorf("leaf started, but failed to save to repository: %v", err)
	}

	return leafID, nil
}

func (l *LeafManager) StopLeaf(stemName, version, leafID string) error {
	// Use StemKey to retrieve the stem
	stemKey := storage.StemKey{Name: stemName, Version: version}
	stem, err := l.StemRepo.FetchStem(stemKey)
	if err != nil {
		return fmt.Errorf("failed to find stem %s: %v", stemKey, err)
	}

	// Find the leaf by its ID
	leaf, exists := stem.LeafInstances[leafID]
	if !exists {
		return fmt.Errorf("leaf with ID %s not found in stem %s", leafID, stemKey)
	}

	// Unbind the leaf from HAProxy
	err = l.HAProxyClient.UnbindLeaf(stem.HAProxyBackend, leaf.HAProxyServer)
	if err != nil {
		return fmt.Errorf("failed to unbind leaf from HAProxy: %v", err)
	}

	// Stop the process by PID
	process, err := os.FindProcess(leaf.PID)
	if err != nil {
		return fmt.Errorf("failed to find process with PID %d: %v", leaf.PID, err)
	}

	err = process.Kill()
	if err != nil {
		return fmt.Errorf("failed to kill process with PID %d: %v", leaf.PID, err)
	}

	// Remove the leaf from the repository
	err = l.LeafRepo.RemoveLeaf(stemKey, leafID)
	if err != nil {
		return fmt.Errorf("failed to remove leaf from repository: %v", err)
	}

	return nil
}

func (l *LeafManager) GetRunningLeafs(key storage.StemKey) ([]models.Leaf, error) {
	// Retrieve the stem using StemKey
	stem, err := l.StemRepo.FetchStem(key)
	if err != nil {
		return nil, fmt.Errorf("failed to find stem %s with version %s: %v", key.Name, key.Version, err)
	}

	// Collect all running leafs
	var runningLeafs []models.Leaf
	for _, leaf := range stem.LeafInstances {
		if leaf.Status == models.StatusRunning {
			runningLeafs = append(runningLeafs, *leaf)
		}
	}

	// Optional: Sort the leafs for consistent order
	sort.Slice(runningLeafs, func(i, j int) bool {
		return runningLeafs[i].ID < runningLeafs[j].ID
	})

	return runningLeafs, nil
}

func (l *LeafManager) startLeafInternal(stemName, stemVersion, leafID string, leafPort int, config *models.StemConfig) (int, error) {
	// Split the command into the executable (first word) and its arguments
	commandParts := strings.Fields(config.Command)
	if len(commandParts) == 0 {
		return 0, fmt.Errorf("command is empty")
	}

	// The first part is the command, the rest are arguments
	executable := commandParts[0]
	args := commandParts[1:]

	// Get log folder from environment variable or fallback to current directory
	logFolder := os.Getenv("PLANTARIUM_LOG_FOLDER")
	if logFolder == "" {
		logFolder = "."
	}

	// Ensure the log folder exists
	if err := os.MkdirAll(logFolder, os.ModePerm); err != nil {
		return 0, fmt.Errorf("failed to create log folder: %v", err)
	}

	// Prepare the log file path
	logFile := fmt.Sprintf("%s/%s.log", logFolder, leafID)

	// Create the command to execute
	cmd := exec.Command(executable, args...)

	// Log the command to be executed
	log.Printf("Executing command: %s %s", executable, strings.Join(args, " "))

	// Open the log file for writing
	logFileHandle, err := os.Create(logFile)
	if err != nil {
		return 0, fmt.Errorf("failed to create log file: %v", err)
	}

	// Set the stdout and stderr of the command to the log file
	cmd.Stdout = logFileHandle
	cmd.Stderr = logFileHandle

	// Start the process
	err = cmd.Start()
	if err != nil {
		logFileHandle.Close() // Ensure file is closed in case of error
		return 0, fmt.Errorf("failed to start leaf process: %v", err)
	}

	go func() {
		// Wait for the process to finish
		err := cmd.Wait()
		if err != nil {
			log.Printf("Process with PID %d finished with error: %v", cmd.Process.Pid, err)
		} else {
			log.Printf("Process with PID %d finished successfully", cmd.Process.Pid)
		}
		time.Sleep(100 * time.Millisecond)
		// Close the log file once the process finishes
		closeErr := logFileHandle.Close()
		if closeErr != nil {
			log.Printf("Failed to close log file %s: %v", logFile, closeErr)
		} else {
			log.Printf("Log file %s successfully closed", logFile)
		}
	}()

	// Return the PID of the running process
	return cmd.Process.Pid, nil
}
