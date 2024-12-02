package manager

import (
	"fmt"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlatformManager_InitializePlatform(t *testing.T) {
	mockStemManager := new(MockStemManager)
	mockLeafManager := new(MockLeafManager)
	basePath := "../../testdata"

	manager := NewPlatformManager(mockStemManager, mockLeafManager, basePath)

	mockStemManager.On("StartSystemStems").Return(nil)
	mockStemManager.On("StartDeploymentStems").Return(nil)

	err := manager.InitializePlatform()
	assert.NoError(t, err)

	mockStemManager.AssertCalled(t, "StartSystemStems")
	mockStemManager.AssertCalled(t, "StartDeploymentStems")
}

func TestPlatformManager_StartSystemStems(t *testing.T) {
	mockStemManager := new(MockStemManager)
	mockLeafManager := new(MockLeafManager)
	basePath := "../../testdata"

	manager := NewPlatformManager(mockStemManager, mockLeafManager, basePath)

	mockStemManager.On("StartSystemStems").Return(nil)

	err := manager.StartSystemStems()
	assert.NoError(t, err)

	mockStemManager.AssertCalled(t, "StartSystemStems")
}

func TestPlatformManager_StartDeploymentStems(t *testing.T) {
	mockStemManager := new(MockStemManager)
	mockLeafManager := new(MockLeafManager)
	basePath := "../../testdata"

	manager := NewPlatformManager(mockStemManager, mockLeafManager, basePath)

	mockStemManager.On("StartDeploymentStems").Return(nil)

	err := manager.StartDeploymentStems()
	assert.NoError(t, err)

	mockStemManager.AssertCalled(t, "StartDeploymentStems")
}

func TestPlatformManager_GetServiceConfigurations(t *testing.T) {
	// Print the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	fmt.Printf("Current working directory: %s\n", currentDir)

	basePath := "../../testdata" // Base directory containing the services structure

	// Create a new PlatformManager instance
	platformManager := NewPlatformManager(nil, nil, basePath)

	// Retrieve service configurations using the PlatformManager's method
	services, err := platformManager.GetServiceConfigurations()
	assert.NoError(t, err, "Failed to get service configurations")
	assert.NotEmpty(t, services, "Expected to find service configurations")

	// Validate the retrieved configuration for hello-service
	assert.Equal(t, 1, len(services), "Expected 1 service configuration")

	helloService := services[0]
	assert.Equal(t, "hello-service", helloService.Config.Name, "Expected service name 'hello-service'")
	assert.Equal(t, "/hello", helloService.Config.URL, "Expected URL '/hello'")
	assert.Equal(t, "java -jar hello-service.jar", helloService.Config.Command, "Expected command to run the service")
	assert.Equal(t, "production", helloService.Config.Env["GLOBAL_VAR"], "Expected GLOBAL_VAR to be 'production'")
	assert.Equal(t, "test", helloService.Config.Dependencies[0].Schema, "Expected dependency schema 'test'")
}
func TestLoadGlobalConfig(t *testing.T) {
	// Set the environment variable to point to the root folder
	testRoot := "../../testdata"
	err := os.Setenv("PLANTARIUM_ROOT_FOLDER", testRoot)
	assert.NoError(t, err, "failed to set environment variable for root folder")

	// Define the expected configuration
	// Define the expected configuration
	expectedConfig := &models.GlobalConfig{
		Plantarium: struct {
			RootFolder string `yaml:"root_folder"`
			LogFolder  string `yaml:"log_folder"`
		}{
			RootFolder: testRoot, // Overwritten by environment variable
			LogFolder:  "/var/log/plantarium",
		},
		HAProxy: struct {
			URL      string `yaml:"url"`
			Login    string `yaml:"login"`
			Password string `yaml:"password"`
		}{
			URL:      "http://localhost:8080",
			Login:    "admin",
			Password: "secure-password",
		},
		Security: struct {
			APIKey string `yaml:"api_key"`
		}{
			APIKey: "super-secure-key",
		},
	}

	// Create the PlatformManager and call LoadGlobalConfig
	manager := &PlatformManager{BasePath: testRoot}
	globalConfig, err := manager.LoadGlobalConfig()

	// Assertions
	assert.NoError(t, err, "failed to load global config")
	assert.Equal(t, expectedConfig.Plantarium.RootFolder, globalConfig.Plantarium.RootFolder)
	assert.Equal(t, expectedConfig.Plantarium.LogFolder, globalConfig.Plantarium.LogFolder)
	assert.Equal(t, expectedConfig.HAProxy.URL, globalConfig.HAProxy.URL)
	assert.Equal(t, expectedConfig.HAProxy.Login, globalConfig.HAProxy.Login)
	assert.Equal(t, expectedConfig.HAProxy.Password, globalConfig.HAProxy.Password)
	assert.Equal(t, expectedConfig.Security.APIKey, globalConfig.Security.APIKey)
}
