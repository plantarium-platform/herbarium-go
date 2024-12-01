package manager

/*
import (
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStemManager_AddStem(t *testing.T) {
	herbariumDB := storage.GetHerbariumDB()
	leafRepo := repos.NewLeafRepository(herbariumDB)
	stemRepo := repos.NewStemRepository(herbariumDB)

	mockHAProxyClient := new(MockHAProxyClient)
	leafManager := NewLeafManager(leafRepo, mockHAProxyClient, stemRepo)

	mockHAProxyClient.On("BindStem", "test-stem-backend").Return(nil)

	stemManager := NewStemManager(stemRepo, leafManager, mockHAProxyClient)

	minInstances := 2
	stemConfig := models.StemConfig{
		Name:         "test-stem",
		URL:          "/test",
		Command:      "./run-test",
		Env:          map[string]string{"ENV_VAR": "test"},
		Version:      "1.0.0",
		MinInstances: &minInstances,
	}

	err := stemManager.AddStem(stemConfig)
	assert.NoError(t, err)

	stemKey := storage.StemKey{Name: "test-stem", Version: "1.0.0"}
	stem, err := stemRepo.FindStem(stemKey)
	assert.NoError(t, err)
	assert.NotNil(t, stem)
	assert.Equal(t, "test-stem", stem.Name)
	assert.Equal(t, "1.0.0", stem.Version)

	assert.Equal(t, *stemConfig.MinInstances, len(stem.LeafInstances))

	for leafID, leaf := range stem.LeafInstances {
		assert.NotNil(t, leaf)
		assert.Equal(t, models.StatusRunning, leaf.Status)
		assert.Equal(t, "test-stem", stem.Name)
		_, err := leafRepo.FindLeafByID(stemKey, leafID)
		assert.NoError(t, err)
	}

	mockHAProxyClient.AssertExpectations(t)
}

func TestStemManager_AddStem_DuplicateError(t *testing.T) {
	herbariumDB := storage.GetHerbariumDB()
	leafRepo := repos.NewLeafRepository(herbariumDB)
	stemRepo := repos.NewStemRepository(herbariumDB)

	mockHAProxyClient := new(MockHAProxyClient)
	leafManager := NewLeafManager(leafRepo, mockHAProxyClient, stemRepo)

	stemManager := NewStemManager(stemRepo, leafManager, mockHAProxyClient)

	stemKey := storage.StemKey{Name: "test-stem", Version: "1.0.0"}
	herbariumDB.Stems[stemKey] = &models.Stem{
		Name:           "test-stem",
		Type:           models.StemTypeDeployment,
		HAProxyBackend: "test-backend",
		Version:        "1.0.0",
		LeafInstances: map[string]*models.Leaf{
			"leaf-1": {
				ID:            "leaf-1",
				Status:        models.StatusRunning,
				Port:          8000,
				PID:           12345,
				HAProxyServer: "haproxy-server",
			},
		},
		Config: &models.StemConfig{
			Name:    "test-stem",
			URL:     "/test",
			Command: "./run-test",
			Version: "1.0.0",
		},
	}

	stemConfig := models.StemConfig{
		Name:         "test-stem",
		URL:          "/test",
		Command:      "./run-test",
		Env:          map[string]string{"ENV_VAR": "test"},
		Version:      "1.0.0",
		MinInstances: nil,
	}

	err := stemManager.AddStem(stemConfig)
	assert.Error(t, err)
	assert.Equal(t, "Stem test-stem already exists in version 1.0.0. Please provide a new version or stop the previous one.", err.Error())
}
*/
