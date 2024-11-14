package manager

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ServiceConfig represents a service configuration structure.
type ServiceConfig struct {
	Services []struct {
		Name         string            `yaml:"name"`
		URL          string            `yaml:"url"`
		Command      string            `yaml:"command"`
		Env          map[string]string `yaml:"env"`
		Dependencies []struct {
			Name   string `yaml:"name"`
			Schema string `yaml:"schema"`
		} `yaml:"dependencies"`
		Version string `yaml:"version"`
	} `yaml:"services"`
}

// Service represents a service with its config and version path.
type Service struct {
	Config     ServiceConfig
	VersionDir string
}

// Manager manages the retrieval of service configurations
type Manager struct {
	BasePath  string
	isWindows bool
}

// NewManager initializes a new Manager instance and detects if the OS is Windows
func NewManager(basePath string) *Manager {
	return &Manager{
		BasePath:  basePath,
		isWindows: runtime.GOOS == "windows",
	}
}

// GetServiceConfigurations reads the configurations for each service in the `current` version directories
// under the base path and returns a slice of Service structs.
func (m *Manager) GetServiceConfigurations() ([]Service, error) {
	var services []Service

	servicesPath := filepath.Join(m.BasePath, "services")
	log.Printf("Starting traversal in base services path: %s", servicesPath)

	entries, err := os.ReadDir(servicesPath)
	if err != nil {
		return nil, fmt.Errorf("error reading services directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Resolve the "current" path for each service
			currentPath, err := m.resolveCurrentPath(servicesPath, entry.Name())
			if err != nil {
				log.Printf("Skipping service directory %s due to error: %v", entry.Name(), err)
				continue
			}

			// Read the service configuration
			configFilePath := filepath.Join(currentPath, "config.yaml")
			configFile, err := os.Open(configFilePath)
			if err != nil {
				log.Printf("Error opening config file %s: %v", configFilePath, err)
				continue
			}
			defer configFile.Close()

			var config ServiceConfig
			if err := yaml.NewDecoder(configFile).Decode(&config); err != nil {
				log.Printf("Error decoding YAML for %s: %v", configFilePath, err)
				continue
			}

			service := Service{
				Config:     config,
				VersionDir: currentPath,
			}
			services = append(services, service)
			log.Printf("Loaded configuration for service: %s, version directory: %s", config.Services[0].Name, currentPath)
		}
	}

	log.Printf("Total services loaded: %d", len(services))
	return services, nil
}

// resolveCurrentPath determines the "current" path based on the OS.
// On Windows, it reads "current" as a file with the directory name inside.
// On Linux/Mac, it treats "current" as a symlink.
func (m *Manager) resolveCurrentPath(basePath, serviceName string) (string, error) {
	currentPath := filepath.Join(basePath, serviceName, "current")

	if m.isWindows {
		content, err := os.ReadFile(currentPath)
		if err != nil {
			return "", fmt.Errorf("unable to read symlink file: %v", err)
		}
		return filepath.Join(filepath.Dir(currentPath), strings.TrimSpace(string(content))), nil
	}

	return filepath.EvalSymlinks(currentPath)
}
