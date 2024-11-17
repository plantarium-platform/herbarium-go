package manager

import (
	"fmt"
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
	assert.Equal(t, "hello-service", helloService.Config.Services[0].Name, "Expected service name 'hello-service'")
	assert.Equal(t, "/hello", helloService.Config.Services[0].URL, "Expected URL '/hello'")
	assert.Equal(t, "java -jar hello-service.jar", helloService.Config.Services[0].Command, "Expected command to run the service")
	assert.Equal(t, "production", helloService.Config.Services[0].Env["GLOBAL_VAR"], "Expected GLOBAL_VAR to be 'production'")
	assert.Equal(t, "test", helloService.Config.Services[0].Dependencies[0].Schema, "Expected dependency schema 'test'")
}
