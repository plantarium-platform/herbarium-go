package manager

import (
	"errors"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlatformManager_StartSystemStems(t *testing.T) {
	mockStemManager := new(MockStemManager)
	mockLeafManager := new(MockLeafManager)

	config := &models.GlobalConfig{
		Plantarium: struct {
			RootFolder string `yaml:"root_folder"`
			LogFolder  string `yaml:"log_folder"`
		}{
			RootFolder: "../../testdata",
			LogFolder:  "/var/log/plantarium",
		},
	}

	manager := NewPlatformManager(mockStemManager, mockLeafManager, config)

	mockStemManager.On("StartSystemStems").Return(nil)

	err := manager.StartSystemStems()
	assert.NoError(t, err, "Expected no error when starting system stems")

	mockStemManager.AssertCalled(t, "StartSystemStems")
}

func TestPlatformManager_StartDeploymentStems(t *testing.T) {
	mockStemManager := new(MockStemManager)
	mockLeafManager := new(MockLeafManager)

	config := &models.GlobalConfig{
		Plantarium: struct {
			RootFolder string `yaml:"root_folder"`
			LogFolder  string `yaml:"log_folder"`
		}{
			RootFolder: "../../testdata",
			LogFolder:  "/var/log/plantarium",
		},
	}

	manager := NewPlatformManager(mockStemManager, mockLeafManager, config)

	mockStemManager.On("StartDeploymentStems").Return(nil)

	err := manager.StartDeploymentStems()
	assert.NoError(t, err, "Expected no error when starting deployment stems")

	mockStemManager.AssertCalled(t, "StartDeploymentStems")
}

func TestPlatformManager_GetServiceConfigurations(t *testing.T) {
	// Set environment variable to testdata folder
	testRoot := "../../testdata"
	err := os.Setenv("PLANTARIUM_ROOT_FOLDER", testRoot)
	assert.NoError(t, err, "Failed to set environment variable for root folder")
	defer os.Unsetenv("PLANTARIUM_ROOT_FOLDER")

	// Ensure the configuration file exists
	configPath := filepath.Join(testRoot, "system", "herbarium", "config.yaml")
	_, err = os.Stat(configPath)
	assert.NoError(t, err, "Configuration file should exist at %s", configPath)

	// Initialize the PlatformManager using the test configuration
	platformManager, err := NewPlatformManagerWithDI()
	assert.NoError(t, err, "Failed to create PlatformManagerWithDI")
	assert.NotNil(t, platformManager, "PlatformManager should not be nil")

	// Retrieve service configurations
	services, err := platformManager.GetServiceConfigurations()
	assert.NoError(t, err, "Failed to get service configurations")
	assert.NotEmpty(t, services, "Expected to find service configurations")

	// Validate the retrieved configuration for a known service
	assert.Equal(t, 1, len(services), "Expected 1 service configuration")

	helloService := services[0]
	assert.Equal(t, "hello-service", helloService.Config.Name, "Expected service name 'hello-service'")
	assert.Equal(t, "/hello", helloService.Config.URL, "Expected URL '/hello'")
	assert.Equal(t, "java -jar hello-service.jar", helloService.Config.Command, "Expected command to run the service")
	assert.Equal(t, "production", helloService.Config.Env["GLOBAL_VAR"], "Expected GLOBAL_VAR to be 'production'")
	assert.Equal(t, "test", helloService.Config.Dependencies[0].Schema, "Expected dependency schema 'test'")
}

// TestPlatformManager_InitializePlatform validates the initialization flow.
func TestPlatformManager_InitializePlatform(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		// Mock dependencies
		mockStemManager := new(MockStemManager)
		mockLeafManager := new(MockLeafManager)
		globalConfig := &models.GlobalConfig{
			HAProxy: struct {
				URL      string `yaml:"url"`
				Login    string `yaml:"login"`
				Password string `yaml:"password"`
			}{
				URL:      "http://localhost:8080",
				Login:    "admin",
				Password: "password",
			},
		}

		// PlatformManager instance
		platformManager := NewPlatformManager(mockStemManager, mockLeafManager, globalConfig)

		// Mock expected behavior
		mockStemManager.On("StartSystemStems").Return(nil)
		mockStemManager.On("StartDeploymentStems").Return(nil)

		// Call InitializePlatform
		err := platformManager.InitializePlatform()
		assert.NoError(t, err)

		// Verify calls
		mockStemManager.AssertCalled(t, "StartSystemStems")
		mockStemManager.AssertCalled(t, "StartDeploymentStems")
	})

	t.Run("system stem initialization failure", func(t *testing.T) {
		// Mock dependencies
		mockStemManager := new(MockStemManager)
		mockLeafManager := new(MockLeafManager)
		globalConfig := &models.GlobalConfig{}

		// PlatformManager instance
		platformManager := NewPlatformManager(mockStemManager, mockLeafManager, globalConfig)

		// Mock expected behavior
		mockStemManager.On("StartSystemStems").Return(errors.New("system stem failure"))

		// Call InitializePlatform
		err := platformManager.InitializePlatform()
		assert.Error(t, err)
		assert.EqualError(t, err, "system stem failure")

		// Verify calls
		mockStemManager.AssertCalled(t, "StartSystemStems")
		mockStemManager.AssertNotCalled(t, "StartDeploymentStems")
	})

	t.Run("deployment stem initialization failure", func(t *testing.T) {
		// Mock dependencies
		mockStemManager := new(MockStemManager)
		mockLeafManager := new(MockLeafManager)
		globalConfig := &models.GlobalConfig{}

		// PlatformManager instance
		platformManager := NewPlatformManager(mockStemManager, mockLeafManager, globalConfig)

		// Mock expected behavior
		mockStemManager.On("StartSystemStems").Return(nil)
		mockStemManager.On("StartDeploymentStems").Return(errors.New("deployment stem failure"))

		// Call InitializePlatform
		err := platformManager.InitializePlatform()
		assert.Error(t, err)
		assert.EqualError(t, err, "deployment stem failure")

		// Verify calls
		mockStemManager.AssertCalled(t, "StartSystemStems")
		mockStemManager.AssertCalled(t, "StartDeploymentStems")
	})
}

func TestNewPlatformManagerWithDI(t *testing.T) {
	// Set the environment variable for the root folder
	testRoot := "../../testdata"
	err := os.Setenv("PLANTARIUM_ROOT_FOLDER", testRoot)
	assert.NoError(t, err, "failed to set environment variable for root folder")
	defer os.Unsetenv("PLANTARIUM_ROOT_FOLDER") // Clean up after test

	// Ensure the configuration file exists
	configPath := filepath.Join(testRoot, "system", "herbarium", "config.yaml")
	_, err = os.Stat(configPath)
	assert.NoError(t, err, "configuration file should exist at %s", configPath)

	// Call the method under test
	platformManager, err := NewPlatformManagerWithDI()
	assert.NoError(t, err, "failed to initialize PlatformManagerWithDI")
	assert.NotNil(t, platformManager, "PlatformManager should not be nil")

	// Validate the loaded configuration
	assert.Equal(t, testRoot, platformManager.Config.Plantarium.RootFolder, "RootFolder should match testRoot")

	// Validate HAProxyClient initialization
	assert.NotNil(t, platformManager.LeafManager, "LeafManager should be initialized")
	assert.NotNil(t, platformManager.StemManager, "StemManager should be initialized")

	// Additional validation can check if the dependencies were wired correctly
	// For example, verify if HAProxyClient or configuration was used as expected.
}
