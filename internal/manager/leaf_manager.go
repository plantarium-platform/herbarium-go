package manager

import (
	"fmt"
	"github.com/plantarium-platform/herbarium-go/internal/haproxy"
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"net"
	"os/exec"
	"runtime"
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
	HAProxyClient haproxy.HAProxyClientInterface
}

// NewLeafManager creates a new LeafManager with the given repository and HAProxy client.
func NewLeafManager(leafRepo repos.LeafRepositoryInterface, haproxyClient haproxy.HAProxyClientInterface) *LeafManager {
	return &LeafManager{
		LeafRepo:      leafRepo,
		HAProxyClient: haproxyClient,
	}
}

// FindAvailablePort starts from a given base port and finds the first available port.
func FindAvailablePort(startPort int) (int, error) {
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
	// Generate a unique leaf ID (you can use a UUID here if preferred)
	leafID := fmt.Sprintf("%s-%s-%d", stemName, version, time.Now().Unix())

	// Define the port for the leaf
	leafPort := 8080 // Ideally, this should be dynamically assigned from a pool of available ports

	// Define the command to run the leaf application (e.g., a "ping" service)
	command := fmt.Sprintf("%s %s", "/path/to/leaf/service", "127.0.0.1")

	// Start the leaf process
	cmd := exec.Command("sh", "-c", command+fmt.Sprintf(" > %s.log", leafID))
	err := cmd.Start()
	if err != nil {
		return "", fmt.Errorf("failed to start leaf process: %v", err)
	}

	// Wait for the command to finish or do something else while the process runs in the background
	// You might want to handle the PID here

	// Save leaf details to the repository
	err = l.LeafRepo.AddLeaf(stemName, leafID, "haproxy-server", cmd.Process.Pid, leafPort, time.Now())
	if err != nil {
		return "", fmt.Errorf("failed to add leaf to repository: %v", err)
	}

	// Return the generated leaf ID
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

// startLeafInternal starts the leaf instance for the given parameters.
// Returns the PID of the started process.
// TODO port provide
func (l *LeafManager) startLeafInternal(stemName, stemVersion, leafID string, leafPort int, config *models.StemConfig) (int, error) {
	// Prepare the command to start the leaf instance
	command := fmt.Sprintf("%s > %s.log 2>&1", config.Command, leafID) // Log to file with leaf ID

	// Determine the platform and adjust the command execution
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// For Windows, use cmd.exe
		cmd = exec.Command("cmd", "/C", command)
	} else {
		// For Unix-like systems (Linux, macOS), use sh -c
		cmd = exec.Command("sh", "-c", command)
	}

	// Start the process
	err := cmd.Start()
	if err != nil {
		return 0, fmt.Errorf("failed to start process: %v", err)
	}

	// Return the PID of the running process
	return cmd.Process.Pid, nil
}
