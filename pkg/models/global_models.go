package models

import "time"

// StemConfig represents the configuration for a service, parsed from a YAML file.
type StemConfig struct {
	Name         string            `yaml:"name"`    // Service name
	URL          string            `yaml:"url"`     // Service URL
	Command      string            `yaml:"command"` // Command to start the service
	Env          map[string]string `yaml:"env"`     // Environment variables
	Dependencies []struct {        // Service dependencies
		Name   string `yaml:"name"`   // Dependency name
		Schema string `yaml:"schema"` // Dependency schema
	} `yaml:"dependencies"`
	Version      string  `yaml:"version"`      // Service version
	MinInstances *int    `yaml:"minInstances"` // Minimum number of instances to keep running (optional)
	StartMessage *string `yaml:"startMessage"` // Message indicating the service has started (optional)
}

// Stem represents a deployment with associated leaf instances and configuration.
type Stem struct {
	Name           string            // Unique name of the deployment
	Type           StemType          // Type of stem (e.g., System, Deployment)
	WorkingURL     string            // Base URL for the stem
	HAProxyBackend string            // HAProxy backend name
	Version        string            // Active version
	Environment    map[string]string // Environment variables (key-value pairs)
	LeafInstances  map[string]*Leaf  // Active leaf instances (keyed by LeafID)
	GraftNodeLeaf  *Leaf             // Placeholder leaf if no real instances exist
	Config         *StemConfig       // Parsed service configuration
}

// Leaf represents a single running instance of a service.
type Leaf struct {
	ID            string     // Unique identifier for the leaf instance
	PID           int        // Process ID of the running leaf
	HAProxyServer string     // HAProxy server name for this leaf
	Port          int        // Port on which the leaf is running
	Status        LeafStatus // Current status of the leaf
	Initialized   time.Time  // Timestamp of when the leaf was initialized
}

// StemType defines the type of a stem, either a system stem or a deployment stem.
type StemType string

const (
	StemTypeSystem     StemType = "SYSTEM"     // System stems
	StemTypeDeployment StemType = "DEPLOYMENT" // User-provided deployments
)

// LeafStatus defines the status of a leaf instance.
type LeafStatus string

const (
	StatusStarting LeafStatus = "STARTING" // The leaf is starting
	StatusRunning  LeafStatus = "RUNNING"  // The leaf is running
	StatusStopping LeafStatus = "STOPPING" // The leaf is stopping
	StatusUnknown  LeafStatus = "UNKNOWN"  // The status of the leaf is unknown
)

type GlobalConfig struct {
	Plantarium struct {
		RootFolder string `yaml:"root_folder"`
		LogFolder  string `yaml:"log_folder"`
	} `yaml:"plantarium"`
	HAProxy struct {
		URL      string `yaml:"url"`
		Login    string `yaml:"login"`
		Password string `yaml:"password"`
	} `yaml:"haproxy"`
	Security struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"security"`
}
