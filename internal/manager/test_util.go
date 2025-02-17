package manager

import (
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/pkg/models"
	"github.com/stretchr/testify/mock"
)

// MockStemManager is a mock implementation of the StemManagerInterface.
// MockStemManager is a mock implementation of the StemManagerInterface.
type MockStemManager struct {
	mock.Mock
}

func (m *MockStemManager) RegisterStem(config models.StemConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockStemManager) UnregisterStem(key storage.StemKey) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockStemManager) FetchStemInfo(key storage.StemKey) (*models.Stem, error) {
	args := m.Called(key)
	if result := args.Get(0); result != nil {
		return result.(*models.Stem), args.Error(1)
	}
	return nil, args.Error(1)
}

// MockLeafManager is a mock implementation of the LeafManagerInterface.
type MockLeafManager struct {
	mock.Mock
}

func (m *MockLeafManager) StartLeaf(stemName, version string, replaceServer *string) (string, error) {
	args := m.Called(stemName, version, replaceServer)
	return args.String(0), args.Error(1)
}

func (m *MockLeafManager) StopLeaf(stemName, version, leafID string) error {
	args := m.Called(stemName, version, leafID)
	return args.Error(0)
}

func (m *MockLeafManager) GetRunningLeafs(key storage.StemKey) ([]models.Leaf, error) {
	args := m.Called(key)
	if leafs, ok := args.Get(0).([]models.Leaf); ok {
		return leafs, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockLeafManager) StartGraftNodeLeaf(stemName, version string) (string, error) {
	args := m.Called(stemName, version)
	return args.String(0), args.Error(1)
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
