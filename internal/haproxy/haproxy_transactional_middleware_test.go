package haproxy

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionMiddleware_Success(t *testing.T) {
	// Initialize the mock HAProxyConfigurationManager
	mockManager := new(MockHAProxyConfigurationManager)

	// Set up the mock methods
	mockManager.On("GetCurrentConfigVersion").Return(int64(1), nil)
	mockManager.On("StartTransaction", int64(1)).Return("txn123", nil)
	mockManager.On("CommitTransaction", "txn123").Return(nil)

	// Define the middleware
	middleware := NewTransactionMiddleware(mockManager)

	// Mock the "next" function to simulate a successful operation
	next := func(transactionID string) error {
		// Simulate some operation that succeeds
		return nil
	}

	// Execute the middleware
	err := middleware(next)()

	// Assert no errors occurred
	assert.NoError(t, err)

	// Assert that the expected methods were called
	mockManager.AssertExpectations(t)
}

func TestTransactionMiddleware_Failure(t *testing.T) {
	// Initialize the mock HAProxyConfigurationManager
	mockManager := new(MockHAProxyConfigurationManager)

	// Set up the mock methods
	mockManager.On("GetCurrentConfigVersion").Return(int64(1), nil)
	mockManager.On("StartTransaction", int64(1)).Return("txn123", nil)
	mockManager.On("RollbackTransaction", "txn123").Return(nil)

	// Define the middleware
	middleware := NewTransactionMiddleware(mockManager)

	// Mock the "next" function to simulate an operation failure
	next := func(transactionID string) error {
		// Simulate some operation that fails
		return errors.New("something went wrong")
	}

	// Execute the middleware
	err := middleware(next)()

	// Assert that an error occurred due to the failure
	assert.Error(t, err)

	// Assert that the rollback was called
	mockManager.AssertExpectations(t)
}

func TestTransactionMiddleware_GetCurrentConfigVersionError(t *testing.T) {
	// Initialize the mock HAProxyConfigurationManager
	mockManager := new(MockHAProxyConfigurationManager)

	// Set up the mock methods
	mockManager.On("GetCurrentConfigVersion").Return(int64(0), errors.New("failed to get version"))

	// Define the middleware
	middleware := NewTransactionMiddleware(mockManager)

	// Mock the "next" function to simulate an operation
	next := func(transactionID string) error {
		return nil
	}

	// Execute the middleware
	err := middleware(next)()

	// Assert that an error occurred due to version fetch failure
	assert.Error(t, err)

	// Assert that the expected methods were called
	mockManager.AssertExpectations(t)
}

func TestTransactionMiddleware_StartTransactionError(t *testing.T) {
	// Initialize the mock HAProxyConfigurationManager
	mockManager := new(MockHAProxyConfigurationManager)

	// Set up the mock methods
	mockManager.On("GetCurrentConfigVersion").Return(int64(1), nil)
	mockManager.On("StartTransaction", int64(1)).Return("", errors.New("failed to start transaction"))

	// Define the middleware
	middleware := NewTransactionMiddleware(mockManager)

	// Mock the "next" function to simulate an operation
	next := func(transactionID string) error {
		return nil
	}

	// Execute the middleware
	err := middleware(next)()

	// Assert that an error occurred due to transaction start failure
	assert.Error(t, err)

	// Assert that the expected methods were called
	mockManager.AssertExpectations(t)
}
