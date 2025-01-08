package haproxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHAProxyClient_BindStem(t *testing.T) {
	// Initialize the mock HAProxyConfigurationManager
	mockManager := new(MockHAProxyConfigurationManager)

	// Set up the mock methods
	mockManager.On("GetCurrentConfigVersion").Return(int64(1), nil)    // Mocking GetCurrentConfigVersion
	mockManager.On("StartTransaction", int64(1)).Return("txn123", nil) // Mock StartTransaction
	mockManager.On("CommitTransaction", "txn123").Return(nil)          // Mock CommitTransaction
	mockManager.On("CreateBackend", "backend1", mock.Anything).Return(nil)

	// Create the HAProxyClient with the mock manager
	client := &HAProxyClient{
		configManager:         mockManager,
		transactionMiddleware: NewTransactionMiddleware(mockManager),
	}

	// Call BindStem
	err := client.BindStem("backend1")

	// Assert no errors occurred
	assert.NoError(t, err)

	// Assert that CreateBackend was called with expected arguments
	mockManager.AssertExpectations(t)
}
func TestHAProxyClient_BindLeaf(t *testing.T) {
	// Initialize the mock HAProxyConfigurationManager
	mockManager := new(MockHAProxyConfigurationManager)

	// Set up the mock methods
	mockManager.On("GetCurrentConfigVersion").Return(int64(1), nil)                             // Mocking GetCurrentConfigVersion
	mockManager.On("StartTransaction", int64(1)).Return("txn123", nil)                          // Mock StartTransaction
	mockManager.On("CommitTransaction", "txn123").Return(nil)                                   // Mock CommitTransaction
	mockManager.On("AddServer", "backend1", "server1", "localhost", 8080, "txn123").Return(nil) // Updated AddServer call

	// Create the HAProxyClient with the mock manager
	client := &HAProxyClient{
		configManager:         mockManager,
		transactionMiddleware: NewTransactionMiddleware(mockManager),
	}

	// Call BindLeaf
	err := client.BindLeaf("backend1", "server1", "localhost", 8080)

	// Assert no errors occurred
	assert.NoError(t, err)

	// Assert that AddServer was called with expected arguments
	mockManager.AssertExpectations(t)
}

func TestHAProxyClient_UnbindLeaf(t *testing.T) {
	// Initialize the mock HAProxyConfigurationManager
	mockManager := new(MockHAProxyConfigurationManager)

	// Set up the mock methods
	mockManager.On("GetCurrentConfigVersion").Return(int64(1), nil)    // Mocking GetCurrentConfigVersion
	mockManager.On("StartTransaction", int64(1)).Return("txn123", nil) // Mock StartTransaction
	mockManager.On("CommitTransaction", "txn123").Return(nil)          // Mock CommitTransaction
	mockManager.On("DeleteServer", "backend1", "server1", mock.Anything).Return(nil)

	// Create the HAProxyClient with the mock manager
	client := &HAProxyClient{
		configManager:         mockManager,
		transactionMiddleware: NewTransactionMiddleware(mockManager),
	}

	// Call UnbindLeaf
	err := client.UnbindLeaf("backend1", "server1")

	// Assert no errors occurred
	assert.NoError(t, err)

	// Assert that DeleteServer was called with expected arguments
	mockManager.AssertExpectations(t)
}
func TestHAProxyClient_ReplaceLeaf(t *testing.T) {
	// Initialize the mock HAProxyConfigurationManager
	mockManager := new(MockHAProxyConfigurationManager)

	// Set up the mock methods
	mockManager.On("GetCurrentConfigVersion").Return(int64(1), nil)                               // Mocking GetCurrentConfigVersion
	mockManager.On("StartTransaction", int64(1)).Return("txn123", nil)                            // Mock StartTransaction
	mockManager.On("CommitTransaction", "txn123").Return(nil)                                     // Mock CommitTransaction
	mockManager.On("DeleteServer", "backend1", "oldServer", "txn123").Return(nil)                 // Updated DeleteServer call
	mockManager.On("AddServer", "backend1", "newServer", "localhost", 8080, "txn123").Return(nil) // Updated AddServer call

	// Create the HAProxyClient with the mock manager
	client := &HAProxyClient{
		configManager:         mockManager,
		transactionMiddleware: NewTransactionMiddleware(mockManager),
	}

	// Call ReplaceLeaf
	err := client.ReplaceLeaf("backend1", "oldServer", "newServer", "localhost", 8080)

	// Assert no errors occurred
	assert.NoError(t, err)

	// Assert that DeleteServer and AddServer were called with expected arguments
	mockManager.AssertExpectations(t)
}

func TestHAProxyClient_UnbindStem(t *testing.T) {
	// Initialize the mock HAProxyConfigurationManager
	mockManager := new(MockHAProxyConfigurationManager)

	// Set up the mock methods
	mockManager.On("GetCurrentConfigVersion").Return(int64(1), nil)    // Mocking GetCurrentConfigVersion
	mockManager.On("StartTransaction", int64(1)).Return("txn123", nil) // Mock StartTransaction
	mockManager.On("CommitTransaction", "txn123").Return(nil)          // Mock CommitTransaction
	mockManager.On("DeleteServer", "backend1", "", mock.Anything).Return(nil)

	// Create the HAProxyClient with the mock manager
	client := &HAProxyClient{
		configManager:         mockManager,
		transactionMiddleware: NewTransactionMiddleware(mockManager),
	}

	// Call UnbindStem
	err := client.UnbindStem("backend1")

	// Assert no errors occurred
	assert.NoError(t, err)

	// Assert that DeleteServer was called with expected arguments
	mockManager.AssertExpectations(t)
}
