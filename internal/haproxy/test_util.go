package haproxy

import (
	"github.com/stretchr/testify/mock"
)

// MockHAProxyConfigurationManager mocks the HAProxyConfigurationManagerInterface
type MockHAProxyConfigurationManager struct {
	mock.Mock
}

// GetCurrentConfigVersion mocks the GetCurrentConfigVersion method
func (m *MockHAProxyConfigurationManager) GetCurrentConfigVersion() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

// StartTransaction mocks the StartTransaction method
func (m *MockHAProxyConfigurationManager) StartTransaction(version int64) (string, error) {
	args := m.Called(version)
	return args.String(0), args.Error(1)
}

// CommitTransaction mocks the CommitTransaction method
func (m *MockHAProxyConfigurationManager) CommitTransaction(transactionID string) error {
	args := m.Called(transactionID)
	return args.Error(0)
}

// RollbackTransaction mocks the RollbackTransaction method
func (m *MockHAProxyConfigurationManager) RollbackTransaction(transactionID string) error {
	args := m.Called(transactionID)
	return args.Error(0)
}

// CreateBackend mocks the CreateBackend method
func (m *MockHAProxyConfigurationManager) CreateBackend(backendName, transactionID string) error {
	args := m.Called(backendName, transactionID)
	return args.Error(0)
}

// AddServer mocks the AddServer method
func (m *MockHAProxyConfigurationManager) AddServer(backendName string, serverData map[string]interface{}, transactionID string) error {
	args := m.Called(backendName, serverData, transactionID)
	return args.Error(0)
}

// DeleteServer mocks the DeleteServer method
func (m *MockHAProxyConfigurationManager) DeleteServer(backendName, serverName, transactionID string) error {
	args := m.Called(backendName, serverName, transactionID)
	return args.Error(0)
}

// GetServersFromBackend mocks the GetServersFromBackend method
func (m *MockHAProxyConfigurationManager) GetServersFromBackend(backendName, transactionID string) ([]HAProxyServer, error) {
	args := m.Called(backendName, transactionID)
	return args.Get(0).([]HAProxyServer), args.Error(1)
}
