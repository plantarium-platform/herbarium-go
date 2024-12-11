package manager

import (
	"errors"
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
	config, err := loadGlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load global configuration: %w", err)
	}

	haproxyConfig := haproxy.HAProxyConfig{
		APIURL:   config.HAProxy.URL,
		Username: config.HAProxy.Login,
		Password: config.HAProxy.Password,
	}

	haproxyConfigManager := haproxy.NewHAProxyConfigurationManager(haproxyConfig)
	haproxyClient := haproxy.NewHAProxyClient(haproxyConfig, haproxyConfigManager)

	herbariumDB := storage.GetHerbariumDB()

	stemRepo := repos.NewStemRepository(herbariumDB)
	leafRepo := repos.NewLeafRepository(herbariumDB)

	leafManager := NewLeafManager(leafRepo, haproxyClient, stemRepo)
	stemManager := NewStemManager(stemRepo, leafManager, haproxyClient)

	return &PlatformManager{
		StemManager: stemManager,
		LeafManager: leafManager,
		BasePath:    config.Plantarium.RootFolder,
		Config:      config,
		isWindows:   runtime.GOOS == "windows",
	}, nil
}

// InitializePlatform initializes the platform by registering system and deployment stems.
func (p *PlatformManager) InitializePlatform() error {
	log.Println("Initializing platform...")

	// Retrieve system and deployment stems
	systemStems, deploymentStems, err := p.GetServiceConfigurations()
	if err != nil {
		log.Printf("Failed to retrieve stem configurations: %v", err)
		return fmt.Errorf("failed to get service configurations: %w", err)
	}

	// Register system stems
	for _, stem := range systemStems {
		log.Printf("Registering system stem: %s", stem.Config.Name)
		if err := p.StemManager.RegisterStem(stem.Config); err != nil {
			log.Printf("Failed to register system stem %s: %v", stem.Config.Name, err)
			return fmt.Errorf("failed to register system stem %s: %w", stem.Config.Name, err)
		}
	}

	// Register deployment stems
	for _, stem := range deploymentStems {
		log.Printf("Registering deployment stem: %s", stem.Config.Name)
		if err := p.StemManager.RegisterStem(stem.Config); err != nil {
			log.Printf("Failed to register deployment stem %s: %v", stem.Config.Name, err)
			return fmt.Errorf("failed to register deployment stem %s: %w", stem.Config.Name, err)
		}
	}

	log.Println("Platform initialized successfully.")
	return nil
}

// GetServiceConfigurations reads the configurations for all services and system components.
func (p *PlatformManager) GetServiceConfigurations() ([]Service, []Service, error) {
	var systemServices, deploymentServices []Service

	// Process system components
	systemPath := filepath.Join(p.BasePath, "system")
	log.Printf("Traversing system path: %s", systemPath)

	systemEntries, err := os.ReadDir(systemPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading system directory: %v", err)
	}

	for _, entry := range systemEntries {
		if entry.IsDir() {
			// Skip the `herbarium` folder as it's not a system stem
			if entry.Name() == "herbarium" {
				log.Printf("Skipping 'herbarium' folder as it's not a system component")
				continue
			}

			// Load service config directly without resolving "current"
			service, err := p.loadServiceConfigForSystem(systemPath, entry.Name())
			if err != nil {
				log.Printf("Skipping system component %s due to error: %v", entry.Name(), err)
				continue
			}
			systemServices = append(systemServices, service)
		}
	}

	// Process deployment services
	servicesPath := filepath.Join(p.BasePath, "services")
	log.Printf("Traversing services path: %s", servicesPath)

	servicesEntries, err := os.ReadDir(servicesPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading services directory: %v", err)
	}

	for _, entry := range servicesEntries {
		if entry.IsDir() {
			service, err := p.loadServiceConfig(servicesPath, entry.Name())
			if err != nil {
				log.Printf("Skipping deployment service %s due to error: %v", entry.Name(), err)
				continue
			}
			deploymentServices = append(deploymentServices, service)
		}
	}

	log.Printf("Loaded %d system services and %d deployment services", len(systemServices), len(deploymentServices))
	return systemServices, deploymentServices, nil
}

// loadServiceConfig loads the service configuration from a directory for deployment services.
func (p *PlatformManager) loadServiceConfig(basePath, serviceName string) (Service, error) {
	currentPath, err := p.resolveCurrentPath(basePath, serviceName)
	if err != nil {
		return Service{}, fmt.Errorf("failed to resolve current version for service %s: %v", serviceName, err)
	}

	return p.loadConfigFromPath(currentPath, serviceName)
}

// loadServiceConfigForSystem loads the service configuration for a system component.
func (p *PlatformManager) loadServiceConfigForSystem(basePath, serviceName string) (Service, error) {
	componentPath := filepath.Join(basePath, serviceName)
	return p.loadConfigFromPath(componentPath, serviceName)
}

// loadConfigFromPath loads configuration from a specific path.
func (p *PlatformManager) loadConfigFromPath(path, serviceName string) (Service, error) {
	configFilePath := filepath.Join(path, "config.yaml")
	configFile, err := os.Open(configFilePath)
	if err != nil {
		return Service{}, fmt.Errorf("error opening config file %s: %v", configFilePath, err)
	}
	defer configFile.Close()

	var config models.StemConfig
	if err := yaml.NewDecoder(configFile).Decode(&config); err != nil {
		return Service{}, fmt.Errorf("error decoding YAML for service %s: %v", serviceName, err)
	}

	return Service{
		Config:     config,
		VersionDir: path,
	}, nil
}

// resolveCurrentPath determines the "current" path for deployment services.
func (p *PlatformManager) resolveCurrentPath(basePath, serviceName string) (string, error) {
	currentPath := filepath.Join(basePath, serviceName, "current")

	if p.isWindows {
		content, err := os.ReadFile(currentPath)
		if err != nil {
			return "", fmt.Errorf("unable to read symlink file for service %s: %v", serviceName, err)
		}
		return filepath.Join(filepath.Dir(currentPath), strings.TrimSpace(string(content))), nil
	}

	resolvedPath, err := filepath.EvalSymlinks(currentPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve symlink for service %s: %v", serviceName, err)
	}

	return resolvedPath, nil
}

// Internal method to load global configuration
func loadGlobalConfig() (*models.GlobalConfig, error) {
	rootFolder := os.Getenv("PLANTARIUM_ROOT_FOLDER")
	if rootFolder == "" {
		return nil, errors.New("PLANTARIUM_ROOT_FOLDER not set")
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
