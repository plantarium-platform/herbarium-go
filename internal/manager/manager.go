package manager

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path/filepath"
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

// Manager manages service configurations and operations.
type Manager struct {
	BasePath string
}

// NewManager creates a new Manager instance with the given base path.
func NewManager(basePath string) *Manager {
	return &Manager{BasePath: basePath}
}

// GetServiceConfigurations reads the configurations for each service in the `current` version directories
// under the base path and returns a slice of Service structs.
func (m *Manager) GetServiceConfigurations() ([]Service, error) {
	var services []Service

	servicesPath := filepath.Join(m.BasePath, "services")
	log.Printf("Starting traversal in base services path: %s", servicesPath)

	// List all directories in the base services path
	entries, err := os.ReadDir(servicesPath)
	if err != nil {
		return nil, fmt.Errorf("error reading services directory: %v", err)
	}

	// Loop through each directory and look for the "current" symlink
	for _, entry := range entries {
		if entry.IsDir() {
			serviceDir := filepath.Join(servicesPath, entry.Name(), "current")
			log.Printf("Checking for 'current' directory or symlink at: %s", serviceDir)

			// Resolve the `current` symlink or path
			resolvedPath, err := filepath.EvalSymlinks(serviceDir)
			if err != nil {
				log.Printf("Skipping %s: not a valid symlink or directory", serviceDir)
				continue
			}

			// Load the configuration file
			configFilePath := filepath.Join(resolvedPath, "config.yaml")
			log.Printf("Looking for config file at: %s", configFilePath)

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

			// Append service with its configuration and version directory
			service := Service{
				Config:     config,
				VersionDir: resolvedPath,
			}
			services = append(services, service)
			log.Printf("Loaded configuration for service: %s, version directory: %s", config.Services[0].Name, resolvedPath)
		}
	}

	log.Printf("Total services loaded: %d", len(services))
	return services, nil
}
