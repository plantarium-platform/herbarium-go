package manager

import (
	"fmt"
	"github.com/plantarium-platform/herbarium-go/internal/haproxy"
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

// LeafManagerInterface defines methods for managing leafs.
type LeafManagerInterface interface {
	StartLeaf(stemName, version string) (string, error)              // Starts a new leaf instance.
	StopLeaf(leafID string) error                                    // Stops a specific leaf instance.
	GetRunningLeafs(stemName, version string) ([]models.Leaf, error) // Retrieves all running leafs for a stem.
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
func (l *LeafManager) StartLeaf(stemName, version string) (string, error) {
	// Generate a unique leaf ID based on the stem name, version, and current timestamp
	leafID := fmt.Sprintf("%s-%s-%d", stemName, version, time.Now().Unix())

	// Find the first available port starting from 8000
	leafPort, err := findAvailablePort(8000)
	if err != nil {
		return "", fmt.Errorf("failed to find an available port: %v", err)
	}

	// Retrieve the stem configuration from the database
	stem, err := l.StemRepo.FindStemByName(stemName)
	if err != nil {
		return "", fmt.Errorf("failed to find stem configuration: %v", err)
	}

	// Start the leaf process and get the PID
	pid, err := l.startLeafInternal(stemName, version, leafID, leafPort, stem.Config)
	if err != nil {
		// If starting the leaf fails, return an error
		return "", fmt.Errorf("failed to start leaf process: %v", err)
	}

	// Bind the leaf service to HAProxy
	err = l.HAProxyClient.BindLeaf(stem.HAProxyBackend, leafID, "127.0.0.1", leafPort)
	if err != nil {
		return "", fmt.Errorf("failed to bind leaf to HAProxy: %v", err)
	}

	// Save leaf details to the repository
	err = l.LeafRepo.AddLeaf(stemName, leafID, leafID, pid, leafPort, time.Now())
	if err != nil {
		// If saving the leaf to the repository fails, handle it as a secondary error
		// Return the leaf ID even if the repository operation fails
		return "", fmt.Errorf("leaf started, but failed to save to repository: %v", err)
	}

	// Return the generated leaf ID upon successful creation
	return leafID, nil
}

// StopLeaf stops a specific leaf instance by its ID.
func (l *LeafManager) StopLeaf(leafID string) error {
	// Method signature only - no implementation here.
	return nil
}

// GetRunningLeafs retrieves all running leafs for a given stem and version.
func (l *LeafManager) GetRunningLeafs(stemName, version string) ([]models.Leaf, error) {
	// Method signature only - no implementation here.
	return nil, nil
}

// StartLeafInternal starts the leaf instance for the given parameters and manages log redirection and cleanup.
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
