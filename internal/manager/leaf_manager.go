package manager

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/plantarium-platform/herbarium-go/internal/haproxy"
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"
)

// Global variables for timeout and sleep interval
const (
	ServiceStartupTimeout = 30 * time.Second
	ServiceCheckInterval  = 50 * time.Millisecond
)

// LeafManagerInterface defines methods for managing leafs.
type LeafManagerInterface interface {
	StartLeaf(stemName, version string, replaceServer *string) (string, error) // Starts a new leaf instance, optionally replacing an existing server in HAProxy.
	StopLeaf(stemName, version, leafID string) error                           // Stops a specific leaf instance.
	GetRunningLeafs(key storage.StemKey) ([]models.Leaf, error)                // Retrieves all running leafs for a stem.
	StartGraftNodeLeaf(stemName, version string) (string, error)               // Starts a graft node leaf and proxies requests to the real instance.
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
// 5. HAProxy binds the leaf to the `ping-backend` backend on `localhost:8000`.
// 6. The repository saves the leaf details under `ping-service-stem`.
// 7. The method returns the leaf ID `ping-service-stem-v1.0-1672574400`.
func (l *LeafManager) StartLeaf(stemName, version string, replaceServer *string) (string, error) {
	log.Printf("Starting leaf for stem: %s, version: %s", stemName, version)

	// Generate a unique leaf ID
	leafID := fmt.Sprintf("%s-%s-%d", stemName, version, time.Now().UnixNano())

	// Find an available port for the leaf
	leafPort, err := findAvailablePort(8000)
	if err != nil {
		log.Printf("Failed to find an available port: %v", err)
		return "", fmt.Errorf("failed to find an available port: %v", err)
	}

	// Retrieve stem configuration
	stemKey := storage.StemKey{Name: stemName, Version: version}
	stem, err := l.StemRepo.FetchStem(stemKey)
	if err != nil {
		log.Printf("Failed to fetch stem configuration for %s version %s: %v", stemName, version, err)
		return "", fmt.Errorf("failed to find stem configuration: %v", err)
	}

	// Start the leaf process
	pid, err := l.startLeafInternal(stemName, version, leafID, leafPort, stem.Config)
	if err != nil {
		log.Printf("Failed to start leaf process for %s version %s: %v", stemName, version, err)
		return "", fmt.Errorf("failed to start leaf process: %v", err)
	}

	// HAProxy integration
	if replaceServer != nil {
		// Replace an existing server in HAProxy
		err = l.HAProxyClient.ReplaceLeaf(stem.HAProxyBackend, *replaceServer, leafID, "localhost", leafPort)
		if err != nil {
			log.Printf("Failed to replace server %s with leaf %s in HAProxy: %v", *replaceServer, leafID, err)
			return "", fmt.Errorf("failed to replace server in HAProxy: %v", err)
		}
	} else {
		// Bind a new server to HAProxy
		err = l.HAProxyClient.BindLeaf(stem.HAProxyBackend, leafID, "localhost", leafPort)
		if err != nil {
			log.Printf("Failed to bind leaf %s to HAProxy: %v", leafID, err)
			return "", fmt.Errorf("failed to bind leaf to HAProxy: %v", err)
		}
	}

	// Save the leaf in the repository
	err = l.LeafRepo.AddLeaf(stemKey, leafID, leafID, pid, leafPort, time.Now())
	if err != nil {
		log.Printf("Leaf %s started but failed to save to repository: %v", leafID, err)
		return "", fmt.Errorf("leaf started, but failed to save to repository: %v", err)
	}

	leafURL := fmt.Sprintf("http://localhost:%d", leafPort)
	log.Printf("Leaf started successfully: ID=%s, URL=%s", leafID, leafURL)

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
func (l *LeafManager) StartGraftNodeLeaf(stemName, version string) (string, error) {
	log.Printf("Starting graft node leaf for stem: %s, version: %s", stemName, version)

	// Retrieve stem configuration
	stemKey := storage.StemKey{Name: stemName, Version: version}
	stem, err := l.StemRepo.FetchStem(stemKey)
	if err != nil {
		log.Printf("Failed to fetch stem configuration for %s version %s: %v", stemName, version, err)
		return "", fmt.Errorf("failed to find stem configuration: %v", err)
	}

	// Check if a graft node already exists
	existingGraftNode, err := l.LeafRepo.GetGraftNode(stemKey)
	if err != nil {
		log.Printf("Error retrieving existing graft node for stem %s: %v", stemName, err)
		return "", fmt.Errorf("failed to retrieve existing graft node: %v", err)
	}
	if existingGraftNode != nil {
		log.Printf("Graft node for stem %s already exists: %s", stemName, existingGraftNode.ID)
		return "", fmt.Errorf("graft node for stem %s already exists", stemName)
	}

	// Generate a unique ID for the graft node leaf
	graftNodeLeafID := fmt.Sprintf("%s-%s-graftnode", stemName, version)

	// Find an available port for the graft node
	graftNodePort, err := findAvailablePort(8000)
	if err != nil {
		log.Printf("Failed to find an available port for graft node: %v", err)
		return "", fmt.Errorf("failed to find an available port: %v", err)
	}

	// Create the graft node leaf object
	graftNodeLeaf := &models.Leaf{
		ID:            graftNodeLeafID,
		PID:           0, // Placeholder, as the process is internal
		HAProxyServer: graftNodeLeafID,
		Port:          graftNodePort,
		Status:        models.StatusRunning,
		Initialized:   time.Now(),
	}

	// Bind the graft node to the HAProxy backend
	err = l.HAProxyClient.BindLeaf(stem.HAProxyBackend, graftNodeLeaf.ID, "localhost", graftNodeLeaf.Port)
	if err != nil {
		log.Printf("Failed to bind graft node to HAProxy backend for stem %s: %v", stemName, err)
		return "", fmt.Errorf("failed to bind graft node to HAProxy backend: %v", err)
	}

	// Create and bind the graft node server
	err = l.createAndBindGraftNodeServer(stem, graftNodeLeaf)
	if err != nil {
		log.Printf("Failed to create and bind graft node for stem %s: %v", stemName, err)
		return "", err
	}

	// Save the graft node in the repository
	err = l.LeafRepo.SetGraftNode(stemKey, graftNodeLeaf)
	if err != nil {
		log.Printf("Failed to save graft node leaf for stem %s: %v", stemName, err)
		return "", fmt.Errorf("failed to save graft node leaf: %v", err)
	}

	log.Printf("Graft node leaf successfully started and bound: ID=%s, Port=%d", graftNodeLeafID, graftNodePort)
	return graftNodeLeafID, nil
}
func (l *LeafManager) createAndBindGraftNodeServer(stem *models.Stem, graftNodeLeaf *models.Leaf) error {
	// Create a new ServeMux and an HTTP server
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", graftNodeLeaf.Port),
		Handler: mux,
	}

	// Define a channel to signal server shutdown
	shutdownChan := make(chan struct{})

	mux.HandleFunc(stem.WorkingURL, func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request for graft node of stem %s", stem.Name)

		// Start the real instance using StartLeaf with graft node replacement
		stemKey := storage.StemKey{Name: stem.Name, Version: stem.Version}
		realLeafID, err := l.StartLeaf(stem.Name, stem.Version, &graftNodeLeaf.ID)
		if err != nil {
			log.Printf("Failed to start real instance for stem %s: %v", stem.Name, err)
			http.Error(w, "Internal Server Error: Unable to start real instance", http.StatusInternalServerError)
			return
		}

		// Retrieve the real leaf details
		realLeaf, err := l.LeafRepo.FindLeafByID(stemKey, realLeafID)
		if err != nil {
			log.Printf("Failed to retrieve real leaf from repository for stem %s: %v", stem.Name, err)
			http.Error(w, "Internal Server Error: Unable to retrieve real instance", http.StatusInternalServerError)
			return
		}

		// Clear the graft node from the repository
		err = l.LeafRepo.ClearGraftNode(stemKey)
		if err != nil {
			log.Printf("Failed to clear graft node for stem %s: %v", stem.Name, err)
			http.Error(w, "Internal Server Error: Unable to clear graft node", http.StatusInternalServerError)
			return
		}

		// Proxy the request to the real instance
		targetURL := fmt.Sprintf("http://localhost:%d%s", realLeaf.Port, r.URL.Path)
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("localhost:%d", realLeaf.Port),
		})
		r.URL.Path = strings.TrimPrefix(r.URL.Path, stem.WorkingURL)
		r.URL.Host = fmt.Sprintf("localhost:%d", realLeaf.Port)
		r.URL.Scheme = "http"
		r.Host = fmt.Sprintf("localhost:%d", realLeaf.Port)

		log.Printf("Forwarding request to real instance: %s%s", targetURL, r.URL.Path)
		proxy.ServeHTTP(w, r)

		// Signal to shutdown the server after the request is handled
		shutdownChan <- struct{}{}
	})

	// Start the graft node server in a goroutine
	go func() {
		log.Printf("Starting graft node server for stem %s on %s", stem.Name, server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Failed to start graft node server for stem %s: %v", stem.Name, err)
		}
	}()

	go func() {
		<-shutdownChan // Wait for the signal to stop
		log.Printf("Shutting down graft node server for stem %s", stem.Name)

		// Use context.Background() instead of nil
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down graft node server for stem %s: %v", stem.Name, err)
		}
	}()
	return nil
}
func (l *LeafManager) startLeafInternal(stemName, stemVersion, leafID string, leafPort int, config *models.StemConfig) (int, error) {
	log.Printf("Starting leaf instance with ID: %s, Stem: %s, Version: %s, Port: %d", leafID, stemName, stemVersion, leafPort)

	// Prepare working directory
	workingDir, err := getWorkingDirectory(stemName, stemVersion)
	if err != nil {
		log.Printf("Failed to get working directory for leaf %s: %v", leafID, err)
		return 0, err
	}

	// Prepare command with placeholders replaced
	command, err := prepareCommandWithTemplate(config.Command, map[string]interface{}{
		"PORT": leafPort,
	})
	if err != nil {
		log.Printf("Failed to prepare command for leaf %s: %v", leafID, err)
		return 0, err
	}

	// Log the full command that will be executed
	log.Printf("Executing command for leaf %s: %s", leafID, command)

	// Parse command
	commandParts := strings.Fields(command)
	executable := commandParts[0]
	args := commandParts[1:]

	// Create and configure the command
	cmd := exec.Command(executable, args...)
	cmd.Dir = workingDir
	cmd.Env = append(os.Environ(), formatEnvVars(config.Env)...)

	// Set up pipes
	stdoutPipe, stderrPipe, err := setupPipes(cmd)
	if err != nil {
		log.Printf("Failed to set up pipes for leaf %s: %v", leafID, err)
		return 0, err
	}

	// Set up log file
	logFile, err := setupLogFile(getLogFolder(), leafID)
	if err != nil {
		log.Printf("Failed to set up log file for leaf %s: %v", leafID, err)
		return 0, err
	}
	defer logFile.Close()

	// Process output and detect readiness
	startMessage := ""
	if config.StartMessage != nil {
		startMessage = *config.StartMessage
	}

	messageChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	// Concurrently log output and detect readiness
	go logAndDetectOutput(stdoutPipe, logFile, leafID, "stdout", startMessage, messageChan, errorChan)
	go logAndDetectOutput(stderrPipe, logFile, leafID, "stderr", startMessage, messageChan, errorChan)

	// Start the process
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start process for leaf %s: %v", leafID, err)
		return 0, fmt.Errorf("failed to start leaf process: %v", err)
	}
	log.Printf("Leaf %s process started with PID: %d", leafID, cmd.Process.Pid)

	// Handle process completion in the background
	go handleProcessCompletion(cmd, logFile, leafID)

	// Wait for readiness (port or start message)
	if err := waitForServiceToStart(leafPort, startMessage, messageChan, errorChan); err != nil {
		log.Printf("Leaf %s service not ready: %v", leafID, err)
		return 0, fmt.Errorf("leaf service not ready: %v", err)
	}

	log.Printf("Leaf %s service successfully started on port %d", leafID, leafPort)
	return cmd.Process.Pid, nil
}
func logAndDetectOutput(pipe io.ReadCloser, logFile *os.File, leafID, pipeType, startMessage string, messageChan chan string, errorChan chan error) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("[Leaf %s %s] %s", leafID, pipeType, line)
		if _, err := logFile.WriteString(line + "\n"); err != nil {
			log.Printf("[Leaf %s] Error writing to log file: %v", leafID, err)
		}
		if startMessage != "" && strings.Contains(line, startMessage) {
			messageChan <- line
		}
	}
	if err := scanner.Err(); err != nil {
		errorChan <- err
	}
}

// prepareCommandWithTemplate processes a command string with placeholders (e.g., `{{.PORT}}`) using the provided data.
func prepareCommandWithTemplate(command string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("command").Parse(command)
	if err != nil {
		return "", fmt.Errorf("failed to parse command template: %w", err)
	}

	var output bytes.Buffer
	err = tmpl.Execute(&output, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute command template: %w", err)
	}

	return output.String(), nil
}

func getLogFolder() string {
	logFolder := os.Getenv("PLANTARIUM_LOG_FOLDER")
	if logFolder == "" {
		logFolder = "."
	}
	return logFolder
}
func formatEnvVars(envVars map[string]string) []string {
	var formatted []string
	for key, value := range envVars {
		formatted = append(formatted, fmt.Sprintf("%s=%s", key, value))
	}
	return formatted
}
func setupLogFile(logFolder, leafID string) (*os.File, error) {
	if err := os.MkdirAll(logFolder, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create log folder: %v", err)
	}
	logFile := fmt.Sprintf("%s/%s.log", logFolder, leafID)
	log.Printf("[Leaf %s] Using log file: %s", leafID, logFile)
	return os.Create(logFile)
}

func getWorkingDirectory(stemName, stemVersion string) (string, error) {
	rootFolder := os.Getenv("PLANTARIUM_ROOT_FOLDER")
	if rootFolder == "" {
		return "", fmt.Errorf("PLANTARIUM_ROOT_FOLDER environment variable is not set")
	}
	workingDir := filepath.Join(rootFolder, "services", stemName, stemVersion)
	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		return "", fmt.Errorf("working directory %s does not exist: %v", workingDir, err)
	}
	return workingDir, nil
}

func setupPipes(cmd *exec.Cmd) (stdout, stderr io.ReadCloser, err error) {
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	stderr, err = cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stderr pipe: %v", err)
	}
	return
}

func handleProcessCompletion(cmd *exec.Cmd, logFile *os.File, leafID string) {
	if err := cmd.Wait(); err != nil {
		if cmd.Process != nil {
			log.Printf("[Leaf %s] Process with PID %d finished with error: %v", leafID, cmd.Process.Pid, err)
		} else {
			log.Printf("[Leaf %s] Process finished with error but PID is unavailable: %v", leafID, err)
		}
	} else {
		if cmd.Process != nil {
			log.Printf("[Leaf %s] Process with PID %d finished successfully", leafID, cmd.Process.Pid)
		} else {
			log.Printf("[Leaf %s] Process finished successfully but PID is unavailable", leafID)
		}
	}

	time.Sleep(ServiceCheckInterval)

	if err := logFile.Close(); err != nil {
		log.Printf("[Leaf %s] Failed to close log file: %v", leafID, err)
	} else {
		log.Printf("[Leaf %s] Log file closed successfully", leafID)
	}
}

func waitForServiceToStart(port int, startMessage string, messageChan chan string, errorChan chan error) error {
	start := time.Now()
	address := fmt.Sprintf("localhost:%d", port)

	for time.Since(start) < ServiceStartupTimeout {
		// Check for start message
		select {
		case msg := <-messageChan:
			if msg != "" {
				log.Printf("Detected start message: %s", msg)
				return nil
			}
		case err := <-errorChan:
			log.Printf("Error while reading logs: %v", err)
			return fmt.Errorf("error while checking start message: %v", err)
		default:
			// Check port availability
			conn, err := net.DialTimeout("tcp", address, ServiceCheckInterval)
			if err == nil {
				_ = conn.Close()
				return nil
			}
		}

		time.Sleep(ServiceCheckInterval)
	}

	return fmt.Errorf("timeout waiting for service on port %d or start message", port)
}
