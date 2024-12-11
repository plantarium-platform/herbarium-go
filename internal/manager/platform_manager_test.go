package manager

import (
	"errors"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"github.com/stretchr/testify/mock"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlatformManager_GetServiceConfigurations(t *testing.T) {
	testRoot := "../../testdata"
	err := os.Setenv("PLANTARIUM_ROOT_FOLDER", testRoot)
	assert.NoError(t, err, "Failed to set PLANTARIUM_ROOT_FOLDER environment variable")
	defer os.Unsetenv("PLANTARIUM_ROOT_FOLDER")

	// Ensure the configuration file exists
	configPath := filepath.Join(testRoot, "system", "herbarium", "config.yaml")
	_, err = os.Stat(configPath)
	assert.NoError(t, err, "Configuration file should exist at %s", configPath)

	// Initialize PlatformManager
	platformManager, err := NewPlatformManagerWithDI()
	assert.NoError(t, err, "Failed to create PlatformManagerWithDI")
	assert.NotNil(t, platformManager, "PlatformManager should not be nil")

	// Retrieve service configurations
	systemServices, deploymentServices, err := platformManager.GetServiceConfigurations()
	assert.NoError(t, err, "Failed to get service configurations")

	// Validate system services
	assert.Len(t, systemServices, 1, "Expected 1 system service configuration")
	planterService := systemServices[0]
	assert.Equal(t, "planter", planterService.Config.Name, "Expected system service name 'planter'")
	assert.Equal(t, "/planter", planterService.Config.URL, "Expected system service URL '/planter'")
	assert.Equal(t, "./planter.sh", planterService.Config.Command, "Expected system service command './planter.sh'")
	assert.Equal(t, "production", planterService.Config.Env["GLOBAL_VAR"], "Expected GLOBAL_VAR to be 'production'")
	assert.Equal(t, "test", planterService.Config.Dependencies[0].Schema, "Expected dependency schema 'test'")

	// Validate deployment services
	assert.Len(t, deploymentServices, 1, "Expected 1 deployment service configuration")
	helloService := deploymentServices[0]
	assert.Equal(t, "hello-service", helloService.Config.Name, "Expected deployment service name 'hello-service'")
	assert.Equal(t, "/hello", helloService.Config.URL, "Expected deployment service URL '/hello'")
	assert.Equal(t, "java -jar hello-service.jar", helloService.Config.Command, "Expected deployment service command 'java -jar hello-service.jar'")
	assert.Equal(t, "production", helloService.Config.Env["GLOBAL_VAR"], "Expected GLOBAL_VAR to be 'production'")
	assert.Equal(t, "test", helloService.Config.Dependencies[0].Schema, "Expected dependency schema 'test'")
}

func TestPlatformManager_InitializePlatform(t *testing.T) {
	// Set environment variable for the testdata folder
	testRoot := "../../testdata"
	err := os.Setenv("PLANTARIUM_ROOT_FOLDER", testRoot)
	assert.NoError(t, err, "Failed to set PLANTARIUM_ROOT_FOLDER environment variable")
	defer os.Unsetenv("PLANTARIUM_ROOT_FOLDER")

	t.Run("successful initialization", func(t *testing.T) {
		// Mock StemManager
		mockStemManager := new(MockStemManager)
		platformManager := NewPlatformManager(mockStemManager, nil, &models.GlobalConfig{
			Plantarium: struct {
				RootFolder string `yaml:"root_folder"`
				LogFolder  string `yaml:"log_folder"`
			}{
				RootFolder: testRoot,
			},
		})

		// Mock RegisterStem behavior
		mockStemManager.On("RegisterStem", mock.Anything).Return(nil)

		// Call InitializePlatform
		err := platformManager.InitializePlatform()
		assert.NoError(t, err, "Expected InitializePlatform to succeed")

		// Verify all stems were registered
		mockStemManager.AssertNumberOfCalls(t, "RegisterStem", 2)
		mockStemManager.AssertCalled(t, "RegisterStem", mock.MatchedBy(func(config models.StemConfig) bool {
			return config.Name == "planter"
		}))
		mockStemManager.AssertCalled(t, "RegisterStem", mock.MatchedBy(func(config models.StemConfig) bool {
			return config.Name == "hello-service"
		}))
	})

	t.Run("system stem initialization failure", func(t *testing.T) {
		// Mock StemManager
		mockStemManager := new(MockStemManager)
		platformManager := NewPlatformManager(mockStemManager, nil, &models.GlobalConfig{
			Plantarium: struct {
				RootFolder string `yaml:"root_folder"`
				LogFolder  string `yaml:"log_folder"`
			}{
				RootFolder: testRoot,
			},
		})

		// Mock RegisterStem behavior for system stems
		mockStemManager.On("RegisterStem", mock.MatchedBy(func(config models.StemConfig) bool {
			return config.Name == "planter"
		})).Return(errors.New("file not found"))

		// Call InitializePlatform
		err := platformManager.InitializePlatform()
		assert.Error(t, err, "Expected error due to system stem failure")
		assert.Contains(t, err.Error(), "failed to register system stem planter", "Error should indicate system stem failure")
		assert.Contains(t, err.Error(), "file not found", "Error should include the root cause")

		// Verify system stem failed and deployment stems were not attempted
		mockStemManager.AssertCalled(t, "RegisterStem", mock.MatchedBy(func(config models.StemConfig) bool {
			return config.Name == "planter"
		}))
		mockStemManager.AssertNotCalled(t, "RegisterStem", mock.MatchedBy(func(config models.StemConfig) bool {
			return config.Name == "hello-service"
		}))
	})

	t.Run("deployment stem initialization failure", func(t *testing.T) {
		// Mock StemManager
		mockStemManager := new(MockStemManager)
		platformManager := NewPlatformManager(mockStemManager, nil, &models.GlobalConfig{
			Plantarium: struct {
				RootFolder string `yaml:"root_folder"`
				LogFolder  string `yaml:"log_folder"`
			}{
				RootFolder: testRoot,
			},
		})

		// Mock RegisterStem behavior for system stems
		mockStemManager.On("RegisterStem", mock.MatchedBy(func(config models.StemConfig) bool {
			return config.Name == "planter"
		})).Return(nil)

		// Mock failure for deployment stems
		mockStemManager.On("RegisterStem", mock.MatchedBy(func(config models.StemConfig) bool {
			return config.Name == "hello-service"
		})).Return(errors.New("insufficient permissions"))

		// Call InitializePlatform
		err := platformManager.InitializePlatform()
		assert.Error(t, err, "Expected error due to deployment stem failure")
		assert.Contains(t, err.Error(), "failed to register deployment stem hello-service", "Error should indicate deployment stem failure")
		assert.Contains(t, err.Error(), "insufficient permissions", "Error should include the root cause")

		// Verify both system and failed deployment stem were attempted
		mockStemManager.AssertCalled(t, "RegisterStem", mock.MatchedBy(func(config models.StemConfig) bool {
			return config.Name == "planter"
		}))
		mockStemManager.AssertCalled(t, "RegisterStem", mock.MatchedBy(func(config models.StemConfig) bool {
			return config.Name == "hello-service"
		}))
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
