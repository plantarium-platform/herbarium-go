package storage

import (
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"time"
)

func initTestStorage() *HerbariumDB {
	// Create fixed timestamp for consistent test data
	fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Initialize storage with direct struct assignment
	return &HerbariumDB{
		Stems: map[string]*models.Stem{
			"system-service": {
				Name:           "system-service",
				Type:           models.StemTypeSystem,
				WorkingURL:     "http://localhost:8080",
				HAProxyBackend: "haproxy-system",
				Version:        "1.0.0",
				Environment:    map[string]string{"ENV": "production"},
				LeafInstances: map[string]*models.Leaf{
					"leaf-1": {
						ID:            "leaf-1",
						PID:           1234,
						HAProxyServer: "haproxy-system",
						Port:          8081,
						Status:        models.StatusUnknown,
						Initialized:   fixedTime,
					},
				},
				Config: &models.ServiceConfig{
					Services: []struct {
						Name         string            `yaml:"name"`
						URL          string            `yaml:"url"`
						Command      string            `yaml:"command"`
						Env          map[string]string `yaml:"env"`
						Dependencies []struct {
							Name   string `yaml:"name"`
							Schema string `yaml:"schema"`
						} `yaml:"dependencies"`
						Version string `yaml:"version"`
					}{
						{
							Name:    "system-service",
							URL:     "http://localhost:8080",
							Command: "./start.sh",
							Env:     map[string]string{"ENV": "production"},
							Version: "1.0.0",
						},
					},
				},
			},
			"user-deployment": {
				Name:           "user-deployment",
				Type:           models.StemTypeDeployment,
				WorkingURL:     "http://localhost:9090",
				HAProxyBackend: "haproxy-user",
				Version:        "1.0.0",
				Environment:    map[string]string{"DEBUG": "true"},
				LeafInstances: map[string]*models.Leaf{
					"leaf-1": {
						ID:            "leaf-1",
						PID:           5678,
						HAProxyServer: "haproxy-user",
						Port:          9091,
						Status:        models.StatusUnknown,
						Initialized:   fixedTime,
					},
				},
				GraftNodeLeaf: &models.Leaf{
					ID:            "graft-leaf",
					PID:           0,
					HAProxyServer: "haproxy-user",
					Port:          9092,
					Status:        models.StatusUnknown,
					Initialized:   fixedTime,
				},
				Config: &models.ServiceConfig{
					Services: []struct {
						Name         string            `yaml:"name"`
						URL          string            `yaml:"url"`
						Command      string            `yaml:"command"`
						Env          map[string]string `yaml:"env"`
						Dependencies []struct {
							Name   string `yaml:"name"`
							Schema string `yaml:"schema"`
						} `yaml:"dependencies"`
						Version string `yaml:"version"`
					}{
						{
							Name:    "user-deployment",
							URL:     "http://localhost:9090",
							Command: "./run.sh",
							Env:     map[string]string{"DEBUG": "true"},
							Version: "1.0.0",
						},
					},
				},
			},
		},
	}
}

// Helper function to get a fresh copy for testing
func GetTestStorage() *HerbariumDB {
	return initTestStorage()
}
