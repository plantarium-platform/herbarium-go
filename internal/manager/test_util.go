package manager

import (
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"github.com/stretchr/testify/mock"
)

// MockStemManager is a mock implementation of the StemManagerInterface.
type MockStemManager struct {
	mock.Mock
}

func (m *MockStemManager) AddStem(name, version string) error {
	args := m.Called(name, version)
	return args.Error(0)
}

func (m *MockStemManager) RemoveStem(name, version string) error {
	args := m.Called(name, version)
	return args.Error(0)
}

func (m *MockStemManager) GetStemInfo(name, version string) (*models.Stem, error) {
	args := m.Called(name, version)
	return args.Get(0).(*models.Stem), args.Error(1)
}

// MockLeafManager is a mock implementation of the LeafManagerInterface.
type MockLeafManager struct {
	mock.Mock
}

func (m *MockLeafManager) StartLeaf(stemName, version string) (string, error) {
	args := m.Called(stemName, version)
	return args.String(0), args.Error(1)
}

func (m *MockLeafManager) StopLeaf(leafID string) error {
	args := m.Called(leafID)
	return args.Error(0)
}

func (m *MockLeafManager) GetRunningLeafs(stemName, version string) ([]models.Leaf, error) {
	args := m.Called(stemName, version)
	return args.Get(0).([]models.Leaf), args.Error(1)
}
