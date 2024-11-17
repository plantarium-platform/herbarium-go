package models

// ServiceConfig represents the configuration for a service, parsed from a YAML file.
type ServiceConfig struct {
	Services []struct {
		Name         string            `yaml:"name"`    // Service name
		URL          string            `yaml:"url"`     // Service URL
		Command      string            `yaml:"command"` // Command to start the service
		Env          map[string]string `yaml:"env"`     // Environment variables
		Dependencies []struct {        // Service dependencies
			Name   string `yaml:"name"`   // Dependency name
			Schema string `yaml:"schema"` // Dependency schema
		} `yaml:"dependencies"`
		Version string `yaml:"version"` // Service version
	} `yaml:"services"`
}
