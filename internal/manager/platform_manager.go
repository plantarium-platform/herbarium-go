package manager

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"gopkg.in/yaml.v2"
)

// PlatformManagerInterface defines the methods for managing the platform lifecycle.
type PlatformManagerInterface interface {
	InitializePlatform() error                                      // Entry point for platform initialization.
	StopPlatform() error                                            // Gracefully stops the platform and cleans up resources.
	StartSystemStems() error                                        // Starts core system stems.
	StartDeploymentStems() error                                    // Starts all deployment stems.
	GetServiceConfigurations() ([]Service, error)                   // Retrieves all service configurations.
	GetServiceConfiguration(name, version string) (*Service, error) // Retrieves a specific service configuration.
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
}

// NewPlatformManager creates a new instance of PlatformManager with the required dependencies.
func NewPlatformManager(stemManager StemManagerInterface, leafManager LeafManagerInterface, basePath string) *PlatformManager {
	return &PlatformManager{
		StemManager: stemManager,
		LeafManager: leafManager,
		BasePath:    basePath,
		isWindows:   runtime.GOOS == "windows",
	}
}

// InitializePlatform initializes the platform by setting up system and deployment stems.
func (p *PlatformManager) InitializePlatform() error {
	if err := p.StartSystemStems(); err != nil {
		return err
	}

	if err := p.StartDeploymentStems(); err != nil {
		return err
	}

	return nil
}

// StopPlatform stops all services and performs platform cleanup.
func (p *PlatformManager) StopPlatform() error {
	// TODO: Implement the logic to gracefully shut down all stems.
	return nil
}

// StartSystemStems starts the core system stems.
func (p *PlatformManager) StartSystemStems() error {
	return nil
}

// StartDeploymentStems starts all configured deployment stems.
func (p *PlatformManager) StartDeploymentStems() error {
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
