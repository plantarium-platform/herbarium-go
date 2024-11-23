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

func (m *MockLeafManager) StopLeaf(stemName, version, leafID string) error {
	args := m.Called(stemName, version, leafID)
	return args.Error(0)
}

func (m *MockLeafManager) GetRunningLeafs(stemName, version string) ([]models.Leaf, error) {
	args := m.Called(stemName, version)
	return args.Get(0).([]models.Leaf), args.Error(1)
}

// MockHAProxyClient is a mock implementation of HAProxyClientInterface.
type MockHAProxyClient struct {
	mock.Mock
}

// BindStem mocks the BindStem method in HAProxyClient.
func (m *MockHAProxyClient) BindStem(backendName string) error {
	args := m.Called(backendName)
	return args.Error(0)
}

// BindLeaf mocks the BindLeaf method in HAProxyClient.
func (m *MockHAProxyClient) BindLeaf(backendName, haProxyServer, serviceAddress string, servicePort int) error {
	args := m.Called(backendName, haProxyServer, serviceAddress, servicePort)
	return args.Error(0)
}

// UnbindLeaf mocks the UnbindLeaf method in HAProxyClient.
func (m *MockHAProxyClient) UnbindLeaf(backendName, haProxyServer string) error {
	args := m.Called(backendName, haProxyServer)
	return args.Error(0)
}

// ReplaceLeaf mocks the ReplaceLeaf method in HAProxyClient.
func (m *MockHAProxyClient) ReplaceLeaf(backendName, oldHAProxyServer, newHAProxyServer, serviceAddress string, servicePort int) error {
	args := m.Called(backendName, oldHAProxyServer, newHAProxyServer, serviceAddress, servicePort)
	return args.Error(0)
}

// UnbindStem mocks the UnbindStem method in HAProxyClient.
func (m *MockHAProxyClient) UnbindStem(backendName string) error {
	args := m.Called(backendName)
	return args.Error(0)
}
