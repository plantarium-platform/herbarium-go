package manager

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/plantarium-platform/herbarium-go/internal/haproxy"
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"gopkg.in/yaml.v2"
)

// PlatformManagerInterface defines the methods for managing the platform lifecycle.
type PlatformManagerInterface interface {
	InitializePlatform() error // Entry point for platform initialization.
	StopPlatform() error       // Gracefully stops the platform and cleans up resources.
}

// Service represents a service with its configuration and version directory.
type Service struct {
	Config     models.StemConfig
	VersionDir string
}

// PlatformManager implements PlatformManagerInterface.
type PlatformManager struct {
	StemManager StemManagerInterface
	LeafManager LeafManagerInterface
	BasePath    string
	isWindows   bool
	Config      *models.GlobalConfig
}

// NewPlatformManager creates a new instance of PlatformManager with the required dependencies (manual DI for tests).
func NewPlatformManager(
	stemManager StemManagerInterface,
	leafManager LeafManagerInterface,
	config *models.GlobalConfig,
) *PlatformManager {
	return &PlatformManager{
		StemManager: stemManager,
		LeafManager: leafManager,
		BasePath:    config.Plantarium.RootFolder,
		Config:      config,
		isWindows:   runtime.GOOS == "windows",
	}
}

// NewPlatformManagerWithDI creates a new PlatformManager instance with all dependencies initialized (production use).
func NewPlatformManagerWithDI() (*PlatformManager, error) {
	// Load global configuration
	config, err := loadGlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load global configuration: %w", err)
	}

	// Shared HAProxyConfig
	haproxyConfig := haproxy.HAProxyConfig{
		APIURL:   config.HAProxy.URL,
		Username: config.HAProxy.Login,
		Password: config.HAProxy.Password,
	}

	// Initialize HAProxyConfigurationManager and HAProxyClient
	haproxyConfigManager := haproxy.NewHAProxyConfigurationManager(haproxyConfig)
	haproxyClient := haproxy.NewHAProxyClient(haproxyConfig, haproxyConfigManager)

	// Initialize in-memory storage
	herbariumDB := storage.GetHerbariumDB()

	// Create repositories
	stemRepo := repos.NewStemRepository(herbariumDB)
	leafRepo := repos.NewLeafRepository(herbariumDB)

	// Initialize managers with dependencies
	leafManager := NewLeafManager(leafRepo, haproxyClient, stemRepo)
	stemManager := NewStemManager(stemRepo, leafManager, haproxyClient)

	// Return a fully initialized PlatformManager
	return &PlatformManager{
		StemManager: stemManager,
		LeafManager: leafManager,
		BasePath:    config.Plantarium.RootFolder,
		Config:      config,
		isWindows:   runtime.GOOS == "windows",
	}, nil
}

// InitializePlatform initializes the platform by setting up system and deployment stems.
func (p *PlatformManager) InitializePlatform() error {
	log.Println("Initializing platform...")

	if err := p.StartSystemStems(); err != nil {
		log.Printf("Failed to start system stems: %v", err)
		return err
	}

	if err := p.StartDeploymentStems(); err != nil {
		log.Printf("Failed to start deployment stems: %v", err)
		return err
	}

	log.Println("Platform initialized successfully.")
	return nil
}

// StartSystemStems starts the core system stems (internal method).
func (p *PlatformManager) StartSystemStems() error {
	log.Println("Starting system stems...")
	// Implementation not provided yet
	return nil
}

// StartDeploymentStems starts all configured deployment stems (internal method).
func (p *PlatformManager) StartDeploymentStems() error {
	log.Println("Starting deployment stems...")
	// Implementation not provided yet
	return nil
}

// StopPlatform stops all services and performs platform cleanup.
func (p *PlatformManager) StopPlatform() error {
	log.Println("Stopping platform...")
	// TODO: Implement the logic to gracefully shut down all stems.
	return nil
}

// GetServiceConfigurations reads the configurations for all services.
func (p *PlatformManager) GetServiceConfigurations() ([]Service, error) {
	var services []Service

	servicesPath := filepath.Join(p.BasePath, "services")
	log.Printf("Starting traversal in base services path: %s", servicesPath)

	entries, err := os.ReadDir(servicesPath)
	if err != nil {
		return nil, fmt.Errorf("error reading services directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			currentPath, err := p.resolveCurrentPath(servicesPath, entry.Name())
			if err != nil {
				log.Printf("Skipping service directory %s due to error: %v", entry.Name(), err)
				continue
			}

			configFilePath := filepath.Join(currentPath, "config.yaml")
			configFile, err := os.Open(configFilePath)
			if err != nil {
				log.Printf("Error opening config file %s: %v", configFilePath, err)
				continue
			}
			defer configFile.Close()

			var config models.StemConfig
			if err := yaml.NewDecoder(configFile).Decode(&config); err != nil {
				log.Printf("Error decoding YAML for %s: %v", configFilePath, err)
				continue
			}

			service := Service{
				Config:     config,
				VersionDir: currentPath,
			}
			services = append(services, service)
			log.Printf("Loaded configuration for service: %s, version directory: %s", config.Name, currentPath)
		}
	}

	log.Printf("Total services loaded: %d", len(services))
	return services, nil
}

// GetServiceConfiguration retrieves a specific service configuration by name and version.
func (p *PlatformManager) GetServiceConfiguration(name, version string) (*Service, error) {
	// TODO: Add logic to retrieve a specific service configuration by name and version.
	return nil, nil
}

// resolveCurrentPath determines the "current" path based on the OS.
func (p *PlatformManager) resolveCurrentPath(basePath, serviceName string) (string, error) {
	currentPath := filepath.Join(basePath, serviceName, "current")

	if p.isWindows {
		content, err := os.ReadFile(currentPath)
		if err != nil {
			return "", fmt.Errorf("unable to read symlink file: %v", err)
		}
		return filepath.Join(filepath.Dir(currentPath), strings.TrimSpace(string(content))), nil
	}

	return filepath.EvalSymlinks(currentPath)
}

// Internal method to load global configuration
func loadGlobalConfig() (*models.GlobalConfig, error) {
	rootFolder := os.Getenv("PLANTARIUM_ROOT_FOLDER")
	if rootFolder == "" {
		rootFolder = "/default/path/to/base" // Default path if the environment variable is not set
	}

	configPath := filepath.Join(rootFolder, "system", "herbarium", "config.yaml")
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read global config at %s: %v", configPath, err)
	}

	var config models.GlobalConfig
	if err := yaml.Unmarshal(configContent, &config); err != nil {
		return nil, fmt.Errorf("failed to parse global config: %v", err)
	}

	config.Plantarium.RootFolder = rootFolder
	return &config, nil
}
